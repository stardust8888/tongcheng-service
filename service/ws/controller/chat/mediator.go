package chat

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/library/snowflake"
	"tongcheng/library/timewheel"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

const (
	Mod               = 64
	BroadCastChannel  = 2048
	RegisterChannel   = 128
	UnRegisterChannel = 128

	HeartBeatWheelInterval = 5 * time.Second
	HeartBeatInterval      = 30 * time.Second
	TimeoutWheelInterval   = 1 * time.Second
	WriteChannelTime       = 5 * time.Second
)

var (
	mediator    *Mediator
	heartbeatTW *timewheel.TimeWheel
	timeoutTW   *timewheel.TimeWheel
)

type TimeoutChannel chan struct{}

func WriteTimeout[T any](ch chan T, value T, timeout time.Duration) {
	if timeout == 0 {
		return
	}

	timeoutChannel := make(TimeoutChannel, 1)
	timeoutTW.AddTask(timeout, 1, snowflake.Generate(), timeoutChannel)
	select {
	case ch <- value:
	case <-timeoutChannel:
		return
	}
}

func InitMediator() {
	mediator = NewMediator()

	timeoutTW = timewheel.New(TimeoutWheelInterval, 128, func(i any) {
		if _, ok := i.(TimeoutChannel); ok {
			i.(TimeoutChannel) <- struct{}{}
		}
	})
	timeoutTW.Start()

	heartbeatTW = timewheel.New(HeartBeatWheelInterval, 128, func(i any) {
		if client, ok := i.(*Client); ok {
			ping := []byte("ping")
			WriteTimeout(client.send, ping, WriteChannelTime)
		}
	})
	heartbeatTW.Start()

	go mediator.Run()
}

type Mediator struct {
	users      []map[int]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	connect    int64
}

func NewMediator() *Mediator {
	users := make([]map[int]*Client, Mod)
	for i := range users {
		users[i] = map[int]*Client{}
	}
	return &Mediator{
		users:      users,
		broadcast:  make(chan []byte, BroadCastChannel),
		register:   make(chan *Client, RegisterChannel),
		unregister: make(chan *Client, UnRegisterChannel),
	}
}

func (m *Mediator) Run() {
	defer func() {
		timeoutTW.Stop()
		heartbeatTW.Stop()
	}()

	for {
		select {
		case client := <-m.register:
			userID := client.user.ID
			mod := userID % Mod
			m.users[mod][userID] = client
			atomic.AddInt64(&m.connect, 1)
		case client := <-m.unregister:
			userID := client.user.ID
			mod := userID % Mod
			if _, ok := m.users[mod][userID]; !ok {
				continue
			}
			delete(m.users[mod], userID)
			_ = client.conn.Close()
			atomic.AddInt64(&m.connect, -1)
		case message := <-m.broadcast:
			logs.Debug(ginx.Context, "Broadcast:", zap.String("msg", string(message)))
			msg := &model.Message{}
			err := msg.Unmarshal(message)
			if err != nil {
				logs.Error(ginx.Context, "message.Unmarshal error:", zap.Error(err))
				continue
			}
			if msg.ContentType == model.ContentTypeReadMessage {
				//设置用户最大已读SeqID
				_ = redis.Set(context.Background(), constant.RedisKey(constant.RedisChatMaxReadSeqID, msg.ChatID, msg.SenderID), msg.SeqID, redis.ForeverTime)
				continue
			}

			usersID := make([]int, 0)
			switch msg.MessageType {
			case model.MessageTypeSingle:
				usersID = []int{msg.ReceiverID, msg.SenderID}
			case model.MessageTypeGroup:
				usersIDStr, _ := redis.SMembers(ginx.Context, constant.RedisKey(constant.RedisChatUserID, msg.ChatID))
				for i := range usersIDStr {
					usersIDInt, _ := strconv.Atoi(usersIDStr[i])
					usersID = append(usersID, usersIDInt)
				}
			}

			for _, userID := range usersID {
				if userID == 0 {
					continue
				}
				mod := userID % Mod
				if client, ok := m.users[mod][userID]; ok {
					if userID == msg.SenderID {
						receipt := *msg
						receipt.ContentType = model.ContentTypeReceiptMessage
						WriteTimeout(client.send, utils.ToJsonByte(receipt), WriteChannelTime)
						//设置最大回执SeqID
						_ = redis.Set(context.Background(), constant.RedisKey(constant.RedisChatMaxReceiptSeqID, msg.ChatID, receipt.SenderID), receipt.SeqID, redis.ForeverTime)
						continue
					}
					WriteTimeout(client.send, message, WriteChannelTime)
				}
			}

			go func(msg *model.Message) {
				SaveToDatabase(msg)
			}(msg)
		}
	}
}

func SaveToDatabase(msg *model.Message) {
	if msg.MessageType <= 0 {
		return
	}
	if msg.ContentType <= 0 {
		return
	}
	if msg.SenderID <= 0 {
		return
	}
	if msg.SeqID <= 0 {
		return
	}
	if len(msg.ChatID) <= 0 {
		return
	}
	if msg.ContentType == model.ContentTypeReceiptMessage {
		return
	}
	if msg.ContentType == model.ContentTypeReadMessage {
		return
	}
	//设置用户最大已读SeqID
	_ = redis.Set(context.Background(), constant.RedisKey(constant.RedisChatMaxReadSeqID, msg.ChatID, msg.SenderID), msg.SeqID, redis.ForeverTime)
	//设置会话的最新消息
	message := utils.ToJson(msg)
	_ = redis.HSet(context.Background(), constant.RedisChatLatestMessage, msg.ChatID, message)
	//设置会话的最近100-200条聊天记录
	MessageQueueKey := constant.RedisKey(constant.RedisChatMessageQueue, msg.ChatID)
	err := redis.ZAdd(context.Background(), MessageQueueKey, redis.Z{Score: float64(msg.SeqID), Member: message})
	if err != nil {
		logs.Error(ginx.Context, "redis.ZAdd error:", zap.Error(err))
	}
	count, err := redis.ZCard(context.Background(), MessageQueueKey)
	if err != nil {
		logs.Error(ginx.Context, "redis.ZCard error:", zap.Error(err))
	}
	if count > 200 {
		_, _ = redis.ZTrim(context.Background(), MessageQueueKey, int64(100))
	}
	_ = redis.Expire(context.Background(), MessageQueueKey, 3*timex.Day)
	//写入mysql
	logs.Debug(ginx.Context, "Create:", zap.String("msg", message))
	err = mysql.Create(ginx.Context, mysql.GetUserDB(), msg)
	if err != nil {
		logs.Error(ginx.Context, "mysql.Create Message error:", zap.Error(err))
	}
}

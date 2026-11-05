package chat

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/constant"
)

type GetListResp struct {
	ChatID   string      `json:"chatID"`
	TalkerID int         `json:"talkerID"`
	Talker   *model.User `json:"talker"`
}

func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	userChats := make([]*model.UserChat, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &userChats, map[string]any{
		"user_id":   userID,
		"is_delete": 0,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	// 所有会话ID
	resp := make([]*GetListResp, len(userChats))
	if len(userChats) == 0 {
		ginx.ResponsePageSuccess(c, resp, 1, 2048, len(resp))
		return
	}

	chatsID := make([]string, 0)
	talkersID := make([]int, 0)
	for i := range userChats {
		resp[i] = &GetListResp{
			ChatID:   userChats[i].ChatID,
			TalkerID: userChats[i].TalkerID,
		}
		chatsID = append(chatsID, userChats[i].ChatID)
		talkersID = append(talkersID, userChats[i].TalkerID)
	}

	//所有talker的用户信息
	talkerMap, err := repo.GetUsersByID(ctx, talkersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	for i := range resp {
		if _, ok := talkerMap[resp[i].TalkerID]; ok {
			resp[i].Talker = talkerMap[resp[i].TalkerID]
		}
	}

	ginx.ResponseSuccess(c, resp)
}

type GetUnreadAndLatestMessageResp struct {
	TotalUnread             int                       `json:"total_unread"`
	ChatUnread              map[string]int            `json:"chat_unread"`
	ChatTalkerReadMaxSeqID  map[string]int            `json:"chat_talker_read_max_seq_id"`
	ChatUserReceiptMaxSeqID map[string]int            `json:"chat_user_receipt_max_seq_id"`
	ChatLatestMessage       map[string]*model.Message `json:"chat_latest_message"`
}

func GetUnreadAndLatestMessage(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)

	result := GetUnreadAndLatestMessageResp{
		ChatUnread:              map[string]int{},
		ChatTalkerReadMaxSeqID:  map[string]int{},
		ChatUserReceiptMaxSeqID: map[string]int{},
		ChatLatestMessage:       map[string]*model.Message{},
	}
	chatsID, err := redis.SMembers(ctx, constant.RedisKey(constant.RedisUserChatID, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ginx.ResponseSuccess(c, result)
			return
		}
		ginx.ResponseServerError(c, err)
		return
	}
	if len(chatsID) == 0 {
		ginx.ResponseSuccess(c, result)
		return
	}

	latestMessages, err := redis.HMGet(ctx, constant.RedisChatLatestMessage, chatsID...)
	if err != nil {
		logs.Error(c, "redis.HMGet error", zap.Error(err))
		ginx.ResponseServerError(c, err)
		return
	}

	for i := range chatsID {
		latestMessage, ok := latestMessages[i].(string)
		if !ok {
			continue
		}

		message := &model.Message{}
		err = message.Unmarshal([]byte(latestMessage))
		if err != nil {
			logs.Error(c, "message.Unmarshal error", zap.Error(err))
			ginx.ResponseServerError(c, err)
			return
		}

		result.ChatLatestMessage[chatsID[i]] = message

		readSeqID := 0
		value, _ := redis.Get(ctx, constant.RedisKey(constant.RedisChatMaxReadSeqID, chatsID[i], userID))
		if value != "" {
			readSeqID, _ = strconv.Atoi(value)
		}
		unread := int(message.SeqID) - readSeqID
		if unread > 0 {
			result.TotalUnread += unread
			result.ChatUnread[chatsID[i]] = unread
		}

		talkerID := message.ReceiverID
		if talkerID == userID {
			talkerID = message.SenderID
		}
		talkerSeqID := 0
		value, _ = redis.Get(ctx, constant.RedisKey(constant.RedisChatMaxReadSeqID, chatsID[i], talkerID))
		if value != "" {
			talkerSeqID, _ = strconv.Atoi(value)
		}
		result.ChatTalkerReadMaxSeqID[chatsID[i]] = talkerSeqID

		receiptSeqID := 0
		value, _ = redis.Get(ctx, constant.RedisKey(constant.RedisChatMaxReceiptSeqID, chatsID[i], userID))
		if value != "" {
			receiptSeqID, _ = strconv.Atoi(value)
		}
		result.ChatUserReceiptMaxSeqID[chatsID[i]] = receiptSeqID
	}

	ginx.ResponseSuccess(c, result)
}

type CreateReq struct {
	TalkerID int `json:"talkerID" binding:"required"`
}

type CreateResp struct {
	Chat   *model.UserChat `json:"chat"`
	Talker *model.User     `json:"talker"`
}

func Create(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := CreateReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.TalkerID == userID {
		ginx.ResponseParamError(c)
		return
	}

	blacked, err := repo.IsBlacked(ctx, req.TalkerID, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if blacked {
		ginx.ResponseNotify(c, "对方已将你加入黑名单")
		return
	}

	result := CreateResp{}
	// 检查对方用户是否存在
	talker := model.User{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &talker, map[string]any{"id": req.TalkerID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "对方用户已注销")
		return
	}
	result.Talker = &talker

	// 检查是否已存在会话
	existChat := model.UserChat{}
	exist, err = mysql.Exist(ctx, mysql.GetUserDB(), &existChat, map[string]any{"user_id": userID, "talker_id": req.TalkerID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if exist {
		result.Chat = &existChat
		ginx.ResponseSuccess(c, result)
		return
	}

	// 创建新会话
	userChats, err := repo.CreateChat(ctx, userID, req.TalkerID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	result.Chat = userChats[0]
	ginx.ResponseSuccess(c, result)
}

type DeleteReq struct {
	ChatID string `json:"chatID"`
}

func Delete(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := DeleteReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userChat := model.UserChat{}
	condition := map[string]any{"chat_id": req.ChatID, "user_id": userID}
	err = mysql.FindOne(ctx, mysql.GetUserDB(), &userChat, condition)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	_, err = mysql.Updates(ctx, mysql.GetUserDB(), &model.UserChat{}, condition, map[string]any{
		"is_delete":   1,
		"user_delete": userID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

package model

import (
	"go.uber.org/zap"

	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/utils/constant"
	"tongcheng/utils/tool/jsoniter"
	"tongcheng/utils/x/timex"
)

const (
	MessageTypeSingle = 1
	MessageTypeGroup  = 2
	MessageTypeSite   = 3
	MessageTypeTip    = 4

	ContentTypeText           = 1
	ContentTypeImage          = 2
	ContentTypeVideo          = 3
	ContentTypeFile           = 4
	ContentTypeReceiptMessage = 5
	ContentTypeReadMessage    = 6
)

type Message struct {
	GormModel
	MsgID       string     `gorm:"column:msg_id;size:64;uniqueIndex;comment:消息ID;default:'';not null" json:"msgID"`
	SeqID       int64      `gorm:"column:seq_id;comment:序列ID;default:0;not null" json:"seqID"`
	ChatID      string     `gorm:"column:chat_id;size:64;index;comment:会话ID;default:'';not null" json:"chatID"`
	SenderID    int        `gorm:"column:sender_id;index;comment:发送者ID;default:0;not null" json:"senderID"`
	ReceiverID  int        `gorm:"column:receiver_id;index;comment:接收者ID;default:0;not null" json:"receiverID"`
	MessageType int        `gorm:"column:message_type;type:tinyint;comment:消息类型 1:2人会话 2:群会话 3:站内信 4:提示;default:0;not null" json:"messageType"`
	ContentType int        `gorm:"column:content_type;type:tinyint;comment:内容类型 1:文本 2:图片 3:视频 4:文件 5:回执 6:读取消息;default:0;not null" json:"contentType"`
	Content     string     `gorm:"column:content;size:1024;comment:内容;not null" json:"content"`
	SendTime    timex.Time `gorm:"column:send_time;type:timestamp;comment:发送时间" json:"sendTime"`
	Status      int        `gorm:"column:status;type:tinyint;comment:状态 0:正常 1:删除 2:撤回 3:修改;default:0;not null" json:"status"`
}

func (*Message) TableName() string {
	return "message"
}

func NewTipMessage(code int) *Message {
	return &Message{
		MsgID:       snowflake.MsgID(),
		MessageType: MessageTypeTip,
		ContentType: ContentTypeText,
		Content:     constant.GetCodeText(code),
		SendTime:    timex.GetBeijingTime(),
	}
}

func (m *Message) Marshal() []byte {
	b, err := jsoniter.Marshal(m)
	if err != nil {
		logs.Error(ginx.Context, "Message Marshal error:", zap.Error(err))
	}
	return b
}

func (m *Message) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, m)
}

func (m *Message) SetMsgID() {
	m.MsgID = snowflake.MsgID()
}

func (m *Message) SetChatID() {
	m.ChatID = snowflake.ChatID()
}

func (m *Message) SetSeqID(chatID string) error {
	seqID, err := redis.Incr(ginx.Context, constant.RedisKey(constant.RedisChatSeqID, chatID))
	if err != nil {
		logs.Error(ginx.Context, "SetSeqID error:", zap.Error(err))
		return err
	}
	m.SeqID = seqID

	return nil
}

func (m *Message) SetContent(content string) {
	m.Content = content
}

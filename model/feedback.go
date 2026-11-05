package model

import "tongcheng/utils/x/timex"

type UserFeedback struct {
	GormModel
	UserID      int        `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	Content     string     `gorm:"column:content;size:1024;comment:反馈内容;default:'';not null" json:"content"`
	Reply       string     `gorm:"column:reply;size:1024;comment:回复内容;default:'';not null" json:"reply"`
	ReplyAt     timex.Time `gorm:"column:reply_at;type:timestamp;comment:回复时间" json:"replyAt"`
	ReplyStatus int        `gorm:"column:reply_status;type:tinyint;comment:回复状态 0:未回复1:已回复;default:0;not null" json:"replyStatus"`
}

func (*UserFeedback) TableName() string {
	return "user_feedback"
}

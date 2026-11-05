package model

import (
	"tongcheng/utils/x/timex"
)

type AdminLoginLog struct {
	GormModel
	AdminID   int        `gorm:"column:admin_id;index;comment:管理员ID;default:0;not null" json:"adminID"`
	LoginIP   string     `gorm:"column:login_ip;size:32;comment:登录IP;default:'';not null" json:"loginIP"`
	LoginTime timex.Time `gorm:"column:login_time;type:timestamp;comment:登录时间;default:null" json:"loginTime"`
	DeviceID  string     `gorm:"column:device_id;size:128;comment:设备ID;default:'';not null" json:"deviceID"`
}

func (*AdminLoginLog) TableName() string {
	return "admin_login_log"
}

type AdminOperationLog struct {
	GormModel
	AdminID   int    `gorm:"column:admin_id;index;comment:管理员ID;default:0;not null" json:"adminID"`
	Operation string `gorm:"column:operation;size:256;comment:操作;default:'';not null" json:"operation"`
	Param     string `gorm:"column:param;size:256;comment:参数;default:'';not null" json:"param"`
	Result    string `gorm:"column:result;size:256;comment:结果;default:'';not null" json:"result"`
}

func (*AdminOperationLog) TableName() string {
	return "admin_operation_log"
}

type UserLoginLog struct {
	GormModel
	UserID    int        `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	LoginIP   string     `gorm:"column:login_ip;size:32;comment:登录IP;default:'';not null" json:"loginIP"`
	LoginTime timex.Time `gorm:"column:login_time;type:timestamp;comment:登录时间;default:null" json:"loginTime"`
	DeviceID  string     `gorm:"column:device_id;size:128;comment:设备;default:'';not null" json:"deviceID"`
}

func (*UserLoginLog) TableName() string {
	return "user_login_log"
}

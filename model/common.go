package model

import (
	"time"

	"tongcheng/utils/x/timex"
)

type TableModel interface {
	TableName() string
}

type GormModel struct {
	ID        int        `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CreatedAt timex.Time `gorm:"column:created_at;type:timestamp;comment:创建时间;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;comment:更新时间;autoUpdateTime" json:"-"` //autoUpdateTime 对标准time类型生效
}

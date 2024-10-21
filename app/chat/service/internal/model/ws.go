package model

import (
	"gorm.io/gorm"
	"time"
)

type GroupBind struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:ID" json:"id"`
	GroupId   string         `gorm:"type:varchar(255);not null;comment:组编号" json:"group_id"`
	MemberId  uint64         `gorm:"column:member_id;type:bigint(20);comment:用户id;" json:"member_id"`
	CreatedAt time.Time      // 创建时间
	UpdatedAt time.Time      // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 删除时间
}

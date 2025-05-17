package model

import (
	"time"

	"gorm.io/gorm"
)

type Traffic struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Date      time.Time      `gorm:"uniqueIndex" json:"date"` // 每日唯一
	Count     int            `json:"count"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

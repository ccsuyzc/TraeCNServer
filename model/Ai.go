package model

import (
	"time"

	"gorm.io/gorm"
)

// AI问答会话表 - 存储用户与AI的对话记录
// 包含问题、回答和token使用统计
type AIConversation struct {
	gorm.Model
	SessionID int `gorm:"index;not null"` // 会话ID
	SerialNumber int       `gorm:"index;"`                 // 序列号
	UserID       uint      `gorm:"index;not null"`         // 用户ID
	Role         string    `gorm:"size:20;default:'user'"` // 角色：user/assistant/system
	Question     string    `gorm:"type:longtext;not null"` // 问题
	Answer       string    `gorm:"type:longtext;not null"` // 回答
	AiModel      string    `gorm:"size:50;not null"`       // 使用的模型
	InputTokens  int       `gorm:"not null"`               // 输入的token数量
	OutputTokens int       `gorm:"not null"`               // 输出的token数量
	TotalTokens  int       `gorm:"index;not null"`         // 总token数量
	CreatedTime  time.Time // 创建时间

	User User `gorm:"foreignKey:UserID"` // 关联用户
}

// AI使用统计表 - 按月记录用户AI使用情况
// 包含token消耗和API调用次数
type AIUsageStatistics struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"uniqueIndex:uidx_user_month;not null"`
	Month       string    `gorm:"uniqueIndex:uidx_user_month;type:char(7);not null"`
	TotalTokens int       `gorm:"not null"`
	APICalls    int       `gorm:"not null"`
	LastUpdate  time.Time // 最后更新时间
}

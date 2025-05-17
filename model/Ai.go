package model

import (
	"gorm.io/gorm"
	"time"
)

type AIModelConfig struct {
	gorm.Model
	// 基础信息
	ModelName   string `gorm:"size:100;uniqueIndex:idx_model_version"` // 模型名称
	Version     string `gorm:"size:50;uniqueIndex:idx_model_version"`  // 版本号
	Provider    string `gorm:"size:50"`                                // 提供商
	ModelType   string `gorm:"size:20"`                                // 类型：text/image/multimodal
	Token       string `gorm:"size:255"`                               // 加密存储的API密钥
	Description string `gorm:"type:text"`                              // 描述信息
	MaxTokens   int    `gorm:"default:2048"`                           // 最大token数
	MaxContext  int    `gorm:"default:4096"`                           // 最大上下文长度
	// 连接配置
	APIEndpoint string `gorm:"size:255"`   // API地址
	APIKey      string `gorm:"type:text"`  // 加密存储的API密钥
	Timeout     int    `gorm:"default:30"` // 超时时间（秒）

	// 运行参数
	DefaultParams int `gorm:"type:int"` // 默认参数

	// 状态管理
	IsDefault   bool `gorm:"default:false"` // 是否默认配置
	IsEnabled   bool `gorm:"default:true"`  // 是否启用
	Maintenance bool `gorm:"default:false"` // 维护状态

	// 版本控制
	BaseVersion string `gorm:"size:50"`   // 基础版本号
	Changelog   string `gorm:"type:text"` // 变更说明

	// 监控指标
	SuccessRate float64 `gorm:"default:0.0"` // 最近成功率
	AvgLatency  float64 `gorm:"default:0.0"` // 平均响应时间（秒）

	// 审计信息
	LastUsed  time.Time // 最后使用时间
	UpdatedBy uint      // 最后修改人
}

// AI问答会话表 - 存储用户与AI的对话记录
// 包含问题、回答和token使用统计
type AIConversation struct {
	gorm.Model
	SessionID    int       `gorm:"index;not null"`         // 会话ID
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

// // AI使用统计表 - 按月记录用户AI使用情况
// // 包含token消耗和API调用次数
// type AIUsageStatistics struct {
// 	ID          uint      `gorm:"primaryKey"`
// 	UserID      uint      `gorm:"uniqueIndex:uidx_user_month;not null"`
// 	Month       string    `gorm:"uniqueIndex:uidx_user_month;type:char(7);not null"`
// 	TotalTokens int       `gorm:"not null"`
// 	APICalls    int       `gorm:"not null"`
// 	LastUpdate  time.Time // 最后更新时间
// }

// deepseek生成的
// 请求结构体
type ChatRequest struct {
	Messages         []MessageAI `json:"messages"`
	Model            string      `json:"model"`
	// FrequencyPenalty float64     `json:"frequency_penalty"`
	MaxTokens        int         `json:"max_tokens"`
	Stream           bool        `json:"stream" default:"false"`
	// 其他字段根据需要添加...
}

type MessageAI struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 大模型响应结构体
type DeepseekResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Choice struct {
	MessageAI struct {
		Content string `json:"content"`
	} `json:"message"`
}


// deepseek第二次生成的方案,非流式传输，然后是在ai_controller2.go中进行处理
type ChatSession struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"`
	SessionUUID string `gorm:"uniqueIndex;size:36;not null"` // 使用UUID作为会话标识
	AiModel     string `gorm:"size:50;not null" json:"model"`
	Title       string `gorm:"size:200"` // 会话标题（可选）
}

type ChatMessage struct {
	gorm.Model
	SessionID    uint   `gorm:"index;not null"`   // 关联ChatSession
	RequestID    string `gorm:"index;size:36"`    // 唯一请求标识
	Role         string `gorm:"size:20;not null"` // user/assistant/system
	Content      string `gorm:"type:longtext;not null"` // 消息内容
	IsResponse   bool   `gorm:"index"` // 是否是模型响应
	PromptTokens int    `gorm:"not null;default:0"` // 输入的token数量

	StreamMode bool `gorm:"index"` // 是否流式模式
	Completed  bool `gorm:"index"` // 是否已完成流式传输

	CompletionTokens int    `gorm:"not null;default:0"` // 输出的token数量
	TotalTokens      int    `gorm:"not null;default:0"` // 总token数量
	StatusCode       int    `gorm:"not null;default:200"` // 响应状态码
	Error            string `gorm:"type:text"`           // 错误信息
}

// 新增流式消息块表
type ChatMessageChunk struct {
	gorm.Model
	MessageID uint   `gorm:"index;not null"` // 关联ChatMessage
	ChunkUUID string `gorm:"index;size:36"`  // 消息块唯一标识
	Content   string `gorm:"type:text;not null"`
	IsLast    bool   `gorm:"index"`    // 是否为最后一个块
	Order     int    `gorm:"not null"` // 块顺序
}

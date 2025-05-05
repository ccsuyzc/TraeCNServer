package model
import (
    "time"
    "gorm.io/gorm"
)

// 用户搜索记录模型
type SearchHistory struct {
    gorm.Model
    UserID       uint      `gorm:"index;not null"`
    SearchContent string   `gorm:"type:text;not null"`
    Timestamp    time.Time `gorm:"index"`
    User         User      `gorm:"foreignKey:UserID"`
}

// 用户阅读记录模型 
type ReadingHistory struct {
    gorm.Model
    UserID      uint      `gorm:"index;not null"` // 用户ID
    ArticleID   uint      `gorm:"index;not null"` // 文章ID
    Duration    int       `gorm:"default:0"`      // 阅读时长
    Timestamp   time.Time `gorm:"index"`         // 阅读时间
    User        User      `gorm:"foreignKey:UserID"`    // 关联用户
    Article     Article   `gorm:"foreignKey:ArticleID"` // 关联文章
}
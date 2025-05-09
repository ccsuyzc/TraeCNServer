package model
import (
    "fmt"
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
    // 创建前验证用户存在
    func (s *SearchHistory) BeforeCreate(tx *gorm.DB) error {
        var user User
        if result := tx.First(&user, s.UserID); result.Error != nil {
            return fmt.Errorf("用户ID %d 不存在", s.UserID)
        }
        return nil
    }

// 用户阅读记录模型 
type ReadingHistory struct {
    gorm.Model
    UserID      uint      `gorm:"index;not null" json:"user_id"`  // 用户ID
    ArticleID   uint      `gorm:"index;not null" json:"article_id"`  // 文章ID
    Duration    int       `gorm:"default:0" json:"duration"`       // 阅读时长（秒）
    Timestamp   time.Time `gorm:"index"`           // 阅读时间
    User        User      `gorm:"foreignKey:UserID"`  // 关联用户
    Article     Article   `gorm:"foreignKey:ArticleID"` // 关联文章
}

    // 创建前验证用户和文章存在
    func (r *ReadingHistory) BeforeCreate(tx *gorm.DB) error {
        var user User
        if result := tx.First(&user, r.UserID); result.Error != nil {
            return fmt.Errorf("用户ID %d 不存在", r.UserID)
        }

        var article Article
        if result := tx.First(&article, r.ArticleID); result.Error != nil {
            return fmt.Errorf("文章ID %d 不存在", r.ArticleID)
        }
        return nil
    }

    // 自动创建时间
    func (r *ReadingHistory) BeforeSave(tx *gorm.DB) error {
        if r.Timestamp.IsZero() {
            r.Timestamp = time.Now()
        }
        return nil
    }
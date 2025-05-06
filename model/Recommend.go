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
    UserID      uint      `gorm:"index;not null"`
    ArticleID   uint      `gorm:"index;not null"`
    Duration    int       `gorm:"default:0"`
    Timestamp   time.Time `gorm:"index"`
    User        User      `gorm:"foreignKey:UserID"`
    Article     Article   `gorm:"foreignKey:ArticleID"`


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
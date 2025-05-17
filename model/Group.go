package model

import (
	"time"

	"gorm.io/gorm"
)

// 帖子模型（替代原Article在群组场景中的使用）
type Post struct {
	gorm.Model
	Content      string  `gorm:"type:text;not null"`          // 帖子内容
	Images       string  `gorm:"type:text"`                   // 图片URL列表，存储为JSON字符串
	GroupID      uint    `gorm:"index"`                       // 所属圈子ID
	UserID       uint    `gorm:"index;not null"`              // 发帖用户ID
	Status       string  `gorm:"size:20;default:'published'"` // 状态: draft/published/deleted
	AuditStatus  string  `gorm:"size:20;default:'pending'"`   // 审核状态: pending/approved/rejected
	LikeCount    uint    `gorm:"default:0"`                   // 点赞数
	CommentCount uint    `gorm:"default:0"`                   // 评论数
	HotScore     float64 `gorm:"default:0"`                   // 热度评分（用于排序）
	IsTop        bool    `gorm:"default:false"`               // 是否置顶
	IsLiked      bool    `gorm:"default:false"`               // 是否已点赞
	// 关联关系
	Group         GroupN         `gorm:"foreignKey:GroupID"`
	Author        User           `gorm:"foreignKey:UserID"`
	GroupComments []GroupComment `gorm:"foreignKey:PostID"` // 帖子评论关系
	GroupLikes    []GroupLike    `gorm:"foreignKey:PostID"` // 帖子点赞关系
}

// 群组表 - 存储用户创建的群组
type GroupN struct {
	gorm.Model
	Name          string `gorm:"size:100;not null;uniqueIndex"` // 圈子名称（唯一）
	Description   string `gorm:"type:text;not null"`            // 圈子描述
	CreatorID     uint   `gorm:"index;not null"`                // 创建者ID
	AvatarURL     string `gorm:"size:255"`                      // 圈子头像
	MemberCount   uint   `gorm:"default:0"`                     // 成员数量
	PostCount     uint   `gorm:"default:0"`                     // 帖子数量
	IsPublic      bool   `gorm:"default:true"`                  // 是否公开
	AuditRequired bool   `gorm:"default:false"`                 // 发帖是否需要审核
	Status        string `gorm:"size:20;default:'Normal'"`      // 状态: Normal/banned
	JoinMethod    string `gorm:"size:20;default:'public'"`      // 加入方式: public/approval/invite
	IsLocked      bool   `gorm:"default:false"`                 // 是否锁定（禁止发帖）
	Reason        string `gorm:"type:text"`                     // 锁定原因
	// 关联关系
	Creator User   `gorm:"foreignKey:CreatorID"`
	Members []User `gorm:"many2many:user_groups;"`
	Posts   []Post `gorm:"foreignKey:GroupID"`
}

// 用户群组中间表 - 用户和群组的多对多关系映射
type UserGroup struct {
	UserID   uint      `gorm:"primaryKey;autoIncrement:false"`
	GroupID  uint      `gorm:"primaryKey;autoIncrement:false"`
	JoinedAt time.Time // 加入时间
	Role     string    `gorm:"size:20;default:'member'"`   // 角色：member/admin
	Status   string    `gorm:"size:20;default:'approved'"` // 状态: pending/approved/rejected

	CanPost   bool `gorm:"default:true"`  // 是否允许发帖
	CanManage bool `gorm:"default:false"` // 是否允许管理

	User    User   `gorm:"foreignKey:UserID"`
	GroupNs GroupN `gorm:"foreignKey:GroupID"`
}

// 帖子评论模型
type GroupComment struct {
	gorm.Model
	Content   string `gorm:"type:text;not null"` // 评论内容
	UserID    uint   `gorm:"index;not null"`     // 评论者ID
	PostID    uint   `gorm:"index;not null"`     // 所属帖子ID
	ParentID  *uint  `gorm:"index"`              // 父评论ID
	Depth     uint   `gorm:"default:0"`          // 评论层级深度
	TreePath  string `gorm:"size:255;index"`     // 树形路径（格式：rootID/.../parentID）
	LikeCount uint   `gorm:"default:0"`          // 点赞数
	IsDeleted bool   `gorm:"default:false"`      // 是否已删除

	// Images     []Image                                  // 图片URL列表

	// 关联关系
	User    User           `gorm:"foreignKey:UserID"`
	Post    Post           `gorm:"foreignKey:PostID"`
	Replies []GroupComment `gorm:"foreignKey:ParentID"` // 子评论
}

type Image struct {
	gorm.Model
	URL       string       `gorm:"size:255;not null"`    // 图片URL
	UserID    uint         `gorm:"index"`                // 上传者ID
	User      User         `gorm:"foreignKey:UserID"`    // 上传者
	PostID    uint         `gorm:"index"`                // 所属帖子ID
	Post      Post         `gorm:"foreignKey:PostID"`    // 所属帖子
	CommentID uint         `gorm:"index"`                // 所属评论ID
	Comment   GroupComment `gorm:"foreignKey:CommentID"` // 所属评论
}

type GroupLike struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex:idx_user_target"`
	PostID    uint `gorm:"uniqueIndex:idx_user_target;index"`
	CommentID uint `gorm:"uniqueIndex:idx_user_target;index"`
}

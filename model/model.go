package model

import (
	"time"

	"gorm.io/gorm"
)

// Message 表 - 存储消息
type Message struct {
	gorm.Model
	SenderID uint `gorm:"index;not null"`

	ReceiverID  uint      `gorm:"index;not null"`
	Content     string    `gorm:"type:text;not null"` // 消息内容
	MessageType string    `gorm:"size:20;default:'text'"`
	SentAt      time.Time `gorm:"index"`
	Read        bool      `gorm:"default:false"`
}

// Conversation 表 - 存储用户之间的会话信息
type Conversation struct {
	gorm.Model
	Participant1  uint      `gorm:"index;not null"`
	Participant2  uint      `gorm:"index;not null"`
	LastMessageID uint      `gorm:"index"`
	UnreadCount   uint      `gorm:"default:0"`
	Archived      bool      `gorm:"default:false"`
	LastActiveAt  time.Time `gorm:"index"`
}

type User struct {
	gorm.Model
	Username             string    `gorm:"uniqueIndex;size:50;not null"` // 用户名
	PasswordHash         string    `gorm:"size:60;not null"`             // 密码
	Phone                string    `gorm:"uniqueIndex;size:15"`          // 手机号
	Email                string    `gorm:"uniqueIndex;size:100;"`        // 邮箱
	EmailVerified        bool      `gorm:"default:false"`                // 邮箱验证状态
	VerificationCode     string    `gorm:"size:6"`                       // 验证码
	VerificationExpiry   time.Time // 验证码过期时间
	AvatarURL            string    `gorm:"size:255"` // 头像URL
	RegistrationTime     time.Time // 注册时间
	LastLoginTime        time.Time // 最后登录时间
	Role                 string    `gorm:"size:10;default:'user'"` // 用户角色 ,管理员root
	IsActive             bool      `gorm:"default:true"`           // 是否激活
	NumberOfFans         uint      `gorm:"default:0"`              // 粉丝数
	NumberOfFollow       uint      `gorm:"default:0"`              // 关注数
	NumberOfLikes        uint      `gorm:"default:0"`              // 点赞数
	NumberOfComments     uint      `gorm:"default:0"`              // 评论数
	NumberOfArticles     uint      `gorm:"default:0"`              // 文章数
	NumberOfFavorites    uint      `gorm:"default:0"`              // 收藏数
	NumberOfViews        uint      `gorm:"default:0"`              // 浏览量
	NumberOfShares       uint      `gorm:"default:0"`              // 分享数
	NumberOfReports      uint      `gorm:"default:0"`              // 举报数
	PersonalSignature    string    `gorm:"size:255"`               // 个性签名
	PersonalIntroduction string    `gorm:"size:255"`               // 个人简介
	PersonalWebsite      string    `gorm:"size:255"`               // 个人网站
	PersonalLocation     string    `gorm:"size:255"`               // 个人位置
	QQ                   string    `gorm:"size:255"`               // QQ
	WeChat               string    `gorm:"size:255"`               // 微信
	Weibo                string    `gorm:"size:255"`               // 微博
	Twitter              string    `gorm:"size:255"`               // Twitter

	SolvedProblems uint `gorm:"default:0"` // 已解决题目数
	SubmitCount    uint `gorm:"default:0"` // 总提交次数
	AcceptCount    uint `gorm:"default:0"` // 通过次数
	// 关联关系
	Articles        []Article         `gorm:"foreignKey:UserID"`   // 文章
	Comments        []Comment         `gorm:"foreignKey:UserID"`    // 评论
	Likes           []Like            `gorm:"foreignKey:UserID"`    // 点赞
	Favorites       []Favorite        `gorm:"foreignKey:UserID"`    // 收藏
	AIConversations []AIConversation  `gorm:"foreignKey:UserID"`    // AI对话
	// AIUsage         AIUsageStatistics `gorm:"foreignKey:UserID"`    // AI使用统计 第一次尝试，先废弃。
	Followers       []User            `gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowedID;joinReferences:FollowerID"` // 粉丝
	Following       []User            `gorm:"many2many:user_follows;foreignKey:ID;joinForeignKey:FollowerID;joinReferences:FollowedID"` // 关注
}

// 文章表 - 存储博客文章内容
// 包含文章元数据、内容和关联关系
const (
	ArticleStatusPublished     = "published"      // 已发布
	ArticleStatusDraft         = "draft"          // 草稿
	ArticleStatusDeleted       = "deleted"        // 已删除
	ArticleStatusPendingReview = "pending_review" // 审核中
	ArticleStatusRejected      = "rejected"       // 驳回
)

type Article struct {
	gorm.Model
	UserID       uint      `gorm:"index;not null"`         // 用户ID
	UserName     string    `gorm:"size:255"`               // 用户名
	CategoryID   uint      `gorm:"index;not null"`         // 分类ID
	Title        string    `gorm:"size:200;not null"`      // 文章标题
	Description  string    `gorm:"size:255"`               // 文章描述
	Cover        string    `gorm:"size:255"`               // 文章封面链接
	Content      string    `gorm:"type:longtext;not null"` // 文章内容
	PublishTime  time.Time `gorm:"index"`                  // 发布时间
	UpdateTime   time.Time // 更新时间
	SubmitTime   time.Time // 审核时间
	IsDeleted    bool      `gorm:"default:false"`            // 是否删除
	Status       string    `gorm:"size:20;default:'draft'"`  // 文章状态 draft为草稿，published为发布，Review为审核中，Rejected为审核未通过
	IsTop        bool      `gorm:"default:false"`            // 是否置顶
	IsOriginal   bool      `gorm:"default:false"`            // 是否原创
	IsPublished  bool      `gorm:"default:false"`            // 是否发布
	IsPost       bool      `gorm:"default:false"`            // 是否是帖子
	PostType     string    `gorm:"size:20;default:'normal'"` // 帖子类型
	ViewCount    int       `gorm:"default:0"`                // 浏览量
	FavoriteCount int       `gorm:"default:0"`                // 收藏数 
	LikeCount    int       `gorm:"default:0"`                // 点赞数
	CommentCount int       `gorm:"default:0"`                // 评论数
	ShareCount   int       `gorm:"default:0"`                // 分享数
	IsDraft      bool      `gorm:"-"`                        // 虚拟字段用于查询
	// GroupID      *uint     `gorm:"index"`                    // 所属圈子ID（指针类型允许nil）
	PostVisible  string    `gorm:"size:20"`                  // 可见性：public/group/members
	PinPosition  int       // 置顶位置

	IsSuperAdmin  bool      `gorm:"default:false"` // 是否超级管理员
	AdminRoles    string    `gorm:"size:255"`      // 管理角色（逗号分隔）
	LastAdminTime time.Time // 最后管理操作时间
	AdminNotes    string    `gorm:"type:text"` // 管理备注
	AdminActions  string    `gorm:"type:text"` // 管理操作记录（JSON格式）
	AdminIP       string    `gorm:"size:50"`   // 管理操作IP

	AuditStatus  string    `gorm:"size:20;default:'pending'"` // 审核状态（pending/approved/rejected）
	AuditTime    time.Time // 审核时间
	AuditorID    *uint     // 审核人ID
	RejectReason string    `gorm:"type:text"`            // 驳回原因
	Auditor      User      `gorm:"foreignKey:AuditorID"` // 审核人
	// 关联关系

	ProblemID  *uint   `gorm:"index"`                // 关联题目ID
	Problem    Problem `gorm:"foreignKey:ProblemID"` // 关联题目
	IsSolution bool    `gorm:"default:false"`        // 是否是题解

	// Group     Group      `gorm:"foreignKey:GroupID"` // 所属圈子（关联关系）
	User      User       `gorm:"foreignKey:UserID"`
	Category  Category   `gorm:"foreignKey:CategoryID"`
	Tags      []Tag      `gorm:"many2many:article_tags;"`
	Comments  []Comment  `gorm:"foreignKey:ArticleID"`
	Likes     []Like     `gorm:"foreignKey:ArticleID"`
	Favorites []Favorite `gorm:"foreignKey:ArticleID"`
}

// 新增管理员操作日志表
type AdminLog struct {
	gorm.Model
	AdminID     uint   `gorm:"index;not null"` // 操作管理员ID
	TargetType  string `gorm:"size:50"`        // 操作对象类型（article/group/user）
	TargetID    uint   `gorm:"index"`          // 操作对象ID
	Action      string `gorm:"size:50"`        // 操作类型（approve/delete/edit）
	Description string `gorm:"type:text"`      // 操作描述
	IPAddress   string `gorm:"size:45"`        // 操作IP地址

	Admin User `gorm:"foreignKey:AdminID"`
}

// 文章标签中间表 - 文章和标签的多对多关系映射
type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey"` // 文章ID
	TagID     uint `gorm:"primaryKey"` // 标签ID
}

// 分类表 - 文章分类管理
// 用于对文章进行分类组织
type Category struct {
	ID          uint   `gorm:"primaryKey"`                   // 分类ID
	Name        string `gorm:"uniqueIndex;size:50;not null"` // 分类名称
	Tap         string `gorm:"size:20"`                      // 分类名称
	Description string `gorm:"size:255"`                     // 分类描述
}

// 标签表 - 文章标签管理
// 用于标记文章主题
type Tag struct {
	ID   uint   `gorm:"primaryKey"`                   // 标签ID
	Name string `gorm:"uniqueIndex;size:50;not null"` // 标签名称
}

// 评论表 - 存储文章评论
// 支持多级回复和评论管理
type Comment struct {
	// gorm.Model
	ID              uint      `gorm:"primaryKey"`         // 评论ID
	ArticleID       uint      `gorm:"index;not null"`     // 文章ID
	UserID          uint      `gorm:"index;not null"`     // 用户ID
	UserName        string    `gorm:"size:255"`           // 用户名
	ParentCommentID *uint     `gorm:"index"`              // 父评论ID
	ParentUserID    *uint     `gorm:"index"`              // 父评论用户ID
	ParentUserName  string    `gorm:"size:255"`           // 父评论用户名
	Content         string    `gorm:"type:text;not null"` // 评论内容
	CommentTime     time.Time // 评论时间
	IsApproved      bool      `gorm:"default:false"` // 是否通过审核
	IsDeleted       bool      `gorm:"default:false"` // 是否删除

	// 自关联关系
	Replies []Comment `gorm:"foreignKey:ParentCommentID"`
	User    User      `gorm:"foreignKey:UserID"`
	Article Article   `gorm:"foreignKey:ArticleID"`
}

// 点赞表 - 记录用户点赞
// 可对文章或评论点赞
type Like struct {
	ID        uint      `gorm:"primaryKey"`     // 点赞ID
	UserID    uint      `gorm:"index;not null"` // 用户ID
	ArticleID *uint     `gorm:"index"`          // 文章ID
	CommentID *uint     `gorm:"index"`          // 评论ID
	LikeTime  time.Time // 点赞时间

	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
	Article Article `gorm:"foreignKey:ArticleID"` // 关联文章
	Comment Comment `gorm:"foreignKey:CommentID"` // 关联评论
}

// 收藏表 - 记录用户收藏的文章
// 用于个人收藏管理
type Favorite struct {
   gorm.Model
	UserID       uint      `gorm:"index;not null"` // 用户ID
	ArticleID    uint      `gorm:"index;not null"` // 文章ID
	FavoriteTime time.Time // 收藏时间

	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
	Article Article `gorm:"foreignKey:ArticleID"` // 关联文章
}





// 用户关注关系表
type UserFollow struct {
	FollowerID uint      `gorm:"primaryKey;autoIncrement:false"` // 关注者ID
	FollowedID uint      `gorm:"primaryKey;autoIncrement:false"` // 被关注者ID
	CreatedAt  time.Time // 关注时间

	Follower User `gorm:"foreignKey:FollowerID"`
	Followed User `gorm:"foreignKey:FollowedID"`
}

// 添加圈子审核表
type GroupApply struct {
	gorm.Model
	GroupID uint   `gorm:"index"`
	UserID  uint   `gorm:"index"`
	Message string `gorm:"type:text"`
	Status  string `gorm:"size:20"` // pending/approved/rejected
}

// 题目表 - 存储编程题目信息
type Problem struct {
	gorm.Model
	Title          string       `gorm:"size:200;not null"`       // 题目标题
	Description    string       `gorm:"type:longtext;not null"`  // 题目描述
	Difficulty     string       `gorm:"size:20;not null;index"`  // 难度等级（easy/medium/hard）
	SampleInput    string       `gorm:"type:text"`               // 示例输入
	SampleOutput   string       `gorm:"type:text"`               // 示例输出
	AuthorID       uint         `gorm:"index;not null"`          // 题目创建者ID
	TimeLimit      int          `gorm:"default:1000"`            // 时间限制（毫秒）
	MemoryLimit    int          `gorm:"default:128"`             // 内存限制（MB）
	Tags           []ProblemTag `gorm:"many2many:problem_tags;"` // 题目标签
	Hint           string       `gorm:"type:text"`               // 解题提示
	TotalAttempts  uint         `gorm:"default:0"`               // 总尝试次数
	AcceptanceRate float64      `gorm:"default:0"`               // 通过率
	IsPublished    bool         `gorm:"default:false"`           // 是否发布
	TestCases      []TestCase   `gorm:"foreignKey:ProblemID"`    // 测试用例
	Solutions      []Article    `gorm:"foreignKey:ProblemID"`    // 关联题解
	Submissions    []Submission `gorm:"foreignKey:ProblemID"`    // 提交记录

	Author User `gorm:"foreignKey:AuthorID"` // 题目作者
}

// 测试用例表 - 存储题目的测试用例
type TestCase struct {
	gorm.Model
	ProblemID      uint   `gorm:"index;not null"`     // 关联题目ID
	Input          string `gorm:"type:text;not null"` // 输入数据
	ExpectedOutput string `gorm:"type:text;not null"` // 期望输出
	IsHidden       bool   `gorm:"default:false"`      // 是否隐藏用例
	Score          int    `gorm:"default:0"`          // 测试用例分数（预留字段）

	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
}

// 提交记录表 - 存储用户代码提交记录
type Submission struct {
	gorm.Model
	UserID        uint   `gorm:"index;not null"`                  // 用户ID
	ProblemID     uint   `gorm:"index;not null"`                  // 题目ID
	Code          string `gorm:"type:longtext;not null"`          // 提交代码
	Language      string `gorm:"size:20;not null"`                // 编程语言
	Status        string `gorm:"size:50;index;default:'pending'"` // 判题状态
	ExecutionTime int    `gorm:"default:0"`                       // 执行时间（ms）
	MemoryUsed    int    `gorm:"default:0"`                       // 内存使用（KB）
	PassedCases   int    `gorm:"default:0"`                       // 通过用例数
	TotalCases    int    `gorm:"default:0"`                       // 总用例数
	ErrorMessage  string `gorm:"type:text"`                       // 错误信息
	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
}

// 题目标签表
type ProblemTag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"uniqueIndex;size:50;not null"` // 标签名称
	Color string `gorm:"size:20"`                      // 标签颜色（前端展示）
}

// 用户题目状态表 - 记录用户解题进度
type UserProblemStatus struct {
	UserID       uint      `gorm:"primaryKey;autoIncrement:false"`
	ProblemID    uint      `gorm:"primaryKey;autoIncrement:false"`
	IsSolved     bool      `gorm:"default:false"` // 是否已解决
	Attempts     uint      `gorm:"default:0"`     // 尝试次数
	LastSubmitAt time.Time // 最后提交时间
	BestTime     int       `gorm:"default:0"` // 最佳用时（ms）
	BestMemory   int       `gorm:"default:0"` // 最佳内存（KB）

	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	a.PublishTime = time.Now()
	return nil
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	c.CommentTime = time.Now()
	return nil
}

func (l *Like) BeforeCreate(tx *gorm.DB) error {
	l.LikeTime = time.Now()
	return nil
}

// // UserGroup的创建/删除钩子
// func (ug *UserGroup) AfterCreate(tx *gorm.DB) error {
// 	return tx.Model(&Group{}).Where("id = ?", ug.GroupID).
// 		Update("member_count", gorm.Expr("member_count + 1")).Error
// }

// func (ug *UserGroup) AfterDelete(tx *gorm.DB) error {
// 	return tx.Model(&Group{}).Where("id = ?", ug.GroupID).
// 		Update("member_count", gorm.Expr("GREATEST(member_count - 1, 0)")).Error
// }

// UserFollow的创建/删除钩子
func (uf *UserFollow) AfterCreate(tx *gorm.DB) error {
	tx.Model(&User{}).Where("id = ?", uf.FollowerID).
		Update("number_of_follow", gorm.Expr("number_of_follow + 1"))
	tx.Model(&User{}).Where("id = ?", uf.FollowedID).
		Update("number_of_fans", gorm.Expr("number_of_fans + 1"))
	return nil
}

func (uf *UserFollow) AfterDelete(tx *gorm.DB) error {
	tx.Model(&User{}).Where("id = ?", uf.FollowerID).
		Update("number_of_follow", gorm.Expr("GREATEST(number_of_follow - 1, 0)"))
	tx.Model(&User{}).Where("id = ?", uf.FollowedID).
		Update("number_of_fans", gorm.Expr("GREATEST(number_of_fans - 1, 0)"))
	return nil
}

// 用户关注操作方法
func (u *User) Follow(db *gorm.DB, userID uint) error {
	return db.Create(&UserFollow{FollowerID: u.ID, FollowedID: userID}).Error
}

// 下面这个是无需验证的版本
func FollowY(db *gorm.DB, userID1, uerID2 uint) error {
	return db.Create(&UserFollow{FollowerID: userID1, FollowedID: uerID2}).Error
}

// 用户取消关注操作方法
func (u *User) Unfollow(db *gorm.DB, userID uint) error {
	return db.Where("follower_id = ? AND followed_id = ?", u.ID, userID).Delete(&UserFollow{}).Error
}

// 无需验证的取消关注的方法
func UnfollowY(db *gorm.DB, userID1, userID2 uint) error {
	return db.Where("follower_id = ? AND followed_id = ?", userID1, userID2).Delete(&UserFollow{}).Error
}

// // Article的创建/删除钩子
// func (a *Article) AfterCreate(tx *gorm.DB) error {
// 	if a.GroupID != nil {
// 		return tx.Model(&Group{}).Where("id = ?", *a.GroupID).
// 			Update("post_count", gorm.Expr("post_count + 1")).Error
// 	}
// 	return nil
// }

// func (a *Article) AfterDelete(tx *gorm.DB) error {
// 	if a.GroupID != nil {
// 		return tx.Model(&Group{}).Where("id = ?", *a.GroupID).
// 			Update("post_count", gorm.Expr("GREATEST(post_count - 1, 0)")).Error
// 	}
// 	return nil
// }

// 文章审核钩子函数
func (a *Article) BeforeUpdate(tx *gorm.DB) error {
	// 当审核状态变化时自动记录时间
	if tx.Statement.Changed("AuditStatus") {
		a.AuditTime = time.Now()
	}
	return nil
}

// // 管理员权限校验方法
// func (u *User) HasAdminPermission(requiredRole string) bool {
//     if u.IsSuperAdmin {
//         return true
//     }

//     roles := strings.Split(u.AdminRoles, ",")
//     for _, role := range roles {
//         if role == requiredRole {
//             return true
//         }
//     }
//     return false
// }

// 提交记录创建后更新统计信息
func (s *Submission) AfterCreate(tx *gorm.DB) error {
	// 更新用户提交统计
	if err := tx.Model(&User{}).Where("id = ?", s.UserID).
		Updates(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
		}).Error; err != nil {
		return err
	}

	// 更新题目统计
	updates := map[string]interface{}{
		"total_attempts": gorm.Expr("total_attempts + 1"),
	}
	if s.Status == "accepted" {
		updates["accepted_submissions"] = gorm.Expr("accepted_submissions + 1")
	}
	return tx.Model(&Problem{}).Where("id = ?", s.ProblemID).Updates(updates).Error
}

// 题目通过率计算钩子
func (p *Problem) BeforeUpdate(tx *gorm.DB) error {
	if p.TotalAttempts > 0 {
		// p.AcceptanceRate = float64(p.AcceptedSubmissions) / float64(p.TotalAttempts) * 100
	}
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},              // 用户表
		&Article{},           // 文章表
		&Category{},          // 分类表
		&Tag{},               // 标签表
		&Comment{},           // 评论表
		&Like{},              // 点赞表
		&Favorite{},          // 收藏表
	
		// &AIUsageStatistics{}, // AI使用统计表 第一次尝试的，先废弃
		&ArticleTag{},        // 文章标签中间表
		// &Group{},             // 群组表
		// &UserGroup{},         // 用户群组中间表
		&AdminLog{},          // 管理员日志表
		&GroupApply{},        // 圈子申请表

		&Problem{},           // 题目表
		&TestCase{},          // 测试用例表
		&Submission{},        // 提交表
		&ProblemTag{},        // 题目标签表
		&UserProblemStatus{}, // 用户题目状态表

		&Message{},    // 消息表
		&UserFollow{}, // 用户关注表

		&SearchHistory{},  // 搜索历史表
		&ReadingHistory{}, // 阅读历史表

		&AIConversation{}, // Deepseek AI对话表 第一次尝试，废弃了的，目前用来保留流式的

		&ChatSession{}, // Deepseek AI对话会话表第二次尝试 非流式的转发
		&ChatMessage{},
       

		// 贴吧
		&GroupN{},   // 圈子
		&UserGroup{}, // 用户圈子中间表
		&Post{},      // 圈子帖子
		&GroupComment{}, // 圈子评论
		&GroupLike{},    // 圈子点赞表

		// 流量记录中间件的表Traffic
		&Traffic{},

		// ai模型记录表
		&AIModelConfig{},
	)
}

// package model

// import (
// 	"time"
// 	//  . "TraeCNServer/config"
// 	"gorm.io/gorm"
// )

// // 用户表 - 存储用户基本信息
// // 包含用户认证信息、个人资料和使用统计
// type User struct {
// 	gorm.Model
// 	Username             string    `gorm:"uniqueIndex;size:50;not null"`  // 用户名
// 	PasswordHash         string    `gorm:"size:60;not null"`              // 密码
// 	Email                string    `gorm:"uniqueIndex;size:100;not null"` // 邮箱
// 	EmailVerified        bool      `gorm:"default:false"`                 // 邮箱验证状态
// 	VerificationCode     string    `gorm:"size:6"`                        // 验证码
// 	VerificationExpiry   time.Time // 验证码过期时间
// 	AvatarURL            string    `gorm:"size:255"` // 头像URL
// 	RegistrationTime     time.Time // 注册时间
// 	LastLoginTime        time.Time // 最后登录时间
// 	Role                 string    `gorm:"size:10;default:'user'"` // 用户角色
// 	IsActive             bool      `gorm:"default:true"`           // 是否激活
// 	NumberOfFans         uint      `gorm:"default:0"`              // 粉丝数
// 	NumberOfFollow       uint      `gorm:"default:0"`              // 关注数
// 	NumberOfLikes        uint      `gorm:"default:0"`              // 点赞数
// 	NumberOfComments     uint      `gorm:"default:0"`              // 评论数
// 	NumberOfArticles     uint      `gorm:"default:0"`              // 文章数
// 	NumberOfFavorites    uint      `gorm:"default:0"`              // 收藏数
// 	NumberOfViews        uint      `gorm:"default:0"`              // 浏览量
// 	NumberOfShares       uint      `gorm:"default:0"`              // 分享数
// 	NumberOfReports      uint      `gorm:"default:0"`              // 举报数
// 	PersonalSignature    string    `gorm:"size:255"`               // 个性签名
// 	PersonalIntroduction string    `gorm:"size:255"`               // 个人简介
// 	PersonalWebsite      string    `gorm:"size:255"`               // 个人网站
// 	PersonalLocation     string    `gorm:"size:255"`               // 个人位置
// 	QQ                   string    `gorm:"size:255"`               // QQ
// 	WeChat               string    `gorm:"size:255"`               // 微信
// 	Weibo                string    `gorm:"size:255"`               // 微博
// 	Twitter              string    `gorm:"size:255"`               // Twitter

// 	// 关联关系
// 	Articles        []Article         `gorm:"foreignKey:UserID"`
// 	Comments        []Comment         `gorm:"foreignKey:UserID"`
// 	Likes           []Like            `gorm:"foreignKey:UserID"`
// 	Favorites       []Favorite        `gorm:"foreignKey:UserID"`
// 	AIConversations []AIConversation  `gorm:"foreignKey:UserID"`
// 	AIUsage         AIUsageStatistics `gorm:"foreignKey:UserID"`
// }

// // 文章表 - 存储博客文章内容
// // 包含文章元数据、内容和关联关系
// const (
// 	ArticleStatusPublished = "published"
// 	ArticleStatusDraft     = "draft"
// )

// type Article struct {
// 	gorm.Model
// 	UserID       uint      `gorm:"index;not null"`         // 用户ID
// 	CategoryID   uint      `gorm:"index;not null"`         // 分类ID
// 	Title        string    `gorm:"size:200;not null"`      // 文章标题
// 	Description  string    `gorm:"size:255"`               // 文章描述
// 	Cover        string    `gorm:"size:255"`               // 文章封面链接
// 	Content      string    `gorm:"type:longtext;not null"` // 文章内容
// 	PublishTime  time.Time `gorm:"index"`                  // 发布时间
// 	UpdateTime   time.Time // 更新时间
// 	IsDeleted    bool      `gorm:"default:false"`            // 是否删除
// 	Status       string    `gorm:"size:20;default:'draft'"`  // 文章状态
// 	IsTop        bool      `gorm:"default:false"`            // 是否置顶
// 	IsOriginal   bool      `gorm:"default:false"`            // 是否原创
// 	IsPublished  bool      `gorm:"default:false"`            // 是否发布
// 	IsPost       bool      `gorm:"default:false"`            // 是否是帖子
// 	PostType     string    `gorm:"size:20;default:'normal'"` // 帖子类型
// 	ViewCount    int       `gorm:"default:0"`                // 浏览量
// 	LikeCount    int       `gorm:"default:0"`                // 点赞数
// 	CommentCount int       `gorm:"default:0"`                // 评论数
// 	ShareCount   int       `gorm:"default:0"`                // 分享数
// 	IsDraft      bool      `gorm:"-"`                        // 虚拟字段用于查询
// 	GroupID      *uint     `gorm:"index"`                    // 所属圈子ID（指针类型允许nil）
// 	PostVisible  string    `gorm:"size:20"`                  // 可见性：public/group/members
// 	PinPosition  int       // 置顶位置

// 	IsSuperAdmin  bool      `gorm:"default:false"` // 是否超级管理员
// 	AdminRoles    string    `gorm:"size:255"`      // 管理角色（逗号分隔）
// 	LastAdminTime time.Time // 最后管理操作时间
// 	AdminNotes    string    `gorm:"type:text"` // 管理备注
// 	AdminActions  string    `gorm:"type:text"` // 管理操作记录（JSON格式）
// 	AdminIP       string    `gorm:"size:50"`   // 管理操作IP

// 	AuditStatus  string    `gorm:"size:20;default:'pending'"` // 审核状态（pending/approved/rejected）
// 	AuditTime    time.Time // 审核时间
// 	AuditorID    *uint     // 审核人ID
// 	RejectReason string    `gorm:"type:text"`            // 驳回原因
// 	Auditor      User      `gorm:"foreignKey:AuditorID"` // 审核人

// 	// 关联关系

// 	Group     Group      `gorm:"foreignKey:GroupID"` // 所属圈子（关联关系）
// 	User      User       `gorm:"foreignKey:UserID"`
// 	Category  Category   `gorm:"foreignKey:CategoryID"`
// 	Tags      []Tag      `gorm:"many2many:article_tags;"`
// 	Comments  []Comment  `gorm:"foreignKey:ArticleID"`
// 	Likes     []Like     `gorm:"foreignKey:ArticleID"`
// 	Favorites []Favorite `gorm:"foreignKey:ArticleID"`
// }

// // 新增管理员操作日志表
// type AdminLog struct {
// 	gorm.Model
// 	AdminID     uint   `gorm:"index;not null"` // 操作管理员ID
// 	TargetType  string `gorm:"size:50"`        // 操作对象类型（article/group/user）
// 	TargetID    uint   `gorm:"index"`          // 操作对象ID
// 	Action      string `gorm:"size:50"`        // 操作类型（approve/delete/edit）
// 	Description string `gorm:"type:text"`      // 操作描述
// 	IPAddress   string `gorm:"size:45"`        // 操作IP地址

// 	Admin User `gorm:"foreignKey:AdminID"`
// }

// // 文章标签中间表 - 文章和标签的多对多关系映射
// type ArticleTag struct {
// 	ArticleID uint `gorm:"primaryKey"` // 文章ID
// 	TagID     uint `gorm:"primaryKey"` // 标签ID
// }

// // 分类表 - 文章分类管理
// // 用于对文章进行分类组织
// type Category struct {
// 	ID          uint   `gorm:"primaryKey"`                   // 分类ID
// 	Name        string `gorm:"uniqueIndex;size:50;not null"` // 分类名称
// 	Description string `gorm:"size:255"`                     // 分类描述
// }

// // 标签表 - 文章标签管理
// // 用于标记文章主题
// type Tag struct {
// 	ID   uint   `gorm:"primaryKey"`                   // 标签ID
// 	Name string `gorm:"uniqueIndex;size:50;not null"` // 标签名称
// }

// // 评论表 - 存储文章评论
// // 支持多级回复和评论管理
// type Comment struct {
// 	ID              uint      `gorm:"primaryKey"`
// 	ParentID        *uint     `gorm:"index"`
// 	ParentType      string    `gorm:"size:50;index"`
// 	ArticleID       uint      `gorm:"index;not null"`
// 	UserID          uint      `gorm:"index;not null"`
// 	ParentCommentID *uint     `gorm:"index"`
// 	Content         string    `gorm:"type:text;not null"` // 评论内容
// 	CommentTime     time.Time // 评论时间
// 	IsApproved      bool      `gorm:"default:false"` // 是否通过审核
// 	IsDeleted       bool      `gorm:"default:false"` // 是否删除

// 	// 自关联关系
// 	Replies []Comment `gorm:"foreignKey:ParentCommentID"`
// 	User    User      `gorm:"foreignKey:UserID"`
// 	Article Article   `gorm:"foreignKey:ArticleID"`
// }

// // 点赞表 - 记录用户点赞
// // 可对文章或评论点赞
// type Like struct {
// 	ID         uint   `gorm:"primaryKey"`
// 	UserID     uint   `gorm:"index;not null"`
// 	ParentID   uint   `gorm:"index"`
// 	ParentType string `gorm:"index;size:50"`
// 	LikeTime   time.Time

// 	User User `gorm:"foreignKey:UserID"`
// }

// // 收藏表 - 记录用户收藏的文章
// // 用于个人收藏管理
// type Favorite struct {
// 	ID           uint      `gorm:"primaryKey"`     // 收藏ID
// 	UserID       uint      `gorm:"index;not null"` // 用户ID
// 	ArticleID    uint      `gorm:"index;not null"` // 文章ID
// 	FavoriteTime time.Time // 收藏时间

// 	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
// 	Article Article `gorm:"foreignKey:ArticleID"` // 关联文章
// }

// // AI问答会话表 - 存储用户与AI的对话记录
// // 包含问题、回答和token使用统计
// type AIConversation struct {
// 	ID           uint      `gorm:"primaryKey"`             // 会话ID
// 	UserID       uint      `gorm:"index;not null"`         // 用户ID
// 	SessionID    string    `gorm:"size:64;index"`          // 会话ID
// 	Question     string    `gorm:"type:longtext;not null"` // 问题
// 	Answer       string    `gorm:"type:longtext;not null"` // 回答
// 	Model        string    `gorm:"size:50;not null"`       // 使用的模型
// 	InputTokens  int       `gorm:"not null"`               // 输入的token数量
// 	OutputTokens int       `gorm:"not null"`               // 输出的token数量
// 	TotalTokens  int       `gorm:"index;not null"`         // 总token数量
// 	CreatedTime  time.Time // 创建时间

// 	User User `gorm:"foreignKey:UserID"` // 关联用户
// }

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

// // 群组表 - 存储用户创建的群组
// type Group struct {
// 	gorm.Model
// 	Name          string `gorm:"size:100;not null;uniqueIndex"` // 圈子名称（唯一）
// 	Description   string `gorm:"type:text;not null"`            // 圈子描述
// 	CreatorID     uint   `gorm:"index;not null"`                // 创建者ID
// 	AvatarURL     string `gorm:"size:255"`                      // 圈子头像
// 	MemberCount   uint   `gorm:"default:0"`                     // 成员数量
// 	PostCount     uint   `gorm:"default:0"`                     // 帖子数量
// 	IsPublic      bool   `gorm:"default:true"`                  // 是否公开
// 	AuditRequired bool   `gorm:"default:false"`                 // 发帖是否需要审核

// 	// 关联关系
// 	Creator User      `gorm:"foreignKey:CreatorID"`
// 	Members []User    `gorm:"many2many:user_groups;"`
// 	Posts   []Article `gorm:"foreignKey:GroupID"`
// }

// // 用户群组中间表 - 用户和群组的多对多关系映射
// type UserGroup struct {
// 	UserID   uint      `gorm:"primaryKey;autoIncrement:false"`
// 	GroupID  uint      `gorm:"primaryKey;autoIncrement:false"`
// 	JoinedAt time.Time // 加入时间
// 	Role     string    `gorm:"size:20;default:'member'"` // 角色：member/admin

// 	CanPost   bool `gorm:"default:true"`  // 是否允许发帖
// 	CanManage bool `gorm:"default:false"` // 是否允许管理

// 	User  User  `gorm:"foreignKey:UserID"`
// 	Group Group `gorm:"foreignKey:GroupID"`
// }

// // 添加圈子审核表
// type GroupApply struct {
// 	gorm.Model
// 	GroupID uint   `gorm:"index"`
// 	UserID  uint   `gorm:"index"`
// 	Message string `gorm:"type:text"`
// 	Status  string `gorm:"size:20"` // pending/approved/rejected
// }

// /*
//    用户刷题模块
// */

// // // 题目表 - 存储编程题目信息
// // type Problem struct {
// // 	gorm.Model
// // 	Title          string       `gorm:"size:200;not null"`       // 题目标题
// // 	Description    string       `gorm:"type:longtext;not null"`  // 题目描述
// // 	Difficulty     string       `gorm:"size:20;not null;index"`  // 难度等级（easy/medium/hard）
// // 	SampleInput    string       `gorm:"type:text"`               // 示例输入
// // 	SampleOutput   string       `gorm:"type:text"`               // 示例输出
// // 	AuthorID       uint         `gorm:"index;not null"`          // 题目创建者ID
// // 	TimeLimit      int          `gorm:"default:1000"`            // 时间限制（毫秒）
// // 	MemoryLimit    int          `gorm:"default:128"`             // 内存限制（MB）
// // 	Tags           []ProblemTag `gorm:"many2many:problem_tags;"` // 题目标签
// // 	Hint           string       `gorm:"type:text"`               // 解题提示
// // 	TotalAttempts  uint         `gorm:"default:0"`               // 总尝试次数
// // 	AcceptanceRate float64      `gorm:"default:0"`               // 通过率
// // 	IsPublished    bool         `gorm:"default:false"`           // 是否发布
// // 	TestCases      []TestCase   `gorm:"foreignKey:ProblemID"`    // 测试用例
// // 	Solutions      []Article    `gorm:"foreignKey:ProblemID"`    // 关联题解
// // 	Submissions    []Submission `gorm:"foreignKey:ProblemID"`    // 提交记录

// // 	Author User `gorm:"foreignKey:AuthorID"` // 题目作者
// // }

// // // 测试用例表 - 存储题目的测试用例
// // type TestCase struct {
// // 	gorm.Model
// // 	ProblemID      uint   `gorm:"index;not null"`     // 关联题目ID
// // 	Input          string `gorm:"type:text;not null"` // 输入数据
// // 	ExpectedOutput string `gorm:"type:text;not null"` // 期望输出
// // 	IsHidden       bool   `gorm:"default:false"`      // 是否隐藏用例
// // 	Score          int    `gorm:"default:0"`          // 测试用例分数（预留字段）

// // 	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
// // }

// // // 提交记录表 - 存储用户代码提交记录
// // type Submission struct {
// // 	gorm.Model
// // 	UserID        uint   `gorm:"index;not null"`                  // 用户ID
// // 	ProblemID     uint   `gorm:"index;not null"`                  // 题目ID
// // 	Code          string `gorm:"type:longtext;not null"`          // 提交代码
// // 	Language      string `gorm:"size:20;not null"`                // 编程语言
// // 	Status        string `gorm:"size:50;index;default:'pending'"` // 判题状态
// // 	ExecutionTime int    `gorm:"default:0"`                       // 执行时间（ms）
// // 	MemoryUsed    int    `gorm:"default:0"`                       // 内存使用（KB）
// // 	PassedCases   int    `gorm:"default:0"`                       // 通过用例数
// // 	TotalCases    int    `gorm:"default:0"`                       // 总用例数
// // 	ErrorMessage  string `gorm:"type:text"`                       // 错误信息
// // 	IsContest     bool   `gorm:"default:false"`                   // 是否比赛提交

// // 	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
// // 	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
// // }

// // // 题目标签表
// // type ProblemTag struct {
// // 	ID    uint   `gorm:"primaryKey"`
// // 	Name  string `gorm:"uniqueIndex;size:50;not null"` // 标签名称
// // 	Color string `gorm:"size:20"`                      // 标签颜色（前端展示）
// // }

// // // 用户题目状态表 - 记录用户解题进度
// // type UserProblemStatus struct {
// // 	UserID       uint      `gorm:"primaryKey;autoIncrement:false"`
// // 	ProblemID    uint      `gorm:"primaryKey;autoIncrement:false"`
// // 	IsSolved     bool      `gorm:"default:false"` // 是否已解决
// // 	Attempts     uint      `gorm:"default:0"`     // 尝试次数
// // 	LastSubmitAt time.Time // 最后提交时间
// // 	BestTime     int       `gorm:"default:0"` // 最佳用时（ms）
// // 	BestMemory   int       `gorm:"default:0"` // 最佳内存（KB）

// // 	User    User    `gorm:"foreignKey:UserID"`    // 关联用户
// // 	Problem Problem `gorm:"foreignKey:ProblemID"` // 关联题目
// // }

// func (a *Article) BeforeCreate(tx *gorm.DB) error {
// 	a.PublishTime = time.Now()
// 	return nil
// }

// func (c *Comment) BeforeCreate(tx *gorm.DB) error {
// 	c.CommentTime = time.Now()
// 	return nil
// }

// func (l *Like) BeforeCreate(tx *gorm.DB) error {
// 	l.LikeTime = time.Now()
// 	return nil
// }

// // UserGroup的创建/删除钩子
// func (ug *UserGroup) AfterCreate(tx *gorm.DB) error {
// 	return tx.Model(&Group{}).Where("id = ?", ug.GroupID).
// 		Update("member_count", gorm.Expr("member_count + 1")).Error
// }

// func (ug *UserGroup) AfterDelete(tx *gorm.DB) error {
// 	return tx.Model(&Group{}).Where("id = ?", ug.GroupID).
// 		Update("member_count", gorm.Expr("GREATEST(member_count - 1, 0)")).Error
// }

// // Article的创建/删除钩子
// func (a *Article) AfterCreate(tx *gorm.DB) error {
// 	if a.GroupID != nil {
// 		return tx.Model(&Group{}).Where("id = ?", *a.GroupID).
// 			Update("post_count", gorm.Expr("post_count + 1")).Error
// 	}
// 	return nil
// }

// func (a *Article) AfterDelete(tx *gorm.DB) error {
// 	if a.GroupID != nil {
// 		return tx.Model(&Group{}).Where("id = ?", *a.GroupID).
// 			Update("post_count", gorm.Expr("GREATEST(post_count - 1, 0)")).Error
// 	}
// 	return nil
// }

// // 文章审核钩子函数
// func (a *Article) BeforeUpdate(tx *gorm.DB) error {
// 	// 当审核状态变化时自动记录时间
// 	if tx.Statement.Changed("AuditStatus") {
// 		a.AuditTime = time.Now()
// 	}
// 	return nil
// }

// // // 管理员权限校验方法
// // func (u *User) HasAdminPermission(requiredRole string) bool {
// //     if u.IsSuperAdmin {
// //         return true
// //     }

// //     roles := strings.Split(u.AdminRoles, ",")
// //     for _, role := range roles {
// //         if role == requiredRole {
// //             return true
// //         }
// //     }
// //     return false
// // }

// // // 提交记录创建后更新统计信息
// // func (s *Submission) AfterCreate(tx *gorm.DB) error {
// // 	// 更新用户提交统计
// // 	if err := tx.Model(&User{}).Where("id = ?", s.UserID).
// // 		Updates(map[string]interface{}{
// // 			"submit_count": gorm.Expr("submit_count + 1"),
// // 		}).Error; err != nil {
// // 		return err
// // 	}

// // 	// 更新题目统计
// // 	updates := map[string]interface{}{
// // 		"total_attempts": gorm.Expr("total_attempts + 1"),
// // 	}
// // 	if s.Status == "accepted" {
// // 		updates["accepted_submissions"] = gorm.Expr("accepted_submissions + 1")
// // 	}
// // 	return tx.Model(&Problem{}).Where("id = ?", s.ProblemID).Updates(updates).Error
// // }

// // // 题目通过率计算钩子
// // func (p *Problem) BeforeUpdate(tx *gorm.DB) error {
// // 	// 计算通过率
// // 	if p.TotalAttempts > 0 {
// // 		// 通过率计算
// // 		// p.AcceptanceRate = float64(p.AcceptedSubmissions) / float64(p.TotalAttempts) * 100
// // 	}
// // 	return nil
// // }

// // // 创建课程钩子
// // func (c *Course) AfterCreate(tx *gorm.DB) error {
// // 	// 自动创建课程目录
// // 	return tx.Model(c).Association("Chapters").Append(&Chapter{
// // 		Title: "默认章节",
// // 	})
// // }

// // 更新学习人数统计
// // func UpdateCourseStudents(courseID uint) error {
// // 	var count int64
// // 	if err := DB.Model(&UserCourseProgress{}).
// // 		Where("course_id = ?", courseID).
// // 		Count(&count).Error; err != nil {
// // 		return err
// // 	}
// // 	return DB.Model(&Course{}).Where("id = ?", courseID).
// // 		Update("student_count", count).Error
// // }

// func AutoMigrate(db *gorm.DB) error {
// 	return db.AutoMigrate(

// 		&User{},              // 用户表
// 		&Article{},           // 文章表
// 		&Category{},          // 分类表
// 		&Tag{},               // 标签表
// 		&Comment{},           // 评论表
// 		&Like{},              // 点赞表
// 		&Favorite{},          // 收藏表
// 		&AIConversation{},    // AI对话表
// 		&AIUsageStatistics{}, // AI使用统计表
// 		&ArticleTag{},        // 文章标签中间表
// 		&Group{},             // 群组表
// 		&UserGroup{},         // 用户群组中间表
// 		&AdminLog{},          // 管理员日志表
// 		&GroupApply{},        // 圈子申请表

// 	)
// }

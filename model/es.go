package model

// ElasticSearch 用户索引结构
// 可根据实际需求调整字段映射和类型

type ESUser struct {
	ID                uint   `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	AvatarURL         string `json:"avatar_url"`
	PersonalSignature string `json:"personal_signature"`
	PersonalIntro     string `json:"personal_intro"`
}

// ElasticSearch 文章索引结构
// 可根据实际需求调整字段映射和类型

type ESArticle struct {
	ID          uint     `json:"id"`
	UserID      uint     `json:"user_id"`
	UserName    string   `json:"user_name"`
	CategoryID  uint     `json:"category_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	PublishTime string   `json:"publish_time"`
	Tags        []string `json:"tags"`
}

// ElasticSearch 标签索引结构
// 可根据实际需求调整字段映射和类型

type ESTag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

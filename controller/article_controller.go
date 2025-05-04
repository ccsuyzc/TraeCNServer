package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ArticleController 文章控制器
type ArticleController struct{}

// Paginate 分页作用域函数
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// CreateArticle 创建文章
func (ac *ArticleController) CreateArticle(c *gin.Context) {
	// 定义请求结构体
	type ArticleRequest struct {
		Title       string `json:"title" binding:"required"`       // 文章标题
		Content     string `json:"content" binding:"required"`     // 文章内容
		Description string `json:"description"`                    // 文章描述
		CategoryID  uint   `json:"category_id" binding:"required"` // 文章分类ID
		TagIDs      []uint `json:"tag_ids"`                        // 文章标签ID列表
		UserID      uint   `json:"user_id" binding:"required"`     // 用户ID
		UserName    string `json:"user_name" binding:"required"`   // 用户名
		Status      string `json:"status" binding:"required"`      // 文章状态
		AvatarLink  string `json:"avatar_link"`                    // 文章封面·链接
		// Token string `json:"token"`
	}

	// 绑定并验证请求数据
	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "msg": "数据解绑失败"})
		return
	}
	log.Print("req:", req)
	// // 从请求头中获取用户ID
	// userID := c.GetHeader("X-User-ID")
	// log.Print("userID:",userID)
	// if userID == "" {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }
	// userIDInt, _ := strconv.Atoi(userID)  // 将字符串转换为整数

	// 从认证信息获取用户ID, 确保用户已认证,c.Get("userID") 从上下文中获取用户ID
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }

	// 检查分类是否存在
	var category model.Category
	if err := DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分类不存在"})
		return
	}

	var validTag model.Tag
	// 检查标签是否存在
	if len(req.TagIDs) > 0 {
		if err := DB.First(&validTag, req.TagIDs[0]).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "第一个标签无效"})
			return
		}
	}

	// 使用事务处理
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 创建文章
		article := model.Article{
			Title:      req.Title,
			Content:    req.Content,
			CategoryID: req.CategoryID,
			// UserID:     uint(userIDInt),
			Description: req.Description,
			UserID:      req.UserID,
			UserName:    req.UserName,
		}
		if err := tx.Create(&article).Error; err != nil {
			return err
		}

		// 关联标签
		if len(req.TagIDs) > 0 {
			if err := tx.Model(&article).Association("Tags").Append(&validTag); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文章失败", "msg": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "文章创建成功"})
}

// GetArticle 获取单篇文章
func (ac *ArticleController) GetArticle(c *gin.Context) {
	ctx := context.Background()
	// 获取文章ID
	id := c.Param("id")

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("article:%s", id)
	cachedData, err := RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var article model.Article
		if json.Unmarshal([]byte(cachedData), &article) == nil {
			c.JSON(http.StatusOK, gin.H{"data": article, "source": "cache"})
			return
		}
	}

	// 缓存未命中，查询数据库
	var article model.Article
	if err := DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 更新阅读量
	DB.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	// 更新缓存数据
	updatedArticle := article
	if jsonData, err := json.Marshal(updatedArticle); err == nil {
		RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{"data": article, "source": "database"})
}

// GetAllArticles 获取所有文章
func (ac *ArticleController) GetAllArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	query := DB.Model(&model.Article{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var articles []model.Article
	query.Scopes(Paginate(page, pageSize)).Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"data":  articles,
		"page":  page,
		"total": total,
	})
}

// GetPublishedArticles 获取指定用户的已发布文章
func (ac *ArticleController) GetPublishedArticles(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("userid"))

	var articles []model.Article
	DB.Where("user_id = ? AND status = ?", userID, "published").Joins("User").Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
	})
}

// UpdateArticle 更新文章
func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	type UpdateArticleRequest struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		CategoryID uint   `json:"category_id" binding:"required"`
		TagIDs     []uint `json:"tag_ids"`
		Status     string `json:"status" binding:"required"`
		AvatarLink string `json:"avatar_link"`
		// Token string `json:"token"`
	}

	// 获取文章ID
	id := c.Param("id")
	var article model.Article
	if err := DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 绑定请求数据
	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从认证信息获取用户ID
	userID, exists := c.Get("userID")
	if !exists || userID.(uint) != article.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无操作权限"})
		return
	}

	// 验证分类是否存在
	var category model.Category
	if err := DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分类不存在"})
		return
	}

	// 验证标签是否存在
	var validTag model.Tag
	if len(req.TagIDs) > 0 {
		if err := DB.First(&validTag, req.TagIDs[0]).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "第一个标签无效"})
			return
		}
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// 更新文章字段
		updateData := map[string]interface{}{
			"Title":       req.Title,
			"Content":     req.Content,
			"CategoryID":  req.CategoryID,
			"Status":      req.Status,
			"AvatarLink":  req.AvatarLink,
			"UpdatedTime": time.Now(),
		}

		if err := tx.Model(&article).Updates(updateData).Error; err != nil {
			return err
		}

		// 更新标签关联
		if len(req.TagIDs) > 0 {
			if err := tx.Model(&article).Association("Tags").Append(&validTag); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章更新成功", "data": article})
}

// DeleteArticle 删除文章
func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	var article model.Article
	id := c.Param("id")
	if err := DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}
	if err := DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

// SearchArticles 搜索文章
func (ac *ArticleController) SearchArticles(c *gin.Context) {
	keyword := c.Query("keyword")
	var articles []model.Article
	if err := DB.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// GetRandomArticles 随机获取文章
func (ac *ArticleController) GetRandomArticles(c *gin.Context) {
	// 从查询参数中获取数量
	quantityStr := c.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	// limitStr := c.DefaultQuery("limit", "5")  // 默认获取5篇文章
	// limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	var articles []model.Article
	if err := DB.Order("RAND()").Limit(quantity).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// 按分类和条数随机获取文章
func (ac *ArticleController) GetRandomArticlesByCategory(c *gin.Context) {
	categoryID := c.Param("id")
	quantityStr := c.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	// limitStr := c.DefaultQuery("limit", "5")
	// limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	var articles []model.Article
	if err := DB.Where("category_id =?", categoryID).Order("RAND()").Limit(quantity).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// GetPopularArticles 获取热门文章
func (ac *ArticleController) GetPopularArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	var articles []model.Article
	if err := DB.Order("view_count desc").Limit(limit).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// 新增分类查询方法
func (ctrl *ArticleController) GetArticlesByCategory(c *gin.Context) {
	categoryID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var articles []model.Article
	err := DB.Preload("Category").Preload("User").
		Where("category_id = ?", categoryID).
		Order("publish_time desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&articles).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(200, gin.H{"data": articles})
}

// GetArticlesByCategoryAndLimit 按分类和条数获取文章
func (ac *ArticleController) GetArticlesByCategoryAndLimit(c *gin.Context) {
	categoryID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "5")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	var articles []model.Article
	err = DB.Preload("Category").
		Where("category_id = ?", categoryID).
		Order("publish_time desc").
		Limit(limit).
		Find(&articles).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// 获取用户全部文章
func (ac *ArticleController) GetUserArticles(c *gin.Context) {
	// 验证用户ID参数
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var articles []model.Article
	result := DB.Preload("Category").Preload("Tags").Where("user_id = ?", userID).Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"pagination": gin.H{
			"currentPage": page,
			"pageSize":    pageSize,
			"total":       result.RowsAffected,
		},
	})
}

// 获取指定用户的全部草稿
func (ac *ArticleController) GetDrafts(c *gin.Context) {
	// // 从认证信息获取用户ID
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }
	userID := c.Param("userid")
	// 转换为 uint
	userIDUint, _ := strconv.ParseUint(userID, 10, 64)
	// 处理分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var drafts []model.Article
	err := DB.Where("user_id = ? AND status = ?", userIDUint, "draft").Preload("Category").Preload("Tags").Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&drafts).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": drafts,
		"pagination": gin.H{
			"currentPage": page,
			"pageSize":    pageSize,
			"total":       len(drafts),
		},
	})
}

// 获取指定用户的指定草稿
func (ac *ArticleController) GetDraft(c *gin.Context) {
	// 参数非空检查
	if c.Param("userid") == "" || c.Param("draftid") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不能为空"})
		return
	}

	// 用户ID转换与校验
	userID, err := strconv.ParseUint(c.Param("userid"), 10, 32)
	if err != nil || userID == 0 {
		log.Printf("非法用户ID参数 | 输入:%s | 类型:%T | 错误:%s", c.Param("userid"), c.Param("userid"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID必须为正整数"})
		return
	}

	// 草稿ID转换与校验
	draftID, err := strconv.ParseUint(c.Param("draftid"), 10, 32)
	if err != nil || draftID == 0 {
		log.Printf("非法草稿ID参数 | 输入:%s | 类型:%T | 错误:%s", c.Param("draftid"), c.Param("draftid"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "草稿ID必须为正整数"})
		return
	}

	var article model.Article
	// 事务处理保证数据一致性
	err = DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND user_id = ? AND LOWER(status) = ?",
			uint(draftID), uint(userID), strings.ToLower("draft")).Preload("Category").Preload("Tags").First(&article)

		if result.Error != nil {
			log.Printf("数据库查询失败 | 用户ID:%d 草稿ID:%d | 错误:%s", userID, draftID, result.Error)
			return result.Error
		}
		return nil
	})

	// if result.Error != nil {
	// 	log.Printf("GetDraft查询失败 | 用户ID:%d(%T) 草稿ID:%d(%T) 状态:%s | 条件: id=%d AND user_id=%d AND status='draft' | 错误:%s",
	// 		userID, c.Param("userid"), draftID, c.Param("draftid"), "draft", draftID, userID, result.Error)
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "草稿不存在或无权访问", "user_id": userID, "draft_id": draftID})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"data":    article,
		"success": true,
	})
}

// UpdateDraft 更新草稿
func (ac *ArticleController) UpdateDraft(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var draft model.Article
	if err := DB.Where("id = ? AND status = 'draft' AND user_id = ?", id, userID).First(&draft).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "草稿不存在或无权修改"})
		return
	}

	if err := c.ShouldBindJSON(&draft); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Save(&draft).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "草稿更新成功", "data": draft})
}

// DeleteDraft 删除草稿
func (ac *ArticleController) DeleteDraft(c *gin.Context) {
	id := c.Param("draftid")
	userID := c.MustGet("userid").(uint)

	if err := DB.Where("id = ? AND status = 'draft' AND user_id = ?", id, userID).Delete(&model.Article{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "草稿删除成功"})
}

// PublishDraft 发布草稿
func (ac *ArticleController) PublishDraft(c *gin.Context) {
	id := c.Param("id") // 获取文章ID
	var article model.Article

	// 检查文章是否存在且是草稿
	if err := DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}
	// 检查文章是否是草稿
	if article.Status != model.ArticleStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft articles can be published"})
		return
	}
	// 更新文章状态为已发布
	tx := DB.Begin()
	article.Status = model.ArticleStatusPublished
	article.PublishTime = time.Now()

	if err := tx.Save(&article).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新用户草稿计数
	if err := tx.Model(&model.User{}).Where("id = ?", article.UserID).
		Update("number_of_drafts", gorm.Expr("number_of_drafts - 1")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "data": article})
}

// SaveDraft 保存为草稿
func (ac *ArticleController) SaveDraft(c *gin.Context) {
	// 定义请求结构体
	type ArticleRequest struct {
		Title       string `json:"title" binding:"required"`       // 文章标题
		Content     string `json:"content" binding:"required"`     // 文章内容
		CategoryID  uint   `json:"category_id" binding:"required"` // 文章分类ID
		TagIDs      []uint `json:"tag_ids"`                        // 文章标签ID列表
		UserID      uint   `json:"user_id" binding:"required"`     // 用户ID
		Status      string `json:"status" binding:"required"`      // 文章状态
		AvatarLink  string `json:"avatar_link"`                    // 文章封面·链接
		Description string `json:"description"`                    // 文章描述
		UserName    string `json:"user_name"`                      // 用户名
		// Token string `json:"token"`
	}

	// 绑定并验证请求数据
	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "msg": "数据解绑失败"})
		return
	}
	log.Print("req:", req)
	// // 从请求头中获取用户ID
	// userID := c.GetHeader("X-User-ID")
	// log.Print("userID:",userID)
	// if userID == "" {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }
	// userIDInt, _ := strconv.Atoi(userID)  // 将字符串转换为整数

	// 从认证信息获取用户ID, 确保用户已认证,c.Get("userID") 从上下文中获取用户ID
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }

	// 检查分类是否存在
	var category model.Category
	if err := DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分类不存在"})
		return
	}

	var validTag model.Tag
	// 检查标签是否存在
	if len(req.TagIDs) > 0 {
		if err := DB.First(&validTag, req.TagIDs[0]).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "第一个标签无效"})
			return
		}
	}

	// 使用事务处理
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 创建文章
		article := model.Article{
			Title:      req.Title,
			Content:    req.Content,
			CategoryID: req.CategoryID,
			// UserID:     uint(userIDInt),
			UserID:      req.UserID,
			UserName:    req.UserName,
			Description: req.Description,
			Status:      "draft", // 设置文章状态为草稿
		}
		if err := tx.Create(&article).Error; err != nil {
			return err
		}

		// 关联标签
		if len(req.TagIDs) > 0 {
			if err := tx.Model(&article).Association("Tags").Append(&validTag); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建草稿失败", "msg": err.Error()})
		return
	}
	// 查询该草稿文章
	var article model.Article
	if err := DB.Where("title =? AND content =? AND category_id =? AND user_id =? AND status =?", req.Title, req.Content, req.CategoryID, req.UserID, "draft").First(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查草稿失败", "msg": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "文章创建成功", "data": article})

}

// // GetPublishedArticles 获取用户已发布的所有文章
// func (ac *ArticleController) GetPublishedArticles(c *gin.Context) {
// }

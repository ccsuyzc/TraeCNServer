package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"TraeCNServer/service"
	"TraeCNServer/service/redis_service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ArticleController 文章控制器
type ArticleController struct{}

// ShelvingArticle 上架文章
func (ac *ArticleController) ShelvingArticle(c *gin.Context) {
	 articleID := c.Param("id")
	// 检查文章是否存在
	var article model.Article
	if err := DB.Preload("Category").Preload("Tags").First(&article, articleID).Error; err!= nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在", "code": 404})
		return
	}

	// 检查文章是否已经上架
	if article.Status == "published" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章已上架", "code": 400 })
		return
	}
	// 更新文章状态为已上架
	if err := DB.Model(&article).Update("status", "published").Error; err!= nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章状态失败", "code": 500 })
		return
	}
}
// DelistingArticle 下架文章
func (ac *ArticleController) DelistingArticle(c *gin.Context) {
	articleID := c.Param("id")
	// 检查文章是否存在
	var article model.Article
	if err := DB.Preload("Category").Preload("Tags").First(&article, articleID).Error; err!= nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在", "code": 404 })
		return
	}
	// 检查文章是否已经下架
	if article.Status == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章已下架", "code": 400 })
		return
	}

	// 更新文章状态为已下架
	if err := DB.Model(&article).Update("status", "rejected").Error; err!= nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章状态失败", "code": 500 })
		return
	}
}

// 获取指定用户的所有收藏文章
func (ac *ArticleController) GetArticleCollection(c *gin.Context) {
	ID := c.Param("userid")
	IDInt, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	userID := uint(IDInt)

	// 查询该用户收藏的所有文章ID
	var favoriteArticles []model.Favorite
	if err := DB.Where("user_id = ?", userID).Find(&favoriteArticles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询收藏失败"})
		return
	}

	// 提取所有被收藏的ArticleID
	articleIDs := make([]uint, 0, len(favoriteArticles))
	for _, fav := range favoriteArticles {
		articleIDs = append(articleIDs, fav.ArticleID)
	}

	if len(articleIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []model.Article{}, "message": "暂无收藏文章", "success": true})
		return
	}

	// 查询这些文章的详细信息，预加载User和Category
	var articles []model.Article
	if err := DB.Preload("User").Preload("Category").Where("id IN ?", articleIDs).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章详情失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles, "message": "获取收藏文章成功", "success": true})
}

// 获取推荐文章
func (ac *ArticleController) RecommendedArticle(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	service := service.RecommendService{DB: DB}  // 使用你的数据库连接实例
	recommendations := service.GenerateRecommendations(c.Request.Context(), uint(userID), 5)

	if len(recommendations) == 0 {
		// 没有推荐时返回随机文章
		var fallback []model.Article
		DB.Order("RAND()").Limit(20).Preload("Category").Preload("Tags").Find(&fallback)
		c.JSON(http.StatusOK, gin.H{
			"data":    fallback,
			"success": true,
			"message": "推荐失败，返回随机文章",
		})
		return
	}

	// 如果长度不够5，就填充随机文章
	if len(recommendations) < 5 {
		var fallback []model.Article
		DB.Order("RAND()").Limit(5 - len(recommendations)).Preload("Category").Preload("Tags").Find(&fallback)
		recommendations = append(recommendations, fallback...)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    recommendations,
		"success": true,
		"message": "推荐成功",
	})
}

// 获取指定用户的全部已发布或审核中或已拒绝文章
func (ac *ArticleController) GetUserAllArticle(c *gin.Context) {
	// 验证用户ID参数
	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 定义响应结构体
	type categorizedArticles struct {
		Published     []model.Article `json:"published"`
		PendingReview []model.Article `json:"pending_review"`
		Rejected      []model.Article `json:"rejected"`
		Pagination    gin.H           `json:"pagination"`
	}

	var result categorizedArticles
	var total int64

	// 并行查询各状态文章
	var wg sync.WaitGroup
	errChan := make(chan error, 3)
	_, cancel := context.WithCancel(c)
	defer cancel()

	// 查询已发布文章
	wg.Add(1)
	go func() {
		defer wg.Done()
		db := DB.Where("user_id = ? AND status = ?", userID, "published")
		if err := db.Preload("Category").Preload("Tags").
			Order("created_at desc").
			Limit(pageSize).Offset((page - 1) * pageSize).
			Find(&result.Published).Error; err != nil {
			errChan <- fmt.Errorf("已发布文章查询失败: %v", err)
		}
	}()

	// 查询审核中文章
	wg.Add(1)
	go func() {
		defer wg.Done()
		db := DB.Where("user_id = ? AND status = ?", userID, "pending_review")
		if err := db.Preload("Category").Preload("Tags").
			Order("created_at desc").
			Limit(pageSize).Offset((page - 1) * pageSize).
			Find(&result.PendingReview).Error; err != nil {
			errChan <- fmt.Errorf("审核中文章查询失败: %v", err)
		}
	}()

	// 查询已拒绝文章
	wg.Add(1)
	go func() {
		defer wg.Done()
		db := DB.Where("user_id = ? AND status = ?", userID, "rejected")
		if err := db.Preload("Category").Preload("Tags").
			Order("created_at desc").
			Limit(pageSize).Offset((page - 1) * pageSize).
			Find(&result.Rejected).Error; err != nil {
			errChan <- fmt.Errorf("已拒绝文章查询失败: %v", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// 处理错误
	if len(errChan) > 0 {
		var errors []string
		for err := range errChan {
			errors = append(errors, err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": strings.Join(errors, "; ")})
		return
	}

	// 获取总条数
	DB.Model(&model.Article{}).Where("user_id = ?", userID).Count(&total)

	// 设置分页信息
	result.Pagination = gin.H{
		"currentPage": page,
		"pageSize":    pageSize,
		"total":       total,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"success": true,
		"message": "查询成功",
	})
}

// LikeArticle 点赞文章
func (ac *ArticleController) LikeArticle(c *gin.Context) {
	var req struct {
		ArticleID uint `json:"article_id" binding:"required"`
		UserID    uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := redis_service.LikeArticle(c.Request.Context(), req.ArticleID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

// UnlikeArticle 取消点赞
func (ac *ArticleController) UnlikeArticle(c *gin.Context) {
	var req struct {
		ArticleID uint `json:"article_id" binding:"required"`
		UserID    uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := redis_service.UnlikeArticle(c.Request.Context(), req.ArticleID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已取消点赞"})
}

// GetArticleLikes 获取文章点赞数
func (ac *ArticleController) GetArticleLikes(c *gin.Context) {
	articleID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	count, err := redis_service.GetArticleLikes(c.Request.Context(), uint(articleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// FavoriteArticle 收藏文章
func (ac *ArticleController) FavoriteArticle(c *gin.Context) {
	var param struct {
		UserID    uint   `json:"user_id" binding:"required"`
		ArticleID string `json:"article_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{"error": err.Error(), "code": 400})
		return
	}
	intArticleID, _ := strconv.Atoi(param.ArticleID)
	// int 转化为 uint
	unintArticleID := uint(intArticleID)
	err := DB.Transaction(func(tx *gorm.DB) error { // 使用事务
		if err := tx.Create(&model.Favorite{UserID: param.UserID, ArticleID: unintArticleID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Article{}).Where("id = ?", unintArticleID).Update("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "收藏失败", "code": 500})
		return
	}

	c.JSON(200, gin.H{"message": "收藏成功", "code": 200})
}

// UnfavoriteArticle 取消收藏
func (ac *ArticleController) UnfavoriteArticle(c *gin.Context) {
	var param struct {
		UserID    uint   `json:"user_id" binding:"required"`
		ArticleID string `json:"article_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	intArticleID, _ := strconv.Atoi(param.ArticleID)
	// int 转化为 uint
	unintArticleID := uint(intArticleID)

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND article_id = ?", param.UserID, unintArticleID).Delete(&model.Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Article{}).Where("id = ?", unintArticleID).Update("favorite_count", gorm.Expr("favorite_count - 1")).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "取消收藏失败"})
		return
	}

	c.JSON(200, gin.H{"message": "取消收藏成功"})
}

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
			Description: req.Description, // 添加描述字段
			UserID:      req.UserID,
			UserName:    req.UserName,
			Status:      "pending_review", // 显式设置初始状态为审核中
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

	c.JSON(http.StatusCreated, gin.H{"message": "文章创建成功", "code": 200, "data": req})
}

// GetArticle 获取单篇文章用来展示
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
	if err := DB.Preload("User").Preload("Category").First(&article, id).Error; err != nil {
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

// GetArticleModify 获取单篇文章用来修改
func (ac *ArticleController) GetArticleModify(c *gin.Context) {
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
	if err := DB.Preload("User").Preload("Category").First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
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

// GetAllPublishedArticles 获取所有已发布文章
func (ac *ArticleController) GetAllPublishedArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var articles []model.Article
	DB.Where("status = ?", "published").
		Joins("User").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": articles,
		"msg":  "获取全部已发布文章成功",
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

	// // 从认证信息获取用户ID
	// userID, exists := c.Get("userID")
	// if !exists || userID.(uint) != article.UserID {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "无操作权限"})
	// 	return
	// }

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
			"Title":      req.Title,
			"Content":    req.Content,
			"CategoryID": req.CategoryID,
			"Status":     req.Status,
			"Cover":      req.AvatarLink,
			// "UpdatedTime": time.Now(),
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": 400})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully", "code": 200})
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
	c.JSON(200, gin.H{"data": articles, "code": 200, "msg": "查询成功"})
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
	// if article.Status != model.ArticleStatusDraft {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "只能发布草稿状态的文章"})
	// 	return
	// }
	// // 更新文章为待审核
	tx := DB.Begin()
	article.Status = model.ArticleStatusPendingReview // 改为待审核状态
	article.SubmitTime = time.Now()                   // 记录提交审核时间

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

// 发布文章
func (ac *ArticleController) Publish(c *gin.Context) {
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
			Cover:       req.AvatarLink,
			Description: req.Description,
			Status:      "pending_review", // 设置文章状态为草稿
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交审核失败", "msg": err.Error()})
		return
	}
	// 查询该文章
	var article model.Article
	if err := DB.Where("title =? AND content =? AND category_id =? AND user_id =? AND status =?", req.Title, req.Content, req.CategoryID, req.UserID, "draft").First(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查是否已经提交审核失败", "msg": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "文章提交成功", "data": article})
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
			Cover:       req.AvatarLink,
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

// UpdatePublishStatus 更新文章发布状态
func (ac *ArticleController) UpdatePublishStatus(c *gin.Context) {
	id := c.Param("id")
	// 转化为uint类型
	articleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID", "code": 400})
		return
	}
	// var req struct {
	// 	StatusCode string `json:"status_code" binding:"required,oneof=published draft deleted pending_review rejected"`
	// }

	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "无效状态码"})
	// 	return
	// }

	// statusMap := map[string]string{
	// 	"published":      model.ArticleStatusPublished,
	// 	"draft":          model.ArticleStatusDraft,
	// 	"deleted":        model.ArticleStatusDeleted,
	// 	"pending_review": model.ArticleStatusPendingReview,
	// 	"rejected":       model.ArticleStatusRejected,
	// }

	// validStatus, exists := statusMap[req.StatusCode]
	// if !exists {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "无效状态码"})
	// 	return
	// }

	var article model.Article
	if err := DB.Model(&article).Where("id = ?", articleID).Update("status", model.ArticleStatusPublished).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "状态更新失败", "code": 400})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功", "code": 200})
}

// 获取待审核文章
func (ctrl *ArticleController) GetPendingArticles(c *gin.Context) {
	var articles []model.Article
	if result := DB.Where("status = ?", "pending_review").Find(&articles); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles, "code": 200, "msg": "获取待审核文章成功"})
}

// 拒绝文章
func (ctrl *ArticleController) RejectArticle(c *gin.Context) {
	// 解析请求参数
	var req struct {
		ArticleID    uint   `json:"article_id" binding:"required"`
		UserID       uint   `json:"user_id" binding:"required"`
		RejectReason string `json:"reject_reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": 400})
		return
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新文章审核信息
	updateData := map[string]interface{}{
		"status":        "rejected",
		"auditor_id":    req.UserID,
		"reject_reason": req.RejectReason,
		"audit_time":    time.Now(),
		"audit_status":  "rejected",
	}

	if err := tx.Model(&model.Article{}).Where("id = ?", req.ArticleID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": 400})
		return
	}

	// 记录管理员日志
	adminLog := model.AdminLog{
		AdminID:     req.UserID,
		TargetType:  "article",
		TargetID:    req.ArticleID,
		Action:      "reject",
		Description: fmt.Sprintf("驳回文章ID %d，原因：%s", req.ArticleID, req.RejectReason),
		IPAddress:   c.ClientIP(),
	}

	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": 400})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": 400})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章驳回操作已完成", "code": 200})
}

// 点赞文章
func (ctrl *ArticleController) Like(c *gin.Context) {
	var like model.Like
	if err := c.ShouldBindJSON(&like); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已点赞
		var existingLike model.Like
		if err := tx.Where("user_id = ? AND article_id = ?", like.UserID, like.ArticleID).First(&existingLike).Error; err == nil {
			return fmt.Errorf("already liked")
		}

		// 创建点赞记录
		if err := tx.Create(&like).Error; err != nil {
			return err
		}

		// 更新文章点赞数
		if err := tx.Model(&model.Article{}).Where("id = ?", like.ArticleID).
			Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "liked successfully"})
}

// 取消点赞
func (ctrl *ArticleController) Unlike(c *gin.Context) {
	var req struct {
		UserID    uint `json:"user_id"`
		ArticleID uint `json:"article_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// 查找并删除点赞记录
		if err := tx.Where("user_id = ? AND article_id = ?", req.UserID, req.ArticleID).
			Delete(&model.Like{}).Error; err != nil {
			return err
		}

		// 更新文章点赞数
		if err := tx.Model(&model.Article{}).Where("id = ?", req.ArticleID).
			Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(400, gin.H{"error": "like record not found"})
		return
	}

	c.JSON(200, gin.H{"message": "unliked successfully"})
}

// 验证文章和用户的收藏关系
func (ctrl *ArticleController) CheckFavoriteStatus(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("userid"))
	articleID, _ := strconv.Atoi(c.Param("articleid"))

	var favorite model.Favorite
	if err := DB.Where("user_id =? AND article_id =?", userID, articleID).First(&favorite).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "未收藏", "is_favorite": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "该用户收藏了该文章", "is_favorite": true})
}

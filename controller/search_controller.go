package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"TraeCNServer/pkg"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchController struct{}

// SearchArticles 文章搜索
func (sc *SearchController) SearchArticles(c *gin.Context) {
	query := c.Query("q")
	userID := c.Query("userid")
	if userID != "" {
		userIDUint, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			log.Printf("用户ID转换错误: %v", err)
		}
		if err == nil {
			history := model.SearchHistory{
				UserID:        uint(userIDUint),
				SearchContent: query,
				Timestamp:     time.Now(),
			}
			if err := DB.Create(&history).Error; err != nil {
				log.Printf("搜索历史保存失败: %v", err)
			}
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	localCh := make(chan []model.Article)
	externalCh := make(chan interface{})

	// 本地数据库查询协程
	go func() {
		var localArticles []model.Article
		db := DB.Where("title LIKE ?", "%"+query+"%")
		db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&localArticles)
		localCh <- localArticles
	}()

	// 外部爬虫协程
	go func() {
		externalArticles := pkg.CrawlerTx(query)
		externalCh <- externalArticles
	}()

	var localArticles []model.Article
	var externalArticles interface{}
	for i := 0; i < 2; i++ {
		select {
		case l := <-localCh:
			localArticles = l
		case e := <-externalCh:
			externalArticles = e
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":      "查询成功",
		"code":     200,
		"local":    localArticles,
		"external": externalArticles,
	})
}

// SearchByTag 标签搜索
func (sc *SearchController) SearchByTag(c *gin.Context) {
	tagName := c.Query("tag")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var articles []model.Article
	db := DB.Joins("JOIN article_tags ON articles.id = article_tags.article_id").
		Joins("JOIN tags ON article_tags.tag_id = tags.id").
		Where("tags.name = ?", tagName)

	var total int64
	db.Model(&model.Article{}).Count(&total)

	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list": articles,
			"pagination": gin.H{
				"total":        total,
				"current_page": page,
				"per_page":     pageSize,
				"mag":          "查询成功",
				"code":         200,
			},
		},
	})
}

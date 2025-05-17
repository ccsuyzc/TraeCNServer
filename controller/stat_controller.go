package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatController struct{}

// 获取用户总数
func (ctrl *StatController) GetUserCount(c *gin.Context) {
	var count int64
	if err := DB.Model(&model.User{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_count": count, "code": 200, "message": "success", "data": count})
}

// 获取文章总数
func (ctrl *StatController) GetArticleCount(c *gin.Context) {
	var count int64
	if err := DB.Model(&model.Article{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article_count": count, "code": 200, "message": "success", "data": count})
}

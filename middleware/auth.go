package middleware

import (
	"net/http"

	. "TraeCNServer/db"
	"TraeCNServer/model"

	"github.com/gin-gonic/gin"
)

func AuthorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uint)
		articleID := c.Param("id")

		var article model.Article
		if err := DB.First(&article, articleID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
			return
		}

		if article.UserID != userID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无操作权限"})
			return
		}

		c.Next()
	}
}

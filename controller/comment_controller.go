package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CommentController 评论控制器
type CommentController struct{}

// CreateComment 创建评论
func (cc *CommentController) CreateComment(c *gin.Context) {
	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 打印评论数据
	fmt.Printf("Received comment: %+v\n", comment)
	// // 从JWT获取用户ID
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
	// 	return
	// }
	// comment.UserID = userID.(uint)

	// 验证文章是否存在
	var article model.Article
	if result := DB.First(&article, comment.ArticleID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "指定文章不存在"})
		return
	}

	if err := DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "评论创建成功", "data": comment})
}

// GetComments 获取评论列表
func (cc *CommentController) GetComments(c *gin.Context) {
	var comments []model.Comment
	articleID := c.Param("id")
	// 转化为数字类型
	articleIDInt, _ := strconv.Atoi(articleID)
	fmt.Print("this is", articleIDInt)
	if err := DB.Where("article_id = ?", articleIDInt).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": comments})
}

// UpdateComment 更新评论
func (cc *CommentController) UpdateComment(c *gin.Context) {
	var comment model.Comment
	id := c.Param("id")
	if err := DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment updated successfully", "data": comment})
}

// DeleteComment 删除评论
func (cc *CommentController) DeleteComment(c *gin.Context) {
	var comment model.Comment
	id := c.Param("id")
	if err := DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err := DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

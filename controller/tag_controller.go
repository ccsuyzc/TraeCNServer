package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TagController 标签控制器
type TagController struct{}

// CreateTag 创建标签
func (tc *TagController) CreateTag(c *gin.Context) {
	var tag model.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "标签创建成功", "data": tag})
}

// GetTag 获取标签
func (tc *TagController) GetTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签未找到"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tag})
}

// UpdateTag 更新标签
func (tc *TagController) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签未找到"})
		return
	}
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "标签更新成功", "data": tag})
}

// DeleteTag 删除标签
func (tc *TagController) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标签未找到"})
		return
	}
	if err := DB.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "标签删除成功"})
}

// GetAllTags 获取所有标签
func (tc *TagController) GetAllTags(c *gin.Context) {
	var tags []model.Tag
	if err := DB.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

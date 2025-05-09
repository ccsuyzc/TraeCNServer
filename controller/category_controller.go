package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryController struct{}

// 创建分类
func (ctrl *CategoryController) CreateCategory(c *gin.Context) {
	var category model.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := DB.Create(&category); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// 获取全部分类
func (ctrl *CategoryController) GetAllCategories(c *gin.Context) {
	var categories []model.Category
	if result := DB.Find(&categories); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// 更新分类
func (ctrl *CategoryController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category model.Category
	if result := DB.First(&category, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Save(&category)
	c.JSON(http.StatusOK, category)
}

// 删除分类
func (ctrl *CategoryController) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if result := DB.Delete(&model.Category{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "分类删除成功"})
}

// 获取分类下的文章数量
func (ctrl *CategoryController) GetArticleCount(c *gin.Context) {
	id := c.Param("id")
	var category model.Category
	if result := DB.First(&category, id); result.Error!= nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}

	var count int64
	if result := DB.Model(&model.Article{}).Where("category_id = ?", category.ID).Count(&count); result.Error!= nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"article_count": count,"code": "200", "message": "success", "data": count})
}
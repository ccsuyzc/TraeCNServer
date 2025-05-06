package controller

import (
	"TraeCNServer/model"
	"net/http"
	"time"
    . "TraeCNServer/db"
	"github.com/gin-gonic/gin"
)

func CreateSearchHistory(c *gin.Context) {
	var history model.SearchHistory
	if err := c.ShouldBindJSON(&history); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history.Timestamp = time.Now()
	result := DB.Create(&history)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录保存失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "搜索记录保存成功"})
}

func CreateReadingHistory(c *gin.Context) {
	var history model.ReadingHistory
	if err := c.ShouldBindJSON(&history); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history.Timestamp = time.Now()
	result := DB.Create(&history)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "阅读记录保存失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "阅读记录保存成功"})
}

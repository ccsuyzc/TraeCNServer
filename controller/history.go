package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"fmt"
	"net/http"
	"time"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"code": 400,"msg":"参数解析错误"})
		return
	}

	history.Timestamp = time.Now()
	fmt.Printf("开始保存阅读记录:%v \n",history.UserID)
	result := DB.Create(&history)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "阅读记录保存失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "阅读记录保存成功"})
}

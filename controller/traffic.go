package controller

import (
	"net/http"
	"time"

	. "TraeCNServer/db"
	. "TraeCNServer/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrafficController struct{}

// GetLast7DaysTraffic 返回最近7天的流量数据
func (tc *TrafficController) GetLast7DaysTraffic(c *gin.Context) {
	var traffics []Traffic
	now := time.Now()
	startDate := now.AddDate(0, 0, -6).Truncate(24 * time.Hour)
	endDate := now.Truncate(24 * time.Hour)

	dbConn := DB
	if dbConn == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库未初始化" ,"code":"500"})
		return
	}

	err := dbConn.Where("date >= ? AND date <= ?", startDate, endDate).Order("date asc").Find(&traffics).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询流量数据失败","code":"500"})
		return
	}

	// 补全没有数据的日期
	result := make([]gin.H, 0, 7)
	dateMap := make(map[string]int)
	for _, t := range traffics {
		dateMap[t.Date.Format("2006-01-02")] = t.Count
	}
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		count := dateMap[dateStr]
		result = append(result, gin.H{"date": dateStr, "count": count})
	}

	c.JSON(http.StatusOK, gin.H{"data": result,"code":"200","msg":"查询成功"})
}

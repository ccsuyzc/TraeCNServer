package middleware

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"time"

	"github.com/gin-gonic/gin"
)

// 流量统计中间件，每日自增
func TrafficMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		today := time.Now().Truncate(24 * time.Hour)
		var traffic model.Traffic
		if err := DB.Where("date = ?", today).First(&traffic).Error; err != nil {
			// 不存在则新建
			traffic = model.Traffic{Date: today, Count: 1}
			DB.Create(&traffic)
		} else {
			// 存在则自增
			DB.Model(&traffic).UpdateColumn("count", traffic.Count+1)
		}
		c.Next()
	}
}

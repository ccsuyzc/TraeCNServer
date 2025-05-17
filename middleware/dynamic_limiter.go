package middleware

// import (
// 	"sync"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/shirou/gopsutil/v3/cpu"
// 	"github.com/shirou/gopsutil/v3/mem"
// 	"go.uber.org/zap"
// 	"golang.org/x/time/rate"
// )

// // SystemLimiter 系统级限流器
// type SystemLimiter struct {
// 	mu           sync.Mutex     
// 	limiter      *rate.Limiter
// 	lastCPULoad  float64
// 	lastMemUsage float64
// 	maxRate      int
// 	minRate      int
// 	logger       *zap.Logger
// }

// // NewSystemLimiter 创建系统级限流器，传入参数：最大请求数，最小请求数，日志记录器
// func NewSystemLimiter(maxRate, minRate int, logger *zap.Logger) *SystemLimiter {
// 	return &SystemLimiter{
// 		limiter: rate.NewLimiter(rate.Limit(maxRate), maxRate*2),
// 		maxRate: maxRate,
// 		minRate: minRate,
// 		logger:  logger,
// 	}
// }
// // 获取系统指标
// func (sl *SystemLimiter) getSystemMetrics() (cpuLoad, memUsage float64) {
// 	percent, err := cpu.Percent(500*time.Millisecond, false)
// 	if err != nil {
// 		sl.logger.Error("获取CPU指标失败", zap.Error(err))
// 	} else if len(percent) > 0 {
// 		cpuLoad = percent[0]
// 	}

// 	vm, err := mem.VirtualMemory()
// 	if err != nil {
// 		sl.logger.Error("获取内存指标失败", zap.Error(err))
// 	} else {
// 		memUsage = vm.UsedPercent
// 	}
// 	return
// }
// // 动态调整限流
// func (sl *SystemLimiter) adjustLimit() {
// 	for {
// 		cpuLoad, memUsage := sl.getSystemMetrics()
// 		sl.mu.Lock()
// 		sl.lastCPULoad = 0.7*sl.lastCPULoad + 0.3*cpuLoad
// 		sl.lastMemUsage = 0.7*sl.lastMemUsage + 0.3*memUsage
// 		currentLoad := (sl.lastCPULoad + sl.lastMemUsage) / 2

// 		newRate := float64(sl.maxRate) * (1 - currentLoad/100)
// 		if newRate < float64(sl.minRate) {
// 			newRate = float64(sl.minRate)
// 		} else if newRate > float64(sl.maxRate) {
// 			newRate = float64(sl.maxRate)
// 		}

// 		sl.limiter.SetLimit(rate.Limit(newRate))
// 		sl.limiter.SetBurst(int(newRate * 1.5))
// 		sl.mu.Unlock()
// 		time.Sleep(3 * time.Second)
// 	}
// }
// // 中间件
// func (sl *SystemLimiter) SystemLimiterMiddleware() gin.HandlerFunc {
// 	go sl.adjustLimit()

// 	return func(c *gin.Context) {
// 		sl.mu.Lock()
// 		defer sl.mu.Unlock()

// 		if !sl.limiter.Allow() {
// 			sl.logger.Warn("系统级限流触发",
// 				zap.Float64("cpu", sl.lastCPULoad),
// 				zap.Float64("mem", sl.lastMemUsage),
// 				zap.Float64("rate", float64(sl.limiter.Limit())),
// 			)
// 			c.AbortWithStatusJSON(429, gin.H{
// 				"code": "SYSTEM_OVERLOAD",
// 				"msg":  "系统资源紧张，请稍后重试",
// 				"metric": gin.H{
// 					"cpu":  sl.lastCPULoad,
// 					"mem":  sl.lastMemUsage,
// 					"rate": sl.limiter.Limit(),
// 				},
// 			})
// 			return
// 		}
// 		c.Next()
// 	}
// }

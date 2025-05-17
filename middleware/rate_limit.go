package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate" 	
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)
type DynamicWindow struct {
	WindowSize  time.Duration
	SegmentNum  int
	Threshold   int
	lastRequest time.Time
	sync.RWMutex
}
// 系统级限流器
func (dw *DynamicWindow) segmentDuration() time.Duration {
	return dw.WindowSize / time.Duration(dw.SegmentNum)
}

// 获取当前时间段的索引
func (dw *DynamicWindow) currentSegment() int {
	elapsed := time.Since(dw.lastRequest)
	return int(elapsed/dw.segmentDuration()) % dw.SegmentNum
}
// 检查当前时间段是否超出阈值

func (dw *DynamicWindow) CalculateRequestStdDev(rdb *redis.Client, key string) float64 {
	ctx := context.Background()
	now := time.Now().UnixNano()
	startTime := now - dw.WindowSize.Nanoseconds()

	timestamps, err := rdb.ZRangeByScore(ctx, key+":timestamps", &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", startTime),
		Max: fmt.Sprintf("%d", now),
	}).Result()
	if err != nil || len(timestamps) < 2 {
		return 0.0
	}

	var sum, mean, sd float64
	for _, ts := range timestamps {
		t, _ := strconv.ParseInt(ts, 10, 64)
		sum += float64(t)
	}
	mean = sum / float64(len(timestamps))

	for _, ts := range timestamps {
		t, _ := strconv.ParseInt(ts, 10, 64)
		sd += math.Pow(float64(t)-mean, 2)
	}
	return math.Sqrt(sd / float64(len(timestamps)))
}
// 调整窗口大小

func (dw *DynamicWindow) AdjustSegments(rdb *redis.Client, key string) {
	dw.Lock()
	defer dw.Unlock()
	stdDev := dw.CalculateRequestStdDev(rdb, key)
	if stdDev > float64(dw.WindowSize.Nanoseconds()/2) {
		dw.SegmentNum *= 2
	} else {
		dw.SegmentNum = int(math.Max(float64(dw.SegmentNum/2), 4))
	}
	dw.SegmentNum = int(math.Min(math.Max(float64(dw.SegmentNum), 4), 64))
}

// 初始化自动调整定时器
func initAutoAdjustTicker(dw *DynamicWindow, rdb *redis.Client) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			keys, _ := rdb.Keys(context.Background(), "rate_limit:*").Result()
			for _, key := range keys {
				dw.AdjustSegments(rdb, key)
			}
		}
	}()
}

// 动态窗口限流中间件
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration, sysLimiter *SystemLimiter) gin.HandlerFunc {
	dw := &DynamicWindow{
		WindowSize:  window,
		SegmentNum:  6,
		Threshold:   limit,
		lastRequest: time.Now(),
	}
	// 初始化自动调整定时器
	initAutoAdjustTicker(dw, rdb)

	return func(c *gin.Context) {
		identifier := c.ClientIP() // 默认使用IP限流
		if userID, exists := c.Get("userID"); exists {
			identifier = fmt.Sprintf("user_%v", userID)
		}

		now := time.Now().UnixNano()
		key := fmt.Sprintf("rate_limit:%s", identifier)
		startTime := now - window.Nanoseconds()

		// 系统级限流优先
		if !sysLimiter.limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{
				"code": "SYSTEM_PRIORITY_LIMIT",
				"msg":  "系统资源限制，请稍后重试",
			})
			return
		}

		// 使用Redis事务
		pipe := rdb.Pipeline()
		pipe.ZRemRangeByScore(context.Background(), key, "0", fmt.Sprintf("%d", startTime))
		timeSegment := dw.currentSegment()
		member := fmt.Sprintf("%d:%d", timeSegment, now)
		pipe.ZAdd(context.Background(), key, redis.Z{Score: float64(now), Member: member})
		pipe.Expire(context.Background(), key, window)
		pipe.ZCard(context.Background(), key)

		cmds, err := pipe.Exec(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		count := cmds[3].(*redis.IntCmd).Val()
		go dw.AdjustSegments(rdb, key)

		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("请求过于频繁，请%d秒后再试", window/time.Second),
			})
			return
		}
		c.Next()
	}
}

// 系统级限流器
type SystemLimiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
	maxRate float64
}

// 初始化系统级限流器
func NewSystemLimiter(maxRatePerSecond int) *SystemLimiter {
	return &SystemLimiter{
		limiter: rate.NewLimiter(rate.Limit(maxRatePerSecond), maxRatePerSecond*2),
		maxRate: float64(maxRatePerSecond),
	}
}

// 动态调整系统级限流阈值（可对接监控系统）
func (sl *SystemLimiter) Adjust(maxRate float64) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.limiter.SetLimit(rate.Limit(maxRate))
}
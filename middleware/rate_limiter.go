package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu              sync.Mutex
	ipLimiters      = make(map[string]*ipLimiter)
	defaultRPS      = 10 // 默认每秒请求数
	defaultBurst    = 20 // 默认突发流量
	cleanupInterval = 10 * time.Minute
)

func init() {
	go cleanupLimiter()
}

func RateLimiter(rps, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		limiter, exists := ipLimiters[ip]
		if !exists {
			limiter = &ipLimiter{
				limiter:  rate.NewLimiter(rate.Limit(rps), burst),
				lastSeen: time.Now(),
			}
			ipLimiters[ip] = limiter
		} else {
			limiter.lastSeen = time.Now()
		}
		mu.Unlock()

		if !limiter.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

func cleanupLimiter() {
	for {
		time.Sleep(cleanupInterval)

		mu.Lock()
		for ip, limiter := range ipLimiters {
			if time.Since(limiter.lastSeen) > cleanupInterval {
				delete(ipLimiters, ip)
			}
		}
		mu.Unlock()
	}
}

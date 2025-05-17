package middleware

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// 安全配置
var (
	blacklistedIPs      = []string{"1.1.1.1", "2.2.2.2"} // 示例黑名单IP
	maliciousUserAgents = regexp.MustCompile(`(?i)(bot|crawler|scanner|spider|hack)`)
)
// SecurityMiddleware 安全中间件
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// IP黑名单检查
		clientIP := c.ClientIP()
		for _, ip := range blacklistedIPs {
			if ip == clientIP {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "访问被拒绝"})
				return
			}
		}

		// User-Agent检查
		userAgent := c.Request.UserAgent()
		if maliciousUserAgents.MatchString(userAgent) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "异常的User-Agent"})
			return
		}

		// XSS防护处理
		c.Writer = &responseWriterWrapper{
			ResponseWriter: c.Writer,
			xssEscaped:     false,
		}

		c.Next()
	}
}

// responseWriterWrapper 自定义响应写入器
type responseWriterWrapper struct {
	gin.ResponseWriter
	xssEscaped bool
}

// Write 重写写入方法，实现XSS防护
func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	if !w.xssEscaped && w.Header().Get("Content-Type") == "text/html" {
		escaped := template.HTMLEscapeString(string(data))
		return w.ResponseWriter.Write([]byte(escaped))
	}
	return w.ResponseWriter.Write(data)
}

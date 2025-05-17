package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// jwtKey 用于JWT签名和验证的密钥
var jwtKey = []byte("yzc")

// Claims 定义了JWT中包含的声明信息
// Username: 存储用户名
// jwt.StandardClaims: 包含JWT标准声明如过期时间等
type Claims struct {
	Username string `json:"username"`
	UserID   uint `json:"userid"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware 创建JWT认证中间件
// 返回值: gin.HandlerFunc - Gin框架的中间件处理函数
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 去除Bearer前缀获取纯token字符串
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		// 解析并验证JWT token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 将用户名存入上下文供后续处理使用
		c.Set("username", claims.Username)
        c.Set("userid", claims.UserID)
		c.Next()
	}
}

// AuthMiddleware 创建JWT认证中间件
// 返回值: gin.HandlerFunc - Gin框架的中间件处理函数
func AuthMiddleware2() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		userID := c.GetHeader("userid")
		if userID != "" {
		  c.Set("userID", userID)
		}
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 去除Bearer前缀获取纯token字符串
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		// 解析并验证JWT token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 将用户名存入上下文供后续处理使用
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GenerateToken 生成JWT token
// 参数: username string - 要包含在token中的用户名
// 返回值: string - 生成的JWT token字符串
//
//	error - 生成过程中遇到的错误
func GenerateToken(username string,userID uint) (string, error) {
	// 设置token过期时间为24小时后
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		UserID:   (userID),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// 使用HS256算法创建带声明的token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用jwtKey签名并返回token字符串
	return token.SignedString(jwtKey)
}

// GenerateAdminToken 生成管理员JWT token
// 参数: username string - 要包含在token中的用户名
// 返回值: string - 生成的JWT token字符串
//
//	error - 生成过程中遇到的错误
func GenerateAdminToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		Role:     "admin",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// LoggerMiddleware 创建日志记录中间件
// 返回值: gin.HandlerFunc - Gin框架的中间件处理函数
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 日志记录逻辑
		c.Next()
	}
}

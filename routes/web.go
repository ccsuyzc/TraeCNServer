package routes

import "github.com/gin-gonic/gin"

func SetupWebRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to My Blog",
		})
	})
}

package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Timestamp time.Time   `json:"timestamp"`
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("[ERROR] %v", err.Error())

			switch e := err.Err.(type) {
			case *gin.Error:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Timestamp: time.Now(),
					Status:    http.StatusBadRequest,
					Message:   "Invalid request",
					Details:   e.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Timestamp: time.Now(),
					Status:    http.StatusInternalServerError,
					Message:   "Internal server error",
				})
			}
		}
	}
}

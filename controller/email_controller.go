package controller

import (
	"net/http"

	"TraeCNServer/pkg"

	"github.com/gin-gonic/gin"
	// "github.com/jordan-wright/email"
)

type EmailController struct{}


// 测试用
func (ec *EmailController) SendDemo(c *gin.Context) {
	var request struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	EM := pkg.EmailPkg{}
	err := EM.SendEmail(request.To, request.Content, request.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}


package controller

// import (
// 	"TraeCNServer/config"
// 	. "TraeCNServer/db"
// 	"TraeCNServer/model"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// // AIController AI问答控制器
// type AIController struct{}

// // APIResponse AI API响应结构体
// type APIResponse struct {
// 	Answer       string `json:"answer"`
// 	InputTokens  int    `json:"inputTokens"`
// 	OutputTokens int    `json:"outputTokens"`
// }

// // AskQuestion 提问接口
// func (ac *AIController) AskQuestion(c *gin.Context) {
// 	// 获取用户ID
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// 解析请求体
// 	var req struct {
// 		Question string `json:"question"`
// 		Model    string `json:"model"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 获取API配置
// 	apiConfig, err := config.GetAIConfig()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI config"})
// 		return
// 	}

// 	// 调用AI API获取回答
// 	client := &http.Client{}
// 	apiResponse, err := callAIAPI(client, apiConfig, req.Question, req.Model)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI API: " + err.Error()})
// 		return
// 	}

// 	answer := apiResponse.Answer
// 	inputTokens := apiResponse.InputTokens
// 	outputTokens := apiResponse.OutputTokens

// 	// 保存对话记录
// 	conversation := model.AIConversation{
// 		UserID:       userID.(uint),
// 		SessionID:    "", // TODO: 生成会话ID
// 		Question:     req.Question,
// 		Answer:       answer,
// 		Model:        req.Model,
// 		InputTokens:  inputTokens,
// 		OutputTokens: outputTokens,
// 		TotalTokens:  inputTokens + outputTokens,
// 		CreatedTime:  time.Now(),
// 	}

// 	if err := DB.Create(&conversation).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// TODO: 更新用户AI使用统计

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Question answered successfully",
// 		"data": gin.H{
// 			"answer": answer,
// 			"tokens": gin.H{
// 				"input":  inputTokens,
// 				"output": outputTokens,
// 				"total":  inputTokens + outputTokens,
// 			},
// 		},
// 	})
// }

// // GetConversationHistory 获取对话历史
// func (ac *AIController) GetConversationHistory(c *gin.Context) {
// 	// 获取用户ID
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// 查询对话历史
// 	var conversations []model.AIConversation
// 	if err := DB.Where("user_id = ?", userID).Order("created_time desc").Find(&conversations).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"data": conversations})
// }

// // GetTokenUsage 获取token使用统计
// func (ac *AIController) GetTokenUsage(c *gin.Context) {
// 	// 获取用户ID
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// 查询token使用统计
// 	var usage model.AIUsageStatistics
// 	if err := DB.Where("user_id = ?", userID).First(&usage).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"data": usage})
// }

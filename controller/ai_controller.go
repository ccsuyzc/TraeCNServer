package controller

// import (
// 	// "TraeCNServer/config"
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"strconv"
// 	"io"

// 	. "TraeCNServer/db"
// 	. "TraeCNServer/model"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// var Token string = "sk-340a96bc42f74a98b33b4bfe49953387" // 替换为实际的Token值
// func ChatHandler(c *gin.Context) {
// 	// 获取用户信息
// 	// userID := c.MustGet("userID").(uint)  jwt获取
// 	// 转换为 uint

// 	userIDstring,_ := strconv.Atoi(c.Param("userid"))
// 	userID := uint(userIDstring)
	
// 	// 解析请求
// 	var req ChatRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
// 		return
// 	}

// 	// 提取用户问题
// 	var question string
// 	for i := len(req.Messages) - 1; i >= 0; i-- {
// 		if req.Messages[i].Role == "user" {
// 			question = req.Messages[i].Content
// 			break
// 		}
// 	}
//     fmt.Println("question:", question)
// 	// 调用大模型
// 	resp, err,body := callAIAPI(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service unavailable"})
// 		return
// 	}

// 	// 解析响应
// 	if len(resp.Choices) == 0 {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "empty AI response"})
// 		return
// 	}
// 	answer := resp.Choices[0].MessageAI.Content

// 	// 生成会话ID（示例实现，需根据业务需求调整）
// 	sessionID := generateSessionID(userID)
// 	serialNumber := getNextSerial(sessionID)

// 	// 保存对话记录
// 	conversation := AIConversation{
// 		SessionID:     int(sessionID),
// 		SerialNumber:  serialNumber,
// 		UserID:       userID,
// 		Role:         "user",
// 		Question:     question,
// 		Answer:       answer,
// 		AiModel:      req.Model,
// 		InputTokens:  resp.Usage.PromptTokens,
// 		OutputTokens: resp.Usage.CompletionTokens,
// 		TotalTokens:  resp.Usage.TotalTokens,
// 		CreatedTime:  time.Now(),
// 	}

// 	if err := DB.Create(&conversation).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save conversation"})
// 		return
// 	}

// 	// c.JSON(http.StatusOK, gin.H{
// 	// 	"answer": answer,
// 	// 	"usage":  resp.Usage,
// 	// })

// 	// 响应
// 	c.JSON(http.StatusOK, body)
// }

// // 生成会话ID的示例实现
// func generateSessionID(userID uint) uint {
// 	// 这里可以实现更复杂的逻辑，比如查询最近未使用的会话
// 	return uint(time.Now().Unix())
// }

// // 获取下一个序列号
// func getNextSerial(sessionID uint) int {
// 	var maxSerial int
// 	DB.Model(&AIConversation{}).
// 		Where("session_id = ?", sessionID).
// 		Select("COALESCE(MAX(serial_number), 0)").
// 		Scan(&maxSerial)
// 	return maxSerial + 1
// }

// func callAIAPI(req ChatRequest) (*DeepseekResponse, error,io.ReadCloser) {
// 	payload, _ := json.Marshal(req)
	
// 	client := &http.Client{Timeout: 30 * time.Second}
// 	aiReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(payload))
	
// 	aiReq.Header.Set("Content-Type", "application/json")
// 	aiReq.Header.Set("Authorization", "Bearer "+Token)
	
// 	resp, err := client.Do(aiReq)
// 	if err != nil {
// 		return nil, err,nil
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("AI API returned status: %d", resp.StatusCode),nil
// 	}

// 	// 解析响应
// 	var aiResp DeepseekResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
// 		return nil, err,nil
// 	}

// 	return &aiResp, nil,resp.Body
// }


// // import (
// // 	"TraeCNServer/config"
// // 	. "TraeCNServer/db"
// // 	"TraeCNServer/model"
// // 	"net/http"
// // 	"time"

// // 	"github.com/gin-gonic/gin"
// // )

// // // AIController AI问答控制器
// // type AIController struct{}

// // // APIResponse AI API响应结构体
// // type APIResponse struct {
// // 	Answer       string `json:"answer"`
// // 	InputTokens  int    `json:"inputTokens"`
// // 	OutputTokens int    `json:"outputTokens"`
// // }

// // // AskQuestion 提问接口
// // func (ac *AIController) AskQuestion(c *gin.Context) {
// // 	// 获取用户ID
// // 	userID, exists := c.Get("userID")
// // 	if !exists {
// // 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// // 		return
// // 	}

// // 	// 解析请求体
// // 	var req struct {
// // 		Question string `json:"question"`
// // 		Model    string `json:"model"`
// // 	}
// // 	if err := c.ShouldBindJSON(&req); err != nil {
// // 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	// 获取API配置
// // 	apiConfig, err := config.GetAIConfig()
// // 	if err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI config"})
// // 		return
// // 	}

// // 	// 调用AI API获取回答
// // 	client := &http.Client{}
// // 	apiResponse, err := callAIAPI(client, apiConfig, req.Question, req.Model)
// // 	if err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI API: " + err.Error()})
// // 		return
// // 	}

// // 	answer := apiResponse.Answer
// // 	inputTokens := apiResponse.InputTokens
// // 	outputTokens := apiResponse.OutputTokens

// // 	// 保存对话记录
// // 	conversation := model.AIConversation{
// // 		UserID:       userID.(uint),
// // 		SessionID:    "", // TODO: 生成会话ID
// // 		Question:     req.Question,
// // 		Answer:       answer,
// // 		Model:        req.Model,
// // 		InputTokens:  inputTokens,
// // 		OutputTokens: outputTokens,
// // 		TotalTokens:  inputTokens + outputTokens,
// // 		CreatedTime:  time.Now(),
// // 	}

// // 	if err := DB.Create(&conversation).Error; err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	// TODO: 更新用户AI使用统计

// // 	c.JSON(http.StatusOK, gin.H{
// // 		"message": "Question answered successfully",
// // 		"data": gin.H{
// // 			"answer": answer,
// // 			"tokens": gin.H{
// // 				"input":  inputTokens,
// // 				"output": outputTokens,
// // 				"total":  inputTokens + outputTokens,
// // 			},
// // 		},
// // 	})
// // }

// // // GetConversationHistory 获取对话历史
// // func (ac *AIController) GetConversationHistory(c *gin.Context) {
// // 	// 获取用户ID
// // 	userID, exists := c.Get("userID")
// // 	if !exists {
// // 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// // 		return
// // 	}

// // 	// 查询对话历史
// // 	var conversations []model.AIConversation
// // 	if err := DB.Where("user_id = ?", userID).Order("created_time desc").Find(&conversations).Error; err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	c.JSON(http.StatusOK, gin.H{"data": conversations})
// // }

// // // GetTokenUsage 获取token使用统计
// // func (ac *AIController) GetTokenUsage(c *gin.Context) {
// // 	// 获取用户ID
// // 	userID, exists := c.Get("userID")
// // 	if !exists {
// // 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// // 		return
// // 	}

// // 	// 查询token使用统计
// // 	var usage model.AIUsageStatistics
// // 	if err := DB.Where("user_id = ?", userID).First(&usage).Error; err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	c.JSON(http.StatusOK, gin.H{"data": usage})
// // }

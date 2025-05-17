package controller

import (
	"encoding/json"
	"fmt"
	"io"

	. "TraeCNServer/db"
	. "TraeCNServer/model"
	. "TraeCNServer/config"
	"bytes"
	"net/http"
	"strconv"
	"time"
 
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 处理聊天请求
func HandleChat(c *gin.Context) {
	// userID := c.MustGet("userID").(uint)
	userIDstring, _ := strconv.Atoi(c.Param("userid"))
	userID := uint(userIDstring)

	// 解析请求参数
	var req struct {
		SessionID string `json:"session_id"` // 可选，新会话自动生成
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
		// 其他参数...
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 会话处理
	var session ChatSession
	if req.SessionID == "" {
		// 创建新会话
		session = ChatSession{
			UserID:      userID,
			SessionUUID: uuid.New().String(),
			AiModel:     req.Model,
			Title:       generateTitle(req.Messages), // 实现标题生成逻辑
		}
		DB.Create(&session)
	} else {
		// 验证会话归属
		if err := DB.Where("session_uuid = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
	}

	// 保存用户消息
	requestID := uuid.New().String()
	for _, msg := range req.Messages {
		if msg.Role == "user" || msg.Role == "system" {
			DB.Create(&ChatMessage{
				SessionID:  session.ID,
				RequestID:  requestID,
				Role:       msg.Role,
				Content:    msg.Content,
				IsResponse: false,
			})
		}
	}

	// 转发到模型API
	// modelReq := buildModelRequest(req) // 构建模型请求
	resp, err := forwardToModelAPI(req)
	if err != nil {
		// 保存错误信息
		DB.Create(&ChatMessage{
			SessionID:  session.ID,
			RequestID:  requestID,
			Role:       "system",
			Content:    err.Error(),
			IsResponse: true,
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model API error"})
		return
	}

	// 处理模型响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "msg": "读取响应错误"})
	}
	defer resp.Body.Close()

	var modelResp struct {
		ID                string `json:"id"`
		Object            string `json:"object"`
		Created           int64  `json:"created"`
		Model             string `json:"model"`
		SystemFingerprint string `json:"system_fingerprint"`
		Choices           []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			Logprobs     interface{} `json:"logprobs"`
			FinishReason string      `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	// 解析模型响应
	if err := json.Unmarshal(bodyBytes, &modelResp); err != nil {
		DB.Create(&ChatMessage{
			SessionID:  session.ID,
			RequestID:  requestID,
			Role:       "system",
			Content:    fmt.Sprintf("解析错误: %v", err),
			IsResponse: true,
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "msg": "解析响应错误"})
		return
	}

	// 保存模型响应
	for _, choice := range modelResp.Choices {
		DB.Create(&ChatMessage{
			SessionID:        session.ID,
			RequestID:        requestID,
			Role:             choice.Message.Role,
			Content:          choice.Message.Content,
			IsResponse:       true,
			PromptTokens:     modelResp.Usage.PromptTokens,
			CompletionTokens: modelResp.Usage.CompletionTokens,
			TotalTokens:      modelResp.Usage.TotalTokens,
			StatusCode:       resp.StatusCode,
		})
	}

	// 返回响应给前端
	c.Data(resp.StatusCode, "application/json", bodyBytes)
}

// 转发请求到模型API
func forwardToModelAPI(request interface{}) (*http.Response, error) {
	reqBody, _ := json.Marshal(request)

	req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Token)

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// 辅助函数：生成会话标题
func generateTitle(messages []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}) string {
	// 实现标题生成逻辑，比如取第一条用户消息的前20个字符
	for _, msg := range messages {
		if msg.Role == "user" {
			if len(msg.Content) > 20 {
				return msg.Content[:20] + "..."
			}
			return msg.Content
		}
	}
	return "New Chat"
}

// request_builder.go
func buildModelRequest(req interface{}) map[string]interface{} {
	// 将前端请求转换为模型API需要的格式
	request := make(map[string]interface{})

	// 使用反射或手动映射字段，这里使用类型断言示例
	if r, ok := req.(map[string]interface{}); ok {
		// 必须字段
		request["messages"] = r["messages"]
		request["model"] = r["model"]

		// 可选字段
		if maxTokens, ok := r["max_tokens"].(float64); ok {
			request["max_tokens"] = int(maxTokens)
		}
		if temperature, ok := r["temperature"].(float64); ok {
			request["temperature"] = temperature
		}
		// 其他字段映射...
	}

	// 设置默认值
	if _, exists := request["temperature"]; !exists {
		request["temperature"] = 1.0
	}

	return request
}

func callAIAPI2(req any) (*DeepseekResponse, error, io.ReadCloser) {
	payload, _ := json.Marshal(req)

	client := &http.Client{Timeout: 30 * time.Second}
	aiReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(payload))

	aiReq.Header.Set("Content-Type", "application/json")
	aiReq.Header.Set("Authorization", "Bearer "+Token)

	resp, err := client.Do(aiReq)
	if err != nil {
		return nil, err, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API returned status: %d", resp.StatusCode), nil
	}

	// 解析响应
	var aiResp DeepseekResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, err, nil
	}

	return &aiResp, nil, resp.Body
}

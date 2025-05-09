package ai

import (
	"TraeCNServer/config"
	// "fmt"
	// "log"
	// . "TraeCNServer/db"
	// . "TraeCNServer/model"
	"bytes"
	"encoding/json"

	// "fmt"
	"io/ioutil"
	// "log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
	// "github.com/google/uuid"
)

// // 保存对话记录
// func saveConversation(sessionID string, userID uint, question string, answer string, model string, inputTokens int, outputTokens int) error {
// 	conversation := AIConversation{
// 		// SessionID:    sessionID,
// 		UserID:   userID,
// 		Question: question,
// 		Answer:   answer,
// 		// Model:        model,
// 		InputTokens:  inputTokens,
// 		OutputTokens: outputTokens,
// 		TotalTokens:  inputTokens + outputTokens,
// 		CreatedTime:  time.Now(),
// 	}
// 	return DB.Create(&conversation).Error
// }

// 处理对话请求
// func DeepSeek(c *gin.Context) {
// 	// // 用户认证
// 	// userID, exists := c.Get("userID")  // 假设用户ID存储在上下文中
// 	// if !exists {
// 	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
// 	// 	return
// 	// }

// 	// 获取配置
// 	apiConfig, err := config.GetAIConfig()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取AI配置失败"})
// 		return
// 	}

// 	// 验证请求参数
// 	var req struct {
// 		UserID    uint   `json:"userID"`
// 		SessionID string `json:"sessionID"` // 会话ID，可选
// 		Question  string `json:"question"`  // 会话内容
// 		Model     string `json:"model"`     // 模型名称，可选
// 		Role      string `json:"role"`      // 角色，可选
// 	}
// 	if err2 := c.ShouldBindJSON(&req); err2 != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
// 		return
// 	}

// 	// 构建动态请求体
// 	messages := []map[string]string{{
// 		"role":    "user",
// 		"content": req.Question,
// 	}}

// 	payloadData := map[string]interface{}{
// 		"model":       req.Model,
// 		"messages":    messages,
// 		"max_tokens":  apiConfig.MaxTokens,
// 		"temperature": apiConfig.Temperature, // 控制随机性
// 	}

// 	payloadBytes, _ := json.Marshal(payloadData)
// 	payload := bytes.NewReader(payloadBytes)

// 	client := &http.Client{
// 		Timeout: 30 * time.Second,
// 	}
// 	method := "POST"
// 	url := apiConfig.APIBaseURL + "/v1/chat/completions"
// 	httpReq, err := http.NewRequest(method, url, payload) // 替换为实际的API URL

// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败: " + err.Error()})
// 		return
// 	}
// 	httpReq.Header.Add("Content-Type", "application/json")
// 	httpReq.Header.Add("Accept", "application/json")
// 	httpReq.Header.Add("Authorization", "Bearer "+apiConfig.APIKey)

// 	res, err := client.Do(httpReq)
// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "API请求失败: " + err.Error()})
// 		return
// 	}
// 	defer res.Body.Close()

// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败: " + err.Error()})
// 		return
// 	}
// 	log.Printf("[DeepSeek] Request:\n%s\nResponse[%d]:\n%s", payloadBytes, res.StatusCode, body)
// 	// 解析API响应
// 	var apiResponse struct {
// 		Choices []struct {
// 			Message struct {
// 				Content string `json:"content"`
// 			} `json:"message"`
// 			Usage struct {
// 				InputTokens  int `json:"input_tokens"`
// 				OutputTokens int `json:"output_tokens"`
// 			} `json:"usage"`
// 		} `json:"choices"`
// 	}

// 	if len(apiResponse.Choices) == 0 {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "API返回无效的choices数组"})
// 		return
// 	}
// 	if err := json.Unmarshal(body, &apiResponse); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "响应解析失败"})
// 		return
// 	}

// 	// 保存对话记录
// 	sessionID := uuid.New().String()
// 	c.Writer.Header().Set("X-Session-ID", sessionID)

// 	if err := saveConversation(sessionID, userID.(uint), req.Question, apiResponse.Choices[0].Message.Content, req.Model,
// 		apiResponse.Choices[0].Usage.InputTokens, apiResponse.Choices[0].Usage.OutputTokens); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录保存失败"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"answer": apiResponse.Choices[0].Message.Content,
// 		"usage": gin.H{
// 			"input":  apiResponse.Choices[0].Usage.InputTokens,
// 			"output": apiResponse.Choices[0].Usage.OutputTokens,
// 			"total":  apiResponse.Choices[0].Usage.InputTokens + apiResponse.Choices[0].Usage.OutputTokens,
// 		},
// 	})
// }

// 新增callDeepSeekAPI函数，用于调用DeepSeek API
func callDeepSeekAPI(apiConfig *config.AIConfig, messages []map[string]string, model string) (string, int, int, error) {
	payloadData := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"max_tokens":  apiConfig.MaxTokens,
		"temperature": apiConfig.Temperature,
	}

	payloadBytes, _ := json.Marshal(payloadData)
	payload := bytes.NewReader(payloadBytes)

	client := &http.Client{Timeout: 60 * time.Second}
	httpReq, err := http.NewRequest("POST", apiConfig.APIBaseURL, payload)
	if err != nil {
		return "", 0, 0, err
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer "+apiConfig.APIKey)

	res, err := client.Do(httpReq)
	if err != nil {
		return "", 0, 0, err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil || len(apiResponse.Choices) == 0 {
		return "", 0, 0, err
	}

	return apiResponse.Choices[0].Message.Content, apiResponse.Usage.PromptTokens, apiResponse.Usage.CompletionTokens, nil
}

// 更新Deemo函数调用逻辑
func Deemo(c *gin.Context) {
	apiConfig, err := config.GetAIConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取AI配置失败"})
		return
	}

	var req struct {
		Msg string `json:"msg"`
	}
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	answer, inputTokens, outputTokens, err := callDeepSeekAPI(apiConfig, []map[string]string{{"role": "user", "content": req.Msg}}, "deepseek-chat")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API调用失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer": answer,
		"usage": gin.H{
			"input":  inputTokens,
			"output": outputTokens,
			"total":  inputTokens + outputTokens,
		},
	})
}

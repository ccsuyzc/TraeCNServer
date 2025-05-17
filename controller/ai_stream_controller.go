package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	. "TraeCNServer/db"
	. "TraeCNServer/model"
	. "TraeCNServer/config"

	"github.com/gin-gonic/gin"
)

// AI流式对话接口
func AIStreamChatHandler(c *gin.Context) {
	// 获取用户ID
	userIDstring := c.Param("userid")
	userID := uint(0)
	if userIDstring != "" {
		userIDint, _ := strconv.Atoi(userIDstring)
		userID = uint(userIDint)
	}

	// 解析请求体
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg":"请求体解析失败","code":"400","error": "invalid request", "detail": err.Error()})
		return
	}

	// 构造请求体
	payload, _ := json.Marshal(req)
	client := &http.Client{Timeout: 60 * time.Second}
	aiReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(payload))
	aiReq.Header.Set("Content-Type", "application/json")
	aiReq.Header.Set("Authorization", "Bearer "+Token) // TODO: 替换为安全Token

	// 发送请求
	resp, err := client.Do(aiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg":"大模型服务不可用", "error": "AI service unavailable", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI API error", "status": resp.StatusCode})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	var fullContent bytes.Buffer
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			fullContent.Write(chunk)
			c.Writer.Write(chunk)
			c.Writer.Flush()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "stream read error"})
			return
		}
	}

	// 解析完整内容并存储到数据库（这里只做简单存储，实际可按需解析）
	conversation := AIConversation{
		SessionID:    int(time.Now().Unix()),
		SerialNumber: 1,
		UserID:       userID,
		Role:         "user",
		Question:     "streamed", // 可根据req.Messages获取
		Answer:       fullContent.String(),
		AiModel:      req.Model,
		InputTokens:  0,
		OutputTokens: 0,
		TotalTokens:  0,
		CreatedTime:  time.Now(),
	}
	_ = DB.Create(&conversation)
}

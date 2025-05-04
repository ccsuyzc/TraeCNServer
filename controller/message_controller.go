package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// MessageController 私信控制器
type MessageController struct {
	DB  *gorm.DB
	Hub *MessageHub
}

// 初始化消息中心
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// MessageHub 管理WebSocket连接
type MessageHub struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan model.Message
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

func NewMessageHub() *MessageHub {
	return &MessageHub{
		Broadcast:  make(chan model.Message),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *MessageHub) Run() {
	// 启动消息中心
	fmt.Println("启动websocket消息中心...")
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				err := client.WriteJSON(message)
				if err != nil {
					client.Close()
					delete(h.Clients, client)
				}
			}
		}
	}
}

// SendMessage 发送私信
func (mc *MessageController) SendMessage(c *gin.Context) {
	// currentUser, _ := c.Get("user")
	// sender := currentUser.(model.User)

	var req struct {
		ReceiverID   uint   `json:"receiver_id" binding:"required"`   // 接收者ID
		ReceiverName string `json:"receiver_name" binding:"required"` // 接收者名称
		SenderID     uint   `json:"sender_id" binding:"required"`     // 发送者ID
		SenderName   string `json:"sender_name" binding:"required"`   // 发送者名称
		Content      string `json:"content" binding:"required"`       // 消息内容
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"code":400})
		return
	}

	message := model.Message{
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		SentAt:     time.Now(),
	}

	if err := DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),"code":400})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":200,
		"message": "私信发送成功",
		"data":    message,
	})
}

// 在发送消息成功后添加广播逻辑
// 新增WebSocket消息处理逻辑
func (mc *MessageController) HandleWebSocket(c *gin.Context) {
	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	mc.Hub.Register <- conn
	defer func() {
		mc.Hub.Unregister <- conn
	}()

	for {
		var msgReq struct {
			Test       bool   `json:"test"`        // 测试标识
			ReceiverID uint   `json:"receiver_id"` // 接收者ID
			Content    string `json:"content"`     // 消息内容
		}

		if err := conn.ReadJSON(&msgReq); err != nil {
			conn.WriteJSON(gin.H{
				"error": "invalid request",
			})
			break
		}

		// 测试连接响应
		if msgReq.Test {
			conn.WriteJSON(gin.H{
				"status":    "success",
				"message":   "WebSocket测试连接成功",
				"timestamp": time.Now().Unix(),
			})
			continue
		}

		currentUser := c.MustGet("user").(model.User)
		msg := model.Message{
			SenderID:   currentUser.ID,
			ReceiverID: msgReq.ReceiverID,
			Content:    msgReq.Content,
			SentAt:     time.Now(),
		}

		if err := mc.DB.Create(&msg).Error; err == nil {
			mc.Hub.Broadcast <- msg
		}
	}
}

// GetMessages 获取对话记录
func (mc *MessageController) GetMessages(c *gin.Context) {
	// currentUser, _ := c.Get("user")  //
	// user := currentUser.(model.User)

	// 从param中的query中获取接收者ID

	// var req struct {
	// 	ReceiverID   uint      `json:"receiver_id" binding:"required"`   // 接收者ID
	// 	ReceiverName string    `json:"receiver_name"` // 接收者名称
	// 	SenderID     uint      `json:"sender_id" binding:"required"`     // 发送者ID
	// 	SenderName   string    `json:"sender_name"`   // 发送者名称
	// 	Content      string    `json:"content" `       // 消息内容
	// }

	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	sendid, _ := strconv.Atoi(c.Param("sendid"))
	receiverid, _ := strconv.Atoi(c.Param("receiverid"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var messages []model.Message
	db := DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		receiverid, sendid, sendid, receiverid).
		Order("sent_at desc")

	var total int64
	db.Model(&model.Message{}).Count(&total)

	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list": messages,
			"pagination": gin.H{
				"total":        total,
				"current_page": page,
				"per_page":     pageSize,
				"total_pages":  (int(total) + pageSize - 1) / pageSize,
			},
		},
	})
}

// MarkAsRead 标记消息为已读
func (mc *MessageController) MarkAsRead(c *gin.Context) {
	messageID := c.Param("id")
	currentUser, _ := c.Get("user")
	user := currentUser.(model.User)

	var message model.Message
	if err := DB.Where("id = ? AND receiver_id = ?", messageID, user.ID).First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "消息未找到"})
		return
	}

	if err := DB.Model(&message).Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "消息已标记为已读"})
}

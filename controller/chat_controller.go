package controller

import (
	"TraeCNServer/model"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ChatController 结构体定义 这是一个结构体，用于管理 WebSocket 连接和消息广播。
type ChatController struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

// NewChatController 函数定义 这是一个函数，用于创建一个新的 ChatController 实例。
func NewChatController() *ChatController {
	return &ChatController{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// HandleWebSocket 函数定义 这是一个函数，用于处理 WebSocket 连接。

func (cc *ChatController) HandleWebSocket(c *gin.Context) {
	conn, err := cc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	cc.mutex.Lock()
	cc.clients[conn] = true
	cc.mutex.Unlock()

	for {
		var msg model.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			cc.removeClient(conn)
			break
		}
		cc.broadcastMessage(msg)
	}
}

// broadcastMessage 函数定义 这是一个函数，用于广播消息。
func (cc *ChatController) broadcastMessage(msg model.Message) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	for client := range cc.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(cc.clients, client)
		}
	}
}

// removeClient 函数定义 这是一个函数，用于移除客户端。
func (cc *ChatController) removeClient(conn *websocket.Conn) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	delete(cc.clients, conn)
}

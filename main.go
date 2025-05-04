package main

import (
	"TraeCNServer/controller"
	config "TraeCNServer/db"
	"TraeCNServer/middleware"
	"TraeCNServer/model"
	"TraeCNServer/routes"

	"github.com/gin-contrib/cors"
	"github.com/gorilla/websocket"

	// "github.com/gin-contrib/cors" // 解决浏览器同源问题
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 配置静态文件路由
	r.Static("/images", "./uploads/images")

	// 配置CORS中间件
	CORSd := cors.DefaultConfig() // 默认配置
	CORSd.AllowAllOrigins = true  // 允许所有跨域
	// CORS.AllowOrigins = []string{"http://localhost:8080"}  // 允许的源
	CORSd.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}                    // 允许的HTTP方法
	CORSd.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "token"} // 允许的请求头
	CORSd.AllowCredentials = true                                                               // 允许携带凭证

	r.Use(cors.New(CORSd))
	r.Use(middleware.ErrorHandler())

	config.InitDB()
	config.InitRedis()

	// 初始化WebSocket Hub
	hub := &controller.MessageHub{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan model.Message),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
	go hub.Run()

	// 路由分组
	api := r.Group("/api")
	routes.SetupApiRoutes(api, hub)

	routes.SetupWebRoutes(r)

	r.Run()
}

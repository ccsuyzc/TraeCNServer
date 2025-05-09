package routes

import (
	"TraeCNServer/controller"
	"TraeCNServer/controller/ai"
	"TraeCNServer/middleware"

	"github.com/gin-gonic/gin"
)

func SetupApiRoutes(r *gin.RouterGroup, hub *controller.MessageHub) {

	// 私信路由
	messageCtrl := controller.MessageController{Hub: hub}
	messageGroup := r.Group("/messages")
	// messageGroup := r.Group("/messages").Use(middleware.AuthMiddleware())
	{
		messageGroup.POST("/send", messageCtrl.SendMessage)               //
		messageGroup.GET("/:sendid/:receiverid", messageCtrl.GetMessages) // 获取与特定用户之间的所有私信
		messageGroup.PUT("/:id/read", messageCtrl.MarkAsRead)             // 标记私信为已读
	}
	// 分类路由
	categoryCtrl := controller.CategoryController{}
	categoriesGroup := r.Group("/categories")
	{
		categoriesGroup.POST("", categoryCtrl.CreateCategory)       // 创建分类
		categoriesGroup.GET("", categoryCtrl.GetAllCategories)      // 获取所有分类
		categoriesGroup.PUT("/:id", categoryCtrl.UpdateCategory)    // 更新分类
		categoriesGroup.DELETE("/:id", categoryCtrl.DeleteCategory) // 删除分类
		// 获取分类下面的文章的数量
		categoriesGroup.GET("/:id/articlecount", categoryCtrl.GetArticleCount)
	}
	// 文章路由
	articleCtrl := controller.ArticleController{}
	articlesGroup := r.Group("/articles")
	{
		articlesGroup.GET("/rarticles/:quantity", articleCtrl.GetRandomArticles)              // 随机获取任意篇文章
		articlesGroup.GET("/rarticle/:id/:quantity", articleCtrl.GetRandomArticlesByCategory) // 随机获取指定分类下的任意篇文章
		articlesGroup.GET("", articleCtrl.GetAllArticles)                                     // 获取所有文章
		articlesGroup.GET("/publishedall", articleCtrl.GetAllPublishedArticles)               // 获取所有已发布文章
		articlesGroup.GET("/:id", articleCtrl.GetArticle)                                     // 获取单篇文章
		articlesGroup.POST("", articleCtrl.CreateArticle)                                     // 创建文章
		articlesGroup.GET("/category/:id", articleCtrl.GetArticlesByCategory)                 // 获取分类下的文章列表

		articlesGroup.GET("/category/:id/limit", articleCtrl.GetArticlesByCategoryAndLimit) // 按分类和条数获取文章
		articlesGroup.PUT("/:id", articleCtrl.UpdateArticle)                                // 更新文章
		articlesGroup.DELETE("/:id", articleCtrl.DeleteArticle)                             // 删除文章
		articlesGroup.PUT("/publish-status/:id", articleCtrl.UpdatePublishStatus)           // 更新文章发布状态 管理员操作
		articlesGroup.POST("/drafts", articleCtrl.SaveDraft)                                // 保存草稿
		articlesGroup.GET("/draft/:userid", articleCtrl.GetDrafts)                          // 获取用户所有草稿
		articlesGroup.GET("/draft/:userid/:draftid", articleCtrl.GetDraft)                  // 获取用户的单个文章草稿
		articlesGroup.PUT("/draft/:userid/:draftid", articleCtrl.UpdateDraft)               // 更新用户的单个文章草稿
		articlesGroup.DELETE("/draft/:userid/:draftid", articleCtrl.DeleteDraft)            // 删除用户的单个文章草稿
		articlesGroup.POST("/publish/:id", articleCtrl.PublishDraft)                        // 用户提交文章
		articlesGroup.GET("/published/:userid", articleCtrl.GetPublishedArticles)           // 获取用户所有已发布文章
		articlesGroup.GET("/userall/:userid", articleCtrl.GetUserAllArticle)                // 获取的除了草稿之外的文章

		articlesGroup.GET("/pending", articleCtrl.GetPendingArticles) // 获取待审核文章
		articlesGroup.POST("/reject", articleCtrl.RejectArticle)      // 拒绝文章
		// articlesGroup.GET("/published/:userid/:articleid", articleCtrl.GetPublishedArticle) // 获取用户的单个已发布文章
		// articlesGroup.PUT("/published/:userid/:articleid", articleCtrl.UpdatePublishedArticle) // 更新用户的单个已发布文章
		articlesGroup.POST("/like", articleCtrl.LikeArticle)     // 点赞文章（需认证）
		articlesGroup.POST("/unlike", articleCtrl.UnlikeArticle) // 取消点赞（需认证）

		articlesGroup.POST("/favorite", articleCtrl.FavoriteArticle)                              // 收藏文章（需认证）
		articlesGroup.POST("/unfavorite", articleCtrl.UnfavoriteArticle)                          // 取消收藏（需认证）
		articlesGroup.GET("isfavorite/:userid/favorite-status/:articleid", articleCtrl.CheckFavoriteStatus) // 验证收藏状态（需认证）
	}

	// 用户路由
	userCtrl := controller.UserController{}
	searchCtrl := controller.SearchController{}
	//r.POST("/users/:id/follow", middleware.AuthMiddleware(), userCtrl.FollowUser)     // 关注用户
	//r.DELETE("/users/:id/follow", middleware.AuthMiddleware(), userCtrl.UnfollowUser) // 取消关注

	r.POST("/users/:Originaluser/follow/:targetId", userCtrl.FollowUser)        // 关注用户
	r.DELETE("/users/:Originaluser/notfollow/:targetId", userCtrl.UnfollowUser) // 取消关注
	r.GET("/usersfollowing/:targetId", userCtrl.GetFollowingList)               // 获取用户关注列表
	r.GET("/users/mutual-follow", userCtrl.CheckMutualFollow)                   // 验证双向关注关系

	r.POST("/register", userCtrl.Register)                     // 保留旧注册方式
	r.POST("/login", userCtrl.Login)                           // 账号密码登录
	r.POST("/loginE", userCtrl.LoginE)                         // 邮件密码登录
	r.POST("/loginW", userCtrl.LoginEW)                        // 邮件验证码登录
	r.GET("/users/:id", userCtrl.GetUserProfile)               // 获取用户信息
	r.GET("/users/all", userCtrl.GetAllUsers)                  // 获取所有用户信息
	r.PUT("/users/:id", userCtrl.UpdateUserProfile)            // 更新用户信息
	r.POST("/auth/sendcode", userCtrl.SendVerificationCode)    // 发送验证码
	r.POST("/auth/verifyregister", userCtrl.VerifyAndRegister) // 验证验证码并注册

	// 用户行为记录路由
	r.POST("/search-history", controller.CreateSearchHistory)   // 创建搜索历史记录
	r.POST("/reading-history", controller.CreateReadingHistory) // 创建阅读历史记录

	// 搜索路由
	r.GET("/search/users", userCtrl.SearchUsers)
	r.GET("/search/articles", searchCtrl.SearchArticles)
	r.GET("/search/tags", searchCtrl.SearchByTag) // 验证注册

	// 评论路由
	commentCtrl := controller.CommentController{}            // 评论路由
	r.POST("/comments", commentCtrl.CreateComment)           // 创建评论
	r.GET("/articles/:id/comments", commentCtrl.GetComments) // 获取评论列表
	r.PUT("/comments/:id", commentCtrl.UpdateComment)        // 更新评论
	r.DELETE("/comments/:id", commentCtrl.DeleteComment)     // 删除评论

	// AI问答路由
	// aiCtrl := controller.AIController{}
	// r.POST("/ai/ask", aiCtrl.AskQuestion)
	// r.GET("/ai/conversations", aiCtrl.GetConversationHistory)
	// r.GET("/ai/usage", aiCtrl.GetTokenUsage)

	// 标签路由
	tagCtrl := controller.TagController{}
	tagsGroup := r.Group("/tags")
	{
		tagsGroup.POST("", tagCtrl.CreateTag)      // 创建标签
		tagsGroup.GET(":id", tagCtrl.GetTag)       // 获取标签
		tagsGroup.PUT(":id", tagCtrl.UpdateTag)    // 更新标签
		tagsGroup.DELETE(":id", tagCtrl.DeleteTag) // 删除标签
		tagsGroup.GET("", tagCtrl.GetAllTags)      // 获取所有标签
	}

	r.GET("/check-auth", middleware.AuthMiddleware(), userCtrl.CheckAuthStatus) // 检查用户是否登录

	// 草稿箱路由
	draftsGroup := r.Group("/drafts")
	draftsGroup.Use(middleware.AuthMiddleware())
	{
		draftsGroup.GET("", articleCtrl.GetDrafts)
		draftsGroup.PUT("/:id", articleCtrl.UpdateDraft)
		draftsGroup.DELETE("/:id", articleCtrl.DeleteDraft)
		draftsGroup.PUT("/publish/:id", articleCtrl.PublishDraft)
	}

	// 上传图片路由
	// 上传图片
	imageCtrl := controller.ImageController{}
	r.POST("/upload-image", imageCtrl.UploadImage)
	r.Static("/images", "./uploads/images")

	// 测试路由

	emailCtrl := controller.EmailController{}
	r.POST("/email/send", emailCtrl.SendDemo) // 发送邮件

	// 注册消息控制器WebSocket路由
	r.GET("/ws/messages", middleware.AuthMiddleware(), messageCtrl.HandleWebSocket)

	// WebSocket测试路由（免认证）
	r.GET("/ws/test", messageCtrl.HandleWebSocket)

	// 测试大模型路由
	// r.POST("/ai/chat", middleware.AuthMiddleware(), ai.DeepSeek)
	// r.POST("/ai/chat", ai.DeepSeek)
	r.POST("/deemo-test", ai.Deemo)
}

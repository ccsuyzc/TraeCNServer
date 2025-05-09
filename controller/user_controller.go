package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/middleware"
	"TraeCNServer/model"
	"TraeCNServer/pkg"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct{}

// SearchUsers 用户搜索
func (uc *UserController) SearchUsers(c *gin.Context) {
	query := c.Query("query")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var users []model.User
	db := DB.Where("username LIKE ?", "%"+query+"%")

	var total int64
	db.Model(&model.User{}).Count(&total)

	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list": users,
			"pagination": gin.H{
				"total":        total,
				"current_page": page,
				"per_page":     pageSize,
				"total_pages":  (int(total) + pageSize - 1) / pageSize,
			},
		},
	})
}

// SendVerificationCode 发送验证码
func (uc *UserController) SendVerificationCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	fmt.Printf("这是req:%v", req)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	fmt.Printf("这是code:%v", code)

	if err := RedisClient.Set(c, "verification:"+req.Email, code, 5*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码存储失败"})
		return
	}

	emailCtrl := pkg.EmailPkg{}
	emailCtrl.SendEmail(req.Email, fmt.Sprintf("您的验证码是：%s，5分钟内有效", code), "验证码")
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送", "code": 200})
}

// VerifyAndRegister 验证并注册
func (uc *UserController) VerifyAndRegister(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storedCode, err := RedisClient.Get(c, "verification:"+req.Email).Result()
	if err != nil || storedCode != req.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码无效或已过期"})
		return
	}

	user := model.User{
		Email:        req.Email,
		Username:     "user_" + randString(8),
		PasswordHash: "123456789",
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	RedisClient.Del(c, "verification:"+req.Email)
	token, _ := middleware.GenerateToken(user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"code":     200,
		"message":  "注册成功",
		"username": user.Username,
		// "password": user.Password,
		"token": token,
	})
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Register 用户注册（保留原方法供其他方式注册）
func (uc *UserController) Register(c *gin.Context) {
	// 创建用户结构体并绑定JSON数据
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	println(user.PasswordHash)
	println("demode")
	// 创建用户
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 生成JWT token
	token, err := middleware.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "data": user, "token": token, "code": 200})
}

// Login 用户登录
func (uc *UserController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 查找用户
	var user model.User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		return
	}
	// 验证密码
	if user.PasswordHash != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	// 生成JWT token
	token, err := middleware.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	// 更新最后登录时间
	DB.Model(&user).Update("last_login_time", time.Now())
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully", "data": user, "token": token, "code": 200})
}

// LoginE 用户邮件密码登录
func (uc *UserController) LoginE(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱未注册"})
		return
	}

	// 密码或验证码验证
	if req.Password != "" {
		if user.PasswordHash != req.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
			return
		}
	} else if req.Code != "" {
		// 这里添加验证码校验逻辑
		if req.Code != "123456" { // 示例验证码
			c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码错误"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要密码或验证码"})
		return
	}

	token, err := middleware.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌生成失败"})
		return
	}

	DB.Model(&user).Update("last_login_time", time.Now())
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data":    user,
		"token":   token,
	})
}

// 邮件验证码登录
func (uc *UserController) LoginEW(c *gin.Context) {

	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"msg":"数据传输有误","code":400})
		return
	}
	
	// 检查验证码是否有效
	storedCode, err := RedisClient.Get(c, "verification:"+req.Email).Result() // 检查验证码是否有效
	if err != nil || storedCode != req.Code { // 检查验证码是否有效
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码无效或已过期","code":400})
		return
	}


	var user model.User
	if err := DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱未注册","code":400})
		return
	}

	RedisClient.Del(c, "verification:"+req.Email)
	token, _ := middleware.GenerateToken(user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"message":  "登录成功",
		// "password": user.Password,
		"data":   user,
		"token": token,
		"code": 200,
	})

}

// GetUserProfile 获取用户资料
func (uc *UserController) GetUserProfile(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	cacheKey := fmt.Sprintf("user:%s", id)
	cachedData, err := RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var user model.User
		if json.Unmarshal([]byte(cachedData), &user) == nil {
			c.JSON(http.StatusOK, gin.H{"data": user, "source": "cache"})
			return
		}
	}

	var user model.User
	if err := DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if jsonData, err := json.Marshal(user); err == nil {
		RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{"data": user, "source": "database"})
}

// UpdateUserProfile 更新用户资料
func (uc *UserController) UpdateUserProfile(c *gin.Context) {
	type UpdateUser struct {
		Username  string `json:"username" binding:"required,min=3,max=20"`
		UserID    uint   `json:"user_id" binding:"required"`
		AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
		// Role                 string `json:"role" binding:"required,oneof=user admin moderator"`
		PersonalSignature    string `json:"personal_signature" gorm:"size:255"`
		PersonalIntroduction string `json:"personal_introduction" gorm:"size:255"`
	}
	var updateUser UpdateUser
	var user model.User
	id := c.Param("id")
	// 转化为数字类型
	idInt, _ := strconv.Atoi(id)

	// 查找用户
	if err := DB.First(&user, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有这个用户", "err": err})
		return
	}

	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}
	// if !strings.Contains(user.Email, "@") {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式无效"})
	// 	return
	// }
	// 字段映射和验证
	if updateUser.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改其他用户资料"})
		return
	}

	// // 检查用户名唯一性
	// if user.Username != updateUser.Username {
	// 	var existingUser model.User
	// 	if err := DB.Where("username = ?", updateUser.Username).First(&existingUser).Error; err == nil {
	// 		c.JSON(http.StatusConflict, gin.H{"error": err, "message": "用户名已存在"})
	// 		return
	// 	}
	// }

	// 更新用户信息
	user.Username = updateUser.Username
	user.AvatarURL = updateUser.AvatarURL
	// user.Role = updateUser.Role
	user.PersonalSignature = updateUser.PersonalSignature
	user.PersonalIntroduction = updateUser.PersonalIntroduction

	fmt.Printf("this is user:%v", user)
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "更新失败"})
		return
	}

	// 清除缓存
	RedisClient.Del(c, fmt.Sprintf("user:%d", user.ID))
	c.JSON(http.StatusOK, gin.H{"message": "User profile updated successfully", "data": user})
}

func (uc *UserController) CheckAuthStatus(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"isLoggedIn": false, "error": "用户未登录"})
		return
	}

	currentUser, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"isLoggedIn": false, "error": "用户信息解析错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isLoggedIn":    true,
		"userID":        currentUser.ID,
		"username":      currentUser.Username,
		"avatar":        currentUser.AvatarURL,
		"role":          currentUser.Role,
		"lastLoginTime": currentUser.LastLoginTime.Format(time.RFC3339),
	})
}

// FollowUser 关注用户
func (uc *UserController) FollowUser(c *gin.Context) {
	// currentUser, _ := c.Get("user")
	// current := currentUser.(model.User)
	targetId, err := strconv.Atoi(c.Param("targetId")) // 关注的用户
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的关注用户ID","code":400})
		return
	}

	Originaluser, err := strconv.Atoi(c.Param("Originaluser")) // 这是原用户
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID","code":400})
		return
	}

	var targetUser model.User
	if err := DB.First(&targetUser, targetId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在","code":400})
		return
	}

	// if err := current.Follow(DB, uint(targetID)); err != nil {
	// 	c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	// 	return
	// }             a关注者，b被关注者
	if err := model.FollowY(DB, uint(Originaluser), uint(targetId)); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(),"code":400})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "关注成功",
		"code":    200,
		"data": gin.H{
			"Originaluser": Originaluser, // 关注者ID
			"targetId":     targetId,     // 被关注者ID
		},
	})
}

// UnfollowUser 取消关注
func (uc *UserController) UnfollowUser(c *gin.Context) {
	// currentUser, _ := c.Get("user")
	// current := currentUser.(model.User)

	// targetID, err := strconv.Atoi(c.Param("id"))
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
	// 	return
	// }

	// if err := current.Unfollow(DB, uint(targetID)); err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	// 	return
	// }
	targetId, err := strconv.Atoi(c.Param("targetId")) // 关注的用户
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的关注用户ID","code":400})
		return
	}

	Originaluser, err := strconv.Atoi(c.Param("Originaluser")) // 这是原用户
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID","code":400})
		return
	}

	var targetUser model.User
	if err := DB.First(&targetUser, targetId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在","code":400})
		return
	}

	// if err := current.Follow(DB, uint(targetID)); err != nil {
	// 	c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	// 	return
	// }             a关注者，b被取消关注者
	if err := model.UnfollowY(DB, uint(Originaluser), uint(targetId)); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(),"code":400})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消关注成功",
		"code":    200,
		"data": gin.H{
			"Originaluser": Originaluser, // 关注者ID
			"targetId":     targetId,     // 被关注者ID
		},
	})
}

// CheckMutualFollow 验证双向关注关系
func (uc *UserController) CheckMutualFollow(c *gin.Context) {
	userAID, _ := strconv.Atoi(c.Query("userA"))
	userBID, _ := strconv.Atoi(c.Query("userB"))

	var aFollowsB bool
	var bFollowsA bool

	// 检查A是否关注B
	DB.Raw("SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = ? AND followed_id = ?)", userAID, userBID).Scan(&aFollowsB)
	// 检查B是否关注A
	DB.Raw("SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = ? AND followed_id = ?)", userBID, userAID).Scan(&bFollowsA)

	statusCode := 0
	statusDesc := ""

	switch {
	case aFollowsB && bFollowsA:
		statusCode = 1
		statusDesc = "互相关注"
	case aFollowsB && !bFollowsA:
		statusCode = 2
		statusDesc = "A单方面关注B"
	case !aFollowsB && bFollowsA:
		statusCode = 3
		statusDesc = "B单方面关注A"
	default:
		statusCode = 0
		statusDesc = "无关注关系"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":        statusCode,
			"description": statusDesc,
			"details": gin.H{
				"a_follows_b": aFollowsB,
				"b_follows_a": bFollowsA,
			},
		},
	})
}

// 新增私有登录方法
func loginByEmail(email, password, code string) (*model.User, error) {
	var user model.User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("邮箱未注册")
	}

	if password != "" {
		if user.PasswordHash != password {
			return nil, fmt.Errorf("密码错误")
		}
	} else if code != "" {
		storedCode, err := RedisClient.Get(context.Background(), "verification:"+email).Result()
		if err != nil || storedCode != code {
			return nil, fmt.Errorf("验证码错误或已过期")
		}
	} else {
		return nil, fmt.Errorf("需要密码或验证码")
	}
	return &user, nil
}

// GetFollowingList 获取用户关注列表
func (uc *UserController) GetFollowingList(c *gin.Context) {
	targetID := c.Param("targetId")

	var follows []model.UserFollow
	if err := DB.Preload("Followed").Where("follower_id = ?", targetID).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	response := make([]gin.H, 0)     // 创建一个空的响应数组
	for _, follow := range follows { // 遍历关注列表
		response = append(response, gin.H{ // 将关注信息添加到响应数组中
			"id": follow.Followed.ID,
			// "username":    follow.Followed.Username,
			// "avatar":      follow.Followed.AvatarURL,
			"followed_at": follow.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"list":  response,
			"total": len(response),
		},
	})
}

// AdminLogin 管理员登录
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"` // 密码登录时需要
	Code     string `json:"code"`     // 验证码登录时需要
}

func AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 复用普通登录验证逻辑
	user, err := loginByEmail(req.Email, req.Password, req.Code)
	if err != nil {
		c.JSON(401, gin.H{"error": "身份验证失败"})
		return
	}

	// 检查管理员权限
	if user.Role != "root" {
		c.JSON(403, gin.H{"error": "权限不足"})
		return
	}

	// 生成管理员专属token
	token, err := middleware.GenerateAdminToken(user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "令牌生成失败"})
		return
	}

	// 返回过滤后的用户信息
	c.JSON(200, gin.H{
		"token": token,
		"user":  filterAdminUserInfo(user),
	})
}
// 这是一个辅助函数，用于过滤用户信息
func filterAdminUserInfo(user *model.User) interface{} {
	return struct {
		ID        uint   `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Role      string `json:"role"`
	}{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
	}
}

// 获取所有用户
func (uc *UserController) GetAllUsers(c *gin.Context) {
	var users []model.User
	if result := DB.Find(&users); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error(),"msg":"查询失败","code":400})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "查询成功", "data": users, "code": 200})
}



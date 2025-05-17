package controller

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GroupController struct {
}

// 获取指定帖子下面的所有评论
func (g *GroupController) GetGroupComment(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效", "code": "500", "message": "请求参数错误"})
		return
	}
	var comments []model.GroupComment
	if err := DB.Where("post_id =? AND parent_id IS NULL", postID).Order("created_at ASC").Preload("User").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	// 递归加载子评论
	for i := range comments {
		loadReplies(&comments[i])
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// 递归加载子评论
func loadReplies(comment *model.GroupComment) {
	var replies []model.GroupComment
	DB.Where("parent_id = ?", comment.ID).Order("created_at ASC").Preload("User").Find(&replies)
	for i := range replies {
		loadReplies(&replies[i])
	}
	comment.Replies = replies
}

// batchLike 批量点赞
func (g *GroupController) BatchLike(c *gin.Context) {
	useridstr := c.Param("userid")
	userid, err1 := strconv.Atoi(useridstr)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "code": "500", "message": "请求参数错误"})
		return
	}
	var req struct {
		Actions []struct {
			PostID string `json:"postId"`
			Action string `json:"action"`
		} `json:"actions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "code": "500", "message": "请求参数错误"})
		return
	}
	userModel := model.User{}
	if err := DB.Where("id =?", userid).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	results := make([]gin.H, 0, len(req.Actions))
	for _, act := range req.Actions {
		postID, err := strconv.Atoi(act.PostID)
		if err != nil {
			results = append(results, gin.H{"postId": act.PostID, "status": "fail", "message": "postId无效"})
			continue
		}
		if act.Action == "like" {
			var existingLike model.GroupLike
			errLike := DB.Where("user_id = ? AND post_id = ?", userModel.ID, postID).First(&existingLike).Error
			if errLike == nil {
				// 已经点过赞，直接返回成功，幂等
				results = append(results, gin.H{"postId": act.PostID, "status": "success", "action": "like"})
				continue
			}
			like := model.GroupLike{UserID: userModel.ID, PostID: uint(postID)}
			err := DB.Create(&like).Error
			if err == nil {
				DB.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + 1"))
				results = append(results, gin.H{"postId": act.PostID, "status": "success", "action": "like"})
			} else {
				results = append(results, gin.H{"postId": act.PostID, "status": "fail", "message": "点赞失败"})
			}
		} else if act.Action == "unlike" {
			var like model.GroupLike
			errLike := DB.Where("user_id = ? AND post_id = ?", userModel.ID, postID).First(&like).Error
			if errLike != nil {
				// 没有点赞记录，直接返回成功，幂等
				results = append(results, gin.H{"postId": act.PostID, "status": "success", "action": "unlike"})
				continue
			}
			err := DB.Where("user_id = ? AND post_id = ?", userModel.ID, postID).Delete(&model.GroupLike{}).Error
			if err == nil {
				DB.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
				results = append(results, gin.H{"postId": act.PostID, "status": "success", "action": "unlike"})
			} else {
				results = append(results, gin.H{"postId": act.PostID, "status": "fail", "message": "取消点赞失败"})
			}
		} else {
			results = append(results, gin.H{"postId": act.PostID, "status": "fail", "message": "未知操作"})
		}
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "code": "200"})
}

// GetUserPostIsLike 检查用户是否点赞了帖子
func (g *GroupController) GetUserPostIsLike(c *gin.Context) {
	userID := c.GetUint("userID")
	postID := c.Param("postID")
	if userID == 0 || postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "code": "500", "message": "请求参数错误"})
		return
	}
	var like model.GroupLike
	if err := DB.Where("user_id =? AND post_id =?", userID, postID).First(&like).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "未点赞", "code": "400"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "已点赞", "code": "200"})
}

// 创建圈子帖子
func (g *GroupController) CreateGroupPost(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err1 := strconv.Atoi(groupIDStr)
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圈子ID无效"})
		return
	}
	var req struct {
		Content string   `json:"content"`
		Images  []string `json:"images"`
		UserID  uint     `json:"userid"`
	}
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2, "code": "500", "message": "请求参数错误"})
		return
	}

	userModel := model.User{}
	if err := DB.Where("id =?", req.UserID).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	imagesJson, err := json.Marshal(req.Images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图片序列化失败", "code": "500", "message": "创建失败"})
		return
	}

	post := model.Post{
		Content: req.Content,
		Images:  string(imagesJson),
		GroupID: uint(groupID),
		UserID:  userModel.ID,
	}
	if err := DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "code": "500", "message": "创建失败"})
		return
	}
	// 查询创建的帖子和评论进行返回
	if err := DB.Preload("Author").First(&post, post.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post, "code": "200", "message": "创建成功"})
}

// 获取圈子帖子列表
func (g *GroupController) GetGroupPosts(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圈子ID无效", "code": "500", "message": "请求参数错误"})
		return
	}
	userIDStr := c.Param("userid")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效", "code": "500", "message": "请求参数错误"})
		return
	}
	var posts []model.Post
	if err := DB.Where("group_id = ? AND status = ?", groupID, "published").Order("is_top DESC, hot_score DESC, created_at DESC").Preload("Author").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	// 遍历帖子，检查用户是否点赞
	for i := range posts {
		var like model.GroupLike
		// 如果点赞了，就把IsLiked设置为true
		if err := DB.Where("user_id =? AND post_id =?", userID, groupID).First(&like).Error; err == nil {
			posts[i].IsLiked = true
		} else {
			posts[i].IsLiked = false
		}
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// 获取最新的10条帖子
func (g *GroupController) GetLatestPosts(c *gin.Context) {
	var posts []model.Post
	if err := DB.Where("status = ?", "published").Order("created_at DESC").Limit(10).Preload("Group").Preload("Author").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// 获取帖子详情（含评论）
func (g *GroupController) GetPostDetail(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效, code: 500, message: 请求参数错误"})
		return
	}

	var post model.Post
	if err := DB.Preload("Author").Preload("GroupComments", func(db *gorm.DB) *gorm.DB {
		return db.Where("parent_id IS NULL").Order("created_at ASC")
	}).Preload("GroupComments.User").First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	// 加载评论的子评论
	for i := range post.GroupComments {
		DB.Model(&post.GroupComments[i]).Preload("User").Preload("Replies.User").Find(&post.GroupComments[i].Replies, "parent_id = ?", post.GroupComments[i].ID)
	}
	// 遍历Post
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// 点赞帖子
func (g *GroupController) LikePost(c *gin.Context) {
	postIDStr := c.Param("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效", "code": "500", "message": "请求参数错误"})
		return
	}
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
	// 	return
	// }
	// userModel := user.(model.User)
	// 下面这一套专门用来替换通过JWT验证的代码的。
	var req struct {
		PostID uint `json:"postid"`
		UserID uint `json:"userid"`
	}
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2, "code": "500", "message": "请求参数错误"})
		return
	}
	userModel := model.User{}
	if err := DB.Where("id =?", req.UserID).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	like := model.GroupLike{UserID: userModel.ID, PostID: uint(postID)}
	if err := DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败", "code": "500", "message": "点赞失败"})
		return
	}
	DB.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	c.JSON(http.StatusOK, gin.H{"message": "点赞成功", "code": "200"})
}

// 取消点赞帖子
func (g *GroupController) UnlikePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效", "code": "500", "message": "请求参数错误"})
		return
	}
	var req struct {
		PostID uint `json:"postid"`
		UserID uint `json:"userid"`
	}
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2, "code": "500", "message": "请求参数错误"})
		return
	}
	userModel := model.User{}
	if err := DB.Where("id =?", req.UserID).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	var like model.GroupLike
	errLike := DB.Where("user_id = ? AND post_id = ?", userModel.ID, postID).First(&like).Error
	if errLike != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未点赞", "code": "400", "message": "未点赞"})
		return
	}
	if err := DB.Where("user_id = ? AND post_id = ?", userModel.ID, postID).Delete(&model.GroupLike{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败", "code": "500", "message": "取消点赞失败"})
		return
	}
	DB.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	c.JSON(http.StatusOK, gin.H{"message": "已取消点赞", "code": "200"})
}

// 发表评论
func (g *GroupController) CommentPost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效"})
		return
	}
	var req struct {
		Content  string `json:"content"`
		ParentID uint   `json:"parent_id"`
		UserID   uint   `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var parentIDPtr *uint
	var depth uint = 0
	treePath := ""
	if req.ParentID != 0 {
		parentIDPtr = &req.ParentID
		// 查询父评论，获取其depth和tree_path
		var parentComment model.GroupComment
		if err := DB.First(&parentComment, req.ParentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err, "code": "400", "message": "父评论不存在"})
			return
		}
		depth = parentComment.Depth + 1
		if parentComment.TreePath != "" {
			treePath = parentComment.TreePath + "/" + strconv.FormatUint(uint64(parentComment.ID), 10)
		} else {
			treePath = strconv.FormatUint(uint64(parentComment.ID), 10)
		}
	} else {
		parentIDPtr = nil
		depth = 0
		treePath = ""
	}
	// 验证用户是否存在
	userModel := model.User{}
	if err := DB.Where("id =?", req.UserID).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err, "code": "400", "message": "用户不存在"})
		return
	}

	// 验证帖子是否存在
	var post model.Post
	if err := DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err, "code": "400", "message": "帖子不存在"})
		return
	}

	comment := model.GroupComment{
		Content:   req.Content,
		UserID:    req.UserID,
		PostID:    uint(postID),
		ParentID:  parentIDPtr,
		Depth:     depth,
		TreePath:  treePath,
		LikeCount: 0,
		IsDeleted: false,
	}
	if err := DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "code": "500", "message": "评论失败"})
		return
	}
	// 更新帖子评论数
	DB.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	// 查询评论并返回
	if err := DB.Preload("User").Preload("Post").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err, "code": "400", "message": "评论不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment, "code": "200", "message": "评论成功"})
}

// 删除帖子（管理员）
func (g *GroupController) DeletePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID无效"})
		return
	}
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
	// 	return
	// }
	// userModel := user.(model.User)
	// if userModel.Role != "root" {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
	// 	return
	// }
	if err := DB.Model(&model.Post{}).Where("id = ?", postID).Update("status", "deleted").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err, "code": "500", "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "code": "200"})
}

// 用户加入圈子
func (g *GroupController) JoinGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圈子ID无效"})
		return
	}
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userModel := user.(model.User)
	var userGroup model.UserGroup
	err = DB.Where("user_id = ? AND group_id = ?", userModel.ID, groupID).First(&userGroup).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已加入该圈子"})
		return
	}
	userGroup = model.UserGroup{
		UserID:   userModel.ID,
		GroupID:  uint(groupID),
		JoinedAt: time.Now(),
		Role:     "member",
		Status:   "approved",
	}
	if err := DB.Create(&userGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加入失败"})
		return
	}
	DB.Model(&model.GroupN{}).Where("id = ?", groupID).UpdateColumn("member_count", gorm.Expr("member_count + 1"))
	c.JSON(http.StatusOK, gin.H{"message": "加入成功"})
}

// 用户退出圈子
func (g *GroupController) QuitGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圈子ID无效"})
		return
	}
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userModel := user.(model.User)
	if err := DB.Where("user_id = ? AND group_id = ?", userModel.ID, groupID).Delete(&model.UserGroup{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "退出失败"})
		return
	}
	DB.Model(&model.GroupN{}).Where("id = ?", groupID).UpdateColumn("member_count", gorm.Expr("GREATEST(member_count - 1, 0)"))
	c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
}

// 获取用户加入的圈子列表
func (g *GroupController) GetUserJoinedGroups(c *gin.Context) {
	userIDStr := c.Param("userid")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效"})
		return
	}
	type result struct {
		groups []model.GroupN
		err    error
	}
	ch1 := make(chan result)
	ch2 := make(chan result)
	// 查询用户加入的圈子
	go func() {
		var joinedGroups []model.GroupN
		err := DB.Joins("JOIN user_groups ON user_groups.group_id = group_ns.id").Where("user_groups.user_id = ? AND group_ns.status != ?", userID, "banned").Find(&joinedGroups).Error
		ch1 <- result{groups: joinedGroups, err: err}
	}()
	// 查询用户创建的圈子
	go func() {
		var createdGroups []model.GroupN
		err := DB.Where("creator_id = ? AND status != ?", userID, "banned").Find(&createdGroups).Error
		ch2 <- result{groups: createdGroups, err: err}
	}()
	res1 := <-ch1
	res2 := <-ch2
	if res1.err != nil && res2.err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	groupMap := make(map[uint]model.GroupN)
	for _, g := range res1.groups {
		groupMap[g.ID] = g
	}
	for _, g := range res2.groups {
		groupMap[g.ID] = g
	}
	mergedGroups := make([]model.GroupN, 0, len(groupMap))
	for _, g := range groupMap {
		mergedGroups = append(mergedGroups, g)
	}
	c.JSON(http.StatusOK, gin.H{"groups": mergedGroups})
}

// 获取用户创建的圈子列表
func (g *GroupController) GetUserCreatedGroups(c *gin.Context) {
	userIDStr := c.Param("userid")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效"})
		return
	}
	var groups []model.GroupN
	if err := DB.Where("creator_id = ?", userID).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// 创建圈子
func (g *GroupController) CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`        // 圈子名称
		Description string `json:"description"` // 圈子描述
		AvatarURL   string `json:"avatar"`      // 圈子头像
		CreatorID   uint   `json:"creator_id"`  // 创建者ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	// token处理的方法
	// userid, exists := c.Get("userid")
	// userID := userid.(uint)
	// user := model.User{}
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
	// 	return
	// }

	userID := req.CreatorID
	// userID, err := strconv.Atoi(userIDstring)
	// if err!= nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效","code": 400})
	// 	return
	// }
	user := model.User{}
	// 查询用户 是否存在
	if err := DB.Where("id =?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "创建圈子的用户不存在", "code": 400})
		return
	}

	group := model.GroupN{
		Name:        req.Name,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		CreatorID:   user.ID,
	}
	if err := DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建圈子失败", "code": 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"group": group, "code": 200})
}

// 获取所有圈子
func (g *GroupController) GetAllGroups(c *gin.Context) {
	var groups []model.GroupN
	// 链表查询
	if err := DB.Preload("Creator").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取圈子失败", "code": 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups, "code": 200})
}

// 封禁圈子（管理员）
func (g *GroupController) BanGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	type req struct {
		Status string `json:"status"` // 封禁状态
		Reason string `json:"reason"` // 封禁理由
		UserID uint   `json:"userID"` // 封禁者ID
	}
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "圈子ID无效"})
		return
	}
	// 解析请求体
	var request req
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "code": 400})
		return
	}
	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
	// 	return
	// }
	// userModel := user.(model.User)
	userModel := model.User{}
	// 查询用户 是否存在
	if err := DB.Where("id =?", request.UserID).First(&userModel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "创建圈子的用户不存在", "code": 400})
		return
	}
	// 检查用户是否为管理员
	if userModel.Role != "root" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	// 封禁圈子
	if err := DB.Model(&model.GroupN{}).Where("id = ?", groupID).Update("status", "banned").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁失败", "code": 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "封禁成功", "code": 200})
}

// 获取用户关注的用户发布的圈子帖子
func (g *GroupController) GetFollowedUsersGroupPosts(c *gin.Context) {
	userIDStr := c.Param("userid")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效"})
		return
	}
	// 查询用户关注的用户ID列表
	var followedIDs []uint
	if err := DB.Table("follows").Where("follower_id = ?", userID).Pluck("followed_id", &followedIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注列表失败"})
		return
	}
	if len(followedIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"posts": []model.Post{}})
		return
	}
	// 查询这些用户发布的圈子帖子
	var posts []model.Post
	if err := DB.Where("user_id IN ? AND status = ?", followedIDs, "published").Order("created_at DESC").Preload("Group").Preload("Author").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取帖子失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// 获取点赞量最多的10条圈子帖子
func (g *GroupController) GetTopLikedGroupPosts(c *gin.Context) {
	var posts []model.Post
	if err := DB.Where("status = ?", "published").Order("like_count DESC").Limit(10).Preload("Group").Preload("Author").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// 检查用户是否加入或创建了指定圈子
func (g *GroupController) CheckUserInGroup(c *gin.Context) {
	groupIDStr := c.Query("groupid")
	userIDStr := c.Query("userid")
	groupID, err1 := strconv.Atoi(groupIDStr)
	userID, err2 := strconv.Atoi(userIDStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	// 检查是否为圈子创建者
	var group model.GroupN
	err := DB.Where("id = ? AND creator_id = ?", groupID, userID).First(&group).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"joined": true, "role": "creator"})
		return
	}
	// 检查是否为圈子成员
	var userGroup model.UserGroup
	err = DB.Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, "approved").First(&userGroup).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"joined": true, "role": userGroup.Role})
		return
	}
	c.JSON(http.StatusOK, gin.H{"joined": false})
}

// 获取用户未加入且未创建的随机两个圈子
func (g *GroupController) GetRandomUnjoinedUncreatedGroups(c *gin.Context) {
	fmt.Println("触发了")
	userIDStr := c.Param("userid")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效"})
		return
	}
	// 查询用户已加入的圈子ID
	var joinedGroupIDs []uint
	DB.Table("user_groups").Where("user_id = ?", userID).Pluck("group_id", &joinedGroupIDs)
	// 查询用户已创建的圈子ID
	var createdGroupIDs []uint
	DB.Table("group_ns").Where("creator_id = ?", userID).Pluck("id", &createdGroupIDs)
	// 合并排除的圈子ID
	excludeIDs := append(joinedGroupIDs, createdGroupIDs...)
	fmt.Println("触发了")
	var groups []model.GroupN
	query := DB.Model(&model.GroupN{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	// 随机排序并限制返回数量，提升效率
	if err := query.Order("RAND()").Limit(2).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups, "code": 200})
}

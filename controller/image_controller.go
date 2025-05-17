package controller

import (
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageController struct{}

// UploadImage 处理图片上传
func (ctrl *ImageController) UploadImage(c *gin.Context) {
	// 1. 验证文件大小（限制5MB）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)

	// 2. 获取上传文件
	file, err := c.FormFile("image") // 假设前端表单字段名为"image"
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}

	// 3. 验证文件类型
	extType := mime.TypeByExtension(filepath.Ext(file.Filename))
	if extType != "image/jpeg" && extType != "image/png" && extType != "image/gif" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "仅支持JPEG/PNG/GIF格式"})
		return
	}

	// 4. 创建存储目录
	uploadPath := "./uploads/images"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "无法创建存储目录"})
		return
	}

	// 5. 生成唯一文件名
	rand.Seed(time.Now().UnixNano())
	newFilename := fmt.Sprintf("%d_%s", rand.Intn(999999), file.Filename)
	dstPath := filepath.Join(uploadPath, newFilename)

	// 6. 保存文件
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("/images/%s", newFilename),
	})
}

func (ctrl *ImageController) DeleteImage(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件名参数"})
		return
	}
	filePath := filepath.Join("./uploads/images", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

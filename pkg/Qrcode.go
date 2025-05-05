package pkg
// // 安装依赖
// // go get github.com/skip2/go-qrcode
// // go get github.com/fogleman/gg
// import (
// 	"github.com/fogleman/gg"
// 	"github.com/skip2/go-qrcode"
// )

// func GenerateShareImage(data any) {
// 	articleID := data
// 	// 1. 获取文章数据
// 	article := GetArticleFromDB(articleID)
  
// 	// 2. 生成二维码
// 	qrCode, _ := qrcode.New(article.URL, qrcode.Medium)
// 	qrImg := qrCode.Image(256)
  
// 	// 3. 合成图片
// 	const width, height = 600, 800
// 	dc := gg.NewContext(width, height)
	
// 	// 背景
// 	dc.SetRGB(1, 1, 1)
// 	dc.Clear()
  
// 	// 添加文字
// 	dc.SetRGB(0, 0, 0)
// 	dc.LoadFontFace("fonts/arial.ttf", 24)
// 	dc.DrawStringWrapped(article.Title, 50, 50, 0, 0, 500, 1.5, gg.AlignLeft)
  
// 	// 插入二维码
// 	dc.DrawImage(qrImg, 50, 200)
  
// 	// 返回图片
// 	c.Header("Content-Type", "image/png")
// 	dc.EncodePNG(c.Writer)
//   }
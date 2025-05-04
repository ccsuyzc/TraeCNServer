package pkg

import (
	"fmt"
	"net/smtp"
	"strings"
	"github.com/jordan-wright/email"
)

type EmailPkg struct{}


func (e *EmailPkg) SendEmail(to,smg,EmailSubject string,) error {
	// targetMailBox := ""          // 目标邮箱
	smtpServer := "smtp.qq.com" // smtp服务器
	emailAddr := "chatyang@foxmail.com"              // 要发件的邮箱地址
	smtpKey := "crgtzjcatrcldcbe"                    // 获取的smtp密钥
    // 检查to是不是邮件地址，则返回错误
	if !strings.Contains(to, "@") {
		return fmt.Errorf("invalid email address: %s", to)
	}

	em := email.NewEmail()

	em.From = fmt.Sprintf("Go论坛博客网 <%s>", emailAddr) // 发件人
	em.To = []string{to}                        // 目标邮箱

	// email title
	em.Subject = EmailSubject // 标题
	// build email content
	em.Text = []byte(smg) // 内容

	// 调用接口发送邮件
    // 此处端口号不一定为25使用对应邮箱时需要具体更换
	em.Send(smtpServer+":587", smtp.PlainAuth("", emailAddr, smtpKey, smtpServer))
	return nil
}


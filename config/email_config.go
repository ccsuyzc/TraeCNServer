package config

type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	// 秘钥
	smtpKey      string
	SMTPFrom     string
}

func LoadEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPServer:   "smtp.qq.com",          // SMTP服务器地址
		SMTPPort:     587,                    // SMTP服务器端口
		smtpKey:      "crgtzjcatrcldcbe",     // SMTP服务器密钥
		SMTPFrom:     "chatyang@foxmail.com", // 发件人邮箱地址
	}
}

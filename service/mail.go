package service

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/tigerowo/infinite-canvas/config"
)

func smtpConfigured() bool {
	return strings.TrimSpace(config.Cfg.SMTPHost) != "" && strings.TrimSpace(config.Cfg.SMTPFrom) != ""
}

func sendAuthMail(to string, subject string, body string) error {
	if !smtpConfigured() {
		return safeMessageError{message: "邮件服务未配置"}
	}
	from := config.Cfg.SMTPFrom
	envelopeFrom := from
	if address, err := mail.ParseAddress(from); err == nil {
		envelopeFrom = address.Address
	}
	auth := smtp.Auth(nil)
	if strings.TrimSpace(config.Cfg.SMTPUser) != "" {
		auth = smtp.PlainAuth("", config.Cfg.SMTPUser, config.Cfg.SMTPPass, config.Cfg.SMTPHost)
	}
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	if strings.TrimSpace(config.Cfg.SMTPReplyTo) != "" {
		headers = append(headers, "Reply-To: "+config.Cfg.SMTPReplyTo)
	}
	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + body
	return smtp.SendMail(config.Cfg.SMTPHost+":"+strconv.Itoa(config.Cfg.SMTPPort), auth, envelopeFrom, []string{to}, []byte(message))
}

func authMailBody(greeting string, link string, action string) string {
	return fmt.Sprintf("你好，%s：\n\n请点击下面的链接%s：\n%s\n\n如果这不是你的操作，请忽略这封邮件。", greeting, action, link)
}

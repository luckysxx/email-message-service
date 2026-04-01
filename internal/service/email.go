package service

import (
	"context"
	"fmt"

	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/email-message/config"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// Sender 是邮件发送的通用接口
type Sender interface {
	SendWelcomeEmail(ctx context.Context, toEmail, username string) error
}

// smtpSender 是 Sender 接口的 SMTP 实现
type smtpSender struct {
	dialer *gomail.Dialer
	from   string
	logger *zap.Logger
}

// NewSMTPSender 是一个构造函数，包含配置项和结构化日志对象的注入
func NewSMTPSender(cfg config.SMTPConfig, logger *zap.Logger) Sender {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.SSL = cfg.SSL
	return &smtpSender{
		dialer: dialer,
		from:   cfg.From,
		logger: logger.With(
			zap.String("component", "smtp_sender"),
			zap.String("smtp_host", cfg.Host),
			zap.Int("smtp_port", cfg.Port),
			zap.Bool("smtp_ssl", cfg.SSL),
		),
	}
}

// SendWelcomeEmail 实现了接口方法
func (s *smtpSender) SendWelcomeEmail(ctx context.Context, toEmail, username string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "欢迎加入我们的平台！")

	body := "<h1>你好，" + username + "！</h1><p>感谢你注册我们的平台，快来体验吧。</p>"
	m.SetBody("text/html", body)

	commonlogger.Ctx(ctx, s.logger).Info("正在连接 SMTP 服务器发送欢迎邮件",
		zap.String("target_email", toEmail),
	)

	// 发送邮件并包装错误信息
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to dial and send email to %s: %w", toEmail, err)
	}

	return nil
}

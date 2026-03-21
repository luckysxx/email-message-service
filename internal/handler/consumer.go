package handler

import (
	"context"
	"encoding/json"

	"github.com/luckysxx/common/trace"
	"github.com/luckysxx/email-message/config"
	"github.com/luckysxx/email-message/internal/event"
	"github.com/luckysxx/email-message/internal/service"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EmailConsumer 负责处理和邮件相关的 Kafka 消息
type EmailConsumer struct {
	reader *kafka.Reader
	sender service.Sender // 依赖注入邮件发送器
	logger *zap.Logger    // 结构化组件日志
}

// NewEmailConsumer 采用依赖注入方式构建消费者
func NewEmailConsumer(cfg config.KafkaConfig, sender service.Sender, logger *zap.Logger) *EmailConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
		Topic:   cfg.Topic,
	})

	return &EmailConsumer{
		reader: r,
		sender: sender,
		logger: logger.With(zap.String("component", "kafka_consumer"), zap.String("topic", cfg.Topic)),
	}
}

// Start 启动阻塞的消费循环，支持通过 context 的取消来实现优雅退出
func (c *EmailConsumer) Start(ctx context.Context) error {
	c.logger.Info("Kafka 邮件消费者已启动")
	for {
		// FetchMessage 阻塞拉取消息
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("Context 已取消，停止消费循环")
				// 业务正常退出，不返回错误
				return nil
			}
			c.logger.Error("从 Kafka 拉取消息失败", zap.Error(err))
			continue
		}

		// 解析 Trace ID
		var traceID string
		for _, h := range m.Headers {
			if h.Key == trace.HeaderTraceID {
				traceID = string(h.Value)
				break
			}
		}
		msgCtx := trace.IntoContext(ctx, traceID)

		// 解析消息
		var evt event.UserRegisteredEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			c.logger.Error("事件数据反序列化失败",
				zap.String("trace_id", traceID),
				zap.Error(err),
				zap.Int64("offset", m.Offset),
				zap.ByteString("value", m.Value),
			)
			// 规范：对于解析失败这种无法恢复的结构体错误，记录日志并跳过 Commit
			c.reader.CommitMessages(ctx, m)
			continue
		}

		// 执行发邮件业务
		if err := c.sender.SendWelcomeEmail(msgCtx, evt.Email, evt.Username); err != nil {
			// 规范：发邮件失败（可能是网络问题），依靠日志追踪，可以选择不 Commit 并重试，或送入死信队列
			c.logger.Error("发送欢迎邮件失败",
				zap.String("trace_id", traceID),
				zap.Error(err),
				zap.String("target_email", evt.Email),
			)
			continue
		}

		c.logger.Info("欢迎邮件发送成功",
			zap.String("trace_id", traceID),
			zap.String("target_email", evt.Email),
			zap.Int64("offset", m.Offset),
		)

		// 业务处理成功后，手动提交 Offset
		c.reader.CommitMessages(ctx, m)
	}
}

// Close 释放网络资源
func (c *EmailConsumer) Close() error {
	c.logger.Info("正在关闭 Kafka 消费者")
	return c.reader.Close()
}

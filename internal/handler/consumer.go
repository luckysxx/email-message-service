package handler

import (
	"context"

	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/mq/bus"
	"github.com/luckysxx/common/mq/busmiddleware"
	"github.com/luckysxx/common/mq/kafkabus"
	"github.com/luckysxx/email-message/config"
	"github.com/luckysxx/email-message/internal/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EmailConsumer 负责挂载各个业务 handler，并启动底层订阅器
type EmailConsumer struct {
	sub     bus.Subscriber
	handler bus.Handler
	logger  *zap.Logger
}

// NewEmailConsumer 构建消费者，这里底层选型可以配置或者直接用 kafkabus
func NewEmailConsumer(cfg config.KafkaConfig, redisClient *redis.Client, sender service.Sender, logger *zap.Logger) *EmailConsumer {
	// 使用我们新加的中间件抽象
	sub := kafkabus.NewSubscriber(cfg.Brokers, cfg.Topic, cfg.GroupID)
	
	return &EmailConsumer{
		sub:     sub,
		handler: busmiddleware.Chain(
			NewEmailHandler(redisClient, sender, logger),
			busmiddleware.WithTrace(),
		),
		logger:  logger.With(zap.String("component", "kafka_consumer"), zap.String("topic", cfg.Topic)),
	}
}

// Start 启动阻塞的消费循环，支持通过 context 的取消来实现优雅退出
func (c *EmailConsumer) Start(ctx context.Context) error {
	commonlogger.Ctx(ctx, c.logger).Info("邮件消费者已启动 (Transport: Kafka)")
	// 复杂的 fetch、offset 提交逻辑等已经被下沉到 kafkabus 里了
	return c.sub.Start(ctx, c.handler)
}

// Close 释放网络资源
func (c *EmailConsumer) Close() error {
	c.logger.Info("正在关闭邮件消费者")
	return c.sub.Close()
}

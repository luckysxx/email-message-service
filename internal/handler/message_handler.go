package handler

import (
	"context"
	"encoding/json"

	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/mq/bus"
	"github.com/luckysxx/email-message/internal/event"
	"github.com/luckysxx/email-message/internal/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EmailHandler 负责解析并处理欢迎邮件事件。
// 它不依赖具体 Broker，便于后续复用到 Kafka、NATS 或测试桩。
type EmailHandler struct {
	redisClient *redis.Client
	sender      service.Sender
	logger      *zap.Logger
}

func NewEmailHandler(redisClient *redis.Client, sender service.Sender, logger *zap.Logger) *EmailHandler {
	return &EmailHandler{
		redisClient: redisClient,
		sender:      sender,
		logger:      logger.With(zap.String("component", "email_handler")),
	}
}

func (h *EmailHandler) Handle(ctx context.Context, msg *bus.Message) error {
	msgCtx := ctx

	var evt event.UserRegisteredEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		commonlogger.Ctx(msgCtx, h.logger).Error("事件数据反序列化失败",
			zap.Error(err),
			zap.String("topic", msg.Topic),
			zap.ByteString("value", msg.Value),
		)
		// 结构化坏消息无法通过重试恢复，直接视为已处理。
		return nil
	}

	if evt.Version == "" {
		evt.Version = event.UserRegisteredVersion
	}
	if evt.Version != event.UserRegisteredVersion {
		commonlogger.Ctx(msgCtx, h.logger).Error("不支持的事件版本，跳过处理",
			zap.String("version", evt.Version),
			zap.String("topic", msg.Topic),
		)
		return nil
	}

	idempotentKey := "email:welcome:" + evt.Email
	isNew, err := h.redisClient.SetNX(ctx, idempotentKey, msgKeyForStore(msg), 0).Result()
	if err != nil {
		commonlogger.Ctx(msgCtx, h.logger).Error("幂等性校验: Redis 请求失败", zap.Error(err))
		return err
	}
	if !isNew {
		commonlogger.Ctx(msgCtx, h.logger).Info("重复的邮件发送请求，触发幂等跳过",
			zap.String("target_email", evt.Email),
			zap.String("topic", msg.Topic),
		)
		return nil
	}

	if err := h.sender.SendWelcomeEmail(msgCtx, evt.Email, evt.Username); err != nil {
		h.redisClient.Del(ctx, idempotentKey)
		commonlogger.Ctx(msgCtx, h.logger).Error("发送欢迎邮件失败",
			zap.Error(err),
			zap.String("target_email", evt.Email),
			zap.String("topic", msg.Topic),
		)
		return err
	}

	commonlogger.Ctx(msgCtx, h.logger).Info("欢迎邮件发送成功",
		zap.String("target_email", evt.Email),
		zap.String("topic", msg.Topic),
	)
	return nil
}

func msgKeyForStore(msg *bus.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Key != "" {
		return msg.Key
	}
	if offset, ok := msg.Metadata["offset"]; ok {
		return offset
	}
	return msg.Topic
}

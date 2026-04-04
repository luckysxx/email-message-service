package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/probe"
	"github.com/luckysxx/common/redis"
	"github.com/luckysxx/email-message/config"
	"github.com/luckysxx/email-message/internal/handler"
	"github.com/luckysxx/email-message/internal/service"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	// 1. 初始化配置
	cfg := config.LoadConfig()

	// 2. 初始化结构化日志
	log := logger.NewLogger("email-message")
	defer log.Sync()

	log.Info("邮件服务启动中...", zap.String("environment", cfg.App.Env))

	// 3. 依赖注入与组件装配
	redisClient := redis.Init(cfg.Redis, log)
	defer redisClient.Close()

	emailSender := service.NewSMTPSender(cfg.SMTP, log)
	consumer := handler.NewEmailConsumer(cfg.Kafka, redisClient, emailSender, log)

	// 4. 应用生命周期和并发管理
	ctx, cancel := context.WithCancel(context.Background())
	eg, egCtx := errgroup.WithContext(ctx)

	// 探针管理端口：/healthz, /readyz, /metrics
	probeShutdown := probe.Serve(ctx, ":9095", log,
		probe.WithRedis(redisClient),
	)
	defer probeShutdown()

	// 在 errgroup 组后台启动 Kafka 消费者
	eg.Go(func() error {
		return consumer.Start(egCtx)
	})

	// 5. 监听操作系统信号，实现主进程级别的优雅退出
	quit := make(chan os.Signal, 1)
	// 捕获 Ctrl+C 和 Docker 发出的 SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("收到停机信号", zap.String("signal", sig.String()))
		cancel() // 主动取消 ctx，通知在后台运行的消费者结束拉取
	case <-egCtx.Done():
		log.Info("服务内部 Context 被取消（可能存在不可恢复的 Worker 错误）")
	}

	// 阻塞并等待所有的后台 Goroutine 清理并安全退出
	if err := eg.Wait(); err != nil {
		log.Error("服务异常退出", zap.Error(err))
	} else {
		log.Info("所有后台任务已安全退出")
	}

	// 6. 清理底层网络与连接资源
	if err := consumer.Close(); err != nil {
		log.Error("关闭底层组件时出错", zap.Error(err))
	}

	log.Info("邮件服务已安全退出")
}

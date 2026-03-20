package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/luckysxx/common/logger"
	"github.com/luckysxx/email-message/config"
	"github.com/luckysxx/email-message/internal/handler"
	"github.com/luckysxx/email-message/internal/service"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	// 1. 初始化配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// 2. 初始化结构化日志
	log := logger.NewLogger("email-message")
	defer log.Sync()

	log.Info("Starting Email Microservice...", zap.String("environment", cfg.App.Env))

	// 3. 依赖注入与组件装配
	emailSender := service.NewSMTPSender(cfg.SMTP, log)
	consumer := handler.NewEmailConsumer(cfg.Kafka, emailSender, log)

	// 4. 应用生命周期和并发管理
	ctx, cancel := context.WithCancel(context.Background())
	eg, egCtx := errgroup.WithContext(ctx)

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
		log.Info("Received OS shutdown signal", zap.String("signal", sig.String()))
		cancel() // 主动取消 ctx，通知在后台运行的消费者结束拉取
	case <-egCtx.Done():
		log.Info("Service context canceled internally (possibly an unrecoverable worker error)")
	}

	// 阻塞并等待所有的后台 Goroutine 清理并安全退出
	if err := eg.Wait(); err != nil {
		log.Error("Service exited abnormally", zap.Error(err))
	} else {
		log.Info("All background workers exited smoothly")
	}

	// 6. 清理底层网络与连接资源
	if err := consumer.Close(); err != nil {
		log.Error("Error closing components during shutdown", zap.Error(err))
	}

	log.Info("Email Microservice stopped gracefully")
}

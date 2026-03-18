package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(env string) (*zap.Logger, error) {
	var config zap.Config

	if env == "development" || env == "dev" {
		// 开发环境：友好的控制台输出，带有颜色，不强制输出完整的、难以阅读的堆栈
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 颜色高亮
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		// 降低开发环境堆栈输出的嘈杂度
		config.DisableStacktrace = true
	} else {
		// 生产环境：结构化的 JSON 输出
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// 生产环境保留 Error 级别的堆栈追踪
		config.DisableStacktrace = false
	}

	// 统一输出到 stdout
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel), // 仅对 Error 级别记录堆栈
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

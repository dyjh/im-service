package core

import (
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
)

// ZapLogger 实现了 Kratos 的 log.Logger 接口
type ZapLogger struct {
	sugar *zap.SugaredLogger
}

// NewZapLoggerAdapter 创建一个 ZapLogger 适配器
func NewZapLoggerAdapter(zapLogger *zap.Logger) log.Logger {
	return &ZapLogger{
		sugar: zapLogger.Sugar(),
	}
}

func (z *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	// Kratos 的日志级别对应 Zap 的日志级别
	switch level {
	case log.LevelDebug:
		z.sugar.Debugw("", keyvals...)
	case log.LevelInfo:
		z.sugar.Infow("", keyvals...)
	case log.LevelWarn:
		z.sugar.Warnw("", keyvals...)
	case log.LevelError:
		z.sugar.Errorw("", keyvals...)
	case log.LevelFatal:
		z.sugar.Fatalw("", keyvals...)
	default:
		z.sugar.Infow("", keyvals...)
	}
	return nil
}

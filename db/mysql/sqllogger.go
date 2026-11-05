package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tongcheng/logs"
)

// ZapLogger 实现 gorm.logger.Interface
type ZapLogger struct {
	zapLogger *zap.Logger
	level     logger.LogLevel
	config    logger.Config
}

// New 构造函数
func New(zapLogger *zap.Logger, config logger.Config) *ZapLogger {
	return &ZapLogger{
		zapLogger: zapLogger,
		level:     config.LogLevel,
		config:    config,
	}
}

// LogMode 切换日志级别
func (l *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

// Info 普通信息
func (l *ZapLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Info {
		l.zapLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 警告
func (l *ZapLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Warn {
		l.zapLogger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 错误
func (l *ZapLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Error {
		l.zapLogger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 核心：打印 SQL
func (l *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}
	sql, rows := fc()
	cost := time.Since(begin)
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("cost", cost),
		zap.Int64("rows", rows),
		zap.String(logs.RequestIDKey, logs.Trace(ctx)),
	}
	// 慢查询单独标记
	if cost > l.config.SlowThreshold {
		fields = append(fields, zap.Bool("slow", true))
	}

	// 错误处理
	switch {
	case err != nil && l.config.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound):
		return
	case err != nil:
		fields = append(fields, zap.Error(err))
		l.zapLogger.Error("gorm sql error", fields...)
	default:
		l.zapLogger.Info("gorm sql", fields...)
	}
}

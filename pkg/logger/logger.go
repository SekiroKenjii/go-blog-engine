package logger

import (
	"sync"

	"go.uber.org/zap"
)

// ILogger interface for structured logging
type ILogger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

type Logger struct {
	zap *zap.Logger
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Default logger instance for backward compatibility
var (
	defaultLogger ILogger
	once          sync.Once
)

// GetDefaultLogger returns the default logger interface
func GetDefaultLogger() ILogger {
	once.Do(func() {
		logger, err := ConfigurableBuilder().Build()
		if err != nil {
			// Fallback to a basic logger if configuration fails
			logger, _ = DevelopmentBuilder().Build()
		}
		defaultLogger = logger
	})

	return defaultLogger
}

func Debug(msg string, fields ...zap.Field) {
	GetDefaultLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetDefaultLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetDefaultLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetDefaultLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetDefaultLogger().Fatal(msg, fields...)
}

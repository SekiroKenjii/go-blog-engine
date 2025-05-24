package logger

import (
	"os"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

func Instance() *zap.Logger {
	once.Do(func() {
		instance = createLogger()
	})

	return instance
}

func createLogger() *zap.Logger {
	encoder := createEncoder()
	writerSync := createWriterSync()
	core := zapcore.NewCore(encoder, writerSync, zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

func createEncoder() zapcore.Encoder {
	encoder := zap.NewProductionEncoderConfig()

	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder.EncodeDuration = zapcore.SecondsDurationEncoder
	encoder.EncodeCaller = zapcore.ShortCallerEncoder
	encoder.TimeKey = "time"

	return zapcore.NewJSONEncoder(encoder)
}

func createWriterSync() zapcore.WriteSyncer {
	logConf := config.Instance().Log
	hook := lumberjack.Logger{
		Filename:   logConf.FileName,
		MaxSize:    logConf.MaxFileSize,
		MaxBackups: logConf.MaxBackups,
		MaxAge:     logConf.MaxAge,
		Compress:   logConf.Compressed,
	}

	fileSyncer := zapcore.AddSync(&hook)
	consoleSyncer := zapcore.AddSync(os.Stdout)

	return zapcore.NewMultiWriteSyncer(consoleSyncer, fileSyncer)
}

func Debug(msg string, fields ...zap.Field) {
	Instance().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Instance().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Instance().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Instance().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Instance().Fatal(msg, fields...)
}

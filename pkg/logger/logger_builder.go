package logger

import (
	"io"
	"os"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerBuilder for constructing loggers with fluent interface
type LoggerBuilder struct {
	level      zapcore.Level
	encoding   string
	writers    []zapcore.WriteSyncer
	options    []zap.Option
	filename   string
	maxSize    int
	maxBackups int
	maxAge     int
	compress   bool
}

// NewLoggerBuilder creates a new logger builder with defaults
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:      zapcore.InfoLevel,
		encoding:   "json",
		writers:    make([]zapcore.WriteSyncer, 0),
		options:    make([]zap.Option, 0),
		maxSize:    100, // MB
		maxBackups: 3,
		maxAge:     28, // days
		compress:   true,
	}
}

// WithLevel sets the log level
func (b *LoggerBuilder) WithLevel(level zapcore.Level) *LoggerBuilder {
	b.level = level
	return b
}

// WithConsoleOutput adds console output
func (b *LoggerBuilder) WithConsoleOutput() *LoggerBuilder {
	b.writers = append(b.writers, zapcore.AddSync(os.Stdout))
	return b
}

// WithFileOutput adds file output with rotation
func (b *LoggerBuilder) WithFileOutput(filename string) *LoggerBuilder {
	b.filename = filename
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    b.maxSize,
		MaxBackups: b.maxBackups,
		MaxAge:     b.maxAge,
		Compress:   b.compress,
	}
	b.writers = append(b.writers, zapcore.AddSync(lumberjackLogger))
	return b
}

// WithWriter adds a custom writer
func (b *LoggerBuilder) WithWriter(writer io.Writer) *LoggerBuilder {
	b.writers = append(b.writers, zapcore.AddSync(writer))
	return b
}

// WithEncoding sets the encoding format
func (b *LoggerBuilder) WithEncoding(encoding string) *LoggerBuilder {
	b.encoding = encoding
	return b
}

// WithStackTrace adds stack traces for the given level and above
func (b *LoggerBuilder) WithStackTrace(level zapcore.Level) *LoggerBuilder {
	b.options = append(b.options, zap.AddStacktrace(level))
	return b
}

// WithRotation configures file rotation settings
func (b *LoggerBuilder) WithRotation(maxSize, maxBackups, maxAge int, compress bool) *LoggerBuilder {
	b.maxSize = maxSize
	b.maxBackups = maxBackups
	b.maxAge = maxAge
	b.compress = compress
	return b
}

// Build creates the logger
func (b *LoggerBuilder) Build() (ILogger, error) {
	if len(b.writers) == 0 {
		b.WithConsoleOutput()
	}

	encoder := b.createEncoder()
	writeSyncer := zapcore.NewMultiWriteSyncer(b.writers...)
	core := zapcore.NewCore(encoder, writeSyncer, b.level)

	zapLog := zap.New(core, b.options...)
	return &Logger{zap: zapLog}, nil
}

// createEncoder creates an encoder based on the configuration
func (b *LoggerBuilder) createEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.TimeKey = "time"

	if b.encoding == "console" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// Predefined builders for common scenarios

// ConfigurableBuilder creates a logger based on config.Instance().Log
func ConfigurableBuilder() *LoggerBuilder {
	logConf := config.Instance().Log

	// Parse log level from config
	var level zapcore.Level
	switch logConf.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	return NewLoggerBuilder().
		WithLevel(level).
		WithEncoding("json").
		WithConsoleOutput().
		WithFileOutput(logConf.FileName).
		WithRotation(logConf.MaxFileSize, logConf.MaxBackups, logConf.MaxAge, logConf.Compressed).
		WithStackTrace(zapcore.ErrorLevel).
		WithCaller()
}

// ProductionBuilder creates a production logger
func ProductionBuilder(filename string) *LoggerBuilder {
	return NewLoggerBuilder().
		WithLevel(zapcore.InfoLevel).
		WithEncoding("json").
		WithConsoleOutput().
		WithFileOutput(filename).
		WithStackTrace(zapcore.ErrorLevel).
		WithCaller()
}

// DevelopmentBuilder creates a development logger
func DevelopmentBuilder() *LoggerBuilder {
	return NewLoggerBuilder().
		WithLevel(zapcore.DebugLevel).
		WithEncoding("console").
		WithConsoleOutput().
		WithStackTrace(zapcore.WarnLevel).
		WithCaller()
}

// TestBuilder creates a test logger
func TestBuilder(writer io.Writer) *LoggerBuilder {
	return NewLoggerBuilder().
		WithLevel(zapcore.DebugLevel).
		WithEncoding("json").
		WithWriter(writer).
		WithCaller()
}

// WithCaller adds caller information to the logger
func (b *LoggerBuilder) WithCaller() *LoggerBuilder {
	b.options = append(b.options, zap.AddCaller())
	return b
}

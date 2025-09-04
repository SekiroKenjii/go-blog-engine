package logger

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const (
	testMessage = "test message"
)

// TestLoggerComponents tests the individual components of the logger
// without using the singleton pattern that depends on external config
func TestLoggerComponents(t *testing.T) {
	t.Run("test logger creation with custom config", func(t *testing.T) {
		// Create encoder similar to the one used in the actual logger
		encoder := createTestEncoder()
		assert.NotNil(t, encoder)

		// Create a test writer (to memory instead of file)
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		// Test logging
		logger.Info(testMessage, zap.String("key", "value"))

		// Verify
		logs := recorded.All()
		assert.Len(t, logs, 1)
		assert.Equal(t, testMessage, logs[0].Message)
	})
}

// createTestEncoder creates an encoder with the same configuration as the production logger
func createTestEncoder() zapcore.Encoder {
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder.EncodeDuration = zapcore.SecondsDurationEncoder
	encoder.EncodeCaller = zapcore.ShortCallerEncoder
	encoder.TimeKey = "time"
	return zapcore.NewJSONEncoder(encoder)
}

// TestLoggerBehaviorWithObserver tests logger behavior using observer pattern
func TestLoggerBehaviorWithObserver(t *testing.T) {
	t.Run("test log levels", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		// Test different log levels
		logger.Debug("debug message") // Should not be recorded (below threshold)
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		// Verify recorded logs
		logs := recorded.All()
		assert.Len(t, logs, 3, "Should record info, warn, and error")

		assert.Equal(t, "info message", logs[0].Message)
		assert.Equal(t, zapcore.InfoLevel, logs[0].Level)

		assert.Equal(t, "warn message", logs[1].Message)
		assert.Equal(t, zapcore.WarnLevel, logs[1].Level)

		assert.Equal(t, "error message", logs[2].Message)
		assert.Equal(t, zapcore.ErrorLevel, logs[2].Level)
	})

	t.Run("test log fields", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		logger.Info("message with fields",
			zap.String("string_field", "test"),
			zap.Int("int_field", 42),
			zap.Bool("bool_field", true),
		)

		logs := recorded.All()
		assert.Len(t, logs, 1)

		log := logs[0]
		assert.Equal(t, "message with fields", log.Message)

		// Check fields
		fields := log.Context
		assert.Len(t, fields, 3)

		// Verify field values
		fieldMap := make(map[string]interface{})
		for _, field := range fields {
			switch field.Type {
			case zapcore.StringType:
				fieldMap[field.Key] = field.String
			case zapcore.Int64Type:
				fieldMap[field.Key] = field.Integer
			case zapcore.BoolType:
				fieldMap[field.Key] = field.Integer == 1
			}
		}

		assert.Equal(t, "test", fieldMap["string_field"])
		assert.Equal(t, int64(42), fieldMap["int_field"])
		assert.Equal(t, true, fieldMap["bool_field"])
	})
}

// TestZapEncoderConfiguration tests the encoder configuration matches expectations
func TestZapEncoderConfiguration(t *testing.T) {
	t.Run("encoder produces correct JSON format", func(t *testing.T) {
		encoder := createTestEncoder()

		// Create a test log entry
		entry := zapcore.Entry{
			Level:   zapcore.InfoLevel,
			Time:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Message: "test message",
			Caller:  zapcore.EntryCaller{Defined: true, File: "test.go", Line: 10},
		}

		fields := []zapcore.Field{
			zap.String("key", "value"),
		}

		buffer, err := encoder.EncodeEntry(entry, fields)
		assert.NoError(t, err)

		jsonStr := buffer.String()
		assert.Contains(t, jsonStr, `"level":"INFO"`)
		assert.Contains(t, jsonStr, `"time":"2023-01-01T12:00:00.000Z"`)
		assert.Contains(t, jsonStr, `"msg":"test message"`)
		assert.Contains(t, jsonStr, `"caller":"test.go:10"`)
		assert.Contains(t, jsonStr, `"key":"value"`)
	})
}

// TestMultiWriteSyncer tests the multi-writer functionality
func TestMultiWriteSyncer(t *testing.T) {
	t.Run("multi writer syncer writes to multiple outputs", func(t *testing.T) {
		// Create temporary file for testing
		tmpFile, err := os.CreateTemp("", "test-log-*.log")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// Create buffer for second writer instead of stdout to avoid sync issues
		buffer := &bytes.Buffer{}

		// Create multi-writer syncer similar to production
		fileSyncer := zapcore.AddSync(tmpFile)
		bufferSyncer := zapcore.AddSync(buffer)
		multiSyncer := zapcore.NewMultiWriteSyncer(bufferSyncer, fileSyncer)

		// Test writing
		testData := []byte("test log entry\n")
		n, err := multiSyncer.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)

		// Sync to ensure data is written (don't sync stdout)
		// Only sync the file syncer, not the multi-syncer to avoid stdout sync
		err = fileSyncer.Sync()
		assert.NoError(t, err)

		// Verify buffer contains the data
		assert.Equal(t, string(testData), buffer.String())

		// Verify file contains the data
		tmpFile.Seek(0, 0)
		fileContent := make([]byte, len(testData))
		n, err = tmpFile.Read(fileContent)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, fileContent)
	})
}

// TestLoggerErrorHandling tests error scenarios
func TestLoggerErrorHandling(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	t.Run("logging empty message", func(t *testing.T) {
		logger.Info("")
		logs := recorded.TakeAll()
		assert.Len(t, logs, 1)
		assert.Equal(t, "", logs[0].Message)
	})

	t.Run("logging with various field types", func(t *testing.T) {
		logger.Info("complex message",
			zap.String("string", "value"),
			zap.Int("int", 42),
			zap.Float64("float", 3.14),
			zap.Bool("bool", true),
			zap.Duration("duration", time.Second),
			zap.Time("time", time.Now()),
			zap.Any("any", map[string]interface{}{"key": "value"}),
		)

		logs := recorded.TakeAll()
		assert.Len(t, logs, 1)
		assert.Equal(t, "complex message", logs[0].Message)
		assert.Len(t, logs[0].Context, 7) // All fields should be present
	})
}

// BenchmarkLogger benchmarks logger performance
func BenchmarkLogger(b *testing.B) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	b.Run("info logging", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("info logging with fields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("number", i),
			)
		}
	})
}

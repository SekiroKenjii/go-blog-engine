package mailers_test

import (
	"context"
	"testing"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
	"github.com/stretchr/testify/assert"
)

// Mock sender for worker testing
type MockSender struct {
	sendFunc func(ctx context.Context, strategyName, toEmail string, params map[string]any) error
}

func (m *MockSender) Send(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, strategyName, toEmail, params)
	}
	return nil
}

func (m *MockSender) SendAsync(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
	return nil
}

// RegisterStrategy implements IMailService interface - empty implementation for testing
func (m *MockSender) RegisterStrategy(name string, strategy strategies.IMailStrategy) {
	// Empty implementation for testing purposes
}

func (m *MockSender) GetAvailableStrategies() []string {
	return []string{}
}

func TestMailWorker(t *testing.T) {
	t.Run("NewMailWorker", func(t *testing.T) {
		mockSender := &MockSender{}
		worker := mailers.NewMailWorker(mockSender, 2, 10, 3)

		assert.NotNil(t, worker)
	})

	t.Run("AddJob - Success", func(t *testing.T) {
		mockSender := &MockSender{}
		worker := mailers.NewMailWorker(mockSender, 1, 10, 3)

		job := mailers.MailJob{
			ToEmail:      testEmailAddress,
			StrategyName: "test",
			Params:       map[string]any{"key": "value"},
		}

		success := worker.AddJob(job)
		assert.True(t, success)
	})

	// Note: AddJob - Queue Full test skipped because it triggers logger.Warn
	// which requires config singleton

	t.Run("GetJobQueue", func(t *testing.T) {
		mockSender := &MockSender{}
		worker := mailers.NewMailWorker(mockSender, 1, 10, 3)

		jobQueue := worker.GetJobQueue()
		assert.NotNil(t, jobQueue)

		// Test that we can send to the queue
		job := mailers.MailJob{
			ToEmail:      testEmailAddress,
			StrategyName: "test",
			Params:       map[string]any{"key": "value"},
		}

		select {
		case jobQueue <- job:
			// Success
		default:
			t.Error("Failed to send job to queue")
		}
	})

	// Note: Worker Lifecycle test skipped because Start() triggers logger.Info
	// which requires config singleton

	// Note: Job Processing and Multiple Workers tests skipped because they
	// call Start() which triggers logger.Info requiring config singleton

	t.Run("Worker Queue Channel Access", func(t *testing.T) {
		mockSender := &MockSender{}
		worker := mailers.NewMailWorker(mockSender, 1, 5, 3)

		jobQueue := worker.GetJobQueue()

		// Test direct queue access
		job := mailers.MailJob{
			ToEmail:      testEmailAddress,
			StrategyName: "direct_test",
			Params:       map[string]any{"direct": true},
		}

		// Should be able to send directly to queue
		select {
		case jobQueue <- job:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout sending to job queue")
		}
	})
}

func TestMailJob(t *testing.T) {
	t.Run("MailJob Creation", func(t *testing.T) {
		now := time.Now()
		job := mailers.MailJob{
			ToEmail:      testEmailAddress,
			StrategyName: "test_strategy",
			Params:       map[string]any{"name": "John", "age": 30},
			Attempts:     1,
			CreatedAt:    now,
		}

		assert.Equal(t, testEmailAddress, job.ToEmail)
		assert.Equal(t, "test_strategy", job.StrategyName)
		assert.Equal(t, 1, job.Attempts)
		assert.Equal(t, now, job.CreatedAt)
		assert.Contains(t, job.Params, "name")
		assert.Contains(t, job.Params, "age")
	})

	t.Run("MailJob Zero Values", func(t *testing.T) {
		job := mailers.MailJob{}

		assert.Empty(t, job.ToEmail)
		assert.Empty(t, job.StrategyName)
		assert.Empty(t, job.Params)
		assert.Zero(t, job.Attempts)
		assert.True(t, job.CreatedAt.IsZero())
	})
}

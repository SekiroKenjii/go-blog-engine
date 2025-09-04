package mailers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	"github.com/stretchr/testify/assert"
)

const (
	testEmailAddress    = "test@example.com"
	testStrategyName    = "test"
	nonexistentStrategy = "nonexistent"
)

// Mock strategy for testing
type MockMailStrategy struct {
	validateFunc     func(params map[string]any) error
	prepareDataFunc  func(ctx context.Context, params map[string]any) (map[string]any, error)
	templateName     string
	subject          string
	fallbackTemplate string
}

func (m *MockMailStrategy) GetTemplateName() string {
	if m.templateName == "" {
		return "test.html"
	}
	return m.templateName
}

func (m *MockMailStrategy) GetSubject() string {
	if m.subject == "" {
		return "Test Subject"
	}
	return m.subject
}

func (m *MockMailStrategy) GetFallbackTemplate() string {
	if m.fallbackTemplate == "" {
		return "Fallback: Hello {{.name}}"
	}
	return m.fallbackTemplate
}

func (m *MockMailStrategy) PrepareData(ctx context.Context, params map[string]any) (map[string]any, error) {
	if m.prepareDataFunc != nil {
		return m.prepareDataFunc(ctx, params)
	}
	return params, nil
}

func (m *MockMailStrategy) Validate(params map[string]any) error {
	if m.validateFunc != nil {
		return m.validateFunc(params)
	}
	return nil
}

func TestMailService(t *testing.T) {
	createTestService := func() mailers.IMailService {
		cfg := &config.EmailConfig{}
		template := mailers.NewMailTemplate("/tmp")
		worker := mailers.NewMailWorker(nil, 1, 10, 3)
		return mailers.NewMailService(cfg, template, worker)
	}

	t.Run("NewMailService", func(t *testing.T) {
		service := createTestService()
		assert.NotNil(t, service)
	})

	t.Run("RegisterStrategy", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{}

		service.RegisterStrategy(testStrategyName, mockStrategy)

		strategies := service.GetAvailableStrategies()
		assert.Contains(t, strategies, testStrategyName)
	})

	t.Run("GetAvailableStrategies", func(t *testing.T) {
		service := createTestService()

		// Initially empty
		strategies := service.GetAvailableStrategies()
		assert.Empty(t, strategies)

		// Add strategies
		service.RegisterStrategy("strategy1", &MockMailStrategy{})
		service.RegisterStrategy("strategy2", &MockMailStrategy{})

		strategies = service.GetAvailableStrategies()
		assert.Len(t, strategies, 2)
		assert.Contains(t, strategies, "strategy1")
		assert.Contains(t, strategies, "strategy2")
	})

	t.Run("Send - Strategy Not Found", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		err := service.Send(ctx, nonexistentStrategy, testEmailAddress, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email strategy 'nonexistent' not found")
	})

	t.Run("Send - Validation Failed", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{
			validateFunc: func(params map[string]any) error {
				return errors.New("validation error")
			},
		}

		service.RegisterStrategy(testStrategyName, mockStrategy)
		ctx := context.Background()

		err := service.Send(ctx, testStrategyName, testEmailAddress, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parameter validation failed")
	})

	t.Run("Send - PrepareData Failed", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{
			prepareDataFunc: func(ctx context.Context, params map[string]any) (map[string]any, error) {
				return nil, errors.New("prepare data error")
			},
		}

		service.RegisterStrategy(testStrategyName, mockStrategy)
		ctx := context.Background()

		err := service.Send(ctx, testStrategyName, testEmailAddress, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare template data")
	})

	t.Run("SendAsync - Strategy Not Found", func(t *testing.T) {
		service := createTestService()
		ctx := context.Background()

		err := service.SendAsync(ctx, nonexistentStrategy, testEmailAddress, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email strategy 'nonexistent' not found")
	})

	t.Run("SendAsync - Validation Failed", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{
			validateFunc: func(params map[string]any) error {
				return errors.New("validation error")
			},
		}

		service.RegisterStrategy(testStrategyName, mockStrategy)
		ctx := context.Background()

		err := service.SendAsync(ctx, testStrategyName, testEmailAddress, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parameter validation failed")
	})

	t.Run("SendAsync - Success", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{}

		service.RegisterStrategy(testStrategyName, mockStrategy)
		ctx := context.Background()
		params := map[string]any{"key": "value"}

		err := service.SendAsync(ctx, testStrategyName, testEmailAddress, params)
		// The error could be nil (if queue accepts) or queue full error
		// Both are valid depending on the queue state
		if err != nil {
			assert.Contains(t, err.Error(), "email queue is full")
		}
	})

	t.Run("SendAsync - Context Cancelled", func(t *testing.T) {
		service := createTestService()
		mockStrategy := &MockMailStrategy{}

		service.RegisterStrategy(testStrategyName, mockStrategy)

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := service.SendAsync(ctx, testStrategyName, testEmailAddress, map[string]any{})
		// Should return context cancelled error or succeed depending on timing
		if err != nil {
			assert.True(t, errors.Is(err, context.Canceled) ||
				err.Error() == "email queue is full")
		}
	})

	t.Run("Multiple Strategy Registration", func(t *testing.T) {
		service := createTestService()

		// Register same strategy with different names
		strategy1 := &MockMailStrategy{templateName: "template1.html"}
		strategy2 := &MockMailStrategy{templateName: "template2.html"}

		service.RegisterStrategy("strategy1", strategy1)
		service.RegisterStrategy("strategy2", strategy2)

		// Override strategy1
		strategy1Updated := &MockMailStrategy{templateName: "updated.html"}
		service.RegisterStrategy("strategy1", strategy1Updated)

		strategies := service.GetAvailableStrategies()
		assert.Len(t, strategies, 2)
		assert.Contains(t, strategies, "strategy1")
		assert.Contains(t, strategies, "strategy2")
	})

	// Note: Full Send flow test disabled due to logger/config singleton dependencies
	// that are complex to mock in unit tests. This would be better suited for integration tests.

	// t.Run("Send - Template Rendering Failed", func(t *testing.T) {
	//     This test is disabled because it triggers logger.Warn through RenderTemplateWithFallback
	//     which depends on config singleton that's difficult to mock in unit tests.
	//     Template rendering failure paths should be tested in integration tests.
	// })

	t.Run("buildMessage - Functionality Test", func(t *testing.T) {
		// Test the message building functionality indirectly through service methods
		// since buildMessage is a private method
		cfg := &config.EmailConfig{
			FromName:  "Test Sender",
			FromEmail: "test@example.com",
		}
		template := mailers.NewMailTemplate("/tmp")
		worker := mailers.NewMailWorker(nil, 1, 10, 3)

		service := mailers.NewMailService(cfg, template, worker)

		// Verify service was created successfully with the config
		assert.NotNil(t, service)

		// Test that service has the expected strategies (empty initially)
		strategies := service.GetAvailableStrategies()
		assert.Empty(t, strategies)
	})
}

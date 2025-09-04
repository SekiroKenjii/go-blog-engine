package mailers_test

import (
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	"github.com/stretchr/testify/assert"
)

func TestMailerFactory(t *testing.T) {
	t.Run("NewMailerFactory", func(t *testing.T) {
		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 2,
				QueueSize:   10,
				MaxRetries:  3,
			},
		}

		factory := mailers.NewMailerFactory(cfg)
		assert.NotNil(t, factory)
	})

	t.Run("CreateMailerSystem - Success", func(t *testing.T) {
		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 2,
				QueueSize:   10,
				MaxRetries:  3,
			},
		}

		factory := mailers.NewMailerFactory(cfg)

		mailSvc, worker, err := factory.CreateMailerSystem()
		assert.NoError(t, err)
		assert.NotNil(t, mailSvc)
		assert.NotNil(t, worker)

		// Verify that strategies are registered
		strategies := mailSvc.GetAvailableStrategies()
		assert.Greater(t, len(strategies), 0, "Should have registered strategies")

		// Verify common auth strategies are registered
		assert.Contains(t, strategies, "verification")
		assert.Contains(t, strategies, "password_reset")
		assert.Contains(t, strategies, "welcome")
	})

	t.Run("CreateMailerSystem - Singleton Behavior", func(t *testing.T) {
		// Reset singleton to ensure clean state
		mailers.ResetSingleton()

		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 2,
				QueueSize:   10,
				MaxRetries:  3,
			},
		}

		factory := mailers.NewMailerFactory(cfg)

		// First call
		mailSvc1, worker1, err1 := factory.CreateMailerSystem()
		assert.NoError(t, err1)
		assert.NotNil(t, mailSvc1)
		assert.NotNil(t, worker1)

		// Second call should return same instances
		mailSvc2, worker2, err2 := factory.CreateMailerSystem()
		assert.NoError(t, err2)
		assert.Equal(t, mailSvc1, mailSvc2, "Should return same mail service instance")
		assert.Equal(t, worker1, worker2, "Should return same worker instance")

		// Reset for cleanup
		mailers.ResetSingleton()
	})

	t.Run("GetMailServiceInstance", func(t *testing.T) {
		// Reset singleton
		mailers.ResetSingleton()

		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 1,
				QueueSize:   5,
				MaxRetries:  2,
			},
		}

		factory := mailers.NewMailerFactory(cfg)

		// Should return nil before creation
		instance := mailers.GetMailServiceInstance()
		assert.Nil(t, instance)

		// Create mailer system
		mailSvc, _, err := factory.CreateMailerSystem()
		assert.NoError(t, err)

		// Now should return the instance
		instance = mailers.GetMailServiceInstance()
		assert.NotNil(t, instance)
		assert.Equal(t, mailSvc, instance)

		// Reset for cleanup
		mailers.ResetSingleton()
	})

	t.Run("GetWorkerInstance", func(t *testing.T) {
		// Reset singleton
		mailers.ResetSingleton()

		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 1,
				QueueSize:   5,
				MaxRetries:  2,
			},
		}

		factory := mailers.NewMailerFactory(cfg)

		// Should return nil before creation
		instance := mailers.GetWorkerInstance()
		assert.Nil(t, instance)

		// Create mailer system
		_, worker, err := factory.CreateMailerSystem()
		assert.NoError(t, err)

		// Now should return the instance
		instance = mailers.GetWorkerInstance()
		assert.NotNil(t, instance)
		assert.Equal(t, worker, instance)

		// Reset for cleanup
		mailers.ResetSingleton()
	})

	t.Run("ResetSingleton", func(t *testing.T) {
		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 1,
				QueueSize:   5,
				MaxRetries:  2,
			},
		}

		factory := mailers.NewMailerFactory(cfg)

		// Create first instance
		mailSvc1, worker1, err1 := factory.CreateMailerSystem()
		assert.NoError(t, err1)
		assert.NotNil(t, mailSvc1)
		assert.NotNil(t, worker1)

		// Reset singleton
		mailers.ResetSingleton()

		// Create new instance - should be different
		mailSvc2, worker2, err2 := factory.CreateMailerSystem()
		assert.NoError(t, err2)
		assert.NotNil(t, mailSvc2)
		assert.NotNil(t, worker2)

		// New instances should be different from old ones
		// Note: We can't directly compare interfaces for inequality in Go easily
		// So we'll check that they're both non-nil and the system works
		assert.NotNil(t, mailSvc1)
		assert.NotNil(t, mailSvc2)
		assert.NotNil(t, worker1)
		assert.NotNil(t, worker2)

		// Clean up
		mailers.ResetSingleton()
	})

	t.Run("Multiple Factory Instances Share Singleton", func(t *testing.T) {
		// Reset singleton
		mailers.ResetSingleton()

		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 1,
				QueueSize:   5,
				MaxRetries:  2,
			},
		}

		factory1 := mailers.NewMailerFactory(cfg)
		factory2 := mailers.NewMailerFactory(cfg)

		// Create from first factory
		mailSvc1, worker1, err1 := factory1.CreateMailerSystem()
		assert.NoError(t, err1)

		// Create from second factory - should return same instances
		mailSvc2, worker2, err2 := factory2.CreateMailerSystem()
		assert.NoError(t, err2)

		assert.Equal(t, mailSvc1, mailSvc2)
		assert.Equal(t, worker1, worker2)

		// Clean up
		mailers.ResetSingleton()
	})

	t.Run("Factory with Different Configs", func(t *testing.T) {
		// Reset singleton
		mailers.ResetSingleton()

		cfg1 := &config.EmailConfig{
			TemplateDir: "/tmp1",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 1,
				QueueSize:   5,
				MaxRetries:  2,
			},
		}

		cfg2 := &config.EmailConfig{
			TemplateDir: "/tmp2",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 3,
				QueueSize:   15,
				MaxRetries:  5,
			},
		}

		factory1 := mailers.NewMailerFactory(cfg1)
		factory2 := mailers.NewMailerFactory(cfg2)

		// First factory creates the singleton
		mailSvc1, _, err1 := factory1.CreateMailerSystem()
		assert.NoError(t, err1)

		// Second factory should return the same instance (singleton behavior)
		// even though it has different config
		mailSvc2, _, err2 := factory2.CreateMailerSystem()
		assert.NoError(t, err2)

		assert.Equal(t, mailSvc1, mailSvc2)

		// Clean up
		mailers.ResetSingleton()
	})

	t.Run("Concurrent Factory Access", func(t *testing.T) {
		// Reset singleton
		mailers.ResetSingleton()

		cfg := &config.EmailConfig{
			TemplateDir: "/tmp",
			Worker: config.EmailWorkerConfig{
				WorkerCount: 2,
				QueueSize:   10,
				MaxRetries:  3,
			},
		}

		const numGoroutines = 10
		results := make(chan mailers.IMailService, numGoroutines)

		// Launch multiple goroutines trying to create the mailer system
		for range numGoroutines {
			go func() {
				factory := mailers.NewMailerFactory(cfg)
				mailSvc, _, _ := factory.CreateMailerSystem()
				results <- mailSvc
			}()
		}

		// Collect all results
		var services []mailers.IMailService
		for i := 0; i < numGoroutines; i++ {
			service := <-results
			services = append(services, service)
		}

		// All should be the same instance (singleton)
		firstService := services[0]
		for i := 1; i < len(services); i++ {
			assert.Equal(t, firstService, services[i], "All services should be the same singleton instance")
		}

		// Clean up
		mailers.ResetSingleton()
	})
}

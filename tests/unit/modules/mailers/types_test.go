package mailers_test

import (
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	"github.com/stretchr/testify/assert"
)

func TestStrategyNames(t *testing.T) {
	t.Run("should return correct strategy names", func(t *testing.T) {
		// Test individual strategy names
		assert.Equal(t, "verification", mailers.Strategies.Verification())
		assert.Equal(t, "password_reset", mailers.Strategies.PasswordReset())
		assert.Equal(t, "welcome", mailers.Strategies.Welcome())
	})

	t.Run("should have consistent strategy names", func(t *testing.T) {
		// Test that strategy names are consistent across calls
		strategies := mailers.StrategyNames{}

		assert.Equal(t, strategies.Verification(), mailers.Strategies.Verification())
		assert.Equal(t, strategies.PasswordReset(), mailers.Strategies.PasswordReset())
		assert.Equal(t, strategies.Welcome(), mailers.Strategies.Welcome())
	})

	t.Run("should not return empty strings", func(t *testing.T) {
		assert.NotEmpty(t, mailers.Strategies.Verification())
		assert.NotEmpty(t, mailers.Strategies.PasswordReset())
		assert.NotEmpty(t, mailers.Strategies.Welcome())
	})

	t.Run("should have unique strategy names", func(t *testing.T) {
		names := []string{
			mailers.Strategies.Verification(),
			mailers.Strategies.PasswordReset(),
			mailers.Strategies.Welcome(),
		}

		// Check uniqueness
		uniqueNames := make(map[string]bool)
		for _, name := range names {
			assert.False(t, uniqueNames[name], "Strategy name '%s' should be unique", name)
			uniqueNames[name] = true
		}
	})
}

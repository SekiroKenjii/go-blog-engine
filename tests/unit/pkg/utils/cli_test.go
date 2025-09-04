package utils

import (
	"os"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// Test constants
const (
	testEnv = "test-env"
)

func TestGetEnvFromArgs(t *testing.T) {
	// Save original args to restore later
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name     string
		args     []string
		fallback []string
		expected string
	}{
		{
			name:     "no arguments provided, no fallback",
			args:     []string{"program"},
			fallback: nil,
			expected: "",
		},
		{
			name:     "no arguments provided, with fallback",
			args:     []string{"program"},
			fallback: []string{"development"},
			expected: "development",
		},
		{
			name:     "environment argument provided",
			args:     []string{"program", "production"},
			fallback: []string{"development"},
			expected: "production",
		},
		{
			name:     "empty environment argument",
			args:     []string{"program", ""},
			fallback: []string{"development"},
			expected: "development",
		},
		{
			name:     "multiple arguments, only first is used",
			args:     []string{"program", "staging", "extra", "args"},
			fallback: []string{"development"},
			expected: "staging",
		},
		{
			name:     "no arguments, multiple fallbacks (only first used)",
			args:     []string{"program"},
			fallback: []string{"development", "production"},
			expected: "development",
		},
		{
			name:     "environment with spaces",
			args:     []string{"program", "test environment"},
			fallback: []string{"development"},
			expected: "test environment",
		},
		{
			name:     "numeric environment name",
			args:     []string{"program", "123"},
			fallback: []string{"development"},
			expected: "123",
		},
		{
			name:     "special characters in environment",
			args:     []string{"program", testEnv + "_1.0"},
			fallback: []string{"development"},
			expected: testEnv + "_1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set os.Args for this test
			os.Args = tt.args

			var result string
			if tt.fallback == nil {
				result = utils.GetEnvFromArgs()
			} else {
				result = utils.GetEnvFromArgs(tt.fallback...)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvFromArgsEdgeCases(t *testing.T) {
	// Save original args to restore later
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	t.Run("empty args slice", func(t *testing.T) {
		os.Args = []string{}
		result := utils.GetEnvFromArgs("default")
		assert.Equal(t, "default", result)
	})

	t.Run("nil args", func(t *testing.T) {
		// This is theoretically possible but unlikely in practice
		os.Args = nil
		result := utils.GetEnvFromArgs("default")
		assert.Equal(t, "default", result)
	})

	t.Run("no fallback and no args", func(t *testing.T) {
		os.Args = []string{"program"}
		result := utils.GetEnvFromArgs()
		assert.Equal(t, "", result)
	})

	t.Run("empty fallback list", func(t *testing.T) {
		os.Args = []string{"program"}
		result := utils.GetEnvFromArgs()
		assert.Equal(t, "", result)
	})
}

func TestGetEnvFromArgsRealWorldScenarios(t *testing.T) {
	// Save original args to restore later
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name        string
		args        []string
		fallback    []string
		expected    string
		description string
	}{
		{
			name:        "development environment",
			args:        []string{"./myapp", "development"},
			fallback:    []string{"production"},
			expected:    "development",
			description: "Typical development setup",
		},
		{
			name:        "production environment",
			args:        []string{"/usr/bin/myapp", "production"},
			fallback:    []string{"development"},
			expected:    "production",
			description: "Production deployment",
		},
		{
			name:        "staging environment",
			args:        []string{"myapp.exe", "staging"},
			fallback:    []string{"development"},
			expected:    "staging",
			description: "Windows staging environment",
		},
		{
			name:        "testing environment",
			args:        []string{"go", "test", "main.go", "test"},
			fallback:    []string{"development"},
			expected:    "test",
			description: "Running with go test command (args[3] is the environment)",
		},
		{
			name:        "default fallback when no env specified",
			args:        []string{"./myapp"},
			fallback:    []string{"development"},
			expected:    "development",
			description: "Default behavior when no environment specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			result := utils.GetEnvFromArgs(tt.fallback...)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestGetEnvFromArgsConsistency(t *testing.T) {
	// Save original args to restore later
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	t.Run("multiple calls return same result", func(t *testing.T) {
		os.Args = []string{"program", testEnv}

		result1 := utils.GetEnvFromArgs("default")
		result2 := utils.GetEnvFromArgs("default")
		result3 := utils.GetEnvFromArgs("different-default")

		assert.Equal(t, testEnv, result1)
		assert.Equal(t, testEnv, result2)
		assert.Equal(t, testEnv, result3)
		assert.Equal(t, result1, result2)
		assert.Equal(t, result2, result3)
	})

	t.Run("changing args affects subsequent calls", func(t *testing.T) {
		os.Args = []string{"program", "env1"}
		result1 := utils.GetEnvFromArgs("default")
		assert.Equal(t, "env1", result1)

		os.Args = []string{"program", "env2"}
		result2 := utils.GetEnvFromArgs("default")
		assert.Equal(t, "env2", result2)

		assert.NotEqual(t, result1, result2)
	})
}

// Benchmark tests
func BenchmarkGetEnvFromArgs(b *testing.B) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"program", "production"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GetEnvFromArgs("development")
	}
}

func BenchmarkGetEnvFromArgsNoArgs(b *testing.B) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"program"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GetEnvFromArgs("development")
	}
}

func BenchmarkGetEnvFromArgsNoFallback(b *testing.B) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"program"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GetEnvFromArgs()
	}
}

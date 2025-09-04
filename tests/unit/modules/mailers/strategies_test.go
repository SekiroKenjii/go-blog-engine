package mailers_test

import (
	"context"
	"testing"

	authStrategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFirstNamePlaceholder = "{{.FirstName}}"
	prepareDataSuccessTest   = "PrepareData - Success"
	validateSuccessTest      = "Validate - Success"
	testToken123             = "test-token-123"
	testFirstNameJohn        = "John"
	testEmailJohn            = "john@example.com"
	testBaseURL              = "https://example.com"
	firstNameRequiredError   = "firstName is required"
	verifyToken456           = "verify-token-456"
	testFirstNameJane        = "Jane"
	testEmailJane            = "jane@example.com"
	testFirstNameBob         = "Bob"
	testEmailBob             = "bob@example.com"
)

func TestPasswordResetEmailStrategy(t *testing.T) {
	t.Run("NewPasswordResetEmailStrategy", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)
		assert.NotNil(t, strategy)
	})

	t.Run("GetTemplateName", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)
		templateName := strategy.GetTemplateName()
		assert.Equal(t, "password_reset.html", templateName)
	})

	t.Run("GetSubject", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)
		subject := strategy.GetSubject()
		assert.Equal(t, "Password Reset Request", subject)
	})

	t.Run("GetFallbackTemplate", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)
		fallback := strategy.GetFallbackTemplate()
		assert.Contains(t, fallback, "reset your password")
		assert.Contains(t, fallback, "{{.ResetURL}}")
		assert.Contains(t, fallback, testFirstNamePlaceholder)
	})

	t.Run(prepareDataSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		params := map[string]any{
			"token":     testToken123,
			"firstName": testFirstNameJohn,
			"email":     testEmailJohn,
		}

		data, err := strategy.PrepareData(context.Background(), params)
		require.NoError(t, err)

		assert.Equal(t, testFirstNameJohn, data["FirstName"])
		assert.Contains(t, data["ResetURL"].(string), testToken123)
		assert.Contains(t, data["ResetURL"].(string), testEmailJohn)
		assert.Equal(t, testBaseURL, data["BaseURL"])
	})

	t.Run("PrepareData - Missing Token", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		params := map[string]any{
			"firstName": testFirstNameJohn,
			"email":     testEmailJohn,
		}

		_, err := strategy.PrepareData(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is required")
	})

	t.Run("PrepareData - Missing FirstName", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		params := map[string]any{
			"token": testToken123,
			"email": testEmailJohn,
		}

		_, err := strategy.PrepareData(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), firstNameRequiredError)
	})

	t.Run("PrepareData - Missing Email", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		params := map[string]any{
			"token":     testToken123,
			"firstName": testFirstNameJohn,
		}

		_, err := strategy.PrepareData(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run(validateSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		params := map[string]any{
			"token":     testToken123,
			"firstName": testFirstNameJohn,
			"email":     testEmailJohn,
		}

		err := strategy.Validate(params)
		assert.NoError(t, err)
	})

	t.Run("Validate - Missing Required Fields", func(t *testing.T) {
		strategy := authStrategies.NewPasswordResetEmailStrategy(testBaseURL)

		testCases := []struct {
			name     string
			params   map[string]any
			errorMsg string
		}{
			{
				name:     "Missing Token",
				params:   map[string]any{"firstName": "John", "email": testEmailJohn},
				errorMsg: "token parameter is required",
			},
			{
				name:     "Missing FirstName",
				params:   map[string]any{"token": "token123", "email": testEmailJohn},
				errorMsg: "firstName parameter is required",
			},
			{
				name:     "Missing Email",
				params:   map[string]any{"token": "token123", "firstName": "John"},
				errorMsg: "email parameter is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := strategy.Validate(tc.params)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			})
		}
	})
}

func TestVerificationEmailStrategy(t *testing.T) {
	t.Run("NewVerificationEmailStrategy", func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)
		assert.NotNil(t, strategy)
	})

	t.Run("GetTemplateName", func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)
		templateName := strategy.GetTemplateName()
		assert.Equal(t, "verification.html", templateName)
	})

	t.Run("GetSubject", func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)
		subject := strategy.GetSubject()
		assert.Equal(t, "Please verify your email address", subject)
	})

	t.Run("GetFallbackTemplate", func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)
		fallback := strategy.GetFallbackTemplate()
		assert.Contains(t, fallback, "verify your email")
		assert.Contains(t, fallback, "{{.VerificationURL}}")
		assert.Contains(t, fallback, testFirstNamePlaceholder)
	})

	t.Run(prepareDataSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)

		params := map[string]any{
			"token":     verifyToken456,
			"firstName": testFirstNameJane,
			"email":     testEmailJane,
		}

		data, err := strategy.PrepareData(context.Background(), params)
		require.NoError(t, err)

		assert.Equal(t, testFirstNameJane, data["FirstName"])
		assert.Contains(t, data["VerificationURL"].(string), verifyToken456)
		assert.Equal(t, testBaseURL, data["BaseURL"])
	})

	t.Run(validateSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewVerificationEmailStrategy(testBaseURL)

		params := map[string]any{
			"token":     verifyToken456,
			"firstName": testFirstNameJane,
			"email":     testEmailJane,
		}

		err := strategy.Validate(params)
		assert.NoError(t, err)
	})
}

func TestWelcomeEmailStrategy(t *testing.T) {
	t.Run("NewWelcomeEmailStrategy", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)
		assert.NotNil(t, strategy)
	})

	t.Run("GetTemplateName", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)
		templateName := strategy.GetTemplateName()
		assert.Equal(t, "welcome.html", templateName)
	})

	t.Run("GetSubject", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)
		subject := strategy.GetSubject()
		assert.Equal(t, "Welcome to Go Blog Engine!", subject)
	})

	t.Run("GetFallbackTemplate", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)
		fallback := strategy.GetFallbackTemplate()
		assert.Contains(t, fallback, "Welcome")
		assert.Contains(t, fallback, testFirstNamePlaceholder)
	})

	t.Run(prepareDataSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)

		params := map[string]any{
			"firstName": testFirstNameBob,
			"email":     testEmailBob,
		}

		data, err := strategy.PrepareData(context.Background(), params)
		require.NoError(t, err)

		assert.Equal(t, testFirstNameBob, data["FirstName"])
		assert.Equal(t, testBaseURL, data["BaseURL"])
	})

	t.Run("PrepareData - Missing FirstName", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)

		params := map[string]any{
			"email": testEmailBob,
		}

		_, err := strategy.PrepareData(context.Background(), params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), firstNameRequiredError)
	})

	t.Run(validateSuccessTest, func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)

		params := map[string]any{
			"firstName": testFirstNameBob,
			"email":     testEmailBob,
		}

		err := strategy.Validate(params)
		assert.NoError(t, err)
	})

	t.Run("Validate - Missing FirstName", func(t *testing.T) {
		strategy := authStrategies.NewWelcomeEmailStrategy(testBaseURL)

		params := map[string]any{
			"email": testEmailBob,
		}

		err := strategy.Validate(params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "firstName parameter is required")
	})
}

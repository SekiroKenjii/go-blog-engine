package mailers

import (
	"context"
	"errors"
	"fmt"

	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
)

const passwordResetFallback = `Hi {{.FirstName}},

We received a request to reset your password for your Go Blog Engine account.

Please click the following link to reset your password:
{{.ResetURL}}

This link will expire in 1 hour for security reasons.

If you didn't request a password reset, please ignore this email.

Best regards,
Go Blog Engine Team`

type PasswordResetEmailStrategy struct {
	baseURL string
}

func NewPasswordResetEmailStrategy(baseURL string) strategies.IMailStrategy {
	return &PasswordResetEmailStrategy{baseURL: baseURL}
}

func (s *PasswordResetEmailStrategy) GetTemplateName() string {
	return "password_reset.html"
}

func (s *PasswordResetEmailStrategy) GetSubject() string {
	return "Password Reset Request"
}

func (s *PasswordResetEmailStrategy) GetFallbackTemplate() string {
	return passwordResetFallback
}

func (s *PasswordResetEmailStrategy) PrepareData(ctx context.Context, params map[string]any) (map[string]any, error) {
	token, ok := params["token"].(string)
	if !ok {
		return nil, fmt.Errorf("token is required for password reset email")
	}

	firstName, ok := params["firstName"].(string)
	if !ok {
		return nil, fmt.Errorf("firstName is required for password reset email")
	}

	email, ok := params["email"].(string)
	if !ok {
		return nil, fmt.Errorf("email is required for password reset email")
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s", s.baseURL, token, email)

	return map[string]any{
		"FirstName": firstName,
		"ResetURL":  resetURL,
		"BaseURL":   s.baseURL,
	}, nil
}

func (s *PasswordResetEmailStrategy) Validate(params map[string]any) error {
	if _, ok := params["token"]; !ok {
		return errors.New("token parameter is required")
	}
	if _, ok := params["firstName"]; !ok {
		return errors.New("firstName parameter is required")
	}
	if _, ok := params["email"]; !ok {
		return errors.New("email parameter is required")
	}
	return nil
}

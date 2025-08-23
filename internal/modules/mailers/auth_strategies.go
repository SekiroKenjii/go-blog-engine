package mailers

import (
	"context"
	"errors"
	"fmt"
)

const (
	ErrTokenRequired     = "token parameter is required"
	ErrFirstNameRequired = "firstName parameter is required"
	ErrEmailRequired     = "email parameter is required"
)

// Simple fallback templates for when HTML templates are not available
const (
	verificationFallback = `Hi {{.FirstName}},

Thank you for registering with Go Blog Engine!

Please click the following link to verify your email address:
{{.VerificationURL}}

If you didn't create this account, please ignore this email.

Best regards,
Go Blog Engine Team`

	passwordResetFallback = `Hi {{.FirstName}},

We received a request to reset your password for your Go Blog Engine account.

Please click the following link to reset your password:
{{.ResetURL}}

This link will expire in 1 hour for security reasons.

If you didn't request a password reset, please ignore this email.

Best regards,
Go Blog Engine Team`

	welcomeFallback = `Hi {{.FirstName}},

Welcome to Go Blog Engine!

Your email has been successfully verified and your account is now active.

You can now start creating and sharing your blog posts.

Visit us at: {{.BaseURL}}

Best regards,
Go Blog Engine Team`
)

// VerificationEmailStrategy handles email verification emails
type VerificationEmailStrategy struct {
	baseURL string
}

func NewVerificationEmailStrategy(baseURL string) IEmailStrategy {
	return &VerificationEmailStrategy{baseURL: baseURL}
}

func (s *VerificationEmailStrategy) GetTemplateName() string {
	return "verification.html"
}

func (s *VerificationEmailStrategy) GetSubject() string {
	return "Please verify your email address"
}

func (s *VerificationEmailStrategy) GetFallbackTemplate() string {
	return verificationFallback
}

func (s *VerificationEmailStrategy) PrepareData(ctx context.Context, params map[string]any) (map[string]any, error) {
	token, ok := params["token"].(string)
	if !ok {
		return nil, fmt.Errorf("token is required for verification email")
	}

	firstName, ok := params["firstName"].(string)
	if !ok {
		return nil, fmt.Errorf("firstName is required for verification email")
	}

	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.baseURL, token)

	return map[string]any{
		"FirstName":       firstName,
		"VerificationURL": verificationURL,
		"BaseURL":         s.baseURL,
	}, nil
}

func (s *VerificationEmailStrategy) Validate(params map[string]any) error {
	if _, ok := params["token"]; !ok {
		return errors.New(ErrTokenRequired)
	}
	if _, ok := params["firstName"]; !ok {
		return errors.New(ErrFirstNameRequired)
	}
	return nil
}

// PasswordResetEmailStrategy handles password reset emails
type PasswordResetEmailStrategy struct {
	baseURL string
}

func NewPasswordResetEmailStrategy(baseURL string) IEmailStrategy {
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
		return errors.New(ErrTokenRequired)
	}
	if _, ok := params["firstName"]; !ok {
		return errors.New(ErrFirstNameRequired)
	}
	if _, ok := params["email"]; !ok {
		return errors.New(ErrEmailRequired)
	}
	return nil
}

// WelcomeEmailStrategy handles welcome emails
type WelcomeEmailStrategy struct {
	baseURL string
}

func NewWelcomeEmailStrategy(baseURL string) IEmailStrategy {
	return &WelcomeEmailStrategy{baseURL: baseURL}
}

func (s *WelcomeEmailStrategy) GetTemplateName() string {
	return "welcome.html"
}

func (s *WelcomeEmailStrategy) GetSubject() string {
	return "Welcome to Go Blog Engine!"
}

func (s *WelcomeEmailStrategy) GetFallbackTemplate() string {
	return welcomeFallback
}

func (s *WelcomeEmailStrategy) PrepareData(ctx context.Context, params map[string]any) (map[string]any, error) {
	firstName, ok := params["firstName"].(string)
	if !ok {
		return nil, fmt.Errorf("firstName is required for welcome email")
	}

	return map[string]any{
		"FirstName": firstName,
		"BaseURL":   s.baseURL,
	}, nil
}

func (s *WelcomeEmailStrategy) Validate(params map[string]any) error {
	if _, ok := params["firstName"]; !ok {
		return errors.New(ErrFirstNameRequired)
	}
	return nil
}

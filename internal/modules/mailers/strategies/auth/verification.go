package mailers

import (
	"context"
	"errors"
	"fmt"

	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
)

const verificationFallback = `Hi {{.FirstName}},

Thank you for registering with Go Blog Engine!

Please click the following link to verify your email address:
{{.VerificationURL}}

If you didn't create this account, please ignore this email.

Best regards,
Go Blog Engine Team`

type VerificationEmailStrategy struct {
	baseURL string
}

func NewVerificationEmailStrategy(baseURL string) strategies.IMailStrategy {
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
		return errors.New("token parameter is required")
	}

	if _, ok := params["firstName"]; !ok {
		return errors.New("firstName parameter is required")
	}

	return nil
}

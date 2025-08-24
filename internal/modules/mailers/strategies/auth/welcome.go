package mailers

import (
	"context"
	"errors"
	"fmt"

	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
)

const welcomeFallback = `Hi {{.FirstName}},

Welcome to Go Blog Engine!

Your email has been successfully verified and your account is now active.

You can now start creating and sharing your blog posts.

Visit us at: {{.BaseURL}}

Best regards,
Go Blog Engine Team`

type WelcomeEmailStrategy struct {
	baseURL string
}

func NewWelcomeEmailStrategy(baseURL string) strategies.IMailStrategy {
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
		return errors.New("firstName parameter is required")
	}

	return nil
}

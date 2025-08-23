package mailers

import "context"

// IEmailStrategy defines how different types of emails should be handled
type IEmailStrategy interface {
	// GetTemplateName returns the template filename for this email type
	GetTemplateName() string

	// GetSubject returns the email subject
	GetSubject() string

	// GetFallbackTemplate returns the fallback template if HTML template fails
	GetFallbackTemplate() string

	// PrepareData prepares the data for template rendering
	PrepareData(ctx context.Context, params map[string]any) (map[string]any, error)

	// Validate validates the required parameters for this email type
	Validate(params map[string]any) error
}

// IEmailSender interface for sending emails using strategy pattern
type IEmailSender interface {
	// SendEmail sends an email synchronously using strategy name
	SendEmail(ctx context.Context, strategyName string, toEmail string, params map[string]any) error

	// SendEmailAsync sends an email asynchronously using strategy name
	SendEmailAsync(ctx context.Context, strategyName string, toEmail string, params map[string]any) error
}

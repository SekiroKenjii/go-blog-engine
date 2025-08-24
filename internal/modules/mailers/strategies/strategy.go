package mailers

import "context"

// IMailStrategy defines how different types of emails should be handled
type IMailStrategy interface {
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

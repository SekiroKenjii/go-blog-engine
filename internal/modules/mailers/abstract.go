package mailers

import "context"

// IMailSender interface for sending emails
type IMailSender interface {
	// Send sends an email synchronously
	Send(ctx context.Context, strategyName string, toEmail string, params map[string]any) error

	// SendAsync sends an email asynchronously
	SendAsync(ctx context.Context, strategyName string, toEmail string, params map[string]any) error
}

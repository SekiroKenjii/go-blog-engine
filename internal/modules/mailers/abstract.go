package mailers

import (
	"context"

	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
)

// IMailService interface for mail service operations
type IMailService interface {
	// RegisterStrategy registers an email strategy
	RegisterStrategy(name string, strategy strategies.IMailStrategy)

	// Send sends an email synchronously
	Send(ctx context.Context, strategyName, toEmail string, params map[string]any) error

	// SendAsync sends an email asynchronously
	SendAsync(ctx context.Context, strategyName, toEmail string, params map[string]any) error

	// GetAvailableStrategies returns list of registered strategies
	GetAvailableStrategies() []string
}

package mailers

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// StrategicMailerConfig holds configuration for StrategicMailer
type StrategicMailerConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// StrategicMailer implements IEmailSender interface using strategy pattern
type StrategicMailer struct {
	config      StrategicMailerConfig
	templateSvc *TemplateService
	emailWorker *StrategicEmailWorker
	strategies  map[string]IEmailStrategy
}

func NewStrategicMailer(
	config StrategicMailerConfig,
	templateSvc *TemplateService,
	emailWorker *StrategicEmailWorker,
) *StrategicMailer {
	return &StrategicMailer{
		config:      config,
		templateSvc: templateSvc,
		emailWorker: emailWorker,
		strategies:  make(map[string]IEmailStrategy),
	}
}

// RegisterStrategy registers an email strategy
func (m *StrategicMailer) RegisterStrategy(name string, strategy IEmailStrategy) {
	m.strategies[name] = strategy
}

// SendEmail sends an email synchronously using the strategy pattern
func (m *StrategicMailer) SendEmail(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
	strategy, exists := m.strategies[strategyName]
	if !exists {
		return fmt.Errorf("email strategy '%s' not found", strategyName)
	}

	// Validate parameters
	if err := strategy.Validate(params); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	// Prepare template data
	templateData, err := strategy.PrepareData(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to prepare template data: %w", err)
	}

	// Render email content
	htmlBody, err := m.templateSvc.RenderTemplateWithFallback(
		strategy.GetTemplateName(),
		templateData,
		strategy.GetFallbackTemplate(),
	)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email
	return m.sendSMTP(toEmail, strategy.GetSubject(), htmlBody)
}

// SendEmailAsync sends an email asynchronously using the strategy pattern
func (m *StrategicMailer) SendEmailAsync(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
	strategy, exists := m.strategies[strategyName]
	if !exists {
		return fmt.Errorf("email strategy '%s' not found", strategyName)
	}

	// Validate parameters
	if err := strategy.Validate(params); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	// Create email job
	job := EmailJob{
		ToEmail:      toEmail,
		StrategyName: strategyName,
		Params:       params,
	}

	// Send to worker
	select {
	case m.emailWorker.GetJobQueue() <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("email queue is full")
	}
}

// sendSMTP handles the actual SMTP sending
func (m *StrategicMailer) sendSMTP(toEmail, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", m.config.SMTPUser, m.config.SMTPPassword, m.config.SMTPHost)

	// Build message
	message := m.buildMessage(toEmail, subject, htmlBody)

	// Send email
	addr := fmt.Sprintf("%s:%s", m.config.SMTPHost, m.config.SMTPPort)
	return smtp.SendMail(addr, auth, m.config.FromEmail, []string{toEmail}, []byte(message))
}

// buildMessage constructs the email message
func (m *StrategicMailer) buildMessage(toEmail, subject, htmlBody string) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", m.config.FromName, m.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	return message.String()
}

// GetAvailableStrategies returns list of registered strategies
func (m *StrategicMailer) GetAvailableStrategies() []string {
	strategies := make([]string, 0, len(m.strategies))
	for name := range m.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

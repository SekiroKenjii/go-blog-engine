package mailers

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/SekiroKenjii/go-blog-engine/config"
	strategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies"
)

type MailService struct {
	config     *config.EmailConfig
	template   *MailTemplate
	mailWorker *MailWorker
	strategies map[string]strategies.IMailStrategy
}

func NewMailService(
	config *config.EmailConfig,
	template *MailTemplate,
	mailWorker *MailWorker,
) IMailService {
	return &MailService{
		config:     config,
		template:   template,
		mailWorker: mailWorker,
		strategies: make(map[string]strategies.IMailStrategy),
	}
}

// RegisterStrategy registers an email strategy
func (m *MailService) RegisterStrategy(name string, strategy strategies.IMailStrategy) {
	m.strategies[name] = strategy
}

// Send sends an email synchronously
func (m *MailService) Send(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
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
	htmlBody, err := m.template.RenderTemplateWithFallback(
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

// SendAsync sends an email asynchronously
func (m *MailService) SendAsync(ctx context.Context, strategyName, toEmail string, params map[string]any) error {
	strategy, exists := m.strategies[strategyName]
	if !exists {
		return fmt.Errorf("email strategy '%s' not found", strategyName)
	}

	// Validate parameters
	if err := strategy.Validate(params); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	// Create email job
	job := MailJob{
		ToEmail:      toEmail,
		StrategyName: strategyName,
		Params:       params,
	}

	// Send to worker
	select {
	case m.mailWorker.GetJobQueue() <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("email queue is full")
	}
}

// sendSMTP handles the actual SMTP sending
func (m *MailService) sendSMTP(toEmail, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", m.config.SMTPUser, m.config.SMTPPassword, m.config.SMTPHost)

	// Build message
	message := m.buildMessage(toEmail, subject, htmlBody)

	// Send email
	addr := fmt.Sprintf("%s:%s", m.config.SMTPHost, m.config.SMTPPort)
	return smtp.SendMail(addr, auth, m.config.FromEmail, []string{toEmail}, []byte(message))
}

// buildMessage constructs the email message
func (m *MailService) buildMessage(toEmail, subject, htmlBody string) string {
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
func (m *MailService) GetAvailableStrategies() []string {
	strategies := make([]string, 0, len(m.strategies))
	for name := range m.strategies {
		strategies = append(strategies, name)
	}

	return strategies
}

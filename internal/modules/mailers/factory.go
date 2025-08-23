package mailers

import (
	"fmt"

	"github.com/SekiroKenjii/go-blog-engine/config"
)

// MailerFactory creates and configures the mailer system with strategy pattern
type MailerFactory struct {
	config *config.Config
}

func NewMailerFactory(config *config.Config) *MailerFactory {
	return &MailerFactory{config: config}
}

// CreateStrategicMailerSystem creates a complete strategic mailer system
func (f *MailerFactory) CreateStrategicMailerSystem() (*StrategicMailer, *StrategicEmailWorker, error) {
	// Create template service
	templateSvc := NewTemplateService(f.config.Email.TemplateDir)

	// Create strategic email worker with default values (since EmailWorker config doesn't exist)
	emailWorker := NewStrategicEmailWorker(
		nil, // Will be set later
		5,   // Default worker count
		100, // Default queue size
		3,   // Default max retries
	)

	// Create strategic mailer config
	mailerConfig := StrategicMailerConfig{
		SMTPHost:     f.config.Email.SMTPHost,
		SMTPPort:     fmt.Sprintf("%d", f.config.Email.SMTPPort),
		SMTPUser:     f.config.Email.Username,
		SMTPPassword: f.config.Email.Password,
		FromEmail:    f.config.Email.FromEmail,
		FromName:     f.config.Email.FromName,
	}

	// Create strategic mailer
	strategicMailer := NewStrategicMailer(mailerConfig, templateSvc, emailWorker)

	// Set the mailer in the worker
	emailWorker.mailer = strategicMailer

	// Register auth strategies
	f.registerAuthStrategies(strategicMailer)

	return strategicMailer, emailWorker, nil
}

// registerAuthStrategies registers all auth-related email strategies
func (f *MailerFactory) registerAuthStrategies(mailer *StrategicMailer) {
	baseURL := f.getBaseURL()

	// Register verification email strategy
	mailer.RegisterStrategy("verification", NewVerificationEmailStrategy(baseURL))

	// Register password reset email strategy
	mailer.RegisterStrategy("password_reset", NewPasswordResetEmailStrategy(baseURL))

	// Register welcome email strategy
	mailer.RegisterStrategy("welcome", NewWelcomeEmailStrategy(baseURL))
}

// getBaseURL constructs the base URL from config
func (f *MailerFactory) getBaseURL() string {
	// Use default localhost for development
	port := fmt.Sprintf("%d", f.config.Server.Port)
	return fmt.Sprintf("http://localhost:%s", port)
}

// StrategyNames provides constants for strategy names
type StrategyNames struct{}

func (StrategyNames) Verification() string  { return "verification" }
func (StrategyNames) PasswordReset() string { return "password_reset" }
func (StrategyNames) Welcome() string       { return "welcome" }

// Strategies provides easy access to strategy names
var Strategies = StrategyNames{}

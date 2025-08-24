package mailers

import (
	"fmt"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/config"

	authStrategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies/auth"
)

var (
	mailerInstance   *Mailer
	workerInstance   *MailWorker
	mailerOnce       sync.Once
	mailerFactoryErr error
)

// MailerFactory creates and configures the mailer system with strategy pattern
type MailerFactory struct {
	config *config.Config
}

func NewMailerFactory(config *config.Config) *MailerFactory {
	return &MailerFactory{config: config}
}

// CreateMailerSystem creates a complete strategic mailer system using singleton pattern
func (f *MailerFactory) CreateMailerSystem() (*Mailer, *MailWorker, error) {
	mailerOnce.Do(func() {
		mailerInstance, workerInstance, mailerFactoryErr = f.createMailerSystem()
	})

	return mailerInstance, workerInstance, mailerFactoryErr
}

// createMailerSystem is the internal method that actually creates the mailer system
func (f *MailerFactory) createMailerSystem() (*Mailer, *MailWorker, error) {
	templateSvc := NewTemplateService(f.config.Email.TemplateDir)

	mailWorker := NewMailWorker(
		nil, // set later
		5,   // Default worker count
		100, // Default queue size
		3,   // Default max retries
	)

	mailerConfig := MailerConfig{
		SMTPHost:     f.config.Email.SMTPHost,
		SMTPPort:     fmt.Sprintf("%d", f.config.Email.SMTPPort),
		SMTPUser:     f.config.Email.Username,
		SMTPPassword: f.config.Email.Password,
		FromEmail:    f.config.Email.FromEmail,
		FromName:     f.config.Email.FromName,
	}

	mailer := NewMailer(mailerConfig, templateSvc, mailWorker)

	mailWorker.mailer = mailer

	f.registerAuthStrategies(mailer)

	return mailer, mailWorker, nil
}

// GetMailerInstance returns the singleton mailer instance
// This method should be called after CreateStrategicMailerSystem has been called at least once
func GetMailerInstance() *Mailer {
	return mailerInstance
}

// GetWorkerInstance returns the singleton worker instance
// This method should be called after CreateStrategicMailerSystem has been called at least once
func GetWorkerInstance() *MailWorker {
	return workerInstance
}

// ResetSingleton resets the singleton instances (useful for testing)
func ResetSingleton() {
	mailerOnce = sync.Once{}
	mailerInstance = nil
	workerInstance = nil
	mailerFactoryErr = nil
}

// registerAuthStrategies registers all auth-related email strategies
func (f *MailerFactory) registerAuthStrategies(mailer *Mailer) {
	baseURL := f.getBaseURL()

	mailer.RegisterStrategy(Strategies.PasswordReset(), authStrategies.NewPasswordResetEmailStrategy(baseURL))
	mailer.RegisterStrategy(Strategies.Verification(), authStrategies.NewVerificationEmailStrategy(baseURL))
	mailer.RegisterStrategy(Strategies.Welcome(), authStrategies.NewWelcomeEmailStrategy(baseURL))
}

// getBaseURL constructs the base URL from config
func (f *MailerFactory) getBaseURL() string {
	// Use default localhost for development
	port := fmt.Sprintf("%d", f.config.Server.Port)
	return fmt.Sprintf("http://localhost:%s", port)
}

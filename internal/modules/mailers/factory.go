package mailers

import (
	"fmt"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/config"

	authStrategies "github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers/strategies/auth"
)

var (
	mailSvcInstance  IMailService
	workerInstance   *MailWorker
	mailerOnce       sync.Once
	mailerFactoryErr error
)

// MailerFactory creates and configures the mailer system with strategy pattern
type MailerFactory struct {
	config *config.EmailConfig
}

func NewMailerFactory(config *config.EmailConfig) *MailerFactory {
	return &MailerFactory{config: config}
}

// CreateMailerSystem creates a complete strategic mailer system using singleton pattern
func (f *MailerFactory) CreateMailerSystem() (IMailService, *MailWorker, error) {
	mailerOnce.Do(func() {
		mailSvcInstance, workerInstance, mailerFactoryErr = f.createMailerSystem()
	})

	return mailSvcInstance, workerInstance, mailerFactoryErr
}

// createMailerSystem is the internal method that actually creates the mailer system
func (f *MailerFactory) createMailerSystem() (IMailService, *MailWorker, error) {
	template := NewMailTemplate(f.config.TemplateDir)

	mailWorker := NewMailWorker(
		nil, // set later
		f.config.Worker.WorkerCount,
		f.config.Worker.QueueSize,
		f.config.Worker.MaxRetries,
	)

	mailSvc := NewMailService(f.config, template, mailWorker)

	mailWorker.sender = mailSvc

	f.registerAuthStrategies(mailSvc)

	return mailSvc, mailWorker, nil
}

// GetMailServiceInstance returns the singleton mail service instance
// This method should be called after CreateStrategicMailerSystem has been called at least once
func GetMailServiceInstance() IMailService {
	return mailSvcInstance
}

// GetWorkerInstance returns the singleton worker instance
// This method should be called after CreateStrategicMailerSystem has been called at least once
func GetWorkerInstance() *MailWorker {
	return workerInstance
}

// ResetSingleton resets the singleton instances (useful for testing)
func ResetSingleton() {
	mailerOnce = sync.Once{}
	mailSvcInstance = nil
	workerInstance = nil
	mailerFactoryErr = nil
}

// registerAuthStrategies registers all auth-related email strategies
func (f *MailerFactory) registerAuthStrategies(mailSvc IMailService) {
	baseURL := f.getBaseURL()

	mailSvc.RegisterStrategy(Strategies.PasswordReset(), authStrategies.NewPasswordResetEmailStrategy(baseURL))
	mailSvc.RegisterStrategy(Strategies.Verification(), authStrategies.NewVerificationEmailStrategy(baseURL))
	mailSvc.RegisterStrategy(Strategies.Welcome(), authStrategies.NewWelcomeEmailStrategy(baseURL))
}

// getBaseURL constructs the base URL from config
func (f *MailerFactory) getBaseURL() string {
	return fmt.Sprintf("https://%s:%d", f.config.SMTPHost, f.config.SMTPPort)
}

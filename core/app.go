package core

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/cache"
	"github.com/SekiroKenjii/go-blog-engine/internal/modules/mailers"
	"github.com/SekiroKenjii/go-blog-engine/internal/router"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	config     *config.Config
	redis      *redis.Client
	router     abstract.IRouter
	mailWorker *mailers.MailWorker
}

// Bootstrap initializes the application components such as configuration, Redis cache, and router.
// It returns an instance of the Application struct.
// This function is typically called at the start of the application to set up the necessary services.
func Bootstrap() *Application {
	cfg := config.Instance()
	redis := cache.RedisInstance()
	router := router.NewRouter()

	factory := mailers.NewMailerFactory(cfg.Email)
	_, mailWorker, err := factory.CreateMailerSystem()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize mail worker: %v", err))
		panic("Failed to initialize mail worker: " + err.Error())
	}

	mailWorker.Start()

	logger.Info("Application startup complete.")

	return &Application{
		config:     cfg,
		redis:      redis,
		router:     router,
		mailWorker: mailWorker,
	}
}

// BuildHttpServer creates and configures an HTTP server using the application configuration and router.
func (a *Application) BuildHttpServer() *http.Server {
	a.router.Configure()

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      a.router.Engine(),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}
}

// Run starts the HTTP server and listens for incoming requests.
// It blocks until the server is shut down or an error occurs.
// If an error occurs while starting the server, it panics with the error message.
// This function is typically called after the server is built to start handling requests.
func (a *Application) Run(httpSrv *http.Server) {
	err := httpSrv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("HTTP Server error: %s", err))
	}

	logger.Info(fmt.Sprintf("Server is running on port: %d", a.config.Server.Port))
}

// Shutdown gracefully shuts down the HTTP server when a termination signal is received.
// It listens for SIGINT and SIGTERM signals, waits for a maximum of 5 seconds for the server to shut down,
// and then closes the server. If the server cannot shut down normally, it logs an error message.
// After the shutdown process is complete, it sends a signal to the done channel to indicate that the shutdown is complete.
// This function is typically called in a separate goroutine to handle graceful shutdowns of the application.
func (a *Application) Shutdown(httpSrv *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("Shutting down Server...")

	if a.mailWorker != nil {
		logger.Info("Stopping mail worker...")
		a.mailWorker.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server could not shutdown normally: %v", err))
	}

	logger.Warn("Server exiting!")

	done <- true
}

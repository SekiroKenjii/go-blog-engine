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
	"github.com/SekiroKenjii/go-blog-engine/internal/router"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	config *config.Config
	redis  *redis.Client
	router abstract.IRouter
}

func Bootstrap() *Application {
	cfg := config.Instance()
	redis := cache.RedisInstance()
	router := router.NewRouter()

	logger.Info("Application startup complete.")

	return &Application{
		config: cfg,
		redis:  redis,
		router: router,
	}
}

func (a *Application) BuildHttpServer() *http.Server {
	a.router.SetupRoutes()

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      a.router.Engine(),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}
}

func (a *Application) Run(httpSrv *http.Server) {
	err := httpSrv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("HTTP Server error: %s", err))
	}

	logger.Info(fmt.Sprintf("Server is running on port: %d", a.config.Server.Port))
}

func (a *Application) Shutdown(httpSrv *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("Shutting down Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server could not shutdown normally: %v", err))
	}

	logger.Warn("Server exiting!")

	done <- true
}

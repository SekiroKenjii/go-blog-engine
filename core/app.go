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
	"github.com/SekiroKenjii/go-blog-engine/internal/router"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
)

type Application struct {
	Config *config.Config
	Router abstract.IRouter
}

func Bootstrap() *Application {
	cfg := config.Instance()
	router := router.Instance()

	logger.Info("Application startup complete.")

	return &Application{
		Config: cfg,
		Router: router,
	}
}

func (a *Application) BuildHttpServer() *http.Server {
	a.Router.SetupRoutes()

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port),
		Handler:      a.Router.GetEngine(),
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

	logger.Info(fmt.Sprintf("Server is running on port: %d", a.Config.Server.Port))
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

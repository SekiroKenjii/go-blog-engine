package auth

import (
	"context"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/db"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
)

// CleanupWorker handles periodic cleanup of expired tokens
type CleanupWorker struct {
	repo     *db.Repository
	interval time.Duration
	stopCh   chan struct{}
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(interval time.Duration) *CleanupWorker {
	return &CleanupWorker{
		repo:     db.RepositoryInstance(),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the cleanup worker
func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	logger.Info("Auth cleanup worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Auth cleanup worker stopped: context cancelled")
			return
		case <-w.stopCh:
			logger.Info("Auth cleanup worker stopped: stop signal received")
			return
		case <-ticker.C:
			w.performCleanup(ctx)
		}
	}
}

// Stop stops the cleanup worker
func (w *CleanupWorker) Stop() {
	close(w.stopCh)
}

// performCleanup performs the actual cleanup operations
func (w *CleanupWorker) performCleanup(ctx context.Context) {
	logger.Debug("Starting auth token cleanup")

	// Clean up expired password reset tokens
	if err := w.repo.DeleteExpiredPasswordResetTokens(ctx); err != nil {
		logger.Error("Failed to cleanup expired password reset tokens: " + err.Error())
	} else {
		logger.Debug("Successfully cleaned up expired password reset tokens")
	}

	// Note: Refresh tokens cleanup would need additional query
	// For now, we rely on the application logic to handle expired refresh tokens

	logger.Debug("Auth token cleanup completed")
}

package worker

import (
	"context"
	"time"

	"github.com/theHinneh/budgeting/internal/application"
	"github.com/theHinneh/budgeting/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// StartTokenCleanupWorker starts a background worker that periodically cleans up expired refresh tokens
func StartTokenCleanupWorker(authService *application.AuthService) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run once every 24 hours
		defer ticker.Stop()

		logger.Info("Starting token cleanup worker")

		// Run cleanup immediately on startup
		if err := cleanupExpiredTokens(authService); err != nil {
			logger.Error("Failed to cleanup expired tokens on startup", zap.Error(err))
		}

		for {
			select {
			case <-ticker.C:
				if err := cleanupExpiredTokens(authService); err != nil {
					logger.Error("Failed to cleanup expired tokens", zap.Error(err))
				}
			}
		}
	}()
}

func cleanupExpiredTokens(authService *application.AuthService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Starting expired token cleanup")

	if err := authService.CleanupExpiredTokens(ctx); err != nil {
		return err
	}

	logger.Info("Completed expired token cleanup")
	return nil
}

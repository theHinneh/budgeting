package worker

import (
	"context"
	"time"

	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/infrastructure/logger"
	"go.uber.org/zap"
)

func StartRecurringExpenseProcessor(expenseService ports.ExpenseServicePort, userService ports.UserRepository) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run every 24 hours
		defer ticker.Stop()
		for range ticker.C {
			logger.Info("Processing recurring expenses...")
			ctx := context.Background()
			userIDs, err := userService.ListAllUserIDs(ctx)
			if err != nil {
				logger.Error("Failed to list all user IDs for recurring expenses", zap.Error(err))
				continue
			}

			for _, userID := range userIDs {
				processedCount, err := expenseService.ProcessDueExpenses(ctx, userID, time.Now().UTC())
				if err != nil {
					logger.Error("Failed to process due expenses for user", zap.String("userID", userID), zap.Error(err))
					continue
				}
				if processedCount > 0 {
					logger.Info("Processed recurring expenses for user", zap.String("userID", userID), zap.Int("count", processedCount))
				}
			}
			logger.Info("Finished processing recurring expenses.")
		}
	}()
}

func StartRecurringIncomeProcessor(incomeService ports.IncomeServicePort, userService ports.UserRepository) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run every 24 hours
		defer ticker.Stop()
		for range ticker.C {
			logger.Info("Processing due income sources...")
			ctx := context.Background()
			userIDs, err := userService.ListAllUserIDs(ctx)
			if err != nil {
				logger.Error("Failed to list all user IDs for recurring incomes", zap.Error(err))
				continue
			}

			for _, userID := range userIDs {
				processedCount, err := incomeService.ProcessDueIncomes(ctx, userID, time.Now().UTC())
				if err != nil {
					logger.Error("Failed to process due incomes for user", zap.String("userID", userID), zap.Error(err))
					continue
				}
				if processedCount > 0 {
					logger.Info("Processed due incomes for user", zap.String("userID", userID), zap.Int("count", processedCount))
				}
			}
			logger.Info("Finished processing due income sources.")
		}
	}()
}

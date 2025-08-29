package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theHinneh/budgeting/internal/application"
	api_http "github.com/theHinneh/budgeting/internal/infrastructure/api/http"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	fbdb "github.com/theHinneh/budgeting/internal/infrastructure/db/firebase"
	"github.com/theHinneh/budgeting/internal/infrastructure/logger"
	"github.com/theHinneh/budgeting/internal/worker"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.InitZaplogger(cfg)

	dbConfig := cfg.GetDatabaseConfig()

	var fbInstance *fbdb.Database

	switch dbConfig.Driver {
	case "firebase":
		logger.Info("Initializing Firebase database adapter")
		fbInstance, err = fbdb.NewDatabase(context.Background(), cfg)
		if err != nil {
			logger.Fatal("Failed to initialize firebase", zap.Error(err))
		}
	default:
		logger.Fatal("Unsupported DB_DRIVER. Only 'firebase' is supported.")
	}

	defer func() {
		if fbInstance != nil {
			if err := fbInstance.Close(); err != nil {
				logger.Error("Failed to close database", zap.Error(err))
			}
		}
	}()

	healthHandler := api_http.NewHealthHandler(cfg, fbInstance.FirestoreClient)

	userService := application.NewUserService(
		fbInstance.UserRepository,
		fbInstance.UserAuthenticator,
	)
	incomeService := application.NewIncomeService(
		fbInstance.IncomeRepository,
	)
	expenseService := application.NewExpenseService(
		fbInstance.ExpenseRepository,
	)
	netWorthService := application.NewNetWorthService(
		fbInstance.IncomeRepository,
		fbInstance.ExpenseRepository,
	)

	authService := application.NewAuthService(
		fbInstance.RefreshTokenRepository,
		fbInstance.TokenAuthenticator,
		fbInstance.TokenGenerator,
	)

	router := api_http.NewRouter(
		healthHandler, userService, incomeService, expenseService, netWorthService, fbInstance.App, authService, cfg,
	)

	serverConfig := cfg.GetServerConfig()
	port := serverConfig.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("Starting server on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Start background workers
	worker.StartRecurringExpenseProcessor(expenseService, fbInstance.UserRepository)
	worker.StartRecurringIncomeProcessor(incomeService, fbInstance.UserRepository)
	worker.StartTokenCleanupWorker(authService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Println("Server exiting")
}

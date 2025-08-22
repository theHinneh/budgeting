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
	"github.com/theHinneh/budgeting/internal/application/ports"
	http3 "github.com/theHinneh/budgeting/internal/infrastructure/api/http"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	fbdb "github.com/theHinneh/budgeting/internal/infrastructure/db/firebase"
	"github.com/theHinneh/budgeting/internal/infrastructure/logger"
	"go.uber.org/zap"
)

func main() {
	logger.InitZaplogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	dbConfig := cfg.GetDatabaseConfig()

	var dbPort ports.DatabasePort
	var fbInstance *fbdb.Database

	switch dbConfig.Driver {
	case "firebase":
		logger.Info("Initializing Firebase database adapter")
		fbInstance, err = fbdb.NewDatabase(context.Background(), cfg)
		if err != nil {
			logger.Fatal("Failed to initialize firebase", zap.Error(err))
		}
		dbPort = fbInstance
		// No migrations for Firebase
	default:
		logger.Fatal("Unsupported DB_DRIVER. Only 'firebase' is supported.")
	}

	defer func() {
		if dbPort != nil {
			if err := dbPort.Close(); err != nil {
				logger.Error("Failed to close database", zap.Error(err))
			}
		}
	}()

	healthHandler := http3.NewHealthHandler(cfg, dbPort)

	userService := application.NewUserService(fbInstance)
	incomeService := application.NewIncomeService(fbInstance)
	expenseService := application.NewExpenseService(fbInstance)

	routes := http3.NewRouter(healthHandler, userService, incomeService, expenseService)

	port := cfg.V.GetString("SERVER_PORT")
	if port == "" {
		port = cfg.V.GetString("server.port")
	}
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: routes,
	}

	go func() {
		logger.Info("Starting server on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http2 "github.com/theHinneh/budgeting/internal/adapters/api/http"
	"github.com/theHinneh/budgeting/internal/adapters/db"
	fbdb "github.com/theHinneh/budgeting/internal/adapters/db/firebase"
	"github.com/theHinneh/budgeting/internal/adapters/db/postgres"
	"github.com/theHinneh/budgeting/internal/core/ports"
	"github.com/theHinneh/budgeting/pkg/config"
	"github.com/theHinneh/budgeting/pkg/logger"
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
	case "postgres":
		logger.Info("Initializing Postgres database adapter")
		pgInstance, err := postgres.NewDatabase(context.Background(), dbConfig)
		if err != nil {
			logger.Fatal("Failed to initialize database", zap.Error(err))
		}
		dbPort = pgInstance

		migration := db.Migrations{
			DB:     pgInstance,
			Models: db.GetModels(),
		}
		db.RunMigrations(migration)
	default:
		logger.Fatal("Unsupported DB_DRIVER. Use 'postgres' or 'firebase'")
	}

	defer func() {
		if dbPort != nil {
			if err := dbPort.Close(); err != nil {
				logger.Error("Failed to close database", zap.Error(err))
			}
		}
	}()

	healthHandler := http2.NewHealthHandler(cfg, dbPort)
	routes := http2.NewRouter(healthHandler, fbInstance)

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

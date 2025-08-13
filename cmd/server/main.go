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
	"github.com/theHinneh/budgeting/internal/adapters/db/postgres"
	"github.com/theHinneh/budgeting/pkg"
	"go.uber.org/zap"
)

func main() {
	pkg.InitZaplogger()

	cfg, err := pkg.Load()
	if err != nil {
		pkg.Fatal("Failed to load configuration", zap.Error(err))
	}

	dbConfig := cfg.GetDatabaseConfig()

	dbInstance, err := postgres.NewDatabase(context.Background(), dbConfig)
	if err != nil {
		pkg.Fatal("Failed to initialize database", zap.Error(err))
	}

	defer func() {
		if err := dbInstance.Close(); err != nil {
			pkg.Error("Failed to close database", zap.Error(err))
		}
	}()

	migration := db.Migrations{
		DB:     dbInstance,
		Models: db.GetModels(),
	}
	db.RunMigrations(migration)

	healthHandler := http2.NewHealthHandler(cfg, dbInstance)
	routes := http2.NewRouter(healthHandler)

	port := cfg.V.GetString("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: routes,
	}

	go func() {
		pkg.Info("Starting server on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	pkg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		pkg.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Println("Server exiting")
}

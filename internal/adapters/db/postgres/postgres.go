package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/theHinneh/budgeting/internal/core/ports"
	"github.com/theHinneh/budgeting/pkg/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

var DB *Database

func NewDatabase(ctx context.Context, cfg config.DatabaseConfig) (ports.DatabasePort, error) {
	if err := validateDBConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid db config: %w", err)
	}

	connector := &PostgresConnector{}
	db, err := connectWithRetry(ctx, connector, &cfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *Database) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.DB.WithContext(ctx).Transaction(fn)
}

type PostgresConnector struct{}

func (pc *PostgresConnector) ConnectWithContext(ctx context.Context, cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	gormConfig := getDefaultGormConfig()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func connectWithRetry(ctx context.Context, connector *PostgresConnector, cfg *config.DatabaseConfig) (*Database, error) {
	maxRetries := 5
	retryDelay := time.Second
	var db *Database

	for i := 0; i < maxRetries; i++ {
		connCtx, cancel := context.WithTimeout(ctx, cfg.ConnTimeout)
		gormDB, err := connector.ConnectWithContext(connCtx, cfg)
		cancel()

		if err == nil {
			db = &Database{DB: gormDB}
			return db, nil
		}

		if i < maxRetries-1 {
			zap.S().Errorf("DB connect attempt %d failed: %v. Retrying in %v", i+1, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		} else {
			return nil, fmt.Errorf("failed to connect to db after %d attempts: %w", maxRetries, err)
		}
	}

	return nil, fmt.Errorf("failed to connect to db")
}

func validateDBConfig(cfg *config.DatabaseConfig) error {
	if cfg.Driver != "postgres" {
		return fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
	if cfg.Host == "" || cfg.Port == "" || cfg.Name == "" || cfg.User == "" {
		return fmt.Errorf("missing required postgres config fields")
	}
	return nil
}

func getDefaultGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	}
}

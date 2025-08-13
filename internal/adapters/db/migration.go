package db

import (
	"github.com/theHinneh/budgeting/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migrations struct {
	DB     DatabasePort
	Models []interface{}
}

type DatabasePort interface {
	AutoMigrate(models ...interface{}) error
	GetDB() *gorm.DB
}

func GetModels() []interface{} {
	return []interface{}{}
}

func RunMigrations(m Migrations) {
	logger.Info("Starting database migrations")

	if len(m.Models) == 0 {
		m.Models = GetModels()
	}

	err := m.DB.AutoMigrate(m.Models...)
	if err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	err = runCustomMigrations(m.DB.GetDB())
	if err != nil {
		logger.Fatal("Failed to run custom migrations", zap.Error(err))
	}

	logger.Info("Database migrations completed successfully")
}

// Custom migrations for more complex operations
func runCustomMigrations(db *gorm.DB) error {
	// Example:
	// return db.Exec("CREATE INDEX idx_users_email ON users(email)").Error
	return nil
}

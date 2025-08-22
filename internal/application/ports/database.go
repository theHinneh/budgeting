package ports

import (
	"context"
)

type DatabasePort interface {
	AutoMigrate(models ...interface{}) error
	Close() error
	Ping(ctx context.Context) error
}

type Database struct {
}

func (d *Database) AutoMigrate(models ...interface{}) error {
	return nil // No GORM, so no migration
}

func (d *Database) Close() error {
	return nil // No GORM, so no close
}

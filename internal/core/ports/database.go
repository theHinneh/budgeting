package ports

import (
	"context"

	"gorm.io/gorm"
)

type DatabasePort interface {
	AutoMigrate(models ...interface{}) error
	GetDB() *gorm.DB
	Close() error
	Ping(ctx context.Context) error
}

type Database struct {
	DB *gorm.DB
}

func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

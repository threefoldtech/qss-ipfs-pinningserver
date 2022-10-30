package database

import (
	"github.com/threefoldtech/tf-pinning-service/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg config.DbConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(cfg.LogLevel)),
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{}, &PinDTO{})
	return db, nil
}

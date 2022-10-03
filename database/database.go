package database

import (
	"github.com/threefoldtech/tf-pinning-service/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() error {
	if DB == nil {
		db, err := gorm.Open(sqlite.Open(config.CFG.Db.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(config.CFG.Db.LogLevel)),
		})
		if err != nil {
			panic("Failed to connect to database!")
		}
		db.AutoMigrate(&User{}, &PinDTO{})
		DB = db
	}
	return nil
}

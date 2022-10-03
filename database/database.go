package database

import (
	svc_logger "github.com/threefoldtech/tf-pinning-service/logger"

	"github.com/threefoldtech/tf-pinning-service/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() error {
	log := svc_logger.GetDefaultLogger()
	if DB == nil {
		db, err := gorm.Open(sqlite.Open(config.CFG.Db.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(config.CFG.Db.LogLevel)),
		})
		if err != nil {
			log.WithFields(svc_logger.Fields{
				"topic":      "Database",
				"from_error": err.Error(),
			}).Panic("Failed to connect to database!")
		}
		db.AutoMigrate(&User{}, &PinDTO{})
		DB = db
	}
	return nil
}

package database

import (
	"github.com/threefoldtech/tf-pinning-service/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	db, err := gorm.Open(sqlite.Open(config.CFG.Db.DSN), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	db.AutoMigrate(&User{}, &PinDTO{})
	DB = db
	return nil
}

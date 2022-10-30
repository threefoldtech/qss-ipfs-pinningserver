package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
)

func loadConfigFromEnv() (config.Config, error) {
	var cfg config.Config

	database_dsn, ok := os.LookupEnv("TFPIN_DB_DSN")
	if !ok {
		database_dsn = "pins.db"
	}
	database_log_level, ok := os.LookupEnv("TFPIN_DB_LOG_LEVEL")
	if !ok {
		database_log_level = "1"
	}

	database_ll_int, err := strconv.Atoi(database_log_level)
	if err != nil || database_ll_int < 1 || database_ll_int > 4 {
		return cfg, errors.New("`TFPIN_DB_LOG_LEVEL` set to invalid value")
	}
	dbc := config.DbConfig{
		DSN:      database_dsn,
		LogLevel: database_ll_int,
	}

	cfg = config.Config{
		Db: dbc,
	}
	return cfg, nil
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("You should provide at lease one command line arguments.")
		fmt.Println("USAGE: ./create_test_users [<UserToken1> ..]")
		os.Exit(1)
	}
	cfg, err := loadConfigFromEnv()
	if err != nil {
		panic(err)
	}
	db, err := database.NewDatabase(cfg.Db)
	if err != nil {
		panic(err)
	}
	usersRepo := database.GetUsersRepository(db)
	ctx := context.Background()
	for _, arg := range os.Args[1:] {
		err := usersRepo.Insert(ctx, arg)
		if err != nil {
			fmt.Println("Can't store new user with the given token.")
			fmt.Println(err.Error())
		} else {
			fmt.Printf("Token `%v` stored successfully.\n", arg)
		}
	}
}

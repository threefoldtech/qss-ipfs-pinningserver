package main

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	db "github.com/threefoldtech/tf-pinning-service/database"
)

func main() {
	config.LoadConfig()
	err := db.ConnectDatabase()
	if err != nil {
		fmt.Print(err)
	}
	usersRepo := database.NewUsersRepository()
	ctx := context.Background()
	usersRepo.Insert(ctx, "MyTestToken")
	usersRepo.Insert(ctx, "MySecretToken")
}

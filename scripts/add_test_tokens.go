package main

import (
	"context"
	"fmt"
	"os"

	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	db "github.com/threefoldtech/tf-pinning-service/database"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("You should provide at lease one command line arguments.")
		fmt.Println("USAGE: ./create_test_users [<UserToken1> ..]")

	}
	config.LoadConfig()
	err := db.ConnectDatabase()
	if err != nil {
		fmt.Print(err)
	}
	usersRepo := database.NewUsersRepository()
	ctx := context.Background()
	for _, arg := range os.Args[1:] {
		usersRepo.Insert(ctx, arg)
	}
}

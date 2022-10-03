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
		os.Exit(1)
	}
	config.LoadConfig()
	err := db.ConnectDatabase()
	if err != nil {
		panic(err)
	}
	usersRepo := database.NewUsersRepository()
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

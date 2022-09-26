/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"log"

	"github.com/threefoldtech/tf-pinning-service/config"
	db "github.com/threefoldtech/tf-pinning-service/database"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
)

func main() {
	log.Printf("Server started")
	config.LoadConfig()
	db.ConnectDatabase()

	router := sw.NewRouter()
	log.Fatal(router.Run(config.CFG.Server.Addr))
}

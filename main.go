/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"log"

	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
	"github.com/threefoldtech/tf-pinning-service/services"
)

func main() {
	log.Printf("Server started")
	config.LoadConfig()
	database.ConnectDatabase()
	services.SetSyncService()
	services.StartInBackground()
	router := sw.NewRouter()
	log.Fatal(router.Run(config.CFG.Server.Addr))
}

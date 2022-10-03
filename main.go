/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	"github.com/threefoldtech/tf-pinning-service/logger"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
	"github.com/threefoldtech/tf-pinning-service/services"
)

func main() {
	log := logger.GetDefaultLogger()
	log.WithFields(logger.Fields{
		"topic": "Main",
	}).Info("Server started")
	config.LoadConfig()
	database.ConnectDatabase()
	services.SetSyncService()
	services.StartInBackground()
	router := sw.NewRouter()
	log.Fatal(router.Run(config.CFG.Server.Addr))
}

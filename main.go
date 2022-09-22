/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"log"

	db "github.com/threefoldtech/tf-pinning-service/database"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()
	db.ConnectDatabase()
	log.Fatal(router.Run(":8000"))
}

/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"log"

	// WARNING!
	// Change this to a fully-qualified import path
	// once you place this file into your project.
	// For example,
	//
	//sw "github.com/GIT_USER_ID/GIT_REPO_ID/go"
	//
	sw "github.com/threefoldtech/tf-pinning-service/go"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()

	log.Fatal(router.Run(":8080"))
}

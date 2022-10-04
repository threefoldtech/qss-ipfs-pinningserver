/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	"github.com/threefoldtech/tf-pinning-service/logger"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
	"github.com/threefoldtech/tf-pinning-service/services"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.GetDefaultLogger()
	loggerContext := log.WithFields(logger.Fields{
		"topic": "Server",
	})
	loggerContext.Info("Server started")
	config.LoadConfig()
	database.ConnectDatabase()
	services.SetSyncService(10) // for now run every 10 minutes
	services.StartInBackground()
	router := sw.NewRouter()
	// log.Fatal(router.Run(config.CFG.Server.Addr))
	srv := &http.Server{
		Addr:    config.CFG.Server.Addr,
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerContext.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	loggerContext.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		loggerContext.Fatal("Server forced to shutdown: ", err)
	}

	loggerContext.Println("Server exiting")
}

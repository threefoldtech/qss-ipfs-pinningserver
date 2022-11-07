/*
 * IPFS Pinning Service API
 *
 */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/morikuni/aec"
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	"github.com/threefoldtech/tf-pinning-service/logger"
	sw "github.com/threefoldtech/tf-pinning-service/pinning-api/controller"
	"github.com/threefoldtech/tf-pinning-service/services"
)

var (
	buildTime string
	version   string
)

const tfpinFigletStr = `
_________ _______  _______ _________ _       
\__   __/(  ____ \(  ____ )\__   __/( (    /|
   ) (   | (    \/| (    )|   ) (   |  \  ( |
   | |   | (__    | (____)|   | |   |   \ | |
   | |   |  __)   |  _____)   | |   | (\ \) |
   | |   | (      | (         | |   | | \   |
   | |   | )      | )      ___) (___| )  \  |
   )_(   |/       |/       \_______/|/    )_)
                                             
`

func printASCIIArt() {
	tfpinLogo := aec.LightGreenF.Apply(tfpinFigletStr)
	fmt.Println(tfpinLogo)
}

func LoadConfigFromEnv() (config.Config, error) {
	var cfg config.Config
	cluster_host, ok := os.LookupEnv("TFPIN_CLUSTER_HOSTNAME")
	if !ok {
		cluster_host = "127.0.0.1"
	}
	cluster_port, ok := os.LookupEnv("TFPIN_CLUSTER_PORT")
	if !ok {
		cluster_port = "9097"
	}
	cluster_username := os.Getenv("TFPIN_CLUSTER_USERNAME")

	cluster_password := os.Getenv("TFPIN_CLUSTER_PASSWORD")
	cluster_timeout, ok := os.LookupEnv("TFPIN_CLUSTER_TIMEOUT")
	var cluster_timeout_int int
	if !ok {
		cluster_timeout_int = 10
	} else {
		cluster_timeout_int, err := strconv.Atoi(cluster_timeout)
		if err != nil || cluster_timeout_int < 5 {
			return cfg, errors.New("`TFPIN_CLUSTER_TIMEOUT` set to invalid value")
		}
	}
	cluster_replication_min, ok := os.LookupEnv("TFPIN_CLUSTER_REPLICA_MIN")
	var cluster_replica_min_int int
	if !ok {
		cluster_replica_min_int = -1
	} else {
		cluster_replica_min_int, err := strconv.Atoi(cluster_replication_min)
		if err != nil || cluster_replica_min_int < 1 {
			return cfg, errors.New("`TFPIN_CLUSTER_REPLICA_MIN` set to invalid value")
		}
	}

	cluster_replication_max, ok := os.LookupEnv("TFPIN_CLUSTER_REPLICA_MAX")
	var cluster_replica_max_int int
	if !ok {
		cluster_replica_max_int = -1
	} else {
		cluster_replica_max_int, err := strconv.Atoi(cluster_replication_max)
		if err != nil || cluster_replica_max_int < cluster_replica_min_int {
			return cfg, errors.New("`TFPIN_CLUSTER_REPLICA_MAX` set to invalid value")
		}
	}

	database_dsn, ok := os.LookupEnv("TFPIN_DB_DSN")
	if !ok {
		database_dsn = "pins.db"
	}
	database_log_level, ok := os.LookupEnv("TFPIN_DB_LOG_LEVEL")
	if !ok {
		database_log_level = "1"
	}
	server_addr, ok := os.LookupEnv("TFPIN_SERVER_ADDR")
	if !ok {
		server_addr = ":8000"
	}
	server_log_level, ok := os.LookupEnv("TFPIN_SERVER_LOG_LEVEL")
	if !ok {
		server_log_level = "3"
	}
	auth_header_key, ok := os.LookupEnv("TFPIN_AUTH_HEADER_KEY")
	if !ok {
		auth_header_key = "Authorization"
	}
	auth_admin_username, ok := os.LookupEnv("TFPIN_AUTH_ADMIN_USERNAME")
	if !ok {
		return cfg, errors.New("`TFPIN_AUTH_ADMIN_USERNAME` need to be set")
	}
	auth_admin_password, ok := os.LookupEnv("TFPIN_AUTH_ADMIN_PASSWORD")
	if !ok {
		return cfg, errors.New("`TFPIN_AUTH_ADMIN_PASSWORD` need to be set")
	}

	cc := config.ClusterConfig{
		Host:                 cluster_host,
		Port:                 cluster_port,
		Username:             cluster_username,
		Password:             cluster_password,
		ReplicationFactorMin: cluster_replica_min_int,
		ReplicationFactorMax: cluster_replica_max_int,
		IpfsClusterTimeout:   cluster_timeout_int,
	}
	database_ll_int, err := strconv.Atoi(database_log_level)
	if err != nil || database_ll_int < 1 || database_ll_int > 4 {
		return cfg, errors.New("`TFPIN_DB_LOG_LEVEL` set to invalid value")
	}
	dbc := config.DbConfig{
		DSN:      database_dsn,
		LogLevel: database_ll_int,
	}
	server_ll_int, err := strconv.Atoi(server_log_level)
	if err != nil || server_ll_int < 0 || server_ll_int > 6 {
		return config.Config{}, errors.New("`TFPIN_SERVER_LOG_LEVEL` set to invalid value")
	}
	sc := config.ServerConfig{
		Addr: server_addr,
	}
	ac := config.AuthConfig{
		ApiKeyHeader:  auth_header_key,
		AdminUserName: auth_admin_username,
		AdminPassword: auth_admin_password,
	}
	lc := config.LoggerConfig{
		LogLevel: server_ll_int,
	}
	cfg = config.Config{
		Cluster: cc,
		Db:      dbc,
		Server:  sc,
		Auth:    ac,
		Logger:  lc,
	}
	return cfg, nil
}

func main() {
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		printASCIIArt()
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		panic("Can't load the service config. caused by\n Error: " + err.Error())
	}
	log := logger.GetDefaultLogger(cfg.Logger)
	loggerContext := log.WithFields(logger.Fields{
		"topic": "Server",
	})
	loggerContext.Info("Config loaded, Server starting..")
	db, err := database.NewDatabase(cfg.Db)
	if err != nil {
		panic("Failed to connect to database. caused by\n Error: " + err.Error())
	}
	pins_repo := database.GetPinsRepository(db)
	users_repo := database.GetUsersRepository(db)
	services.SetSyncService(1, log, pins_repo, cfg.Cluster) // for now run every 10 minutes
	services.SetDagService(1, log, pins_repo, cfg.Cluster)  // for now run every 10 minutes
	services.StartInBackground()
	handlers := &sw.Handlers{
		Log:       log,
		PinsRepo:  pins_repo,
		UsersRepo: users_repo,
		Config:    cfg,
	}
	router := sw.NewRouter(handlers)
	// log.Fatal(router.Run(config.CFG.Server.Addr))
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		loggerContext.Fatal("Server forced to shutdown: ", err)
	}

	loggerContext.Println("Server exiting")
}

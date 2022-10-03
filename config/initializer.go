package config

import (
	"os"
	"strconv"

	"github.com/threefoldtech/tf-pinning-service/logger"
)

var CFG Config

type authConfig struct {
	ApiKeyHeader string // ex. "Authorization"
}

type clusterConfig struct {
	Host string
	Port string
}

type serverConfig struct {
	Addr string // ex. ":8080" or "0.0.0.0:8000"
}

type dbConfig struct {
	DSN      string // ex. "pins.db" for sqlite
	LogLevel int    // could be 1 to 4, meaning Silent, Error, Warn, Info
}

type Config struct {
	Auth    authConfig
	Cluster clusterConfig
	Server  serverConfig
	Db      dbConfig
}

func LoadConfig() {
	cluster_host, ok := os.LookupEnv("TFPIN_CLUSTER_HOSTNAME")
	if !ok {
		panic("`TFPIN_CLUSTER_HOSTNAME` Not present in the environment!\nPlease make sure to set all required environment variables.")
	}
	cluster_port, ok := os.LookupEnv("TFPIN_CLUSTER_PORT")
	if !ok {
		panic("`TFPIN_CLUSTER_PORT` Not present in the environment!\nPlease make sure to set all required environment variables.")
	}
	database_dsn, ok := os.LookupEnv("TFPIN_DB_DSN")
	if !ok {
		panic("`TFPIN_DB_DSN` Not present in the environment!\nPlease make sure to set all required environment variables.")
	}
	database_log_level, ok := os.LookupEnv("TFPIN_DB_LOG_LEVEL")
	if !ok {
		database_log_level = "1"
	}
	server_addr, ok := os.LookupEnv("TFPIN_SERVER_ADDR")
	if !ok {
		panic("`TFPIN_SERVER_ADDR` Not present in the environment!\nPlease make sure to set all required environment variables.")
	}
	auth_header_key, ok := os.LookupEnv("TFPIN_AUTH_HEADER_KEY")
	if !ok {
		panic("`TFPIN_AUTH_HEADER_KEY` Not present in the environment!\nPlease make sure to set all required environment variables.")
	}
	cc := clusterConfig{
		Host: cluster_host,
		Port: cluster_port,
	}
	database_ll_int, err := strconv.Atoi(database_log_level)
	if err != nil || database_ll_int < 0 || database_ll_int > 4 {
		panic("`TFPIN_DB_LOG_LEVEL` set to invalid value!")
	}
	dbc := dbConfig{
		DSN:      database_dsn,
		LogLevel: database_ll_int,
	}
	sc := serverConfig{
		Addr: server_addr,
	}
	ac := authConfig{
		ApiKeyHeader: auth_header_key,
	}
	CFG = Config{
		Cluster: cc,
		Db:      dbc,
		Server:  sc,
		Auth:    ac,
	}
	log := logger.GetDefaultLogger()
	log.WithFields(logger.Fields{
		"topic": "Config",
	}).Info("Configurations have been successfully loaded")
}

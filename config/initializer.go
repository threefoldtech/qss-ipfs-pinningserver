package config

import (
	"errors"
	"os"
	"strconv"
)

var CFG Config

type authConfig struct {
	ApiKeyHeader string // ex. "Authorization"
}

type clusterConfig struct {
	Host                 string
	Port                 string
	Username             string
	Password             string
	ReplicationFactorMin int
	ReplicationFactorMax int
}

type serverConfig struct {
	Addr     string // ex. ":8080" or "0.0.0.0:8000"
	LogLevel int    // could be 0 to 6, meaning PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel
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

func LoadConfig() error {
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

	cluster_replication_min, ok := os.LookupEnv("TFPIN_CLUSTER_REPLICA_MIN")
	var cluster_replica_min_int int
	if !ok {
		cluster_replica_min_int = -1
	} else {
		cluster_replica_min_int, err := strconv.Atoi(cluster_replication_min)
		if err != nil || cluster_replica_min_int < 1 {
			return errors.New("`TFPIN_CLUSTER_REPLICA_MIN` set to invalid value")
		}
	}

	cluster_replication_max, ok := os.LookupEnv("TFPIN_CLUSTER_REPLICA_MAX")
	var cluster_replica_max_int int
	if !ok {
		cluster_replica_max_int = -1
	} else {
		cluster_replica_max_int, err := strconv.Atoi(cluster_replication_max)
		if err != nil || cluster_replica_max_int < cluster_replica_min_int {
			return errors.New("`TFPIN_CLUSTER_REPLICA_MAX` set to invalid value")
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

	cc := clusterConfig{
		Host:                 cluster_host,
		Port:                 cluster_port,
		Username:             cluster_username,
		Password:             cluster_password,
		ReplicationFactorMin: cluster_replica_min_int,
		ReplicationFactorMax: cluster_replica_max_int,
	}
	database_ll_int, err := strconv.Atoi(database_log_level)
	if err != nil || database_ll_int < 1 || database_ll_int > 4 {
		return errors.New("`TFPIN_DB_LOG_LEVEL` set to invalid value")
	}
	dbc := dbConfig{
		DSN:      database_dsn,
		LogLevel: database_ll_int,
	}
	server_ll_int, err := strconv.Atoi(server_log_level)
	if err != nil || server_ll_int < 0 || server_ll_int > 6 {
		return errors.New("`TFPIN_SERVER_LOG_LEVEL` set to invalid value")
	}
	sc := serverConfig{
		Addr:     server_addr,
		LogLevel: server_ll_int,
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
	return nil
}

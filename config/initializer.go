package config

import (
	"os"
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
	DSN string // ex. "pins.db" for sqlite
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
	dbc := dbConfig{
		DSN: database_dsn,
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
}

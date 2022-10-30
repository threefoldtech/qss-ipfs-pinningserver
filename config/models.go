package config

var CFG Config

type AuthConfig struct {
	ApiKeyHeader string // ex. "Authorization"
}

type ClusterConfig struct {
	Host                 string
	Port                 string
	Username             string
	Password             string
	ReplicationFactorMin int
	ReplicationFactorMax int
	IpfsClusterTimeout   int
}

type ServerConfig struct {
	Addr string // ex. ":8080" or "0.0.0.0:8000"
}

type DbConfig struct {
	DSN      string // ex. "pins.db" for sqlite
	LogLevel int    // could be 1 to 4, meaning Silent, Error, Warn, Info
}

type LoggerConfig struct {
	LogLevel int // could be 0 to 6, meaning PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel
}

type Config struct {
	Auth    AuthConfig
	Cluster ClusterConfig
	Server  ServerConfig
	Db      DbConfig
	Logger  LoggerConfig
}

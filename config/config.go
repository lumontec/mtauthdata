package config

type ServerConfig struct {
	Upstreamurl        string
	ExposedPort        string
	PostgresConfig     string
	Opaurl             string
	HttpCallTimeoutSec int64
}

type LoggerConfig struct {
	LevelModules      string
	EnableJSONLogging bool
	DisableAllLogging bool
	DisableStackTrace bool
	DisableCaller     bool
	Development       bool
}

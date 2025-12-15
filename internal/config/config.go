package config

// Config holds application configuration values.
type Config struct {
	Port         string
	DBDriver     string
	DBDSN        string
	LogLevel     string
	OTLPEndpoint string
}

// Load returns configuration with defaults.
func Load() Config {
	return Config{
		Port:         "8080",
		DBDriver:     "postgres",
		DBDSN:        "",
		LogLevel:     "info",
		OTLPEndpoint: "",
	}
}

package config

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	defaultLogLevel = "info"
	configFilePath  = "config.yaml"
)

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Driver             string
	Host               string
	Port               string
	Name               string
	User               string
	Password           string
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
	ConnTimeout        time.Duration
	PgSchema           string
	PgSearchPath       string
	PgSSLCert          string
	PgSSLKey           string
	PgSSLRootCert      string
}

type Configuration struct {
	V *viper.Viper
}

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
}

// Load reads configuration from the config file and environment variables
func Load() (*Configuration, error) {
	LoadEnv()

	v := viper.New()
	v.SetDefault("log.level", defaultLogLevel)
	v.SetConfigFile(configFilePath)

	if err := v.ReadInConfig(); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			log.Printf("Config file %s not found, using environment variables only", configFilePath)
		} else {
			return nil, err
		}
	}

	v.AutomaticEnv()

	return &Configuration{V: v}, nil
}

// GetDatabaseConfig extracts database configuration. It supports both ENV-style keys
// (e.g., DB_DRIVER) and YAML nested keys (e.g., database.driver) as fallbacks.
func (c *Configuration) GetDatabaseConfig() DatabaseConfig {
	getStr := func(primary string, fallbacks ...string) string {
		if v := c.V.GetString(primary); v != "" {
			return v
		}
		for _, fb := range fallbacks {
			if v := c.V.GetString(fb); v != "" {
				return v
			}
		}
		return ""
	}
	getInt := func(primary string, fallbacks ...string) int {
		if v := c.V.GetInt(primary); v != 0 {
			return v
		}
		for _, fb := range fallbacks {
			if v := c.V.GetInt(fb); v != 0 {
				return v
			}
		}
		return 0
	}
	getDur := func(primary string, fallbacks ...string) time.Duration {
		if v := c.V.GetDuration(primary); v != 0 {
			return v
		}
		for _, fb := range fallbacks {
			if v := c.V.GetDuration(fb); v != 0 {
				return v
			}
		}
		return 0
	}

	return DatabaseConfig{
		Driver:             getStr("DB_DRIVER", "database.driver"),
		Host:               getStr("DB_HOST", "database.host"),
		Port:               getStr("DB_PORT", "database.port"),
		Name:               getStr("DB_NAME", "database.name"),
		User:               getStr("DB_USER", "database.user"),
		Password:           getStr("DB_PASSWORD", "database.password"),
		SSLMode:            getStr("DB_SSLMODE", "database.sslmode"),
		MaxOpenConnections: getInt("DB_MAX_OPEN_CONNS", "database.max_open_connections"),
		MaxIdleConnections: getInt("DB_MAX_IDLE_CONNS", "database.max_idle_connections"),
		ConnMaxLifetime:    getDur("DB_CONN_MAX_LIFETIME", "database.connection_max_lifetime"),
		ConnTimeout:        getDur("DB_CONN_TIMEOUT", "database.connection_timeout"),
		PgSchema:           getStr("DB_PG_SCHEMA", "database.postgres.schema"),
		PgSearchPath:       getStr("DB_PG_SEARCH_PATH", "database.postgres.search_path"),
		PgSSLCert:          getStr("DB_PG_SSLCERT", "database.postgres.sslcert"),
		PgSSLKey:           getStr("DB_PG_SSLKEY", "database.postgres.sslkey"),
		PgSSLRootCert:      getStr("DB_PG_SSLROOTCERT", "database.postgres.sslrootcert"),
	}
}

// GetPathToConfig returns the path of the config file in use
func (c *Configuration) GetPathToConfig() string {
	return c.V.ConfigFileUsed()
}

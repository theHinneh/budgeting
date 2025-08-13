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

// GetDatabaseConfig extracts PostgreSQL configuration from the loaded config
func (c *Configuration) GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:             c.V.GetString("DB_DRIVER"),
		Host:               c.V.GetString("DB_HOST"),
		Port:               c.V.GetString("DB_PORT"),
		Name:               c.V.GetString("DB_NAME"),
		User:               c.V.GetString("DB_USER"),
		Password:           c.V.GetString("DB_PASSWORD"),
		SSLMode:            c.V.GetString("DB_SSLMODE"),
		MaxOpenConnections: c.V.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConnections: c.V.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime:    c.V.GetDuration("DB_CONN_MAX_LIFETIME"),
		ConnTimeout:        c.V.GetDuration("DB_CONN_TIMEOUT"),
		PgSchema:           c.V.GetString("DB_PG_SCHEMA"),
		PgSearchPath:       c.V.GetString("DB_PG_SEARCH_PATH"),
		PgSSLCert:          c.V.GetString("DB_PG_SSLCERT"),
		PgSSLKey:           c.V.GetString("DB_PG_SSLKEY"),
		PgSSLRootCert:      c.V.GetString("DB_PG_SSLROOTCERT"),
	}
}

// GetPathToConfig returns the path of the config file in use
func (c *Configuration) GetPathToConfig() string {
	return c.V.ConfigFileUsed()
}

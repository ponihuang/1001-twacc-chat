package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort            = "8080"
	defaultMySQLPort       = "3306"
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 60 * time.Second
)

// Config holds application runtime settings.
type Config struct {
	Port                   string
	Server                 ServerConfig
	MySQL                  MySQLConfig
	EnableMySQL            bool
	AutoMigrate            bool
	MigrationsDir          string
	IntegrationSharedToken string
	LoginTokenTTL          time.Duration
}

// ServerConfig groups HTTP server settings.
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// MySQLConfig groups database connection settings.
type MySQLConfig struct {
	DSN             string
	Host            string
	Port            string
	Database        string
	User            string
	Password        string
	Params          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Load reads runtime configuration from environment variables.
func Load() Config {
	return Config{
		Port: envOrDefault("PORT", defaultPort),
		Server: ServerConfig{
			ReadTimeout:  durationEnv("HTTP_READ_TIMEOUT", defaultReadTimeout),
			WriteTimeout: durationEnv("HTTP_WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:  durationEnv("HTTP_IDLE_TIMEOUT", defaultIdleTimeout),
		},
		MySQL: MySQLConfig{
			DSN:             strings.TrimSpace(os.Getenv("MYSQL_DSN")),
			Host:            envOrDefault("MYSQL_HOST", "127.0.0.1"),
			Port:            envOrDefault("MYSQL_PORT", defaultMySQLPort),
			Database:        strings.TrimSpace(os.Getenv("MYSQL_DATABASE")),
			User:            strings.TrimSpace(os.Getenv("MYSQL_USER")),
			Password:        os.Getenv("MYSQL_PASSWORD"),
			Params:          envOrDefault("MYSQL_PARAMS", "parseTime=true&charset=utf8mb4&loc=Local"),
			MaxOpenConns:    intEnv("MYSQL_MAX_OPEN_CONNS", defaultMaxOpenConns),
			MaxIdleConns:    intEnv("MYSQL_MAX_IDLE_CONNS", defaultMaxIdleConns),
			ConnMaxLifetime: durationEnv("MYSQL_CONN_MAX_LIFETIME", defaultConnMaxLifetime),
		},
		EnableMySQL:            boolEnv("ENABLE_MYSQL", false),
		AutoMigrate:            boolEnv("AUTO_MIGRATE", true),
		MigrationsDir:          envOrDefault("MIGRATIONS_DIR", filepath.Join("db", "migrations")),
		IntegrationSharedToken: strings.TrimSpace(os.Getenv("INTEGRATION_SHARED_TOKEN")),
		LoginTokenTTL:          durationEnv("LOGIN_TOKEN_TTL", 24*time.Hour),
	}
}

// Enabled returns true when MySQL should be initialized at startup.
func (c MySQLConfig) Enabled() bool {
	return c.DSN != "" || (c.Host != "" && c.Database != "" && c.User != "")
}

// FormattedDSN returns the configured DSN, or builds one from discrete fields.
func (c MySQLConfig) FormattedDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Database, c.Params)
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

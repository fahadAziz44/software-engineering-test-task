package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

// DatabaseConfig holds database connection parameters
type DatabaseConfig struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	Name     string `envconfig:"POSTGRES_DB" default:"postgres"`
	SSLMode  string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `envconfig:"PORT" default:"8080"`
}

// LoadFromEnv loads all configuration from environment variables using envconfig.
// envconfig automatically:
// - Reads environment variables based on struct tags
// - Validates required fields (required:"true" tag)
// - Sets default values (default:"value" tag)
// - Converts types (string -> int, bool, etc.)
// - Provides clear error messages
func LoadFromEnv() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &cfg, nil
}

// BuildDSN builds the PostgreSQL connection string from loaded configuration
func (c *Config) BuildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetEnvironment is used to distinguish between different runtime environments
// Returns "production" if GIN_MODE is "release", otherwise "development"
func GetEnvironment() string {
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		return "production"
	}
	return "development"
}

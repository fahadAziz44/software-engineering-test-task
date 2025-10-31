package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Name    string `yaml:"name"`
	SSLMode string `yaml:"sslmode"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// BuildDSN makes database connection string and throws an error if required environment variables are not set
func (c *Config) BuildDSN() (string, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if user == "" {
		return "", fmt.Errorf("DB_USER environment variable is required")
	}
	if password == "" {
		return "", fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		user,
		password,
		c.Database.Name,
		c.Database.SSLMode,
	)

	return dsn, nil
}

// GetEnvironment returns the current runtime environment
// Returns "production" if GIN_MODE is "release", otherwise "development"
func GetEnvironment() string {
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		return "production"
	}
	return "development"
}

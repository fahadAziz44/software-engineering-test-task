package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
	// TODO: Review this solution before using in production
	// Can not do  os.ReadFile(configPath) because it can be used to read files outside the application directory (prevents path traversal - G304)
	// Get the working directory to create a scoped root
	// -- NOTE: This is a alternate solution to prevent path traversal. Identified by gosec.
	// see docs/SECURITY_FIX_PATH_TRAVERSAL.md for more detail
	// -- COPIED: Should be reviewed before using in production ---
	// --- COPY START ---
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// This prevents reading files outside the application directory (prevents path traversal - G304)
	root := os.DirFS(workDir)

	// Clean the path to prevent directory traversal attempts like "../../../etc/passwd"
	cleanPath := filepath.Clean(configPath)

	// Read file using scoped filesystem (prevents path traversal)
	data, err := fs.ReadFile(root, cleanPath)
	// --- COPY END ---

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

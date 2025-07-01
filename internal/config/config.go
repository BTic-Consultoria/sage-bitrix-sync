package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our app
type Config struct {
	// Database config
	SageDB SageDBConfig `json:"sage_db"`
	// Bitrix24 config
	Bitrix BitrixConfig `json:"bitrix"`
	// API config
	API APIConfig `json:"api"`
	// Sync configuration
	Sync SyncConfig `json:"sync"`
}

// SageDBConfig represents SQL Server connection details
type SageDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// BitrixConfig represents Bitrix24 API settings
type BitrixConfig struct {
	Endpoint   string `json:"endpoint"`
	ClientCode string `json:"client_code"`
}

// APIConfig represents web API settings
type APIConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// SyncConfig represents synchronization settings
type SyncConfig struct {
	IntervalMinutes int  `json:"interval_minutes"`
	PackEmpresa     bool `json:"pack_empresa"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		SageDB: SageDBConfig{
			Host:     getEnv("SAGE_DB_HOST", "localhost"),
			Port:     getEnvAsInt("SAGE_DB_PORT", 1433),
			Database: getEnv("SAGE_DB_NAME", ""),
			Username: getEnv("SAGE_DB_USER", ""),
			Password: getEnv("SAGE_DB_PASSWORD", ""),
		},
		Bitrix: BitrixConfig{
			Endpoint:   getEnv("BITRIX_ENDPOINT", ""),
			ClientCode: getEnv("BITRIX_CLIENT_CODE", ""),
		},
		API: APIConfig{
			Host: getEnv("API_HOST", "0.0.0.0"),
			Port: getEnvAsInt("API_PORT", 8080),
		},
		Sync: SyncConfig{
			IntervalMinutes: getEnvAsInt("SYNC_INTERVAL_MINUTES", 5),
			PackEmpresa:     getEnvAsBool("PACK_EMPRESA", true),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", &err)
	}

	return config, nil
}

// Validate checks if all required configuration is present.
func (c *Config) Validate() error {
	if c.SageDB.Host == "" {
		return fmt.Errorf("SAGE_DB_HOST is required")
	}
	if c.SageDB.Password == "" {
		return fmt.Errorf("SAGE_DB_PASSWORD is required")
	}
	if c.Bitrix.Endpoint == "" {
		return fmt.Errorf("BITRIX_ENDPOINT is required")
	}
	return nil
}

// GetConnectionString builds SQL Server connection string.
func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable",
		c.SageDB.Host,
		c.SageDB.Port,
		c.SageDB.Database,
		c.SageDB.Username,
		c.SageDB.Password,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

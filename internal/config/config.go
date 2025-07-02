// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
// In Go, we use structs instead of classes
type Config struct {
	// Database configuration - equivalent to your App.config DB settings
	SageDB SageDBConfig `json:"sage_db"`

	// License configuration
	License LicenseConfig `json:"license"`

	// Bitrix24 configuration
	Bitrix BitrixConfig `json:"bitrix"`

	// Company mapping configuration
	Company CompanyMappingConfig `json:"company"`

	// API configuration
	API APIConfig `json:"api"`

	// Sync configuration
	Sync SyncConfig `json:"sync"`
}

// SageDBConfig represents SQL Server connection details
// Note: In Go, field names starting with capital letters are "exported" (public)
// This is equivalent to your DB_HOST, DB_PORT, etc. from App.config
type SageDBConfig struct {
	Host     string `json:"host"` // Can include named instance like "SERVER\\INSTANCE"
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LicenseConfig represents licensing information
type LicenseConfig struct {
	ID string `json:"id"`
}

// BitrixConfig represents Bitrix24 API settings
type BitrixConfig struct {
	Endpoint   string `json:"endpoint"`
	ClientCode string `json:"client_code"`
}

// CompanyMappingConfig represents company mapping between Bitrix and Sage
type CompanyMappingConfig struct {
	BitrixCode string `json:"bitrix_code"`
	SageCode   string `json:"sage_code"`
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
// In Go, functions that can fail return an error as the last return value
func Load() (*Config, error) {
	// Load .env file if it exists (similar to your App.config)
	_ = godotenv.Load()

	config := &Config{
		SageDB: SageDBConfig{
			Host:     getEnv("SAGE_DB_HOST", "SRVSAGE\\SAGEEXPRESS"),
			Port:     getEnvAsInt("SAGE_DB_PORT", 64952),
			Database: getEnv("SAGE_DB_NAME", "STANDARD"),
			Username: getEnv("SAGE_DB_USER", "LOGIC"),
			Password: getEnv("SAGE_DB_PASSWORD", ""),
		},
		License: LicenseConfig{
			ID: getEnv("LICENSE_ID", ""),
		},
		Bitrix: BitrixConfig{
			Endpoint:   getEnv("BITRIX_ENDPOINT", ""),
			ClientCode: getEnv("BITRIX_CLIENT_CODE", "test"),
		},
		Company: CompanyMappingConfig{
			BitrixCode: getEnv("EMPRESA_BITRIX", "test"),
			SageCode:   getEnv("EMPRESA_SAGE", "1"),
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
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate checks if all required configuration is present
// This is a method on the Config struct (like a method in your C# class)
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
	if c.License.ID == "" {
		return fmt.Errorf("LICENSE_ID is required")
	}
	return nil
}

// GetConnectionString builds SQL Server connection string
// This handles named instances properly (like SRVSAGE\\SAGEEXPRESS)
func (c *Config) GetConnectionString() string {
	// For SQL Server named instances, we need to format properly
	// The Go mssql driver expects: server=host\\instance;port=port;database=db;user id=user;password=pass
	return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=disable;trustServerCertificate=true",
		c.SageDB.Host,     // This can include named instance like "SRVSAGE\\SAGEEXPRESS"
		c.SageDB.Port,     // Your non-standard port 64952
		c.SageDB.Database, // STANDARD
		c.SageDB.Username, // LOGIC
		c.SageDB.Password, // Your password
	)
}

// Helper functions for environment variable parsing
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

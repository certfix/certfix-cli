package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitConfig initializes the configuration
func InitConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// Create .certfix directory if it doesn't exist
		configDir := filepath.Join(home, ".certfix")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".certfix" (without extension)
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // Read in environment variables that match

	// Set defaults
	viper.SetDefault("endpoint", "https://api.certfix.io")
	viper.SetDefault("timeout", 30)
	viper.SetDefault("retry_attempts", 3)

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		// Config file found and successfully parsed
	}
}

// Set sets a configuration value
func Set(key, value string) error {
	viper.Set(key, value)

	// Save to config file
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// If no config file is in use, create one
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(home, ".certfix", "config.yaml")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Get retrieves a configuration value
func Get(key string) (string, error) {
	if !viper.IsSet(key) {
		return "", fmt.Errorf("configuration key '%s' not found", key)
	}

	return viper.GetString(key), nil
}

// List returns all configuration values
func List() (map[string]interface{}, error) {
	return viper.AllSettings(), nil
}

// GetDefaultEndpoint returns the default API endpoint
func GetDefaultEndpoint() string {
	return viper.GetString("endpoint")
}

// GetTimeout returns the configured timeout
func GetTimeout() int {
	return viper.GetInt("timeout")
}

// GetRetryAttempts returns the configured retry attempts
func GetRetryAttempts() int {
	return viper.GetInt("retry_attempts")
}

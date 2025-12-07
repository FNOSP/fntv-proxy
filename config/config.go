package config

import (
	"fmt"
	"fntv-proxy/logger"
	"log"
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	Port int `ini:"port"`
}

// LoadConfig loads the configuration file
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	// Check if config file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create default config file
		err = createDefaultConfig(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create default config file: %v", err)
		}
		logger.StdoutLogger.Printf("Created default config file: %s", filename)
	}

	// Load config file
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	port, err := cfg.Section("server").Key("port").Int()
	if err != nil {
		log.Printf("Invalid port format in config file, using default port 1999")
		port = 1999
	}

	config.Port = port
	return config, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(filename string) error {
	cfg := ini.Empty()
	cfg.Section("server").Key("port").SetValue("1999")
	return cfg.SaveTo(filename)
}

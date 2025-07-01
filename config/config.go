package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Event represents a single event definition.
type Event struct {
	EventType   string `yaml:"event_type"`
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
	
}

// Config holds the application's configuration.
type Config struct {
	Events []Event `yaml:"events"`
}

// LoadConfig reads the configuration file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

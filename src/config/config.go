package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Proxy struct {
		Port       int    `yaml:"port"`
		BackendURL string `yaml:"backend_url"`
	} `yaml:"proxy"`
	Analyzer struct {
		Port        int `yaml:"port"`
		MaxExamples int `yaml:"max_examples"`
	} `yaml:"analyzer"`
}

// Load loads the configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.Proxy.Port <= 0 {
		return fmt.Errorf("proxy port must be positive")
	}
	if c.Analyzer.Port <= 0 {
		return fmt.Errorf("analyzer port must be positive")
	}
	if c.Proxy.BackendURL == "" {
		return fmt.Errorf("backend URL is required")
	}
	if c.Analyzer.MaxExamples <= 0 {
		return fmt.Errorf("max examples must be positive")
	}
	return nil
}

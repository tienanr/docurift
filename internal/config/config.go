package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	minPort = 1
	maxPort = 65535
)

// Config represents the DocuRift configuration structure
type Config struct {
	Proxy struct {
		Port       int    `yaml:"port"`
		BackendURL string `yaml:"backend-url"`
	} `yaml:"proxy"`

	Analyzer struct {
		Port            int      `yaml:"port"`
		MaxExamples     int      `yaml:"max-examples"`
		NoExampleFields []string `yaml:"no-example-fields"`
	} `yaml:"analyzer"`
}

// validatePort checks if the port is within valid range
func validatePort(port int, service string) error {
	if port < minPort || port > maxPort {
		return fmt.Errorf("%s port must be between %d and %d", service, minPort, maxPort)
	}
	return nil
}

// LoadConfig loads the configuration from the specified file path
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate proxy port
	if err := validatePort(config.Proxy.Port, "proxy"); err != nil {
		return nil, err
	}

	// Validate analyzer port
	if err := validatePort(config.Analyzer.Port, "analyzer"); err != nil {
		return nil, err
	}

	// Check for port conflict
	if config.Proxy.Port == config.Analyzer.Port {
		return nil, fmt.Errorf("proxy and analyzer cannot use the same port (%d)", config.Proxy.Port)
	}

	// Validate other required fields
	if config.Proxy.BackendURL == "" {
		return nil, fmt.Errorf("backend-url is required")
	}
	if config.Analyzer.MaxExamples <= 0 {
		return nil, fmt.Errorf("max-examples must be greater than 0")
	}

	return &config, nil
}

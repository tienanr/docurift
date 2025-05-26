package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
proxy:
    port: 9876
    backend-url: http://localhost:8080

analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
        - api_key
        - password
    storage:
        path: /tmp
        frequency: 5
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading valid config
	config, err := LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 9876, config.Proxy.Port)
	assert.Equal(t, "http://localhost:8080", config.Proxy.BackendURL)
	assert.Equal(t, 9877, config.Analyzer.Port)
	assert.Equal(t, 10, config.Analyzer.MaxExamples)
	assert.Equal(t, []string{"Authorization", "api_key", "password"}, config.Analyzer.RedactedFields)
	assert.Equal(t, "/tmp", config.Analyzer.Storage.Path)
	assert.Equal(t, 5, config.Analyzer.Storage.Frequency)

	// Test loading config with default storage values
	defaultStorageConfig := `
proxy:
    port: 9876
    backend-url: http://localhost:8080

analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
`
	if err := os.WriteFile(tmpfile.Name(), []byte(defaultStorageConfig), 0644); err != nil {
		t.Fatal(err)
	}

	config, err = LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, ".", config.Analyzer.Storage.Path)     // Default path
	assert.Equal(t, 10, config.Analyzer.Storage.Frequency) // Default frequency

	// Test cases for invalid configurations
	testCases := []struct {
		name     string
		config   string
		errorMsg string
	}{
		{
			name: "invalid proxy port range",
			config: `
proxy:
    port: 70000
    backend-url: http://localhost:8080
analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 5
`,
			errorMsg: "proxy port must be between 1 and 65535",
		},
		{
			name: "invalid analyzer port range",
			config: `
proxy:
    port: 9876
    backend-url: http://localhost:8080
analyzer:
    port: 0
    max-examples: 10
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 5
`,
			errorMsg: "analyzer port must be between 1 and 65535",
		},
		{
			name: "port conflict",
			config: `
proxy:
    port: 9876
    backend-url: http://localhost:8080
analyzer:
    port: 9876
    max-examples: 10
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 5
`,
			errorMsg: "proxy and analyzer cannot use the same port (9876)",
		},
		{
			name: "missing backend url",
			config: `
proxy:
    port: 9876
    backend-url: ""
analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 5
`,
			errorMsg: "backend-url is required",
		},
		{
			name: "invalid max examples",
			config: `
proxy:
    port: 9876
    backend-url: http://localhost:8080
analyzer:
    port: 9877
    max-examples: 0
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 5
`,
			errorMsg: "max-examples must be greater than 0",
		},
		{
			name: "invalid storage frequency",
			config: `
proxy:
    port: 9876
    backend-url: http://localhost:8080
analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
    storage:
        path: /tmp
        frequency: 0
`,
			errorMsg: "", // Should not error, should use default value
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(tmpfile.Name(), []byte(tc.config), 0644); err != nil {
				t.Fatal(err)
			}

			config, err := LoadConfig(tmpfile.Name())
			if tc.errorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				// For the invalid storage frequency test, verify default value is used
				if tc.name == "invalid storage frequency" {
					assert.Equal(t, 10, config.Analyzer.Storage.Frequency)
				}
			}
		})
	}
}

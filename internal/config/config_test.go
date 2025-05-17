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
`,
			errorMsg: "max-examples must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(tmpfile.Name(), []byte(tc.config), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadConfig(tmpfile.Name())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorMsg)
		})
	}
}

package analyzer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToolsSpecGeneration(t *testing.T) {
	// Create a new analyzer
	analyzer := NewAnalyzer()
	analyzer.SetMaxExamples(2)

	// Create a test request and response
	req, _ := http.NewRequest("POST", "/api/users", nil)
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
	}

	// Process a sample request/response
	analyzer.ProcessRequest("POST", "/api/users", req, resp, []byte(`{"name": "John", "age": 30}`), []byte(`{"id": 1, "name": "John", "age": 30}`))

	// Test generic tools spec
	t.Run("Generic Tools Spec", func(t *testing.T) {
		spec := analyzer.GenerateToolsSpec()
		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}

		tool := spec.Tools[0]
		if tool.Name != "post_api_users" {
			t.Errorf("Expected name 'post_api_users', got '%s'", tool.Name)
		}

		if tool.Description != "Creates data from /api/users" {
			t.Errorf("Expected description 'Creates data from /api/users', got '%s'", tool.Description)
		}

		// Check parameters
		params := tool.Parameters
		if params["type"] != "object" {
			t.Errorf("Expected type 'object', got '%v'", params["type"])
		}

		properties := params["properties"].(map[string]interface{})
		if len(properties) == 0 {
			t.Error("Expected properties to be non-empty")
		}
	})

	// Test OpenAI tools spec
	t.Run("OpenAI Tools Spec", func(t *testing.T) {
		spec := analyzer.GenerateOpenAIToolsSpec()
		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}

		tool := spec.Tools[0]
		if tool.Type != "function" {
			t.Errorf("Expected type 'function', got '%s'", tool.Type)
		}

		if tool.Function.Name != "post_api_users" {
			t.Errorf("Expected function name 'post_api_users', got '%s'", tool.Function.Name)
		}
	})

	// Test Anthropic tools spec
	t.Run("Anthropic Tools Spec", func(t *testing.T) {
		spec := analyzer.GenerateAnthropicToolsSpec()
		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}

		tool := spec.Tools[0]
		if tool.Name != "post_api_users" {
			t.Errorf("Expected name 'post_api_users', got '%s'", tool.Name)
		}

		if tool.InputSchema["type"] != "object" {
			t.Errorf("Expected input_schema type 'object', got '%v'", tool.InputSchema["type"])
		}
	})
}

func TestToolsEndpoints(t *testing.T) {
	// Create a new analyzer and server
	analyzer := NewAnalyzer()
	server := NewServer(analyzer)

	// Create a test request and response
	req, _ := http.NewRequest("POST", "/api/users", nil)
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
	}

	// Process a sample request/response
	analyzer.ProcessRequest("POST", "/api/users", req, resp, []byte(`{"name": "John", "age": 30}`), []byte(`{"id": 1, "name": "John", "age": 30}`))

	// Test generic tools endpoint
	t.Run("Generic Tools Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tools.json", nil)
		w := httptest.NewRecorder()
		server.handleTools(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var spec ToolsSpec
		if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}
	})

	// Test OpenAI tools endpoint
	t.Run("OpenAI Tools Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tools/openai.json", nil)
		w := httptest.NewRecorder()
		server.handleOpenAITools(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var spec OpenAIToolsSpec
		if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}
	})

	// Test Anthropic tools endpoint
	t.Run("Anthropic Tools Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tools/anthropic.json", nil)
		w := httptest.NewRecorder()
		server.handleAnthropicTools(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var spec AnthropicToolsSpec
		if err := json.NewDecoder(w.Body).Decode(&spec); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(spec.Tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(spec.Tools))
		}
	})

	// Test method not allowed
	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/tools.json", nil)
		w := httptest.NewRecorder()
		server.handleTools(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

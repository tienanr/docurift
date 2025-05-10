package analyzer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAnalyzer(t *testing.T) {
	a := NewAnalyzer()
	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
		return
	}
	if a.maxExamples != 10 {
		t.Errorf("Expected maxExamples to be 10, got %d", a.maxExamples)
	}
}

func TestSetMaxExamples(t *testing.T) {
	a := NewAnalyzer()
	a.SetMaxExamples(5)
	if a.maxExamples != 5 {
		t.Errorf("Expected maxExamples to be 5, got %d", a.maxExamples)
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "https://example.com/api/users",
			expected: "/api/users",
		},
		{
			name:     "with numeric ID",
			input:    "https://example.com/api/users/123",
			expected: "/api/users/{id}",
		},
		{
			name:     "with UUID",
			input:    "https://example.com/api/users/123e4567-e89b-12d3-a456-426614174000",
			expected: "/api/users/{uuid}",
		},
		{
			name:     "with query params",
			input:    "https://example.com/api/users?page=1&limit=10",
			expected: "/api/users",
		},
		{
			name:     "root path",
			input:    "https://example.com/",
			expected: "/",
		},
		{
			name:     "no protocol",
			input:    "example.com/api/users",
			expected: "example.com/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid UUID",
			input:    "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
		{
			name:     "invalid UUID",
			input:    "123e4567-e89b-12d3-a456",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUUID(tt.input)
			if result != tt.expected {
				t.Errorf("isUUID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessRequest(t *testing.T) {
	// Create test request
	reqBody := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "https://example.com/api/users?page=1", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("X-Custom-Header", "test-value")

	// Create test response
	respBody := map[string]interface{}{
		"id":   1,
		"name": "John Doe",
	}
	respBodyBytes, _ := json.Marshal(respBody)
	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type":      []string{"application/json"},
			"X-Response-Header": []string{"test-value"},
		},
		Body: io.NopCloser(bytes.NewBuffer(respBodyBytes)),
	}

	// Create analyzer and process request
	a := NewAnalyzer()
	a.ProcessRequest("POST", "https://example.com/api/users?page=1", req, resp, reqBodyBytes, respBodyBytes)

	// Get processed data
	data := a.GetData()
	key := "POST /api/users"
	endpoint, exists := data[key]

	if !exists {
		t.Fatalf("Expected endpoint %s to exist", key)
	}

	// Verify request headers
	if len(endpoint.RequestHeaders.Examples["X-Custom-Header"]) == 0 {
		t.Error("Expected X-Custom-Header to be processed")
	}

	// Verify URL parameters
	if len(endpoint.URLParameters.Examples["page"]) == 0 {
		t.Error("Expected URL parameter 'page' to be processed")
	}

	// Verify response status
	if _, exists := endpoint.ResponseStatuses[200]; !exists {
		t.Error("Expected response status 200 to be processed")
	}
}

func TestSchemaStore(t *testing.T) {
	store := NewSchemaStore()

	// Test adding values
	store.AddValue("test.path", "value1")
	store.AddValue("test.path", "value2")

	// Test duplicate value handling
	store.AddValue("test.path", "value1")

	// Test optional flag
	store.SetOptional("test.path", true)

	// Verify values
	if len(store.Examples["test.path"]) != 2 {
		t.Errorf("Expected 2 unique values, got %d", len(store.Examples["test.path"]))
	}

	if !store.Optional["test.path"] {
		t.Error("Expected path to be marked as optional")
	}
}

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "email address",
			input:    "user@example.com",
			expected: "john.doe@example.com",
		},
		{
			name:     "phone number",
			input:    "+1-555-123-4567",
			expected: "+1-555-123-4567",
		},
		{
			name:     "credit card",
			input:    "4111-1111-1111-1111",
			expected: "4111-1111-1111-1111",
		},
		{
			name:     "non-sensitive string",
			input:    "regular string",
			expected: "regular string",
		},
		{
			name:     "non-string value",
			input:    123,
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeValue(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessJSONPayload(t *testing.T) {
	store := NewSchemaStore()

	tests := []struct {
		name     string
		payload  interface{}
		expected map[string][]interface{}
	}{
		{
			name: "simple object",
			payload: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expected: map[string][]interface{}{
				"name": {"John"},
				"age":  {30},
			},
		},
		{
			name: "nested object",
			payload: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"address": map[string]interface{}{
						"city": "New York",
					},
				},
			},
			expected: map[string][]interface{}{
				"user.name":         {"John"},
				"user.address.city": {"New York"},
			},
		},
		{
			name: "array of objects",
			payload: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"name": "John",
						"age":  30,
					},
					map[string]interface{}{
						"name": "Jane",
						"age":  25,
					},
				},
			},
			expected: map[string][]interface{}{
				"users[].name": {"John", "Jane"},
				"users[].age":  {30, 25},
			},
		},
		{
			name: "mixed types",
			payload: map[string]interface{}{
				"string": "text",
				"number": 42,
				"bool":   true,
				"null":   nil,
			},
			expected: map[string][]interface{}{
				"string": {"text"},
				"number": {42},
				"bool":   {true},
				"null":   {nil},
			},
		},
		{
			name: "array of primitives",
			payload: map[string]interface{}{
				"tags": []interface{}{"tag1", "tag2", "tag3"},
			},
			expected: map[string][]interface{}{
				"tags[]": {"tag1", "tag2", "tag3"},
			},
		},
		{
			name: "deep nested arrays of objects",
			payload: map[string]interface{}{
				"invoices": []interface{}{
					map[string]interface{}{
						"id": 1,
						"line_items": []interface{}{
							map[string]interface{}{
								"product_id": 1,
								"quantity":   2,
								"unit_price": 999.99,
								"tax_info": []interface{}{
									map[string]interface{}{
										"jurisdiction": "CA",
										"tax_rate":     8.5,
										"description":  "California State Tax",
									},
									map[string]interface{}{
										"jurisdiction": "LA",
										"tax_rate":     2.0,
										"description":  "Los Angeles County Tax",
									},
								},
							},
							map[string]interface{}{
								"product_id": 2,
								"quantity":   1,
								"unit_price": 29.99,
								"tax_info": []interface{}{
									map[string]interface{}{
										"jurisdiction": "CA",
										"tax_rate":     8.5,
										"description":  "California State Tax",
									},
								},
							},
						},
					},
					map[string]interface{}{
						"id": 2,
						"line_items": []interface{}{
							map[string]interface{}{
								"product_id": 3,
								"quantity":   3,
								"unit_price": 75.00,
								"tax_info": []interface{}{
									map[string]interface{}{
										"jurisdiction": "TX",
										"tax_rate":     6.25,
										"description":  "Texas State Tax",
									},
								},
							},
						},
					},
				},
			},
			expected: map[string][]interface{}{
				"invoices[].id":                                   {1, 2},
				"invoices[].line_items[].product_id":              {1, 2, 3},
				"invoices[].line_items[].quantity":                {2, 1, 3},
				"invoices[].line_items[].unit_price":              {999.99, 29.99, 75.00},
				"invoices[].line_items[].tax_info[].jurisdiction": {"CA", "LA", "TX"},
				"invoices[].line_items[].tax_info[].tax_rate":     {8.5, 2.0, 6.25},
				"invoices[].line_items[].tax_info[].description": {
					"California State Tax",
					"Los Angeles County Tax",
					"Texas State Tax",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the store for each test
			store = NewSchemaStore()

			// Process the payload
			processJSONPayload(store, "", tt.payload)

			// Verify the results
			for path, expectedValues := range tt.expected {
				values, exists := store.Examples[path]
				if !exists {
					t.Errorf("Expected path %s to exist", path)
					continue
				}

				if len(values) != len(expectedValues) {
					t.Errorf("Expected %d values for path %s, got %d", len(expectedValues), path, len(values))
					continue
				}

				for i, expected := range expectedValues {
					if values[i] != expected {
						t.Errorf("For path %s, expected value %v at index %d, got %v", path, expected, i, values[i])
					}
				}
			}

			// Verify no unexpected paths exist
			for path := range store.Examples {
				if _, exists := tt.expected[path]; !exists {
					t.Errorf("Unexpected path found: %s", path)
				}
			}
		})
	}
}

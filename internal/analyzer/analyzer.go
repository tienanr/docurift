package analyzer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// SchemaStore represents a store for tracking JSON schema paths and their values
type SchemaStore struct {
	mu          sync.RWMutex
	Examples    map[string][]interface{} // path -> []values
	Optional    map[string]bool          // path -> isOptional
	maxExamples int                      // Maximum number of examples to keep per field
}

// NewSchemaStore creates a new SchemaStore
func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		Examples:    make(map[string][]interface{}),
		Optional:    make(map[string]bool),
		maxExamples: 10, // Set default max examples
	}
}

// AddValue adds a value to the schema store for a given path
func (s *SchemaStore) AddValue(path string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Examples[path]; !exists {
		s.Examples[path] = make([]interface{}, 0)
		s.Optional[path] = true
	}

	// Check if value already exists
	for _, v := range s.Examples[path] {
		if areValuesEqual(v, value) {
			return // Skip duplicate values
		}
	}

	// Add value if we haven't reached the limit
	if len(s.Examples[path]) < s.maxExamples {
		s.Examples[path] = append(s.Examples[path], value)
	}
}

// areValuesEqual compares two interface{} values for equality
func areValuesEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch v1 := a.(type) {
	case map[string]interface{}:
		// If a is a map, b must also be a map
		v2, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(v1) != len(v2) {
			return false
		}
		for k, val1 := range v1 {
			val2, exists := v2[k]
			if !exists {
				return false
			}
			if !areValuesEqual(val1, val2) {
				return false
			}
		}
		return true

	case []interface{}:
		// If a is a slice, b must also be a slice
		v2, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if !areValuesEqual(v1[i], v2[i]) {
				return false
			}
		}
		return true

	case float64:
		// Handle float64 specifically to avoid precision issues
		v2, ok := b.(float64)
		if !ok {
			return false
		}
		return v1 == v2

	case int:
		// Handle int specifically
		v2, ok := b.(int)
		if !ok {
			return false
		}
		return v1 == v2

	case string:
		// Handle string specifically
		v2, ok := b.(string)
		if !ok {
			return false
		}
		return v1 == v2

	case bool:
		// Handle bool specifically
		v2, ok := b.(bool)
		if !ok {
			return false
		}
		return v1 == v2

	default:
		// For other types, use type assertion and direct comparison
		return a == b
	}
}

// SetOptional marks a path as optional
func (s *SchemaStore) SetOptional(path string, optional bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Optional[path] = optional
}

// EndpointData represents the data structure for a specific endpoint
type EndpointData struct {
	Method           string
	URL              string
	RequestHeaders   *SchemaStore
	RequestPayload   *SchemaStore
	URLParameters    *SchemaStore // New field for URL parameters
	ResponseStatuses map[int]*ResponseData
}

// ResponseData represents response data for a specific status code
type ResponseData struct {
	Headers *SchemaStore
	Payload *SchemaStore
}

// Analyzer is the main analyzer structure
type Analyzer struct {
	mu          sync.RWMutex
	endpoints   map[string]*EndpointData // key: method+url
	maxExamples int                      // Maximum number of examples to keep per field
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		endpoints:   make(map[string]*EndpointData),
		maxExamples: 10, // Default value
	}
}

// SetMaxExamples sets the maximum number of examples to keep per field
func (a *Analyzer) SetMaxExamples(max int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maxExamples = max
}

// Common HTTP headers to exclude from documentation
var excludedHeaders = map[string]bool{
	"Content-Length":    true,
	"Content-Type":      true,
	"Date":              true,
	"Server":            true,
	"Connection":        true,
	"Keep-Alive":        true,
	"Transfer-Encoding": true,
	"Accept":            true,
	"Accept-Encoding":   true,
	"Accept-Language":   true,
	"User-Agent":        true,
	"Host":              true,
}

// sensitivePatterns defines regex patterns for sensitive data
var sensitivePatterns = map[string]string{
	// Email pattern
	`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`: "john.doe@example.com",
	// Phone number pattern (supports various formats)
	`^\+?[0-9]{10,15}$`: "+1-555-123-4567",
	// Credit card pattern (supports various formats)
	`^[0-9]{4}[- ]?[0-9]{4}[- ]?[0-9]{4}[- ]?[0-9]{4}$`: "4111-1111-1111-1111",
	// SSN pattern
	`^[0-9]{3}[- ]?[0-9]{2}[- ]?[0-9]{4}$`: "123-45-6789",
	// Password pattern (any string containing "password" or "pass" or "pwd")
	`(?i).*(password|pass|pwd).*`: "********",
}

// sanitizeValue replaces sensitive data with dummy values
func sanitizeValue(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		for pattern, replacement := range sensitivePatterns {
			matched, _ := regexp.MatchString(pattern, str)
			if matched {
				return replacement
			}
		}
	}
	return value
}

// normalizeURL removes the host name from a URL and generalizes path parameters
func normalizeURL(url string) string {
	// Find the last occurrence of "://"
	protocolIndex := strings.LastIndex(url, "://")
	if protocolIndex == -1 {
		return url
	}

	// Find the first "/" after the protocol
	pathIndex := strings.Index(url[protocolIndex+3:], "/")
	if pathIndex == -1 {
		return "/"
	}

	// Get the path part
	path := url[protocolIndex+3+pathIndex:]

	// Remove query parameters
	if queryIndex := strings.Index(path, "?"); queryIndex != -1 {
		path = path[:queryIndex]
	}

	// Split path into segments
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		// Skip empty segments
		if segment == "" {
			continue
		}

		// Check if segment is a numeric ID
		if _, err := strconv.Atoi(segment); err == nil {
			segments[i] = "{id}"
			continue
		}

		// Check if segment is a UUID
		if isUUID(segment) {
			segments[i] = "{uuid}"
			continue
		}
	}

	// Rejoin segments
	return strings.Join(segments, "/")
}

// isUUID checks if a string is a valid UUID
func isUUID(s string) bool {
	// UUID pattern: 8-4-4-4-12 hexadecimal digits
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(s))
	return matched
}

// ProcessRequest processes a request and response pair
func (a *Analyzer) ProcessRequest(method, url string, req *http.Request, resp *http.Response, reqBody, respBody []byte) {
	// Skip invalid responses
	if resp.StatusCode >= 400 {
		return
	}

	// Process URL parameters before normalizing the URL
	urlParams := make(map[string][]string)
	for key, values := range req.URL.Query() {
		urlParams[key] = values
	}

	// Normalize the URL by removing the host name and query parameters
	normalizedURL := normalizeURL(url)
	key := method + " " + normalizedURL

	a.mu.Lock()
	endpoint, exists := a.endpoints[key]
	if !exists {
		endpoint = &EndpointData{
			Method:           method,
			URL:              normalizedURL,
			RequestHeaders:   NewSchemaStore(),
			RequestPayload:   NewSchemaStore(),
			URLParameters:    NewSchemaStore(), // Initialize URL parameters store
			ResponseStatuses: make(map[int]*ResponseData),
		}
		a.endpoints[key] = endpoint
	}
	a.mu.Unlock()

	// Process URL parameters
	for key, values := range urlParams {
		for _, value := range values {
			endpoint.URLParameters.AddValue(key, value)
		}
		// Mark as optional if not present in all requests
		endpoint.URLParameters.SetOptional(key, true)
	}

	// Process request headers
	for key, values := range req.Header {
		if !excludedHeaders[key] {
			for _, value := range values {
				endpoint.RequestHeaders.AddValue(key, value)
			}
		}
	}

	// Process request payload if present
	if len(reqBody) > 0 {
		var payload interface{}
		if err := json.Unmarshal(reqBody, &payload); err == nil {
			processJSONPayload(endpoint.RequestPayload, "", payload)
		}
	}

	// Process response
	status := resp.StatusCode
	a.mu.Lock()
	responseData, exists := endpoint.ResponseStatuses[status]
	if !exists {
		responseData = &ResponseData{
			Headers: NewSchemaStore(),
			Payload: NewSchemaStore(),
		}
		endpoint.ResponseStatuses[status] = responseData
	}
	a.mu.Unlock()

	// Process response headers
	for key, values := range resp.Header {
		if !excludedHeaders[key] {
			for _, value := range values {
				responseData.Headers.AddValue(key, value)
			}
		}
	}

	// Process response payload if present
	if len(respBody) > 0 {
		if resp.Header.Get("Content-Encoding") == "gzip" {
			b := bytes.NewReader(respBody)
			reader, err := gzip.NewReader(b)
			defer reader.Close()
			if err == nil {
				respBody, _ = io.ReadAll(reader)
			}
		}

		var payload interface{}
		if err := json.Unmarshal(respBody, &payload); err == nil {
			processJSONPayload(responseData.Payload, "", payload)
		}
	}
}

// processJSONPayload recursively processes a JSON payload to extract schema paths
func processJSONPayload(store *SchemaStore, basePath string, value interface{}) {
	if basePath == "" && value == nil {
		return
	}

	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPath := basePath
			if newPath != "" {
				newPath += "."
			}
			newPath += key
			if val == nil {
				fmt.Printf("AddValue: %s = nil\n", newPath)
				store.AddValue(newPath, nil)
			} else {
				processJSONPayload(store, newPath, val)
			}
		}
	case []interface{}:
		if len(v) == 0 {
			if basePath != "" && !strings.Contains(basePath, "]") {
				fmt.Printf("AddValue: %s[] = nil\n", basePath)
				store.AddValue(basePath+"[]", nil)
			}
			return
		}

		if isObjectArray(v) {
			// Recursively process each object in the array with the correct path
			for _, item := range v {
				processJSONPayload(store, basePath+"[]", item)
			}
		} else {
			arrayPath := basePath + "[]"
			for _, val := range v {
				if basePath != "" && !strings.Contains(basePath, "]") {
					fmt.Printf("AddValue: %s = %v\n", arrayPath, val)
					store.AddValue(arrayPath, val)
				}
			}
		}
	default:
		fmt.Printf("AddValue: %s = %v\n", basePath, value)
		store.AddValue(basePath, value)
	}
}

// isObjectArray checks if an array contains objects
func isObjectArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}
	_, ok := arr[0].(map[string]interface{})
	return ok
}

// GetData returns the current state of the analyzer
func (a *Analyzer) GetData() map[string]*EndpointData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.endpoints
}

package analyzer

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// SchemaStore represents a store for tracking JSON schema paths and their values
type SchemaStore struct {
	mu       sync.RWMutex
	Examples map[string][]interface{} // path -> []values
	Optional map[string]bool          // path -> isOptional
}

// NewSchemaStore creates a new SchemaStore
func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		Examples: make(map[string][]interface{}),
		Optional: make(map[string]bool),
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
	s.Examples[path] = append(s.Examples[path], value)
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
	ResponseStatuses map[int]*ResponseData
}

// ResponseData represents response data for a specific status code
type ResponseData struct {
	Headers *SchemaStore
	Payload *SchemaStore
}

// Analyzer is the main analyzer structure
type Analyzer struct {
	mu        sync.RWMutex
	endpoints map[string]*EndpointData // key: method+url
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		endpoints: make(map[string]*EndpointData),
	}
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

// normalizeURL removes the host name from a URL
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

	// Return the path part
	return url[protocolIndex+3+pathIndex:]
}

// ProcessRequest processes a request and response pair
func (a *Analyzer) ProcessRequest(method, url string, req *http.Request, resp *http.Response, reqBody, respBody []byte) {
	// Normalize the URL by removing the host name
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
			ResponseStatuses: make(map[int]*ResponseData),
		}
		a.endpoints[key] = endpoint
	}
	a.mu.Unlock()

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
		var payload interface{}
		if err := json.Unmarshal(respBody, &payload); err == nil {
			processJSONPayload(responseData.Payload, "", payload)
		}
	}
}

// processJSONPayload recursively processes a JSON payload to extract schema paths
func processJSONPayload(store *SchemaStore, basePath string, value interface{}) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPath := basePath
			if newPath != "" {
				newPath += "."
			}
			newPath += key
			processJSONPayload(store, newPath, val)
		}
	case []interface{}:
		for _, val := range v {
			processJSONPayload(store, basePath+"[]", val)
		}
	default:
		// Sanitize the value before storing it
		sanitizedValue := sanitizeValue(v)
		store.AddValue(basePath, sanitizedValue)
	}
}

// GetData returns the current state of the analyzer
func (a *Analyzer) GetData() map[string]*EndpointData {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.endpoints
}

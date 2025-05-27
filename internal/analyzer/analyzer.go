package analyzer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SchemaStore represents a store for tracking JSON schema paths and their values
type SchemaStore struct {
	mu          sync.RWMutex
	Examples    map[string][]interface{} // path -> []values
	Optional    map[string]bool          // path -> isOptional
	maxExamples int                      // Maximum number of examples to keep per field
	analyzer    *Analyzer                // Reference to parent analyzer for accessing noExampleFields
}

// NewSchemaStore creates a new SchemaStore
func NewSchemaStore() *SchemaStore {
	return &SchemaStore{
		Examples:    make(map[string][]interface{}),
		Optional:    make(map[string]bool),
		maxExamples: 10, // Set default max examples
	}
}

// SetAnalyzer sets the parent analyzer reference
func (s *SchemaStore) SetAnalyzer(a *Analyzer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analyzer = a
}

// AddValue adds a value to the schema store for a given path
func (s *SchemaStore) AddValue(path string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If this is a redacted field, store "REDACTED" instead of the actual value
	if s.analyzer != nil && s.analyzer.shouldRedact(path) {
		value = "REDACTED"
	}

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
	mu               sync.RWMutex
	endpoints        map[string]*EndpointData // key: method+url
	maxExamples      int                      // Maximum number of examples to keep per field
	redactedFields   []string                 // Fields to redact in documentation
	stopChan         chan struct{}            // Channel to signal stop for persistence goroutine
	storageLocation  string                   // Path where analyzer.json is stored
	storageFrequency int                      // Frequency of state persistence in seconds
	proxyPort        int                      // Proxy server port
	backendURL       string                   // Backend URL for proxy
	analyzerPort     int                      // Analyzer server port
}

// SchemaVersion represents the current version of the analyzer schema
const SchemaVersion = "1.0"

// PersistedState represents the structure of the saved analyzer state
type PersistedState struct {
	Version   string                   `json:"version"`
	Endpoints map[string]*EndpointData `json:"endpoints"`
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(storageLocation string, storageFrequency int) *Analyzer {
	// Set default values if not provided
	if storageLocation == "" {
		storageLocation = "."
	}
	if storageFrequency <= 0 {
		storageFrequency = 10
	}

	a := &Analyzer{
		endpoints:        make(map[string]*EndpointData),
		maxExamples:      10, // Default value
		redactedFields:   make([]string, 0),
		stopChan:         make(chan struct{}),
		storageLocation:  storageLocation,
		storageFrequency: storageFrequency,
	}

	// Load existing data if available
	a.loadState()

	// Start persistence goroutine
	go a.startPersistence()

	return a
}

// startPersistence starts a goroutine that saves the analyzer state periodically
func (a *Analyzer) startPersistence() {
	ticker := time.NewTicker(time.Duration(a.storageFrequency) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.saveState()
		case <-a.stopChan:
			return
		}
	}
}

// saveState saves the current state of the analyzer to analyzer.json
func (a *Analyzer) saveState() {
	a.mu.RLock()
	state := PersistedState{
		Version:   SchemaVersion,
		Endpoints: a.endpoints,
	}
	a.mu.RUnlock()

	jsonData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	filePath := filepath.Join(a.storageLocation, "analyzer.json")
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return
	}
}

// loadState loads the analyzer state from analyzer.json if it exists and version matches
func (a *Analyzer) loadState() {
	filePath := filepath.Join(a.storageLocation, "analyzer.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[INFO] No saved state found at %s", filePath)
		}
		return
	}

	var state PersistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	// Only load if version matches
	if state.Version != SchemaVersion {
		log.Printf("[INFO] Saved state version mismatch: found %s, expected %s", state.Version, SchemaVersion)
		return
	}

	a.mu.Lock()
	a.endpoints = state.Endpoints
	a.mu.Unlock()
}

// Stop stops the persistence goroutine
func (a *Analyzer) Stop() {
	close(a.stopChan)
}

// SetMaxExamples sets the maximum number of examples to keep per field
func (a *Analyzer) SetMaxExamples(max int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maxExamples = max
}

// SetRedactedFields sets the list of fields to redact in documentation
func (a *Analyzer) SetRedactedFields(fields []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.redactedFields = fields
}

// shouldRedact checks if a field should be redacted
func (a *Analyzer) shouldRedact(field string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, redactedField := range a.redactedFields {
		if strings.EqualFold(field, redactedField) {
			return true
		}
	}
	return false
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
		// Set analyzer reference for all schema stores
		endpoint.RequestHeaders.SetAnalyzer(a)
		endpoint.RequestPayload.SetAnalyzer(a)
		endpoint.URLParameters.SetAnalyzer(a)
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
		// Set analyzer reference for response schema stores
		responseData.Headers.SetAnalyzer(a)
		responseData.Payload.SetAnalyzer(a)
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
			if err == nil {
				defer reader.Close()
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
				store.AddValue(newPath, nil)
			} else {
				processJSONPayload(store, newPath, val)
			}
		}
	case []interface{}:
		if len(v) == 0 {
			if basePath != "" && !strings.Contains(basePath, "]") {
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
					store.AddValue(arrayPath, val)
				}
			}
		}
	default:
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

// GetConfig returns the current configuration of the analyzer
func (a *Analyzer) GetConfig() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"maxExamples":      a.maxExamples,
		"redactedFields":   a.redactedFields,
		"storageLocation":  a.storageLocation,
		"storageFrequency": a.storageFrequency,
		"endpointCount":    len(a.endpoints),
		"port":             a.analyzerPort,
	}
}

// SetProxyConfig sets the proxy configuration
func (a *Analyzer) SetProxyConfig(port int, backendURL string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.proxyPort = port
	a.backendURL = backendURL
}

// GetProxyPort returns the proxy port
func (a *Analyzer) GetProxyPort() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.proxyPort
}

// GetBackendURL returns the backend URL
func (a *Analyzer) GetBackendURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.backendURL
}

// SetAnalyzerPort sets the analyzer port
func (a *Analyzer) SetAnalyzerPort(port int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.analyzerPort = port
}

// GetAnalyzerPort returns the analyzer port
func (a *Analyzer) GetAnalyzerPort() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.analyzerPort
}

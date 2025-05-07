package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PostmanCollection represents a Postman collection
type PostmanCollection struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      string `json:"schema"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanItem represents a request or folder in a Postman collection
type PostmanItem struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Request     *PostmanRequest `json:"request,omitempty"`
	Item        []PostmanItem   `json:"item,omitempty"`
}

// PostmanRequest represents a request in a Postman collection
type PostmanRequest struct {
	Method      string              `json:"method"`
	Header      []PostmanHeader     `json:"header"`
	URL         PostmanURL          `json:"url"`
	Body        *PostmanRequestBody `json:"body,omitempty"`
	Description string              `json:"description,omitempty"`
}

// PostmanHeader represents a header in a Postman request
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// PostmanURL represents a URL in a Postman request
type PostmanURL struct {
	Raw      string         `json:"raw"`
	Protocol string         `json:"protocol"`
	Host     []string       `json:"host"`
	Path     []string       `json:"path"`
	Query    []PostmanQuery `json:"query,omitempty"`
}

// PostmanQuery represents a query parameter in a Postman request
type PostmanQuery struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanRequestBody represents a request body in a Postman request
type PostmanRequestBody struct {
	Mode    string                 `json:"mode"`
	Raw     string                 `json:"raw,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GeneratePostmanCollection generates a Postman collection from analyzer data
func (a *Analyzer) GeneratePostmanCollection() *PostmanCollection {
	a.mu.RLock()
	defer a.mu.RUnlock()

	collection := &PostmanCollection{}
	collection.Info.Name = "API Collection"
	collection.Info.Description = "Generated API collection from analyzer data"
	collection.Info.Schema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

	// Group endpoints by base path
	endpointsByPath := make(map[string][]*EndpointData)
	for _, endpoint := range a.endpoints {
		path := strings.Split(endpoint.URL, "/")[1] // Get the first segment after /
		endpointsByPath[path] = append(endpointsByPath[path], endpoint)
	}

	// Create items for each group
	for path, endpoints := range endpointsByPath {
		item := PostmanItem{
			Name:        path,
			Description: fmt.Sprintf("Endpoints for %s", path),
			Item:        make([]PostmanItem, 0),
		}

		// Add each endpoint as a request
		for _, endpoint := range endpoints {
			request := createPostmanRequest(endpoint)
			if request != nil {
				item.Item = append(item.Item, PostmanItem{
					Name:        fmt.Sprintf("%s %s", endpoint.Method, endpoint.URL),
					Description: fmt.Sprintf("%s request for %s", endpoint.Method, endpoint.URL),
					Request:     request,
				})
			}
		}

		collection.Item = append(collection.Item, item)
	}

	return collection
}

// createPostmanRequest creates a Postman request from an endpoint
func createPostmanRequest(endpoint *EndpointData) *PostmanRequest {
	request := &PostmanRequest{
		Method: endpoint.Method,
		Header: make([]PostmanHeader, 0),
		URL: PostmanURL{
			Raw:      endpoint.URL,
			Protocol: "http",
			Host:     []string{"localhost:8080"},
			Path:     strings.Split(endpoint.URL, "/"),
		},
	}

	// Add headers
	if endpoint.RequestHeaders != nil {
		for header, values := range endpoint.RequestHeaders.Examples {
			if len(values) > 0 {
				request.Header = append(request.Header, PostmanHeader{
					Key:   header,
					Value: fmt.Sprintf("%v", values[0]),
					Type:  "text",
				})
			}
		}
	}

	// Add query parameters
	if endpoint.URLParameters != nil {
		for param, values := range endpoint.URLParameters.Examples {
			if len(values) > 0 {
				request.URL.Query = append(request.URL.Query, PostmanQuery{
					Key:   param,
					Value: fmt.Sprintf("%v", values[0]),
				})
			}
		}
	}

	// Add request body if exists
	if endpoint.RequestPayload != nil && len(endpoint.RequestPayload.Examples) > 0 {
		// Convert the first example to JSON
		example := createExampleFromStore(endpoint.RequestPayload)
		if example != nil {
			jsonData, err := json.MarshalIndent(example, "", "  ")
			if err == nil {
				request.Body = &PostmanRequestBody{
					Mode: "raw",
					Raw:  string(jsonData),
					Options: map[string]interface{}{
						"raw": map[string]interface{}{
							"language": "json",
						},
					},
				}
			}
		}
	}

	return request
}

// createExampleFromStore creates an example object from a SchemaStore
func createExampleFromStore(store *SchemaStore) interface{} {
	if store == nil || len(store.Examples) == 0 {
		return nil
	}

	// Create a map to hold the example
	example := make(map[string]interface{})

	// Process each path and its examples
	for path, values := range store.Examples {
		if len(values) == 0 {
			continue
		}

		// Split the path into parts
		parts := strings.Split(path, ".")
		current := example

		// Navigate through the path
		for i, part := range parts {
			isLast := i == len(parts)-1
			isArray := strings.HasSuffix(part, "[]")

			if isArray {
				part = strings.TrimSuffix(part, "[]")
				if _, exists := current[part]; !exists {
					current[part] = make([]interface{}, 0)
				}
				if isLast {
					arr := current[part].([]interface{})
					arr = append(arr, values[0])
					current[part] = arr
				} else {
					arr := current[part].([]interface{})
					if len(arr) == 0 {
						arr = append(arr, make(map[string]interface{}))
					}
					current = arr[0].(map[string]interface{})
				}
			} else {
				if isLast {
					current[part] = values[0]
				} else {
					if _, exists := current[part]; !exists {
						current[part] = make(map[string]interface{})
					}
					current = current[part].(map[string]interface{})
				}
			}
		}
	}

	return example
}

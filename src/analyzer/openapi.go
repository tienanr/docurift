package analyzer

import (
	"fmt"
	"strings"
)

// OpenAPI represents the OpenAPI 3.0 specification
type OpenAPI struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	Summary     string              `json:"summary"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Schema      Schema `json:"schema"`
}

type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Header struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type        string            `json:"type,omitempty"`
	Format      string            `json:"format,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Description string            `json:"description,omitempty"`
	Example     interface{}       `json:"example,omitempty"`
	Examples    []interface{}     `json:"examples,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// GenerateOpenAPI generates OpenAPI specification from analyzer data
func (a *Analyzer) GenerateOpenAPI() *OpenAPI {
	a.mu.RLock()
	defer a.mu.RUnlock()

	openAPI := &OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:   "API Documentation",
			Version: "1.0.0",
		},
		Paths:      make(map[string]PathItem),
		Components: Components{Schemas: make(map[string]Schema)},
	}

	for key, endpoint := range a.endpoints {
		// Split method and path
		parts := strings.SplitN(key, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method, path := parts[0], parts[1]

		// Create or get path item
		pathItem, exists := openAPI.Paths[path]
		if !exists {
			pathItem = PathItem{}
		}

		// Create operation
		operation := &Operation{
			Summary:   fmt.Sprintf("%s %s", method, path),
			Responses: make(map[string]Response),
		}

		// Add request parameters from headers
		for header, store := range endpoint.RequestHeaders.Examples {
			param := Parameter{
				Name:        header,
				In:          "header",
				Required:    !endpoint.RequestHeaders.Optional[header],
				Description: fmt.Sprintf("Header: %s", header),
				Schema: Schema{
					Type:     "string",
					Examples: store,
				},
			}
			operation.Parameters = append(operation.Parameters, param)
		}

		// Add request body schema if exists
		if len(endpoint.RequestPayload.Examples) > 0 {
			requestBody := &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: generateSchemaFromStore(endpoint.RequestPayload),
					},
				},
			}
			operation.RequestBody = requestBody
		}

		// Add responses
		for status, responseData := range endpoint.ResponseStatuses {
			response := Response{
				Description: fmt.Sprintf("Status %d", status),
				Content: map[string]MediaType{
					"application/json": {
						Schema: generateSchemaFromStore(responseData.Payload),
					},
				},
				Headers: make(map[string]Header),
			}

			// Add response headers
			for header, store := range responseData.Headers.Examples {
				response.Headers[header] = Header{
					Schema: Schema{
						Type:     "string",
						Examples: store,
					},
				}
			}

			operation.Responses[fmt.Sprintf("%d", status)] = response
		}

		// Add operation to path item
		switch method {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		}

		openAPI.Paths[path] = pathItem
	}

	return openAPI
}

// generateSchemaFromStore generates OpenAPI schema from SchemaStore
func generateSchemaFromStore(store *SchemaStore) Schema {
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
	}

	for path, examples := range store.Examples {
		// Skip empty paths (root object)
		if path == "" {
			continue
		}

		// Determine if field is required
		if !store.Optional[path] {
			schema.Required = append(schema.Required, path)
		}

		// Create schema for the field
		fieldSchema := Schema{
			Examples: examples,
		}

		// Determine type based on examples
		if len(examples) > 0 {
			switch examples[0].(type) {
			case string:
				fieldSchema.Type = "string"
			case float64:
				fieldSchema.Type = "number"
			case bool:
				fieldSchema.Type = "boolean"
			case []interface{}:
				fieldSchema.Type = "array"
				fieldSchema.Items = &Schema{Type: "object"}
			case map[string]interface{}:
				fieldSchema.Type = "object"
			}
		}

		schema.Properties[path] = fieldSchema
	}

	return schema
}

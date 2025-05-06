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

		// Add path parameters
		segments := strings.Split(path, "/")
		for _, segment := range segments {
			if segment == "{id}" {
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        "id",
					In:          "path",
					Required:    true,
					Description: "Resource ID",
					Schema: Schema{
						Type: "integer",
					},
				})
			} else if segment == "{uuid}" {
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        "uuid",
					In:          "path",
					Required:    true,
					Description: "Resource UUID",
					Schema: Schema{
						Type:   "string",
						Format: "uuid",
					},
				})
			}
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
	// First pass: collect all paths and their parts
	paths := make(map[string][]string)
	for path := range store.Examples {
		if path == "" {
			continue
		}
		parts := strings.Split(path, ".")
		paths[path] = parts
	}

	// If we have array paths, create an array schema
	if len(paths) > 0 {
		// Get the first path to check if it's an array
		var firstPath string
		for p := range paths {
			firstPath = p
			break
		}
		parts := strings.Split(firstPath, ".")
		if strings.HasSuffix(parts[0], "[]") {
			// Create array schema
			arraySchema := Schema{
				Type: "array",
				Items: &Schema{
					Type:       "object",
					Properties: make(map[string]Schema),
				},
			}

			// Add properties to the array items schema
			for path, parts := range paths {
				fieldName := strings.TrimSuffix(parts[0], "[]")
				if len(parts) > 1 {
					fieldName = parts[1]
				}

				// Create property schema
				examples := store.Examples[path]
				if len(examples) > 0 {
					propertySchema := createPropertySchema(examples)
					// Add to array items schema
					arraySchema.Items.Properties[fieldName] = propertySchema

					// Add to required fields if not optional
					if !store.Optional[path] {
						arraySchema.Items.Required = append(arraySchema.Items.Required, fieldName)
					}
				}
			}

			return arraySchema
		}
	}

	// Handle regular object schema
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
	}

	// Group paths by their first part to handle nested objects
	groupedPaths := make(map[string]map[string][]string)
	for _, parts := range paths {
		firstPart := parts[0]
		if _, exists := groupedPaths[firstPart]; !exists {
			groupedPaths[firstPart] = make(map[string][]string)
		}
		// Store the remaining parts for nested objects
		if len(parts) > 1 {
			groupedPaths[firstPart][strings.Join(parts[1:], ".")] = parts[1:]
		} else {
			groupedPaths[firstPart][""] = parts
		}
	}

	// Add properties to the schema
	for fieldName, nestedPaths := range groupedPaths {
		// Check if this is a nested object
		if len(nestedPaths) > 1 || (len(nestedPaths) == 1 && nestedPaths[""] == nil) {
			// Create a nested object schema
			nestedSchema := Schema{
				Type:       "object",
				Properties: make(map[string]Schema),
			}

			// Process nested paths
			for nestedPath, parts := range nestedPaths {
				if nestedPath == "" {
					continue // Skip empty paths
				}
				examples := store.Examples[fieldName+"."+nestedPath]
				if len(examples) > 0 {
					propertySchema := createPropertySchema(examples)
					nestedSchema.Properties[parts[0]] = propertySchema

					if !store.Optional[fieldName+"."+nestedPath] {
						nestedSchema.Required = append(nestedSchema.Required, parts[0])
					}
				}
			}

			schema.Properties[fieldName] = nestedSchema
		} else {
			// Handle regular field
			examples := store.Examples[fieldName]
			if len(examples) > 0 {
				propertySchema := createPropertySchema(examples)
				schema.Properties[fieldName] = propertySchema

				if !store.Optional[fieldName] {
					schema.Required = append(schema.Required, fieldName)
				}
			}
		}
	}

	return schema
}

// createPropertySchema creates a schema for a property based on its examples
func createPropertySchema(examples []interface{}) Schema {
	propertySchema := Schema{}
	if len(examples) > 0 {
		switch examples[0].(type) {
		case string:
			propertySchema.Type = "string"
		case float64:
			propertySchema.Type = "number"
		case bool:
			propertySchema.Type = "boolean"
		case []interface{}:
			propertySchema.Type = "array"
			propertySchema.Items = &Schema{Type: "object"}
		case map[string]interface{}:
			propertySchema.Type = "object"
		}
		propertySchema.Examples = examples
	}
	return propertySchema
}

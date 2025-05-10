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
	Enum        []string          `json:"enum,omitempty"`
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

		// Add URL parameters if they exist
		if endpoint.URLParameters != nil {
			for param, store := range endpoint.URLParameters.Examples {
				// Skip common parameters that are handled separately
				if param == "page" || param == "page_size" || param == "sort_by" || param == "order" || param == "search" {
					continue
				}

				// Determine parameter type based on examples
				paramType := "string"
				if len(store) > 0 {
					switch store[0].(type) {
					case bool:
						paramType = "boolean"
					case float64:
						paramType = "number"
					case int:
						paramType = "integer"
					}
				}

				// Create parameter
				param := Parameter{
					Name:        param,
					In:          "query",
					Required:    !endpoint.URLParameters.Optional[param],
					Description: fmt.Sprintf("Query parameter: %s", param),
					Schema: Schema{
						Type:     paramType,
						Examples: store,
					},
				}
				operation.Parameters = append(operation.Parameters, param)
			}

			// Add common query parameters
			commonParams := []struct {
				name        string
				description string
				type_       string
			}{
				{"page", "Page number for pagination", "integer"},
				{"page_size", "Number of items per page", "integer"},
				{"sort_by", "Field to sort by", "string"},
				{"order", "Sort order (asc/desc)", "string"},
				{"search", "Search query", "string"},
			}

			for _, cp := range commonParams {
				if store, exists := endpoint.URLParameters.Examples[cp.name]; exists {
					operation.Parameters = append(operation.Parameters, Parameter{
						Name:        cp.name,
						In:          "query",
						Required:    !endpoint.URLParameters.Optional[cp.name],
						Description: cp.description,
						Schema: Schema{
							Type:     cp.type_,
							Examples: store,
						},
					})
				}
			}
		}

		// Add request parameters from headers
		if endpoint.RequestHeaders != nil {
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
		}

		// Add request body schema if exists
		if endpoint.RequestPayload != nil && len(endpoint.RequestPayload.Examples) > 0 {
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
			if responseData.Headers != nil {
				for header, store := range responseData.Headers.Examples {
					response.Headers[header] = Header{
						Schema: Schema{
							Type:     "string",
							Examples: store,
						},
					}
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
	if store == nil || len(store.Examples) == 0 {
		return Schema{Type: "object"}
	}

	// Collect all top-level keys' prefixes
	var (
		arrayKey string
		allArray = true
		first    = true
	)
	for path := range store.Examples {
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			if strings.HasSuffix(parts[0], "[]") {
				if first {
					arrayKey = parts[0]
					first = false
				} else if parts[0] != arrayKey {
					allArray = false
					break
				}
			} else {
				allArray = false
				break
			}
		}
	}

	// Only treat as root array if all top-level keys start with the same array key
	if arrayKey != "" && allArray {
		itemStore := &SchemaStore{
			Examples: make(map[string][]interface{}),
			Optional: make(map[string]bool),
		}
		for path, examples := range store.Examples {
			parts := strings.Split(path, ".")
			if len(parts) > 1 {
				if strings.HasSuffix(parts[0], "[]") {
					newPath := strings.Join(parts[1:], ".")
					itemStore.Examples[newPath] = examples
					if optional, exists := store.Optional[path]; exists {
						itemStore.Optional[newPath] = optional
					}
				}
			}
		}
		itemSchema := buildObjectSchemaFromStore(itemStore)
		if itemSchema.Type == "" {
			itemSchema.Type = "object"
		}
		if itemSchema.Type == "object" && itemSchema.Properties == nil {
			itemSchema.Properties = make(map[string]Schema)
		}
		schema := Schema{
			Type:  "array",
			Items: &itemSchema,
		}
		return schema
	}

	// Otherwise, build as an object
	return buildObjectSchemaFromStore(store)
}

// createPropertySchema creates a schema for a property based on its examples
func createPropertySchema(examples []interface{}) Schema {
	propertySchema := Schema{}
	if len(examples) > 0 {
		switch examples[0].(type) {
		case string:
			propertySchema.Type = "string"
			// Check if we have a limited set of unique string values
			uniqueValues := make(map[string]bool)
			for _, ex := range examples {
				if str, ok := ex.(string); ok {
					uniqueValues[str] = true
				}
			}
			// If we have less than 5 unique values, add them as enum
			if len(uniqueValues) > 0 && len(uniqueValues) < 5 {
				enumValues := make([]string, 0, len(uniqueValues))
				for val := range uniqueValues {
					enumValues = append(enumValues, val)
				}
				propertySchema.Enum = enumValues
			}
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

// buildObjectSchemaFromStore builds an object schema from a SchemaStore
func buildObjectSchemaFromStore(store *SchemaStore) Schema {
	type node struct {
		children map[string]*node
		leaf     bool
		path     string
	}

	root := &node{children: make(map[string]*node)}

	// Build the tree
	for path := range store.Examples {
		parts := strings.Split(path, ".")
		cur := root
		for i, part := range parts {
			if _, ok := cur.children[part]; !ok {
				cur.children[part] = &node{children: make(map[string]*node)}
			}
			cur = cur.children[part]
			if i == len(parts)-1 {
				cur.leaf = true
				cur.path = path
			}
		}
	}

	var build func(n *node, isRoot bool) Schema
	build = func(n *node, isRoot bool) Schema {
		if n.leaf {
			examples := store.Examples[n.path]
			return createPropertySchema(examples)
		}

		// Only check for all-arrays if not at root
		if !isRoot {
			allArrays := true
			for k := range n.children {
				if !strings.HasSuffix(k, "[]") {
					allArrays = false
					break
				}
			}
			if allArrays && len(n.children) > 0 {
				objSchema := Schema{
					Type:       "object",
					Properties: make(map[string]Schema),
				}
				for k, child := range n.children {
					name := strings.TrimSuffix(k, "[]")
					childSchema := build(child, false)
					if childSchema.Type == "" {
						childSchema.Type = "object"
					}
					objSchema.Properties[name] = Schema{
						Type:  "array",
						Items: &childSchema,
					}
				}
				return objSchema
			}
		}

		// Otherwise, it's an object node
		objSchema := Schema{
			Type:       "object",
			Properties: make(map[string]Schema),
		}

		for k, child := range n.children {
			name := k
			if strings.HasSuffix(name, "[]") {
				name = strings.TrimSuffix(name, "[]")
				childSchema := build(child, false)
				if childSchema.Type == "" {
					childSchema.Type = "object"
				}
				objSchema.Properties[name] = Schema{
					Type:  "array",
					Items: &childSchema,
				}
			} else {
				childSchema := build(child, false)
				if childSchema.Type == "" {
					childSchema.Type = "object"
				}
				objSchema.Properties[name] = childSchema
			}

			fullPath := child.path
			if fullPath == "" {
				var pathParts []string
				cur := child
				for cur != nil && cur.path == "" && len(cur.children) == 1 {
					for kk := range cur.children {
						pathParts = append(pathParts, kk)
						cur = cur.children[kk]
						break
					}
				}
				if cur != nil && cur.path != "" {
					fullPath = cur.path
				} else if len(pathParts) > 0 {
					fullPath = strings.Join(pathParts, ".")
				}
			}
			if fullPath != "" && !store.Optional[fullPath] {
				objSchema.Required = append(objSchema.Required, name)
			}
		}
		return objSchema
	}

	schema := build(root, true)
	if schema.Type == "" {
		schema.Type = "object"
	}
	return schema
}

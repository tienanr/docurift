package analyzer

import (
	"fmt"
	"strings"
	"unicode"
)

// ToolSpec represents the specification for a single API tool
type ToolSpec struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Examples    []Example              `json:"examples,omitempty"`
}

// Example represents a request/response example for a tool
type Example struct {
	Request  interface{} `json:"request"`
	Response interface{} `json:"response"`
}

// ToolsSpec represents the complete tools specification
type ToolsSpec struct {
	Tools []ToolSpec `json:"tools"`
}

// OpenAIToolsSpec represents the OpenAI function calling format
type OpenAIToolsSpec struct {
	Tools []OpenAITool `json:"tools"`
}

// OpenAITool represents a tool in OpenAI's function calling format
type OpenAITool struct {
	Type     string   `json:"type"`
	Function ToolSpec `json:"function"`
}

// AnthropicToolsSpec represents the Anthropic tools format
type AnthropicToolsSpec struct {
	Tools []AnthropicTool `json:"tools"`
}

// AnthropicTool represents a tool in Anthropic's format
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Helper to generate a tool/function name from method and path
func generateToolName(method, path string) string {
	// Convert method to lowercase and path to a valid identifier
	name := strings.ToLower(method)

	// Split path into segments and join with underscores
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for _, segment := range segments {
		// Remove any non-alphanumeric characters
		segment = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '_'
		}, segment)
		name += "_" + segment
	}

	// Replace any double underscores with single
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	// Ensure the name starts with a letter
	if len(name) > 0 && !unicode.IsLetter(rune(name[0])) {
		name = "api_" + name
	}

	return name
}

// GenerateToolsSpec generates a tools specification from the analyzer data
func (a *Analyzer) GenerateToolsSpec() *ToolsSpec {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tools := make([]ToolSpec, 0)

	for key, endpoint := range a.endpoints {
		parts := strings.SplitN(key, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method := parts[0]
		path := parts[1]

		// Create tool spec
		tool := ToolSpec{
			Name:        generateToolName(method, path),
			Description: generateDescription(method, path, endpoint),
			Parameters:  generateParameters(endpoint),
			Examples:    generateExamples(endpoint),
		}

		tools = append(tools, tool)
	}

	return &ToolsSpec{
		Tools: tools,
	}
}

// GenerateOpenAIToolsSpec generates a tools specification in OpenAI's function calling format
func (a *Analyzer) GenerateOpenAIToolsSpec() *OpenAIToolsSpec {
	baseSpec := a.GenerateToolsSpec()
	tools := make([]OpenAITool, 0)

	for _, tool := range baseSpec.Tools {
		// Enhance the tool spec for OpenAI
		enhancedTool := enhanceToolForOpenAI(tool)

		openAITool := OpenAITool{
			Type:     "function",
			Function: enhancedTool,
		}
		tools = append(tools, openAITool)
	}

	return &OpenAIToolsSpec{
		Tools: tools,
	}
}

// enhanceToolForOpenAI enhances a tool specification with OpenAI-specific features
func enhanceToolForOpenAI(tool ToolSpec) ToolSpec {
	// Create a copy of the tool
	enhanced := tool

	// Enhance the description
	enhanced.Description = enhanceDescription(tool.Description)

	// Enhance parameters
	enhanced.Parameters = enhanceParameters(tool.Parameters)

	return enhanced
}

// enhanceDescription makes the description more suitable for OpenAI
func enhanceDescription(desc string) string {
	// Add more context and make it more action-oriented
	if strings.HasPrefix(desc, "Creates") {
		return desc + ". Use this function to add new data to the system."
	} else if strings.HasPrefix(desc, "Updates") {
		return desc + ". Use this function to modify existing data."
	} else if strings.HasPrefix(desc, "Deletes") {
		return desc + ". Use this function to remove data from the system."
	} else if strings.HasPrefix(desc, "Retrieves") {
		return desc + ". Use this function to get data from the system."
	}
	return desc
}

// enhanceParameters enhances parameter descriptions and adds OpenAI-specific features
func enhanceParameters(params map[string]interface{}) map[string]interface{} {
	enhanced := make(map[string]interface{})
	for k, v := range params {
		enhanced[k] = v
	}

	// Enhance properties
	if properties, ok := params["properties"].(map[string]interface{}); ok {
		enhancedProperties := make(map[string]interface{})
		for name, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				// Enhance property description
				if desc, ok := propMap["description"].(string); ok {
					propMap["description"] = enhancePropertyDescription(name, desc, propMap["type"].(string))
				}

				// Add format hints for common field types
				if propMap["type"] == "string" {
					if isEmail(name) {
						propMap["format"] = "email"
					} else if isURL(name) {
						propMap["format"] = "uri"
					} else if isDate(name) {
						propMap["format"] = "date-time"
					}
				}

				// Add min/max for numeric types
				if propMap["type"] == "number" || propMap["type"] == "integer" {
					if isAge(name) {
						propMap["minimum"] = 0
						propMap["maximum"] = 150
					} else if isCount(name) {
						propMap["minimum"] = 0
					}
				}

				enhancedProperties[name] = propMap
			}
		}
		enhanced["properties"] = enhancedProperties
	}

	return enhanced
}

// enhancePropertyDescription enhances a property description with more context
func enhancePropertyDescription(name, desc, type_ string) string {
	// Add type information to the description
	typeInfo := ""
	switch type_ {
	case "string":
		typeInfo = " (text)"
	case "number":
		typeInfo = " (number)"
	case "boolean":
		typeInfo = " (true/false)"
	case "array":
		typeInfo = " (list)"
	case "object":
		typeInfo = " (object)"
	}

	// Add common field hints
	if isEmail(name) {
		return desc + typeInfo + ". Must be a valid email address."
	} else if isURL(name) {
		return desc + typeInfo + ". Must be a valid URL."
	} else if isDate(name) {
		return desc + typeInfo + ". Must be a valid date/time."
	} else if isAge(name) {
		return desc + typeInfo + ". Must be a number between 0 and 150."
	} else if isCount(name) {
		return desc + typeInfo + ". Must be a non-negative number."
	}

	return desc + typeInfo
}

// Helper functions to identify common field types
func isEmail(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "email") || strings.Contains(name, "mail")
}

func isURL(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "url") || strings.Contains(name, "link") || strings.Contains(name, "uri")
}

func isDate(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "date") || strings.Contains(name, "time") || strings.Contains(name, "at")
}

func isAge(name string) bool {
	name = strings.ToLower(name)
	return name == "age"
}

func isCount(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "count") || strings.Contains(name, "quantity") || strings.Contains(name, "number")
}

// generateDescription creates a human-readable description for the endpoint
func generateDescription(method, path string, endpoint *EndpointData) string {
	// Use the most common response status to determine if it's a success endpoint
	maxCount := 0
	for _, data := range endpoint.ResponseStatuses {
		if len(data.Payload.Examples) > maxCount {
			maxCount = len(data.Payload.Examples)
		}
	}

	action := "Retrieves"
	switch method {
	case "POST":
		action = "Creates"
	case "PUT":
		action = "Updates"
	case "DELETE":
		action = "Deletes"
	case "PATCH":
		action = "Modifies"
	}

	return action + " data from " + path
}

// generateParameters creates the parameters specification from the endpoint data
func generateParameters(endpoint *EndpointData) map[string]interface{} {
	params := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}

	// Add URL parameters
	if endpoint.URLParameters != nil {
		for path, examples := range endpoint.URLParameters.Examples {
			paramName := strings.TrimPrefix(path, ".")
			paramType := inferType(examples)
			params["properties"].(map[string]interface{})[paramName] = map[string]interface{}{
				"type":        paramType,
				"description": "URL parameter: " + paramName,
			}
			if !endpoint.URLParameters.Optional[path] {
				params["required"] = append(params["required"].([]string), paramName)
			}
		}
	}

	// Add request payload parameters
	if endpoint.RequestPayload != nil {
		for path, examples := range endpoint.RequestPayload.Examples {
			paramName := strings.TrimPrefix(path, ".")
			paramType := inferType(examples)
			params["properties"].(map[string]interface{})[paramName] = map[string]interface{}{
				"type":        paramType,
				"description": "Request body parameter: " + paramName,
			}
			if !endpoint.RequestPayload.Optional[path] {
				params["required"] = append(params["required"].([]string), paramName)
			}
		}
	}

	return params
}

// generateExamples creates example request/response pairs
func generateExamples(endpoint *EndpointData) []Example {
	examples := make([]Example, 0)

	// Find the most common success response
	var successData *ResponseData
	maxCount := 0
	for _, data := range endpoint.ResponseStatuses {
		if len(data.Payload.Examples) > maxCount {
			successData = data
			maxCount = len(data.Payload.Examples)
		}
	}

	if successData != nil && len(successData.Payload.Examples) > 0 {
		// Use the first example for now
		example := Example{
			Request:  endpoint.RequestPayload.Examples[""],
			Response: successData.Payload.Examples[""],
		}
		examples = append(examples, example)
	}

	return examples
}

// inferType determines the JSON schema type from example values
func inferType(examples []interface{}) string {
	if len(examples) == 0 {
		return "string"
	}

	// Check first example to determine type
	switch examples[0].(type) {
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}

// GenerateSchema generates a JSON schema from the stored examples
func (s *SchemaStore) GenerateSchema() map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   make([]string, 0),
	}

	// Process each path and its examples
	for path, examples := range s.Examples {
		if len(examples) == 0 {
			continue
		}

		example := examples[0]
		propSchema := s.inferSchema(example)

		// Add description if we have examples
		if len(examples) > 0 {
			desc := "Examples: "
			for i, ex := range examples {
				if i > 0 {
					desc += ", "
				}
				desc += fmt.Sprintf("%v", ex)
			}
			propSchema["description"] = desc
		}

		schema["properties"].(map[string]interface{})[path] = propSchema
		if s.Optional != nil && !s.Optional[path] {
			schema["required"] = append(schema["required"].([]string), path)
		}
	}

	return schema
}

// inferSchema infers the JSON schema type from a value
func (s *SchemaStore) inferSchema(value interface{}) map[string]interface{} {
	schema := make(map[string]interface{})

	switch v := value.(type) {
	case string:
		schema["type"] = "string"
		if isUUID(v) {
			schema["format"] = "uuid"
		}
	case float64:
		schema["type"] = "number"
	case int:
		schema["type"] = "number"
	case bool:
		schema["type"] = "boolean"
	case []interface{}:
		schema["type"] = "array"
		if len(v) > 0 {
			schema["items"] = s.inferSchema(v[0])
		}
	case map[string]interface{}:
		schema["type"] = "object"
		props := make(map[string]interface{})
		for k, val := range v {
			props[k] = s.inferSchema(val)
		}
		schema["properties"] = props
	case nil:
		schema["type"] = "null"
	}

	return schema
}

// GenerateAnthropicToolsSpec generates a tools specification in Anthropic's format
func (a *Analyzer) GenerateAnthropicToolsSpec() *AnthropicToolsSpec {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tools := make([]AnthropicTool, 0)

	for _, endpoint := range a.endpoints {
		tool := AnthropicTool{
			Name:        generateToolName(endpoint.Method, endpoint.URL),
			Description: endpoint.Method + " " + endpoint.URL + " endpoint",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		}
		if endpoint.RequestPayload != nil {
			schema := endpoint.RequestPayload.GenerateSchema()
			if props, ok := schema["properties"].(map[string]interface{}); ok {
				tool.InputSchema["properties"] = props
				if required, ok := schema["required"].([]string); ok {
					tool.InputSchema["required"] = required
				}
			}
		}
		tools = append(tools, tool)
	}

	return &AnthropicToolsSpec{
		Tools: tools,
	}
}

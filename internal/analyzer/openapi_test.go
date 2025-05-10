package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOpenAPI(t *testing.T) {
	// Create a test analyzer with sample data
	a := &Analyzer{
		endpoints: map[string]*EndpointData{
			"GET /users": {
				URLParameters: &SchemaStore{
					Examples: map[string][]interface{}{
						"page":      {1, 2},
						"page_size": {10, 20},
						"search":    {"john", "jane"},
					},
					Optional: map[string]bool{
						"page":      false,
						"page_size": false,
						"search":    true,
					},
				},
				ResponseStatuses: map[int]*ResponseData{
					200: {
						Payload: &SchemaStore{
							Examples: map[string][]interface{}{
								"id":       {1, 2},
								"name":     {"John", "Jane"},
								"email":    {"john@example.com", "jane@example.com"},
								"active":   {true, false},
								"metadata": {map[string]interface{}{"role": "admin"}},
							},
							Optional: map[string]bool{
								"id":       false,
								"name":     false,
								"email":    false,
								"active":   true,
								"metadata": true,
							},
						},
					},
				},
			},
			"POST /users": {
				RequestPayload: &SchemaStore{
					Examples: map[string][]interface{}{
						"name":     {"John", "Jane"},
						"email":    {"john@example.com", "jane@example.com"},
						"password": {"secret123", "secret456"},
					},
					Optional: map[string]bool{
						"name":     false,
						"email":    false,
						"password": false,
					},
				},
				ResponseStatuses: map[int]*ResponseData{
					201: {
						Payload: &SchemaStore{
							Examples: map[string][]interface{}{
								"id": {1, 2},
							},
							Optional: map[string]bool{
								"id": false,
							},
						},
					},
				},
			},
			"POST /invoices": {
				RequestPayload: &SchemaStore{
					Examples: map[string][]interface{}{
						"[].due_date":                             {"2025-06-08T22:16:33.230621-07:00", "2025-06-08T22:16:33.232269-07:00", "2025-06-08T22:16:33.232583-07:00"},
						"[].id":                                   {230, 3156, 603},
						"[].invoice_number":                       {"INV-001", "INV-002"},
						"[].issue_date":                           {"2025-05-09T22:16:33.230621-07:00", "2025-05-09T22:16:33.232269-07:00", "2025-05-09T22:16:33.232583-07:00"},
						"[].line_items[].description":             {"High-end laptop", "Wireless mouse"},
						"[].line_items[].id":                      {0},
						"[].line_items[].product_id":              {1, 2},
						"[].line_items[].quantity":                {2, 1},
						"[].line_items[].tax_info[].description":  {"California State Tax", "Los Angeles County Tax"},
						"[].line_items[].tax_info[].id":           {0},
						"[].line_items[].tax_info[].jurisdiction": {"CA", "LA", "NY"},
						"[].line_items[].tax_info[].tax_amount":   {169.9983, 39.9996, 2.54915, 8.5, 8.875},
						"[].line_items[].tax_info[].tax_rate":     {8.5, 2, 8.875},
						"[].line_items[].total_price":             {1999.98, 29.99, 100},
						"[].line_items[].unit_price":              {999.99, 29.99, 100, 50},
						"[].metadata.currency":                    {"USD"},
						"[].metadata.payment_method":              {"credit_card"},
						"[].notes":                                {"Net 30 payment terms"},
						"[].order_id":                             {1, 2},
						"[].payment_terms":                        {"Due upon receipt"},
						"[].status":                               {"pending", "paid"},
						"[].subtotal":                             {2029.97, 100},
						"[].total":                                {2242.51705, 108.5, 108.875},
						"[].total_tax":                            {212.54705, 8.5, 8.875},
						"[].user_id":                              {1},
					},
					Optional: map[string]bool{
						"[].due_date":                             true,
						"[].id":                                   true,
						"[].invoice_number":                       true,
						"[].issue_date":                           true,
						"[].line_items[].description":             true,
						"[].line_items[].id":                      true,
						"[].line_items[].product_id":              true,
						"[].line_items[].quantity":                true,
						"[].line_items[].tax_info[].description":  true,
						"[].line_items[].tax_info[].id":           true,
						"[].line_items[].tax_info[].jurisdiction": true,
						"[].line_items[].tax_info[].tax_amount":   true,
						"[].line_items[].tax_info[].tax_rate":     true,
						"[].line_items[].total_price":             true,
						"[].line_items[].unit_price":              true,
						"[].metadata.currency":                    true,
						"[].metadata.payment_method":              true,
						"[].notes":                                true,
						"[].order_id":                             true,
						"[].payment_terms":                        true,
						"[].status":                               true,
						"[].subtotal":                             true,
						"[].total":                                true,
						"[].total_tax":                            true,
						"[].user_id":                              true,
					},
				},
				ResponseStatuses: map[int]*ResponseData{
					201: {
						Payload: &SchemaStore{
							Examples: map[string][]interface{}{
								"invoice_id": {1001},
							},
							Optional: map[string]bool{
								"invoice_id": false,
							},
						},
					},
				},
			},
		},
	}

	// Generate OpenAPI spec
	openAPI := a.GenerateOpenAPI()

	// Test basic structure
	assert.Equal(t, "3.0.0", openAPI.OpenAPI)
	assert.Equal(t, "API Documentation", openAPI.Info.Title)
	assert.Equal(t, "1.0.0", openAPI.Info.Version)

	// Test GET /users endpoint
	getUsersPath, exists := openAPI.Paths["/users"]
	assert.True(t, exists)
	assert.NotNil(t, getUsersPath.Get)

	// Test query parameters
	getOp := getUsersPath.Get
	assert.Len(t, getOp.Parameters, 3)

	// Verify page parameter
	var pageParam *Parameter
	for _, p := range getOp.Parameters {
		if p.Name == "page" {
			pageParam = &p
			break
		}
	}
	assert.NotNil(t, pageParam)
	assert.Equal(t, "query", pageParam.In)
	assert.True(t, pageParam.Required)
	assert.Equal(t, "integer", pageParam.Schema.Type)

	// Test response schema
	response200 := getOp.Responses["200"]
	assert.NotNil(t, response200)
	assert.Equal(t, "Status 200", response200.Description)

	// Test POST /users endpoint
	postUsersPath, exists := openAPI.Paths["/users"]
	assert.True(t, exists)
	assert.NotNil(t, postUsersPath.Post)

	// Test request body
	postOp := postUsersPath.Post
	assert.NotNil(t, postOp.RequestBody)
	assert.True(t, postOp.RequestBody.Required)

	// Verify request body schema
	content := postOp.RequestBody.Content["application/json"]
	assert.NotNil(t, content.Schema)
	assert.Equal(t, "object", content.Schema.Type)
	assert.NotEmpty(t, content.Schema.Properties)
	assert.Contains(t, content.Schema.Properties, "name")
	assert.Contains(t, content.Schema.Properties, "email")
	assert.Contains(t, content.Schema.Properties, "password")

	// Test POST /invoices endpoint (deep nested object, root array)
	postInvoicesPath, exists := openAPI.Paths["/invoices"]
	assert.True(t, exists)
	assert.NotNil(t, postInvoicesPath.Post)

	postInvoicesOp := postInvoicesPath.Post
	assert.NotNil(t, postInvoicesOp.RequestBody)
	invoiceArraySchema := postInvoicesOp.RequestBody.Content["application/json"].Schema
	assert.Equal(t, "array", invoiceArraySchema.Type)
	assert.NotNil(t, invoiceArraySchema.Items)

	// Check the array item schema
	invoiceSchema := invoiceArraySchema.Items
	assert.Equal(t, "object", invoiceSchema.Type)
	assert.NotNil(t, invoiceSchema.Properties)
	assert.Contains(t, invoiceSchema.Properties, "id")
	assert.Contains(t, invoiceSchema.Properties, "invoice_number")
	assert.Contains(t, invoiceSchema.Properties, "issue_date")
	assert.Contains(t, invoiceSchema.Properties, "due_date")
	assert.Contains(t, invoiceSchema.Properties, "status")
	assert.Contains(t, invoiceSchema.Properties, "subtotal")
	assert.Contains(t, invoiceSchema.Properties, "total_tax")
	assert.Contains(t, invoiceSchema.Properties, "total")
	assert.Contains(t, invoiceSchema.Properties, "payment_terms")
	assert.Contains(t, invoiceSchema.Properties, "notes")
	assert.Contains(t, invoiceSchema.Properties, "user_id")
	assert.Contains(t, invoiceSchema.Properties, "order_id")
	assert.Contains(t, invoiceSchema.Properties, "line_items")
	assert.Contains(t, invoiceSchema.Properties, "metadata")

	// Check line_items array schema
	lineItemsSchema := invoiceSchema.Properties["line_items"]
	assert.Equal(t, "array", lineItemsSchema.Type)
	assert.NotNil(t, lineItemsSchema.Items)
	assert.Equal(t, "object", lineItemsSchema.Items.Type)
	assert.NotNil(t, lineItemsSchema.Items.Properties)
	assert.Contains(t, lineItemsSchema.Items.Properties, "id")
	assert.Contains(t, lineItemsSchema.Items.Properties, "product_id")
	assert.Contains(t, lineItemsSchema.Items.Properties, "quantity")
	assert.Contains(t, lineItemsSchema.Items.Properties, "unit_price")
	assert.Contains(t, lineItemsSchema.Items.Properties, "total_price")
	assert.Contains(t, lineItemsSchema.Items.Properties, "description")
	assert.Contains(t, lineItemsSchema.Items.Properties, "tax_info")

	// Check tax_info array schema
	taxInfoSchema := lineItemsSchema.Items.Properties["tax_info"]
	assert.Equal(t, "array", taxInfoSchema.Type)
	assert.NotNil(t, taxInfoSchema.Items)
	assert.Equal(t, "object", taxInfoSchema.Items.Type)
	assert.NotNil(t, taxInfoSchema.Items.Properties)
	assert.Contains(t, taxInfoSchema.Items.Properties, "id")
	assert.Contains(t, taxInfoSchema.Items.Properties, "description")
	assert.Contains(t, taxInfoSchema.Items.Properties, "jurisdiction")
	assert.Contains(t, taxInfoSchema.Items.Properties, "tax_rate")
	assert.Contains(t, taxInfoSchema.Items.Properties, "tax_amount")

	// Check metadata object schema
	metadataSchema := invoiceSchema.Properties["metadata"]
	assert.Equal(t, "object", metadataSchema.Type)
	assert.NotNil(t, metadataSchema.Properties)
	assert.Contains(t, metadataSchema.Properties, "currency")
	assert.Contains(t, metadataSchema.Properties, "payment_method")
}

func TestGenerateSchemaFromStore(t *testing.T) {
	// Test array schema
	arrayStore := &SchemaStore{
		Examples: map[string][]interface{}{
			"items[].id":   {1, 2},
			"items[].name": {"John", "Jane"},
		},
		Optional: map[string]bool{
			"items[].id":   false,
			"items[].name": false,
		},
	}

	arraySchema := generateSchemaFromStore(arrayStore)
	assert.Equal(t, "array", arraySchema.Type)
	assert.NotNil(t, arraySchema.Items)
	assert.Equal(t, "object", arraySchema.Items.Type)
	assert.Len(t, arraySchema.Items.Properties, 2)
	assert.Contains(t, arraySchema.Items.Properties, "id")
	assert.Contains(t, arraySchema.Items.Properties, "name")

	// Test nested object schema
	nestedStore := &SchemaStore{
		Examples: map[string][]interface{}{
			"user.id":           {1, 2},
			"user.name":         {"John", "Jane"},
			"user.address.city": {"New York", "London"},
		},
		Optional: map[string]bool{
			"user.id":           false,
			"user.name":         false,
			"user.address.city": true,
		},
	}

	nestedSchema := generateSchemaFromStore(nestedStore)
	assert.Equal(t, "object", nestedSchema.Type)
	assert.Contains(t, nestedSchema.Properties, "user")

	userSchema := nestedSchema.Properties["user"]
	assert.Equal(t, "object", userSchema.Type)
	assert.Contains(t, userSchema.Properties, "id")
	assert.Contains(t, userSchema.Properties, "name")
	assert.Contains(t, userSchema.Properties, "address")
}

func TestCreatePropertySchema(t *testing.T) {
	tests := []struct {
		name     string
		examples []interface{}
		wantType string
	}{
		{
			name:     "string property",
			examples: []interface{}{"test", "example"},
			wantType: "string",
		},
		{
			name:     "number property",
			examples: []interface{}{1.5, 2.7},
			wantType: "number",
		},
		{
			name:     "boolean property",
			examples: []interface{}{true, false},
			wantType: "boolean",
		},
		{
			name:     "array property",
			examples: []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}},
			wantType: "array",
		},
		{
			name:     "object property",
			examples: []interface{}{map[string]interface{}{"key": "value"}},
			wantType: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := createPropertySchema(tt.examples)
			assert.Equal(t, tt.wantType, schema.Type)
			assert.Equal(t, tt.examples, schema.Examples)
		})
	}
}

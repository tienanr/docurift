package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var baseURL = getBaseURL()

func getBaseURL() string {
	if url := os.Getenv("PROXY_URL"); url != "" {
		return url
	}
	return "http://localhost:9876"
}

// --- Utility Functions ---

func doGet(t *testing.T, path string) *http.Response {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	return resp
}

func doPost(t *testing.T, path string, body any) *http.Response {
	jsonData, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	return resp
}

// --- Test Cases ---

func TestGetProducts(t *testing.T) {
	resp := doGet(t, "/products")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestCreateAndDeleteProduct(t *testing.T) {
	// Create multiple products with different fields
	products := []map[string]interface{}{
		{
			"name":     "Laptop",
			"price":    999.99,
			"category": "Electronics",
			"inStock":  true,
		},
		{
			"name":        "Smartphone",
			"price":       699.99,
			"category":    "Electronics",
			"inStock":     false,
			"description": "Latest model",
		},
		{
			"name":     "Headphones",
			"price":    199.99,
			"category": "Audio",
			"inStock":  true,
			"color":    "Black",
		},
	}

	for _, product := range products {
		resp := doPost(t, "/products", product)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 from POST, got %d", resp.StatusCode)
		}
	}
}

func TestGetUsers(t *testing.T) {
	resp := doGet(t, "/users")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /users failed: %d", resp.StatusCode)
	}
}

func TestCreateUsers(t *testing.T) {
	// Create multiple users with different fields
	users := []map[string]interface{}{
		{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		},
		{
			"name":    "Jane Smith",
			"email":   "jane@example.com",
			"age":     25,
			"phone":   "123-456-7890",
			"address": "123 Main St",
		},
		{
			"name":     "Bob Wilson",
			"email":    "bob@example.com",
			"age":      40,
			"company":  "Tech Corp",
			"position": "Developer",
		},
	}

	for _, user := range users {
		resp := doPost(t, "/users", user)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /users failed: %d", resp.StatusCode)
		}
	}
}

func TestGetUserByID(t *testing.T) {
	resp := doGet(t, "/users/1")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /users/1 failed: %d", resp.StatusCode)
	}
}

func TestOrders(t *testing.T) {
	// Create multiple orders with different fields
	orders := []map[string]interface{}{
		{
			"user_id":    1,
			"product_id": 1,
			"quantity":   2,
		},
		{
			"user_id":    2,
			"product_id": 2,
			"quantity":   1,
			"notes":      "Gift wrapping requested",
			"shipping": map[string]interface{}{
				"address": "123 Main St",
				"city":    "New York",
			},
		},
		{
			"user_id":    3,
			"product_id": 1,
			"quantity":   3,
			"priority":   true,
			"shipping": map[string]interface{}{
				"address": "456 Oak Ave",
				"city":    "Los Angeles",
				"zip":     "90001",
			},
		},
	}

	for _, order := range orders {
		resp := doPost(t, "/orders", order)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /orders failed: %d", resp.StatusCode)
		}
	}
}

func TestInvalidMethodsViaProxy(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPut, "/orders"},
		{http.MethodPut, "/products/1"},
		{http.MethodPatch, "/users/1"},
		{http.MethodPatch, "/categories/1"},
		{http.MethodPut, "/reviews/1"},
		{http.MethodPatch, "/addresses/1"},
		{http.MethodPut, "/payment-methods/1"},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, baseURL+test.path, bytes.NewReader([]byte(`{}`)))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("%s %s failed to send request: %v", test.method, test.path, err)
			continue
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405 for %s %s, got %d", test.method, test.path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func TestReviews(t *testing.T) {
	// Test creating a review
	review := Review{
		UserID:       1,
		ProductID:    1,
		Rating:       5,
		Comment:      "Great product!",
		Title:        "Excellent quality",
		HelpfulVotes: 10,
		Metadata: map[string]string{
			"verified_purchase": "true",
			"platform":          "web",
		},
	}
	reviewJSON, _ := json.Marshal(review)
	resp, err := http.Post(baseURL+"/reviews", "application/json", bytes.NewBuffer(reviewJSON))
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting reviews
	resp, err = http.Get(baseURL + "/reviews")
	if err != nil {
		t.Fatalf("Failed to get reviews: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestCategories(t *testing.T) {
	// Test getting categories
	resp, err := http.Get(baseURL + "/categories")
	if err != nil {
		t.Fatalf("Failed to get categories: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test creating a category with optional fields
	category := Category{
		Name:        "Books",
		Description: "Books and publications",
		ParentID:    1,
		ImageURL:    "https://example.com/books.jpg",
		Attributes: map[string]string{
			"type":   "physical",
			"format": "paperback",
		},
	}
	categoryJSON, _ := json.Marshal(category)
	resp, err = http.Post(baseURL+"/categories", "application/json", bytes.NewBuffer(categoryJSON))
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestAddresses(t *testing.T) {
	// Test creating an address with optional fields
	address := Address{
		UserID:      1,
		Street:      "123 Main St",
		City:        "New York",
		State:       "NY",
		Country:     "USA",
		PostalCode:  "10001",
		Apartment:   "4B",
		IsDefault:   true,
		PhoneNumber: "+1-555-0123",
	}
	addressJSON, _ := json.Marshal(address)
	resp, err := http.Post(baseURL+"/addresses", "application/json", bytes.NewBuffer(addressJSON))
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting addresses
	resp, err = http.Get(baseURL + "/addresses")
	if err != nil {
		t.Fatalf("Failed to get addresses: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestPaymentMethods(t *testing.T) {
	// Test creating a payment method with sensitive data
	payment := PaymentMethod{
		UserID:           1,
		CardNumber:       "4111111111111111",
		ExpiryDate:       "12/25",
		CardholderName:   "John Doe",
		CVV:              "123",
		IsDefault:        true,
		BillingAddressID: 1,
	}
	paymentJSON, _ := json.Marshal(payment)
	resp, err := http.Post(baseURL+"/payment-methods", "application/json", bytes.NewBuffer(paymentJSON))
	if err != nil {
		t.Fatalf("Failed to create payment method: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting payment methods
	resp, err = http.Get(baseURL + "/payment-methods")
	if err != nil {
		t.Fatalf("Failed to get payment methods: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestSensitiveData(t *testing.T) {
	// Test user with sensitive data
	user := User{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secret123",   // Sensitive data
		SSN:      "123-45-6789", // Sensitive data
	}
	userJSON, _ := json.Marshal(user)
	resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to create user with sensitive data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test payment method with sensitive data
	payment := PaymentMethod{
		UserID:         1,
		CardNumber:     "4111111111111111", // Sensitive data
		ExpiryDate:     "12/25",
		CardholderName: "John Doe",
		CVV:            "123", // Sensitive data
	}
	paymentJSON, _ := json.Marshal(payment)
	resp, err = http.Post(baseURL+"/payment-methods", "application/json", bytes.NewBuffer(paymentJSON))
	if err != nil {
		t.Fatalf("Failed to create payment method with sensitive data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestOptionalFields(t *testing.T) {
	// Test product with optional fields
	product := Product{
		ID:          1,
		Name:        "Test Product",
		Price:       99.99,
		Description: "A test product with optional fields",
		Category:    "Test",
		Tags:        []string{"test", "optional"},
		Metadata: map[string]string{
			"color": "red",
			"size":  "medium",
		},
	}
	productJSON, _ := json.Marshal(product)
	resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(productJSON))
	if err != nil {
		t.Fatalf("Failed to create product with optional fields: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test review with optional fields
	review := Review{
		UserID:       1,
		ProductID:    1,
		Rating:       5,
		Comment:      "Great product!",
		Title:        "Optional title",
		HelpfulVotes: 5,
		Metadata: map[string]string{
			"verified": "true",
		},
	}
	reviewJSON, _ := json.Marshal(review)
	resp, err = http.Post(baseURL+"/reviews", "application/json", bytes.NewBuffer(reviewJSON))
	if err != nil {
		t.Fatalf("Failed to create review with optional fields: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestGetProductByID(t *testing.T) {
	// First create a product
	product := Product{
		Name:     "Test Product",
		Price:    99.99,
		Category: "Test",
	}
	productJSON, _ := json.Marshal(product)
	resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(productJSON))
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created product's ID
	var createdProduct Product
	if err := json.NewDecoder(resp.Body).Decode(&createdProduct); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Test getting the product by ID
	resp, err = http.Get(fmt.Sprintf("%s/products/%d", baseURL, createdProduct.ID))
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting non-existent product
	resp, err = http.Get(fmt.Sprintf("%s/products/99999", baseURL))
	if err != nil {
		t.Fatalf("Failed to get non-existent product: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
	}
}

func TestGetCategoryByID(t *testing.T) {
	// First create a category
	category := Category{
		Name:        "Test Category",
		Description: "Test Description",
		ParentID:    1,
	}
	categoryJSON, _ := json.Marshal(category)
	resp, err := http.Post(baseURL+"/categories", "application/json", bytes.NewBuffer(categoryJSON))
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created category's ID
	var createdCategory Category
	if err := json.NewDecoder(resp.Body).Decode(&createdCategory); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Test getting the category by ID
	resp, err = http.Get(fmt.Sprintf("%s/categories/%d", baseURL, createdCategory.ID))
	if err != nil {
		t.Fatalf("Failed to get category: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting non-existent category
	resp, err = http.Get(fmt.Sprintf("%s/categories/99999", baseURL))
	if err != nil {
		t.Fatalf("Failed to get non-existent category: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
	}
}

func TestGetPaymentMethodByID(t *testing.T) {
	// First create a payment method
	payment := PaymentMethod{
		UserID:         1,
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CardholderName: "John Doe",
	}
	paymentJSON, _ := json.Marshal(payment)
	resp, err := http.Post(baseURL+"/payment-methods", "application/json", bytes.NewBuffer(paymentJSON))
	if err != nil {
		t.Fatalf("Failed to create payment method: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created payment method's ID
	var createdPayment PaymentMethod
	if err := json.NewDecoder(resp.Body).Decode(&createdPayment); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Test getting the payment method by ID
	resp, err = http.Get(fmt.Sprintf("%s/payment-methods/%d", baseURL, createdPayment.ID))
	if err != nil {
		t.Fatalf("Failed to get payment method: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting non-existent payment method
	resp, err = http.Get(fmt.Sprintf("%s/payment-methods/99999", baseURL))
	if err != nil {
		t.Fatalf("Failed to get non-existent payment method: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
	}
}

func TestGetReviewByID(t *testing.T) {
	// First create a review
	review := Review{
		UserID:    1,
		ProductID: 1,
		Rating:    5,
		Comment:   "Test review",
	}
	reviewJSON, _ := json.Marshal(review)
	resp, err := http.Post(baseURL+"/reviews", "application/json", bytes.NewBuffer(reviewJSON))
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created review's ID
	var createdReview Review
	if err := json.NewDecoder(resp.Body).Decode(&createdReview); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Test getting the review by ID
	resp, err = http.Get(fmt.Sprintf("%s/reviews/%d", baseURL, createdReview.ID))
	if err != nil {
		t.Fatalf("Failed to get review: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting non-existent review
	resp, err = http.Get(fmt.Sprintf("%s/reviews/99999", baseURL))
	if err != nil {
		t.Fatalf("Failed to get non-existent review: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
	}
}

func TestGetAddressByID(t *testing.T) {
	// First create an address
	address := Address{
		UserID:     1,
		Street:     "123 Test St",
		City:       "Test City",
		State:      "TS",
		Country:    "Test Country",
		PostalCode: "12345",
	}
	addressJSON, _ := json.Marshal(address)
	resp, err := http.Post(baseURL+"/addresses", "application/json", bytes.NewBuffer(addressJSON))
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created address's ID
	var createdAddress Address
	if err := json.NewDecoder(resp.Body).Decode(&createdAddress); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Test getting the address by ID
	resp, err = http.Get(fmt.Sprintf("%s/addresses/%d", baseURL, createdAddress.ID))
	if err != nil {
		t.Fatalf("Failed to get address: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Test getting non-existent address
	resp, err = http.Get(fmt.Sprintf("%s/addresses/99999", baseURL))
	if err != nil {
		t.Fatalf("Failed to get non-existent address: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
	}
}

func TestInvalidIDFormats(t *testing.T) {
	// Test invalid ID formats for all endpoints
	endpoints := []string{
		"/products",
		"/categories",
		"/reviews",
		"/addresses",
		"/payment-methods",
	}

	invalidIDs := []string{
		"abc",  // non-numeric
		"1.23", // decimal
		"-1",   // negative
		"0",    // zero
		"",     // empty
		"1abc", // mixed
	}

	for _, endpoint := range endpoints {
		for _, invalidID := range invalidIDs {
			resp, err := http.Get(fmt.Sprintf("%s%s/%s", baseURL, endpoint, invalidID))
			if err != nil {
				t.Errorf("Failed to get %s with invalid ID %s: %v", endpoint, invalidID, err)
				continue
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status BadRequest for %s with invalid ID %s, got %v", endpoint, invalidID, resp.StatusCode)
			}
			resp.Body.Close()
		}
	}
}

func TestProductURLParameters(t *testing.T) {
	// Create test products
	products := []Product{
		{Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true, Description: "High-end laptop"},
		{Name: "Smartphone", Price: 699.99, Category: "Electronics", InStock: true, Description: "Latest model"},
		{Name: "Headphones", Price: 199.99, Category: "Audio", InStock: false, Description: "Noise cancelling"},
		{Name: "Keyboard", Price: 79.99, Category: "Electronics", InStock: true, Description: "Mechanical keyboard"},
		{Name: "Mouse", Price: 49.99, Category: "Electronics", InStock: true, Description: "Wireless mouse"},
	}

	// Create products first
	for _, p := range products {
		productJSON, _ := json.Marshal(p)
		resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(productJSON))
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
		resp.Body.Close()
	}

	// Test pagination
	t.Run("Pagination", func(t *testing.T) {
		resp := doGet(t, "/products?page=1&page_size=2")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Product
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 products, got %d", len(result))
		}
	})

	// Test sorting
	t.Run("Sorting", func(t *testing.T) {
		resp := doGet(t, "/products?sort_by=price&order=desc")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Product
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if len(result) < 2 {
			t.Fatalf("Expected at least 2 products")
		}
		if result[0].Price < result[1].Price {
			t.Error("Products not sorted by price in descending order")
		}
	})

	// Test filtering
	t.Run("Filtering", func(t *testing.T) {
		resp := doGet(t, "/products?filter_category=Electronics&filter_in_stock=true")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Product
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, p := range result {
			if p.Category != "Electronics" || !p.InStock {
				t.Errorf("Product does not match filter criteria: %+v", p)
			}
		}
	})

	// Test searching
	t.Run("Searching", func(t *testing.T) {
		resp := doGet(t, "/products?search=laptop")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Product
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if len(result) == 0 {
			t.Error("Expected to find products matching 'laptop'")
		}
		for _, p := range result {
			if !strings.Contains(strings.ToLower(p.Name), "laptop") &&
				!strings.Contains(strings.ToLower(p.Description), "laptop") {
				t.Errorf("Product does not match search criteria: %+v", p)
			}
		}
	})
}

func TestReviewURLParameters(t *testing.T) {
	// Create test reviews
	reviews := []Review{
		{UserID: 1, ProductID: 1, Rating: 5, Comment: "Great product!"},
		{UserID: 1, ProductID: 2, Rating: 4, Comment: "Good but expensive"},
		{UserID: 2, ProductID: 1, Rating: 3, Comment: "Average product"},
		{UserID: 2, ProductID: 2, Rating: 5, Comment: "Excellent!"},
	}

	// Create reviews first
	for _, r := range reviews {
		reviewJSON, _ := json.Marshal(r)
		resp, err := http.Post(baseURL+"/reviews", "application/json", bytes.NewBuffer(reviewJSON))
		if err != nil {
			t.Fatalf("Failed to create review: %v", err)
		}
		resp.Body.Close()
	}

	// Test filtering by user
	t.Run("FilterByUser", func(t *testing.T) {
		resp := doGet(t, "/reviews?filter_user_id=1")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Review
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, r := range result {
			if r.UserID != 1 {
				t.Errorf("Review does not belong to user 1: %+v", r)
			}
		}
	})

	// Test filtering by product
	t.Run("FilterByProduct", func(t *testing.T) {
		resp := doGet(t, "/reviews?filter_product_id=1")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Review
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, r := range result {
			if r.ProductID != 1 {
				t.Errorf("Review does not belong to product 1: %+v", r)
			}
		}
	})

	// Test filtering by rating
	t.Run("FilterByRating", func(t *testing.T) {
		resp := doGet(t, "/reviews?filter_rating=5")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Review
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, r := range result {
			if r.Rating != 5 {
				t.Errorf("Review does not have rating 5: %+v", r)
			}
		}
	})

	// Test pagination
	t.Run("Pagination", func(t *testing.T) {
		resp := doGet(t, "/reviews?page=1&page_size=2")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Review
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 reviews, got %d", len(result))
		}
	})
}

func TestInvoices(t *testing.T) {
	// Create test products first
	products := []Product{
		{Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true},
		{Name: "Mouse", Price: 29.99, Category: "Electronics", InStock: true},
	}

	for _, p := range products {
		productJSON, _ := json.Marshal(p)
		resp, err := http.Post(baseURL+"/products", "application/json", bytes.NewBuffer(productJSON))
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
		resp.Body.Close()
	}

	// Test creating an invoice
	invoice := Invoice{
		UserID:        1,
		OrderID:       1,
		InvoiceNumber: "INV-001",
		IssueDate:     time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		Status:        "pending",
		LineItems: []LineItem{
			{
				ProductID:   1,
				Quantity:    2,
				UnitPrice:   999.99,
				Description: "High-end laptop",
				TaxInfo: []TaxInfo{
					{
						Jurisdiction: "CA",
						TaxRate:      8.5,
						Description:  "California State Tax",
					},
					{
						Jurisdiction: "LA",
						TaxRate:      2.0,
						Description:  "Los Angeles County Tax",
					},
				},
			},
			{
				ProductID:   2,
				Quantity:    1,
				UnitPrice:   29.99,
				Description: "Wireless mouse",
				TaxInfo: []TaxInfo{
					{
						Jurisdiction: "CA",
						TaxRate:      8.5,
						Description:  "California State Tax",
					},
				},
			},
		},
		Notes:        "Net 30 payment terms",
		PaymentTerms: "Due upon receipt",
		Metadata: map[string]string{
			"payment_method": "credit_card",
			"currency":       "USD",
		},
	}

	invoiceJSON, _ := json.Marshal(invoice)
	resp, err := http.Post(baseURL+"/invoices", "application/json", bytes.NewBuffer(invoiceJSON))
	if err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Decode the response to get the created invoice
	var createdInvoice Invoice
	if err := json.NewDecoder(resp.Body).Decode(&createdInvoice); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	// Verify the invoice calculations
	expectedSubtotal := (999.99 * 2) + 29.99
	if createdInvoice.Subtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %v, got %v", expectedSubtotal, createdInvoice.Subtotal)
	}

	// Calculate expected tax for laptop (2 items)
	laptopTax := (999.99 * 2) * (0.085 + 0.02)
	// Calculate expected tax for mouse (1 item)
	mouseTax := 29.99 * 0.085
	expectedTax := laptopTax + mouseTax

	// Use epsilon for floating-point comparison
	const epsilon = 0.000001
	if math.Abs(createdInvoice.TotalTax-expectedTax) > epsilon {
		t.Errorf("Expected total tax %v, got %v", expectedTax, createdInvoice.TotalTax)
	}

	expectedTotal := expectedSubtotal + expectedTax
	if math.Abs(createdInvoice.Total-expectedTotal) > epsilon {
		t.Errorf("Expected total %v, got %v", expectedTotal, createdInvoice.Total)
	}

	// Test getting the invoice by ID
	resp, err = http.Get(fmt.Sprintf("%s/invoices/%d", baseURL, createdInvoice.ID))
	if err != nil {
		t.Fatalf("Failed to get invoice: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
	resp.Body.Close()

	// Test filtering invoices
	resp, err = http.Get(fmt.Sprintf("%s/invoices?filter_user_id=1", baseURL))
	if err != nil {
		t.Fatalf("Failed to get filtered invoices: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
	resp.Body.Close()

	// Test pagination
	resp, err = http.Get(fmt.Sprintf("%s/invoices?page=1&page_size=10", baseURL))
	if err != nil {
		t.Fatalf("Failed to get paginated invoices: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestInvoiceURLParameters(t *testing.T) {
	// Create test invoices with different statuses
	invoices := []Invoice{
		{
			UserID:        1,
			OrderID:       1,
			Status:        "pending",
			InvoiceNumber: "INV-001",
			LineItems: []LineItem{
				{
					ProductID: 1,
					Quantity:  1,
					UnitPrice: 100.00,
					TaxInfo: []TaxInfo{
						{Jurisdiction: "CA", TaxRate: 8.5},
					},
				},
			},
		},
		{
			UserID:        1,
			OrderID:       2,
			Status:        "paid",
			InvoiceNumber: "INV-002",
			LineItems: []LineItem{
				{
					ProductID: 2,
					Quantity:  2,
					UnitPrice: 50.00,
					TaxInfo: []TaxInfo{
						{Jurisdiction: "NY", TaxRate: 8.875},
					},
				},
			},
		},
		{
			UserID:        2,
			OrderID:       3,
			Status:        "overdue",
			InvoiceNumber: "INV-003",
			LineItems: []LineItem{
				{
					ProductID: 3,
					Quantity:  3,
					UnitPrice: 75.00,
					TaxInfo: []TaxInfo{
						{Jurisdiction: "TX", TaxRate: 6.25},
					},
				},
			},
		},
	}

	// Create invoices first
	for _, inv := range invoices {
		invoiceJSON, _ := json.Marshal(inv)
		resp, err := http.Post(baseURL+"/invoices", "application/json", bytes.NewBuffer(invoiceJSON))
		if err != nil {
			t.Fatalf("Failed to create invoice: %v", err)
		}
		resp.Body.Close()
	}

	// Test filtering by user
	t.Run("FilterByUser", func(t *testing.T) {
		resp := doGet(t, "/invoices?filter_user_id=1")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Invoice
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, inv := range result {
			if inv.UserID != 1 {
				t.Errorf("Invoice does not belong to user 1: %+v", inv)
			}
		}
	})

	// Test filtering by status
	t.Run("FilterByStatus", func(t *testing.T) {
		resp := doGet(t, "/invoices?filter_status=pending")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Invoice
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, inv := range result {
			if inv.Status != "pending" {
				t.Errorf("Invoice does not have status 'pending': %+v", inv)
			}
		}
	})

	// Test filtering by order
	t.Run("FilterByOrder", func(t *testing.T) {
		resp := doGet(t, "/invoices?filter_order_id=2")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Invoice
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		for _, inv := range result {
			if inv.OrderID != 2 {
				t.Errorf("Invoice does not belong to order 2: %+v", inv)
			}
		}
	})

	// Test pagination
	t.Run("Pagination", func(t *testing.T) {
		resp := doGet(t, "/invoices?page=1&page_size=2")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		var result []Invoice
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 invoices, got %d", len(result))
		}
	})
}

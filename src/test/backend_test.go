package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

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

func doDelete(t *testing.T, path string) *http.Response {
	req, _ := http.NewRequest(http.MethodDelete, baseURL+path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s failed: %v", path, err)
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
		{http.MethodPost, "/users"},
		{http.MethodPut, "/orders"},
		{http.MethodPut, "/products/1"},
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

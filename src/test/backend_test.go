package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
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
	p := Product{Name: "ProxyTest", Price: 49.99, InStock: true, Category: "Test"}
	resp := doPost(t, "/products", p)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from POST, got %d", resp.StatusCode)
	}

	var created Product
	json.NewDecoder(resp.Body).Decode(&created)

	resp = doGet(t, "/products/"+strconv.Itoa(created.ID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET created product failed: %d", resp.StatusCode)
	}

	resp = doDelete(t, "/products/"+strconv.Itoa(created.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE product failed: %d", resp.StatusCode)
	}
}

func TestGetUsers(t *testing.T) {
	resp := doGet(t, "/users")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /users failed: %d", resp.StatusCode)
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
	order := Order{
		UserID:    1,
		ProductID: 1,
		Quantity:  2,
	}

	resp := doPost(t, "/orders", order)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /orders failed: %d", resp.StatusCode)
	}

	var created Order
	json.NewDecoder(resp.Body).Decode(&created)
	if created.Total <= 0 || created.CreatedAt.After(time.Now()) {
		t.Errorf("Unexpected order fields: %+v", created)
	}

	resp = doGet(t, "/orders")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /orders failed: %d", resp.StatusCode)
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
		req, _ := http.NewRequest(test.method, baseURL+test.path, strings.NewReader(`{}`))
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
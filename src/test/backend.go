package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"in_stock"`
	Category string  `json:"category"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var products = []Product{
	{ID: 1, Name: "Laptop", Price: 1299.99, InStock: true, Category: "Electronics"},
	{ID: 2, Name: "Sneakers", Price: 89.99, InStock: true, Category: "Footwear"},
}
var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}
var orders = []Order{}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/products", handleProducts)
	mux.HandleFunc("/products/", handleProductByID)
	mux.HandleFunc("/users", handleUsers)
	mux.HandleFunc("/users/", handleUserByID)
	mux.HandleFunc("/orders", handleOrders)

	log.Println("Backend API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

// ==== HANDLERS ====

func handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, products)
	case http.MethodPost:
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		p.ID = rand.Intn(10000)
		products = append(products, p)
		respondJSON(w, p)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		for _, p := range products {
			if p.ID == id {
				respondJSON(w, p)
				return
			}
		}
		http.NotFound(w, r)
	case http.MethodDelete:
		for i, p := range products {
			if p.ID == id {
				products = append(products[:i], products[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, users)
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	for _, u := range users {
		if u.ID == id {
			respondJSON(w, u)
			return
		}
	}
	http.NotFound(w, r)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, orders)
	case http.MethodPost:
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		// Normally you'd validate user/product existence
		o.ID = rand.Intn(10000)
		o.Total = float64(o.Quantity) * findProductPrice(o.ProductID)
		o.CreatedAt = time.Now()
		orders = append(orders, o)
		respondJSON(w, o)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ==== UTILS ====

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
func findProductPrice(productID int) float64 {
	for _, p := range products {
		if p.ID == productID {
			return p.Price
		}
	}
	return 0.0
}

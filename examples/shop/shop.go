package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
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
	// Optional fields
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
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
	// Optional fields
	Password string `json:"password,omitempty"`
	SSN      string `json:"ssn,omitempty"`
}

type Review struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	// Optional fields
	Title        string            `json:"title,omitempty"`
	HelpfulVotes int               `json:"helpful_votes,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Optional fields
	ParentID   int               `json:"parent_id,omitempty"`
	ImageURL   string            `json:"image_url,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Address struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	// Optional fields
	Apartment   string `json:"apartment,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type PaymentMethod struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	CardNumber     string `json:"card_number"`
	ExpiryDate     string `json:"expiry_date"`
	CardholderName string `json:"cardholder_name"`
	// Optional fields
	CVV              string `json:"cvv,omitempty"`
	IsDefault        bool   `json:"is_default,omitempty"`
	BillingAddressID int    `json:"billing_address_id,omitempty"`
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
var reviews = []Review{}
var categories = []Category{
	{ID: 1, Name: "Electronics", Description: "Electronic devices and accessories"},
	{ID: 2, Name: "Clothing", Description: "Apparel and fashion items"},
}
var addresses = []Address{}
var paymentMethods = []PaymentMethod{}

// QueryParams represents common URL parameters
type QueryParams struct {
	Page     int
	PageSize int
	SortBy   string
	Order    string
	Search   string
	Filter   map[string]string
}

// parseQueryParams extracts common URL parameters from the request
func parseQueryParams(r *http.Request) QueryParams {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := make(map[string]string)
	for k, v := range q {
		if strings.HasPrefix(k, "filter_") {
			filter[strings.TrimPrefix(k, "filter_")] = v[0]
		}
	}

	return QueryParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   q.Get("sort_by"),
		Order:    q.Get("order"),
		Search:   q.Get("search"),
		Filter:   filter,
	}
}

// paginateResults returns a paginated slice of results
func paginateResults[T any](items []T, page, pageSize int) []T {
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(items) {
		return []T{}
	}
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// filterProducts applies filters to products
func filterProducts(products []Product, filter map[string]string) []Product {
	if len(filter) == 0 {
		return products
	}

	var filtered []Product
	for _, p := range products {
		match := true
		for k, v := range filter {
			switch k {
			case "category":
				if p.Category != v {
					match = false
				}
			case "in_stock":
				if strconv.FormatBool(p.InStock) != v {
					match = false
				}
			case "min_price":
				minPrice, _ := strconv.ParseFloat(v, 64)
				if p.Price < minPrice {
					match = false
				}
			case "max_price":
				maxPrice, _ := strconv.ParseFloat(v, 64)
				if p.Price > maxPrice {
					match = false
				}
			}
		}
		if match {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// searchProducts searches products by name and description
func searchProducts(products []Product, query string) []Product {
	if query == "" {
		return products
	}
	query = strings.ToLower(query)
	var results []Product
	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) {
			results = append(results, p)
		}
	}
	return results
}

// sortProducts sorts products by the specified field
func sortProducts(products []Product, sortBy, order string) []Product {
	if sortBy == "" {
		return products
	}

	sort.Slice(products, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = products[i].Name < products[j].Name
		case "price":
			less = products[i].Price < products[j].Price
		case "category":
			less = products[i].Category < products[j].Category
		default:
			return false
		}
		if order == "desc" {
			return !less
		}
		return less
	})
	return products
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	// Add health check endpoint
	http.HandleFunc("/health", handleHealth)

	// Add other endpoints
	http.HandleFunc("/products", handleProducts)
	http.HandleFunc("/products/", handleProductByID)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUserByID)
	http.HandleFunc("/orders", handleOrders)
	http.HandleFunc("/reviews", handleReviews)
	http.HandleFunc("/reviews/", handleReviewByID)
	http.HandleFunc("/categories", handleCategories)
	http.HandleFunc("/categories/", handleCategoryByID)
	http.HandleFunc("/addresses", handleAddresses)
	http.HandleFunc("/addresses/", handleAddressByID)
	http.HandleFunc("/payment-methods", handlePaymentMethods)
	http.HandleFunc("/payment-methods/", handlePaymentMethodByID)

	log.Println("Backend API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ==== HANDLERS ====

func handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		params := parseQueryParams(r)

		// Apply filters
		filtered := filterProducts(products, params.Filter)

		// Apply search
		searched := searchProducts(filtered, params.Search)

		// Apply sorting
		sorted := sortProducts(searched, params.SortBy, params.Order)

		// Apply pagination
		paginated := paginateResults(sorted, params.Page, params.PageSize)

		respondJSON(w, paginated)
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

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid product ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid product ID format", http.StatusBadRequest)
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
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, users)
	case http.MethodPost:
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		user.ID = rand.Intn(10000)
		users = append(users, user)
		respondJSON(w, user)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	// Check method first
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/users/")

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
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

func handleReviews(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		params := parseQueryParams(r)

		// Filter reviews
		var filtered []Review
		for _, review := range reviews {
			match := true
			for k, v := range params.Filter {
				switch k {
				case "user_id":
					userID, _ := strconv.Atoi(v)
					if review.UserID != userID {
						match = false
					}
				case "product_id":
					productID, _ := strconv.Atoi(v)
					if review.ProductID != productID {
						match = false
					}
				case "rating":
					rating, _ := strconv.Atoi(v)
					if review.Rating != rating {
						match = false
					}
				}
			}
			if match {
				filtered = append(filtered, review)
			}
		}

		// Apply pagination
		paginated := paginateResults(filtered, params.Page, params.PageSize)

		respondJSON(w, paginated)
	case http.MethodPost:
		var review Review
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		review.ID = rand.Intn(10000)
		review.CreatedAt = time.Now()
		reviews = append(reviews, review)
		respondJSON(w, review)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleReviewByID(w http.ResponseWriter, r *http.Request) {
	// Check method first
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/reviews/")

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid review ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid review ID format", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		for _, r := range reviews {
			if r.ID == id {
				respondJSON(w, r)
				return
			}
		}
		http.NotFound(w, r)
	case http.MethodDelete:
		for i, r := range reviews {
			if r.ID == id {
				reviews = append(reviews[:i], reviews[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}
}

func handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, categories)
	case http.MethodPost:
		var category Category
		if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		category.ID = rand.Intn(10000)
		categories = append(categories, category)
		respondJSON(w, category)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	// Check method first
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid category ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid category ID format", http.StatusBadRequest)
		return
	}

	for _, c := range categories {
		if c.ID == id {
			respondJSON(w, c)
			return
		}
	}
	http.NotFound(w, r)
}

func handleAddresses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, addresses)
	case http.MethodPost:
		var address Address
		if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		address.ID = rand.Intn(10000)
		addresses = append(addresses, address)
		respondJSON(w, address)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAddressByID(w http.ResponseWriter, r *http.Request) {
	// Check method first
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/addresses/")

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid address ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid address ID format", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		for _, a := range addresses {
			if a.ID == id {
				respondJSON(w, a)
				return
			}
		}
		http.NotFound(w, r)
	case http.MethodDelete:
		for i, a := range addresses {
			if a.ID == id {
				addresses = append(addresses[:i], addresses[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}
}

func handlePaymentMethods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, paymentMethods)
	case http.MethodPost:
		var payment PaymentMethod
		if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		payment.ID = rand.Intn(10000)
		paymentMethods = append(paymentMethods, payment)
		respondJSON(w, payment)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePaymentMethodByID(w http.ResponseWriter, r *http.Request) {
	// Check method first
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/payment-methods/")

	// Validate ID format
	if idStr == "" || !isValidID(idStr) {
		http.Error(w, "invalid payment method ID format", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid payment method ID format", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		for _, p := range paymentMethods {
			if p.ID == id {
				respondJSON(w, p)
				return
			}
		}
		http.NotFound(w, r)
	case http.MethodDelete:
		for i, p := range paymentMethods {
			if p.ID == id {
				paymentMethods = append(paymentMethods[:i], paymentMethods[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}
}

// isValidID checks if a string is a valid positive integer
func isValidID(idStr string) bool {
	// Check if empty
	if idStr == "" {
		return false
	}

	// Check if it's a valid positive integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return false
	}

	// Check if positive
	return id > 0
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

package analyzer

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server represents the analyzer HTTP server
type Server struct {
	analyzer *Analyzer
}

// NewServer creates a new analyzer server
func NewServer(analyzer *Analyzer) *Server {
	return &Server{
		analyzer: analyzer,
	}
}

// Start starts the analyzer server
func (s *Server) Start(addr string) error {
	// API endpoints
	http.HandleFunc("/api/health", s.handleHealth)
	http.HandleFunc("/api/analyzer", s.handleAnalyzer)
	http.HandleFunc("/api/openapi.json", s.handleOpenAPI)
	http.HandleFunc("/api/postman.json", s.handlePostman)
	http.HandleFunc("/swagger", s.handleSwaggerUI)

	// Handle OPTIONS requests for CORS
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	// Serve static UI files
	fs := http.FileServer(http.Dir("internal/analyzer/ui"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for an API endpoint, return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// For all other requests, serve the UI
		// If the path doesn't exist, serve index.html for client-side routing
		path := filepath.Join("internal/analyzer/ui", r.URL.Path)
		if _, err := os.Stat(path); err != nil {
			http.ServeFile(w, r, "internal/analyzer/ui/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	log.Printf("Analyzer server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleAnalyzer handles requests to the analyzer endpoint
func (s *Server) handleAnalyzer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	data := s.analyzer.GetData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleOpenAPI handles requests to the OpenAPI endpoint
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	openAPI := s.analyzer.GenerateOpenAPI()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openAPI)
}

// handlePostman handles requests to the Postman collection endpoint
func (s *Server) handlePostman(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	collection := s.analyzer.GeneratePostmanCollection()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=api-collection.json")
	json.NewEncoder(w).Encode(collection)
}

// handleHealth handles requests to the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

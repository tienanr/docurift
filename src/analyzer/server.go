package analyzer

import (
	"encoding/json"
	"log"
	"net/http"
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
	http.HandleFunc("/analyzer", s.handleAnalyzer)
	http.HandleFunc("/openapi.json", s.handleOpenAPI)
	http.HandleFunc("/postman.json", s.handlePostman)
	http.HandleFunc("/", s.handleSwaggerUI)
	log.Printf("Analyzer server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleAnalyzer handles requests to the analyzer endpoint
func (s *Server) handleAnalyzer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	collection := s.analyzer.GeneratePostmanCollection()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=api-collection.json")
	json.NewEncoder(w).Encode(collection)
}

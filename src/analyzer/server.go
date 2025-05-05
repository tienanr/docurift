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

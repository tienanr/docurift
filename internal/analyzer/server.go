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
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/analyzer", s.handleAnalyzer)
	http.HandleFunc("/openapi.json", s.handleOpenAPI)
	http.HandleFunc("/postman.json", s.handlePostman)
	http.HandleFunc("/tools.json", s.handleTools)
	http.HandleFunc("/tools/openai.json", s.handleOpenAITools)
	http.HandleFunc("/tools/anthropic.json", s.handleAnthropicTools)
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

// handleHealth handles requests to the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleTools handles requests to the tools specification endpoint
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.analyzer.GenerateToolsSpec()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

// handleOpenAITools handles requests to the OpenAI tools specification endpoint
func (s *Server) handleOpenAITools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.analyzer.GenerateOpenAIToolsSpec()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

// handleAnthropicTools handles requests to the Anthropic tools specification endpoint
func (s *Server) handleAnthropicTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.analyzer.GenerateAnthropicToolsSpec()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

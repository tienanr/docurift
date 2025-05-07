package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/tienanr/docurift/analyzer"
	"github.com/tienanr/docurift/config"
	"github.com/vulcand/oxy/forward"
)

// customResponseWriter captures the response for logging
type customResponseWriter struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (w *customResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	w.buf.Write(b) // Capture response
	return w.ResponseWriter.Write(b)
}

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize analyzer with max examples from config
	analyzerInstance := analyzer.NewAnalyzer()
	analyzerInstance.SetMaxExamples(cfg.Analyzer.MaxExamples)
	analyzerServer := analyzer.NewServer(analyzerInstance)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Analyzer.Port)
		if err := analyzerServer.Start(addr); err != nil {
			log.Fatalf("Failed to start analyzer server: %v", err)
		}
	}()

	// Parse backend URL
	backendURL, err := url.Parse(cfg.Proxy.BackendURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}

	fwd, err := forward.New(forward.PassHostHeader(true))
	if err != nil {
		log.Fatalf("Failed to create forwarder: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Capture request body
		var reqBody []byte
		if req.Body != nil {
			reqBody, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		req.URL.Scheme = backendURL.Scheme
		req.URL.Host = backendURL.Host

		log.Printf("→ Forwarding request: %s %s", req.Method, req.URL.String())

		crw := &customResponseWriter{ResponseWriter: w, statusCode: 200}
		fwd.ServeHTTP(crw, req)

		// Log response after it's been written
		log.Printf("← Response status: %d\n← Body: %s", crw.statusCode, crw.buf.String())

		// Process request/response with analyzer
		analyzerInstance.ProcessRequest(
			req.Method,
			req.URL.String(),
			req,
			&http.Response{
				StatusCode: crw.statusCode,
				Header:     crw.Header(),
			},
			reqBody,
			crw.buf.Bytes(),
		)
	})

	addr := fmt.Sprintf(":%d", cfg.Proxy.Port)
	log.Printf("Proxy listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/tienanr/docurift/internal/analyzer"
	"github.com/tienanr/docurift/internal/config"
	"github.com/vulcand/oxy/forward"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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

func printUsage() {
	fmt.Printf("DocuRift - Automatic API Documentation Generator\n\n")
	fmt.Printf("Usage: docurift -config <config-file>\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -config string    Path to configuration file (required)\n")
	fmt.Printf("  -version         Show version information\n")
	fmt.Printf("\nExample:\n")
	fmt.Printf("  docurift -config config.yaml\n")
}

// checkPortAvailable checks if a port is available for use
func checkPortAvailable(port int, service string) error {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("%s port %d is already in use: %w", service, port, err)
	}
	ln.Close()
	return nil
}

func main() {
	// Define command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")

	// Parse flags
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("DocuRift version %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Show usage if no arguments provided
	if len(os.Args) == 1 {
		printUsage()
		return
	}

	// Load configuration
	if *configPath == "" {
		log.Fatal("Configuration file path is required")
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if ports are available
	if err := checkPortAvailable(cfg.Proxy.Port, "proxy"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	if err := checkPortAvailable(cfg.Analyzer.Port, "analyzer"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Starting DocuRift with proxy port %d and analyzer port %d", cfg.Proxy.Port, cfg.Analyzer.Port)

	// Initialize analyzer with configuration
	analyzerInstance := analyzer.NewAnalyzer(cfg.Analyzer.Storage.Path, cfg.Analyzer.Storage.Frequency)
	analyzerInstance.SetMaxExamples(cfg.Analyzer.MaxExamples)
	analyzerInstance.SetRedactedFields(cfg.Analyzer.RedactedFields)
	analyzerInstance.SetProxyConfig(cfg.Proxy.Port, cfg.Proxy.BackendURL)
	analyzerInstance.SetAnalyzerPort(cfg.Analyzer.Port)
	analyzerServer := analyzer.NewServer(analyzerInstance)

	// Start analyzer server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Analyzer.Port)
		log.Printf("Starting analyzer server on %s", addr)
		if err := analyzerServer.Start(addr); err != nil {
			log.Fatalf("Failed to start analyzer server: %v", err)
		}
	}()

	// Parse backend URL
	backendURLParsed, err := url.Parse(cfg.Proxy.BackendURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}

	log.Printf("Using backend URL: %s", backendURLParsed.String())

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

		req.URL.Scheme = backendURLParsed.Scheme
		req.URL.Host = backendURLParsed.Host

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
	log.Printf("Starting proxy server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}

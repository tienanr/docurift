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
	"github.com/vulcand/oxy/forward"
)

const (
	// Port range constants
	MinPort = 1024
	MaxPort = 65535
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
	fmt.Printf("Usage: docurift [options]\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -proxy-port int        Proxy server port (default 9876)\n")
	fmt.Printf("  -analyzer-port int     Analyzer server port (default 9877)\n")
	fmt.Printf("  -backend-url string    Backend API URL (default http://localhost:8080)\n")
	fmt.Printf("  -max-examples int      Maximum number of examples per endpoint (default 10)\n")
	fmt.Printf("  -version              Show version information\n")
	fmt.Printf("\nExample:\n")
	fmt.Printf("  docurift -proxy-port 9876 -analyzer-port 9877 -backend-url http://localhost:8080 -max-examples 20\n")
}

// validatePort checks if a port is within the valid range
func validatePort(port int, service string) error {
	if port < MinPort {
		return fmt.Errorf("%s port %d is below minimum allowed port (%d)", service, port, MinPort)
	}
	if port > MaxPort {
		return fmt.Errorf("%s port %d is above maximum allowed port (%d)", service, port, MaxPort)
	}
	return nil
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
	proxyPort := flag.Int("proxy-port", 9876, "Proxy server port")
	analyzerPort := flag.Int("analyzer-port", 9877, "Analyzer server port")
	backendURL := flag.String("backend-url", "http://localhost:8080", "Backend API URL")
	maxExamples := flag.Int("max-examples", 10, "Maximum number of examples per endpoint")
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

	// Validate ports
	if err := validatePort(*proxyPort, "proxy"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	if err := validatePort(*analyzerPort, "analyzer"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Check for port conflicts
	if *proxyPort == *analyzerPort {
		log.Fatalf("Invalid configuration: proxy port (%d) cannot be the same as analyzer port", *proxyPort)
	}

	// Check if ports are available
	if err := checkPortAvailable(*proxyPort, "proxy"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	if err := checkPortAvailable(*analyzerPort, "analyzer"); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Validate other parameters
	if *backendURL == "" {
		log.Fatal("Invalid configuration: backend URL is required")
	}
	if *maxExamples <= 0 {
		log.Fatal("Invalid configuration: max examples must be positive")
	}

	log.Printf("Starting DocuRift with proxy port %d and analyzer port %d", *proxyPort, *analyzerPort)

	// Initialize analyzer with max examples
	analyzerInstance := analyzer.NewAnalyzer()
	analyzerInstance.SetMaxExamples(*maxExamples)
	analyzerServer := analyzer.NewServer(analyzerInstance)

	// Start analyzer server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", *analyzerPort)
		log.Printf("Starting analyzer server on %s", addr)
		if err := analyzerServer.Start(addr); err != nil {
			log.Fatalf("Failed to start analyzer server: %v", err)
		}
	}()

	// Parse backend URL
	backendURLParsed, err := url.Parse(*backendURL)
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

	addr := fmt.Sprintf(":%d", *proxyPort)
	log.Printf("Starting proxy server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}

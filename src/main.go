package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/tienanr/docurift/analyzer"
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
	// Initialize analyzer
	analyzerInstance := analyzer.NewAnalyzer()
	analyzerServer := analyzer.NewServer(analyzerInstance)
	go func() {
		if err := analyzerServer.Start(":8082"); err != nil {
			log.Fatalf("Failed to start analyzer server: %v", err)
		}
	}()

	backendURL, err := url.Parse("http://localhost:8081")
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

	log.Println("Proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

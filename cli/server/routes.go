package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// buildRouter creates the HTTP mux with all routes and middleware.
func (s *Server) buildRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/validate/", s.handleValidatePrefix)
	mux.HandleFunc("/schemas", s.handleSchemas)
	mux.HandleFunc("/health", s.handleHealth)

	var handler http.Handler = mux

	// Apply middleware in reverse order (outermost first)
	handler = recoveryMiddleware(handler)
	handler = loggingMiddleware(handler)
	if s.config.CORS {
		handler = corsMiddleware(handler)
	}

	return handler
}

// handleValidatePrefix dispatches /validate/:name requests by extracting the name
// from the URL path.
func (s *Server) handleValidatePrefix(w http.ResponseWriter, r *http.Request) {
	// Extract schema name from /validate/<name>
	path := strings.TrimPrefix(r.URL.Path, "/validate/")
	if path == "" || strings.Contains(path, "/") {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	s.handleValidateNamed(w, r, path)
}

// loggingMiddleware logs each request to stdout.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		fmt.Printf("%s %s %d %s\n", r.Method, r.URL.Path, rw.status, duration.Round(time.Millisecond))
	})
}

// recoveryMiddleware recovers from panics and returns a 500 error.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("PANIC: %v\n", err)
				writeJSONError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers to responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

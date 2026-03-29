package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	anyvali "github.com/BetterCorp/AnyVali/sdk/go"
)

// Version is the server/CLI version. Set by the main package or cmd.
var Version = "0.0.1"

// Config holds server configuration.
type Config struct {
	Host       string
	Port       int
	SchemasDir string
	CORS       bool
}

// Server is the AnyVali HTTP validation server.
type Server struct {
	config  Config
	schemas map[string]anyvali.Schema
	httpSrv *http.Server
}

// New creates a new server with the given config.
func New(cfg Config) (*Server, error) {
	s := &Server{
		config:  cfg,
		schemas: make(map[string]anyvali.Schema),
	}

	if cfg.SchemasDir != "" {
		if err := s.loadSchemas(cfg.SchemasDir); err != nil {
			return nil, fmt.Errorf("failed to load schemas: %w", err)
		}
	}

	return s, nil
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	mux := s.buildRouter()

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Printf("AnyVali server listening on %s\n", addr)
	if s.config.SchemasDir != "" {
		fmt.Printf("Loaded %d schemas from %s\n", len(s.schemas), s.config.SchemasDir)
	}

	err := s.httpSrv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Handler returns the HTTP handler for use in tests.
func (s *Server) Handler() http.Handler {
	return s.buildRouter()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpSrv.Shutdown(ctx)
	}
}

// GetSchemas returns the loaded schema names.
func (s *Server) GetSchemas() []string {
	names := make([]string, 0, len(s.schemas))
	for name := range s.schemas {
		names = append(names, name)
	}
	return names
}

func (s *Server) loadSchemas(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read schemas directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read schema file %s: %w", entry.Name(), err)
		}

		schema, err := anyvali.ImportJSON(data)
		if err != nil {
			return fmt.Errorf("invalid schema in %s: %w", entry.Name(), err)
		}

		name := strings.TrimSuffix(entry.Name(), ".json")
		s.schemas[name] = schema
	}

	return nil
}

// handleValidate handles POST /validate requests.
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !isJSONContentType(r) {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var body struct {
		Schema    json.RawMessage `json:"schema"`
		SchemaRef string          `json:"schemaRef"`
		Input     json.RawMessage `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if body.Input == nil {
		writeJSONError(w, http.StatusBadRequest, "missing 'input' field")
		return
	}

	var schema anyvali.Schema
	var err error

	if body.SchemaRef != "" {
		var ok bool
		schema, ok = s.schemas[body.SchemaRef]
		if !ok {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown schema ref: %q", body.SchemaRef))
			return
		}
	} else if body.Schema != nil {
		schema, err = anyvali.ImportJSON(body.Schema)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid schema: "+err.Error())
			return
		}
	} else {
		writeJSONError(w, http.StatusBadRequest, "must provide 'schema' or 'schemaRef'")
		return
	}

	var input any
	if err := json.Unmarshal(body.Input, &input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid input JSON: "+err.Error())
		return
	}

	result := schema.SafeParse(input)
	writeValidationResult(w, result)
}

// handleValidateNamed handles POST /validate/:name requests.
func (s *Server) handleValidateNamed(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !isJSONContentType(r) {
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	schema, ok := s.schemas[name]
	if !ok {
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("schema %q not found", name))
		return
	}

	var input any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	result := schema.SafeParse(input)
	writeValidationResult(w, result)
}

// handleSchemas handles GET /schemas requests.
func (s *Server) handleSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	names := s.GetSchemas()
	writeJSON(w, http.StatusOK, map[string]any{
		"schemas": names,
	})
}

// handleHealth handles GET /health requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": Version,
	})
}

func writeValidationResult(w http.ResponseWriter, result anyvali.ParseResult) {
	resp := map[string]any{
		"valid": result.Success,
	}

	if result.Success {
		resp["data"] = result.Data
	} else {
		issues := make([]map[string]any, len(result.Issues))
		for i, issue := range result.Issues {
			issueMap := map[string]any{
				"code":    issue.Code,
				"message": issue.Message,
			}
			if len(issue.Path) > 0 {
				issueMap["path"] = issue.Path
			}
			if issue.Expected != "" {
				issueMap["expected"] = issue.Expected
			}
			if issue.Received != "" {
				issueMap["received"] = issue.Received
			}
			issues[i] = issueMap
		}
		resp["issues"] = issues
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}

func isJSONContentType(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	// Accept "application/json" or "application/json; charset=utf-8" etc.
	return strings.HasPrefix(strings.ToLower(ct), "application/json")
}

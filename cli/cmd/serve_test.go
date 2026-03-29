package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/BetterCorp/AnyVali/cli/server"
)

func setupTestServer(t *testing.T, cors bool) (*httptest.Server, *server.Server) {
	t.Helper()
	dir := t.TempDir()

	// Write a test schema file
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"name": map[string]any{"kind": "string", "minLength": 1},
			},
			"required":    []any{"name"},
			"unknownKeys": "reject",
		},
	}
	data, _ := json.Marshal(schema)
	os.WriteFile(filepath.Join(dir, "user.json"), data, 0644)

	cfg := server.Config{
		Host:       "127.0.0.1",
		Port:       0,
		SchemasDir: dir,
		CORS:       cors,
	}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	return ts, srv
}

func TestHealthEndpoint(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
	if body["version"] != server.Version {
		t.Errorf("expected version=%s, got %v", server.Version, body["version"])
	}
}

func TestSchemasEndpoint(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/schemas")
	if err != nil {
		t.Fatalf("GET /schemas: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	schemas, ok := body["schemas"].([]any)
	if !ok {
		t.Fatalf("expected schemas array, got %T", body["schemas"])
	}

	found := false
	for _, s := range schemas {
		if s == "user" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'user' in schemas list, got %v", schemas)
	}
}

func TestValidateInlineSchema(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	reqBody := map[string]any{
		"schema": map[string]any{
			"anyvaliVersion": "1.0",
			"schemaVersion":  "1",
			"root": map[string]any{
				"kind": "string",
			},
		},
		"input": "hello",
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["valid"] != true {
		t.Errorf("expected valid=true, got %v", body["valid"])
	}
}

func TestValidateSchemaRef(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	reqBody := map[string]any{
		"schemaRef": "user",
		"input": map[string]any{
			"name": "Alice",
		},
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["valid"] != true {
		t.Errorf("expected valid=true, got %v", body["valid"])
	}
}

func TestValidateNamedEndpoint(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	input := map[string]any{"name": "Bob"}
	data, _ := json.Marshal(input)

	resp, err := http.Post(ts.URL+"/validate/user", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /validate/user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["valid"] != true {
		t.Errorf("expected valid=true, got %v", body["valid"])
	}
}

func TestValidateInvalidContentType(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/validate", "text/plain", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", resp.StatusCode)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestValidateUnknownSchemaRef(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	reqBody := map[string]any{
		"schemaRef": "nonexistent",
		"input":     "test",
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestValidateNamedNotFound(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/validate/nonexistent", "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("POST /validate/nonexistent: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 404, got %d, body: %s", resp.StatusCode, string(body))
	}
}

func TestCORSHeaders(t *testing.T) {
	ts, _ := setupTestServer(t, true)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected CORS origin=*, got %q", origin)
	}
}

func TestCORSPreflight(t *testing.T) {
	ts, _ := setupTestServer(t, true)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/validate", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS /validate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", resp.StatusCode)
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestNoCORSByDefault(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Errorf("expected no CORS header, got %q", origin)
	}
}

func TestServerValidateInvalidData(t *testing.T) {
	ts, _ := setupTestServer(t, false)
	defer ts.Close()

	reqBody := map[string]any{
		"schemaRef": "user",
		"input":     map[string]any{},
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(ts.URL+"/validate", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /validate: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["valid"] != false {
		t.Errorf("expected valid=false, got %v", body["valid"])
	}

	issues, ok := body["issues"].([]any)
	if !ok || len(issues) == 0 {
		t.Error("expected at least one issue")
	}
}

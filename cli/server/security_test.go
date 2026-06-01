package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestServer creates a Server with optional test schemas loaded from a temp dir.
func newTestServer(t *testing.T, cors bool, withSchemas bool) *Server {
	t.Helper()
	dir := ""
	if withSchemas {
		dir = t.TempDir()
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
	}

	cfg := Config{Host: "localhost", Port: 0, SchemasDir: dir, CORS: cors}
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	return srv
}

// decodeJSONResponse reads and decodes a JSON response body from a recorder.
func decodeJSONResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v\nbody: %s", err, rr.Body.String())
	}
	return body
}

// ---------------------------------------------------------------------------
// CVE-2023-44487 / CWE-400: Request body size limits
// ---------------------------------------------------------------------------
// The server currently uses json.NewDecoder(r.Body).Decode() without wrapping
// r.Body in an io.LimitReader. These tests document that behaviour and verify
// the server at least remains functional (does not crash) when receiving large
// payloads.

func TestRequestBodySizeNoLimit(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	// Build a ~1 MB JSON payload. The server has no configured limit, so it
	// must still parse or reject this gracefully.
	bigValue := strings.Repeat("A", 1<<20) // 1 MiB string
	body := map[string]any{
		"schema": map[string]any{
			"anyvaliVersion": "1.0",
			"schemaVersion":  "1",
			"root":           map[string]any{"kind": "string"},
		},
		"input": bigValue,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// The server should not crash; it should return a parseable JSON response.
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for large body, got %d", rr.Code)
	}
	resp := decodeJSONResponse(t, rr)
	if resp["valid"] != true {
		t.Errorf("expected valid=true for a string input, got %v", resp["valid"])
	}
}

func TestRequestBodySizeDocumentNoLimitReader(t *testing.T) {
	// SECURITY FINDING: The server does not use io.LimitReader to cap the
	// request body size. An attacker could send an extremely large payload
	// to exhaust server memory (CWE-400 / CVE-2023-44487).
	//
	// Recommended fix: wrap r.Body with io.LimitReader(r.Body, maxBytes)
	// where maxBytes is a sensible upper bound (e.g. 1 MB), and return 413
	// Request Entity Too Large if exceeded.
	//
	// This test intentionally passes to document the gap.
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	// 10 MB payload - server should ideally reject this with 413.
	bigValue := strings.Repeat("X", 10<<20)
	body := fmt.Sprintf(`{"schema":{"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"string"}},"input":"%s"}`, bigValue)

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Currently succeeds (no limit). When a limit is added, change this to
	// expect http.StatusRequestEntityTooLarge (413).
	if rr.Code != http.StatusOK {
		t.Logf("server returned %d for 10MB body (limit may have been added)", rr.Code)
	}
}

// REVIEW: The test above documents the missing limit but still passes when the
// server accepts an oversized body. This companion test is the enforceable
// security expectation after request-body limiting is implemented.
func TestRequestBodySizeLimitEnforced(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	bigValue := strings.Repeat("X", 10<<20)
	body := fmt.Sprintf(`{"schema":{"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"string"}},"input":"%s"}`, bigValue)

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected oversized request body to be rejected with 413, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// CWE-400: Slowloris / timeout configuration
// ---------------------------------------------------------------------------

func TestServerTimeoutsNotConfigured(t *testing.T) {
	// SECURITY FINDING: The http.Server in Start() does not set ReadTimeout,
	// WriteTimeout, or IdleTimeout. This leaves the server vulnerable to
	// slowloris-style denial-of-service attacks (CWE-400).
	//
	// Recommended fix:
	//   s.httpSrv = &http.Server{
	//       Addr:         addr,
	//       Handler:      mux,
	//       ReadTimeout:  10 * time.Second,
	//       WriteTimeout: 30 * time.Second,
	//       IdleTimeout:  120 * time.Second,
	//   }

	cfg := Config{Host: "localhost", Port: 0, SchemasDir: "", CORS: false}
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	// The httpSrv field is only created in Start(), which actually listens.
	// We verify the gap by checking the field is nil before Start.
	if srv.httpSrv != nil {
		t.Error("expected httpSrv to be nil before Start()")
	}

	// Document: when Start() creates the http.Server, it sets no timeouts.
	// This test passes to document the finding. When timeouts are added,
	// add assertions like:
	//   if srv.httpSrv.ReadTimeout == 0 { t.Error("ReadTimeout not set") }
	t.Log("SECURITY: http.Server timeouts (ReadTimeout, WriteTimeout, IdleTimeout) are not configured")
}

// REVIEW: The test above only logs the missing timeout configuration. This
// companion test asserts the production http.Server has non-zero timeouts.
func TestServerTimeoutsAreConfigured(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: -1, SchemasDir: "", CORS: false}
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	_ = srv.Start()

	if srv.httpSrv == nil {
		t.Fatal("expected Start to initialize httpSrv")
	}
	if srv.httpSrv.ReadTimeout == 0 {
		t.Error("ReadTimeout must be configured")
	}
	if srv.httpSrv.WriteTimeout == 0 {
		t.Error("WriteTimeout must be configured")
	}
	if srv.httpSrv.IdleTimeout == 0 {
		t.Error("IdleTimeout must be configured")
	}
}

// ---------------------------------------------------------------------------
// CWE-22: Path traversal in /validate/:name
// ---------------------------------------------------------------------------

func TestPathTraversal(t *testing.T) {
	srv := newTestServer(t, false, true)
	handler := srv.Handler()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name: "parent directory traversal",
			path: "/validate/..%2Fsecret",
			// Go's default mux cleans the path and issues a 301 redirect,
			// which prevents the request from reaching the handler. This is
			// a safe outcome.
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "slash in schema name",
			path:       "/validate/foo/bar",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "double dot without encoding",
			path: "/validate/../secret",
			// Go's default mux cleans ".." path segments and redirects.
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "double dot with double encoding",
			path:       "/validate/..%252Fsecret",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "backslash traversal",
			path:       "/validate/..\\secret",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "null byte injection",
			path:       "/validate/user%00.json",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "deeply nested traversal",
			path:       "/validate/a/b/c/d/e",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "valid schema name",
			path:       "/validate/user",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"name": "test"}`
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("path %q: expected status %d, got %d (body: %s)",
					tt.path, tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestPathTraversalResponseBody(t *testing.T) {
	srv := newTestServer(t, false, true)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/validate/foo/bar", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	resp := decodeJSONResponse(t, rr)
	errMsg, ok := resp["error"].(string)
	if !ok {
		t.Fatalf("expected error string in response, got %v", resp)
	}

	// The error message should NOT leak internal path info.
	if strings.Contains(errMsg, "foo/bar") || strings.Contains(errMsg, "..") {
		t.Logf("error message may leak path info: %q", errMsg)
	}
}

// ---------------------------------------------------------------------------
// CWE-20: Malformed JSON
// ---------------------------------------------------------------------------

func TestMalformedJSON(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	tests := []struct {
		name string
		body string
	}{
		{"truncated object", `{"schema": {`},
		{"trailing comma", `{"schema": {"kind": "string",}, "input": "x"}`},
		{"raw text", `hello world`},
		{"empty body", ``},
		{"only whitespace", `   `},
		{"xml instead of json", `<root><kind>string</kind></root>`},
		{"binary garbage", "\x00\x01\x02\x03\x04\x05"},
		{"array instead of object", `[1, 2, 3]`},
		{"deeply nested braces", strings.Repeat("{", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %q, got %d (body: %s)",
					tt.name, rr.Code, rr.Body.String())
			}

			// Response should be valid JSON even for malformed input.
			resp := decodeJSONResponse(t, rr)
			if _, hasError := resp["error"]; !hasError {
				t.Errorf("expected 'error' field in response for %q", tt.name)
			}
		})
	}
}

func TestMalformedJSONNamedEndpoint(t *testing.T) {
	srv := newTestServer(t, false, true)
	handler := srv.Handler()

	tests := []struct {
		name string
		body string
	}{
		{"truncated json", `{"name": `},
		{"not json at all", `not json`},
		{"empty body", ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validate/user", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %q, got %d", tt.name, rr.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CWE-20: Content-Type validation
// ---------------------------------------------------------------------------

func TestContentTypeValidation(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	tests := []struct {
		name        string
		contentType string
		wantStatus  int
	}{
		{"text/plain", "text/plain", http.StatusUnsupportedMediaType},
		{"text/html", "text/html", http.StatusUnsupportedMediaType},
		{"application/xml", "application/xml", http.StatusUnsupportedMediaType},
		{"multipart/form-data", "multipart/form-data", http.StatusUnsupportedMediaType},
		{"application/x-www-form-urlencoded", "application/x-www-form-urlencoded", http.StatusUnsupportedMediaType},
		{"empty content type", "", http.StatusUnsupportedMediaType},
		{"application/json", "application/json", http.StatusBadRequest}, // 400 because body has missing fields
		{"application/json with charset", "application/json; charset=utf-8", http.StatusBadRequest},
		{"APPLICATION/JSON uppercase", "APPLICATION/JSON", http.StatusBadRequest},
	}

	body := `{"input": "hello"}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Content-Type %q: expected status %d, got %d (body: %s)",
					tt.contentType, tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestContentTypeValidationNamedEndpoint(t *testing.T) {
	srv := newTestServer(t, false, true)
	handler := srv.Handler()

	tests := []struct {
		name        string
		contentType string
		wantStatus  int
	}{
		{"text/plain rejected", "text/plain", http.StatusUnsupportedMediaType},
		{"application/json accepted", "application/json", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"name": "Alice"}`
			req := httptest.NewRequest(http.MethodPost, "/validate/user", strings.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Content-Type %q: expected %d, got %d", tt.contentType, tt.wantStatus, rr.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CWE-20: Schema injection via request body
// ---------------------------------------------------------------------------

func TestSchemaInjectionUnknownKind(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	tests := []struct {
		name   string
		schema map[string]any
	}{
		{
			name: "unknown schema kind",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
				"root": map[string]any{
					"kind": "__proto__",
				},
			},
		},
		{
			name: "constructor pollution kind",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
				"root": map[string]any{
					"kind": "constructor",
				},
			},
		},
		{
			name: "empty kind",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
				"root": map[string]any{
					"kind": "",
				},
			},
		},
		{
			name: "numeric kind",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
				"root": map[string]any{
					"kind": 12345,
				},
			},
		},
		{
			name: "deeply nested malicious schema",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
				"root": map[string]any{
					"kind": "object",
					"properties": map[string]any{
						"evil": map[string]any{
							"kind": "exec_command",
						},
					},
				},
			},
		},
		{
			name: "missing root entirely",
			schema: map[string]any{
				"anyvaliVersion": "1.0",
				"schemaVersion":  "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]any{
				"schema": tt.schema,
				"input":  "anything",
			}
			data, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			// The server must reject unknown/malicious schema kinds with 400,
			// not panic or succeed.
			if rr.Code != http.StatusBadRequest {
				t.Errorf("schema kind %q: expected 400, got %d (body: %s)",
					tt.name, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestSchemaInjectionRecursiveRef(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	// Attempt a self-referencing schema to trigger infinite recursion.
	body := map[string]any{
		"schema": map[string]any{
			"anyvaliVersion": "1.0",
			"schemaVersion":  "1",
			"definitions": map[string]any{
				"self": map[string]any{
					"kind": "ref",
					"ref":  "self",
				},
			},
			"root": map[string]any{
				"kind": "ref",
				"ref":  "self",
			},
		},
		"input": "anything",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// The server should not hang or panic. It should return an error (400)
	// or at worst a 500 caught by recovery middleware.
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("recursive ref: unexpected status %d", rr.Code)
	}
}

// REVIEW: The test above allows 200 OK, so a recursive-ref payload can be
// treated as a successful request. This companion test forbids success.
func TestSchemaInjectionRecursiveRefDoesNotSucceed(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	body := map[string]any{
		"schema": map[string]any{
			"anyvaliVersion": "1.0",
			"schemaVersion":  "1",
			"definitions": map[string]any{
				"self": map[string]any{
					"kind": "ref",
					"ref":  "self",
				},
			},
			"root": map[string]any{
				"kind": "ref",
				"ref":  "self",
			},
		},
		"input": "anything",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Fatalf("recursive ref payload must not return 200 OK: %s", rr.Body.String())
	}
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
		t.Fatalf("recursive ref: expected 400 or 500, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// CWE-942: CORS headers
// ---------------------------------------------------------------------------

func TestCORSHeadersPresent(t *testing.T) {
	srv := newTestServer(t, true, false)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	tests := []struct {
		header string
		want   string
	}{
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Methods", "GET, POST, OPTIONS"},
		{"Access-Control-Allow-Headers", "Content-Type"},
	}

	for _, tt := range tests {
		got := rr.Header().Get(tt.header)
		if got != tt.want {
			t.Errorf("header %q: expected %q, got %q", tt.header, tt.want, got)
		}
	}
}

func TestCORSWildcardOriginDocumented(t *testing.T) {
	// SECURITY FINDING: The CORS middleware sets Access-Control-Allow-Origin
	// to "*", which allows any origin to make requests. For a validation
	// server this may be acceptable, but if sensitive schemas are loaded,
	// consider restricting to specific origins (CWE-942).
	srv := newTestServer(t, true, false)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected wildcard origin, got %q", origin)
	}
	t.Log("SECURITY: CORS allows all origins (*). Consider restricting for production deployments.")
}

func TestCORSPreflightRequest(t *testing.T) {
	srv := newTestServer(t, true, false)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodOptions, "/validate", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rr.Code)
	}

	// Verify the response includes CORS headers.
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("preflight missing Access-Control-Allow-Origin")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("preflight missing Access-Control-Allow-Methods")
	}
}

func TestCORSDisabledNoHeaders(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected no CORS headers when disabled, got origin=%q", origin)
	}
}

// ---------------------------------------------------------------------------
// CWE-755: Recovery middleware (panic handling)
// ---------------------------------------------------------------------------

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	// Create a handler that panics, wrap it in recovery middleware.
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic from handler")
	})

	handler := recoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/trigger-panic", nil)
	rr := httptest.NewRecorder()

	// This should NOT propagate the panic to the test.
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", rr.Code)
	}

	resp := decodeJSONResponse(t, rr)
	errMsg, ok := resp["error"].(string)
	if !ok {
		t.Fatalf("expected error in response, got %v", resp)
	}
	if errMsg != "internal server error" {
		t.Errorf("expected generic error message, got %q", errMsg)
	}
}

func TestRecoveryMiddlewareDoesNotLeakPanicDetails(t *testing.T) {
	// Panic with a sensitive string to verify it is not echoed to the client.
	sensitive := "database password is s3cret"
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(sensitive)
	})

	handler := recoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/trigger-panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, sensitive) {
		t.Errorf("panic detail leaked to client: %s", body)
	}
	if strings.Contains(body, "s3cret") {
		t.Errorf("sensitive data leaked to client: %s", body)
	}
}

func TestRecoveryMiddlewareNilPanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})

	handler := recoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/trigger-panic", nil)
	rr := httptest.NewRecorder()

	// Should not propagate; nil panic is still caught by recover().
	handler.ServeHTTP(rr, req)

	// With Go 1.21+ panic(nil) is wrapped, so recover() returns non-nil.
	// The middleware should either return 500 or pass through cleanly.
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusOK {
		t.Errorf("unexpected status %d after nil panic", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// HTTP method enforcement
// ---------------------------------------------------------------------------

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer(t, false, true)
	handler := srv.Handler()

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/validate"},
		{http.MethodPut, "/validate"},
		{http.MethodDelete, "/validate"},
		{http.MethodPatch, "/validate"},
		{http.MethodGet, "/validate/user"},
		{http.MethodPut, "/validate/user"},
		{http.MethodPost, "/schemas"},
		{http.MethodDelete, "/schemas"},
		{http.MethodPost, "/health"},
		{http.MethodDelete, "/health"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: expected 405, got %d", tt.method, tt.path, rr.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Response header checks
// ---------------------------------------------------------------------------

func TestResponseContentTypeIsJSON(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/health", ""},
		{http.MethodGet, "/schemas", ""},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			var req *http.Request
			if ep.body != "" {
				req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			ct := rr.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				t.Errorf("expected Content-Type application/json, got %q", ct)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Missing required fields
// ---------------------------------------------------------------------------

func TestValidateMissingInputField(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	body := `{"schema": {"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"string"}}}`
	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing input, got %d", rr.Code)
	}

	resp := decodeJSONResponse(t, rr)
	errMsg, _ := resp["error"].(string)
	if !strings.Contains(errMsg, "input") {
		t.Errorf("error should mention missing 'input' field, got: %q", errMsg)
	}
}

func TestValidateMissingSchemaAndRef(t *testing.T) {
	srv := newTestServer(t, false, false)
	handler := srv.Handler()

	body := `{"input": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing schema/schemaRef, got %d", rr.Code)
	}

	resp := decodeJSONResponse(t, rr)
	errMsg, _ := resp["error"].(string)
	if !strings.Contains(errMsg, "schema") {
		t.Errorf("error should mention missing schema, got: %q", errMsg)
	}
}

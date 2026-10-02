package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsLiteralTypeAndPrecisionBypasses(t *testing.T) {
	srv := newTestServer(t, false, false)
	for _, tc := range []struct {
		name string
		body string
	}{
		{"boolean as string", `{"schema":{"root":{"kind":"literal","value":true}},"input":"true"}`},
		{"number as string", `{"schema":{"root":{"kind":"enum","values":[1,2]}},"input":"1"}`},
		{"rounded literal", `{"schema":{"root":{"kind":"literal","value":9007199254740992}},"input":9007199254740993}`},
		{"rounded int max", `{"schema":{"root":{"kind":"int64","max":9007199254740992}},"input":9007199254740993}`},
		{"rounded int min", `{"schema":{"root":{"kind":"int64","min":9007199254740993}},"input":9007199254740992}`},
		{"unsigned overflow", `{"schema":{"root":{"kind":"uint64"}},"input":18446744073709551616}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status %d: %s", rr.Code, rr.Body.String())
			}
			if response := decodeJSONResponse(t, rr); response["valid"] != false {
				t.Fatalf("validation bypassed: %s", rr.Body.String())
			}
		})
	}
}

func TestValidateRejectsOverflowedLengthConstraints(t *testing.T) {
	srv := newTestServer(t, false, false)
	for _, body := range []string{
		`{"schema":{"root":{"kind":"string","minLength":100000000000000000000}},"input":""}`,
		`{"schema":{"root":{"kind":"array","items":{"kind":"any"},"minItems":100000000000000000000}},"input":[]}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("overflowed constraint accepted: %s", rr.Body.String())
		}
	}
}

func TestValidateRejectsUnrepresentableFloatConstraint(t *testing.T) {
	srv := newTestServer(t, false, false)
	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"schema":{"root":{"kind":"number","min":1e400}},"input":0}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unrepresentable minimum was ignored: %d %s", rr.Code, rr.Body.String())
	}
}

func TestNamedSchemaPreservesIntegerPrecision(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "id.json"), []byte(`{"root":{"kind":"literal","value":9007199254740992}}`), 0600); err != nil {
		t.Fatal(err)
	}
	srv, err := New(Config{SchemasDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		input string
		valid bool
	}{
		{`9007199254740992`, true},
		{`9007199254740993`, false},
	} {
		req := httptest.NewRequest(http.MethodPost, "/validate/id", strings.NewReader(tc.input))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK || decodeJSONResponse(t, rr)["valid"] != tc.valid {
			t.Fatalf("input %s: status=%d body=%s", tc.input, rr.Code, rr.Body.String())
		}
	}
}

func TestValidatePreservesExactUint64(t *testing.T) {
	srv := newTestServer(t, false, false)
	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"schema":{"root":{"kind":"uint64"}},"input":18446744073709551615}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", rr.Code, rr.Body.String())
	}
	if response := decodeJSONResponse(t, rr); response["valid"] != true {
		t.Fatalf("valid uint64 rejected: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `18446744073709551615`) {
		t.Fatalf("uint64 lost precision in response: %s", rr.Body.String())
	}
}

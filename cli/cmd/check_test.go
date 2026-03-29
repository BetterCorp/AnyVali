package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckValidSchema(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"name":  map[string]any{"kind": "string"},
				"email": map[string]any{"kind": "string", "format": "email"},
				"age":   map[string]any{"kind": "int64"},
				"admin": map[string]any{"kind": "bool"},
			},
			"required":    []any{"name", "email", "age", "admin"},
			"unknownKeys": "reject",
		},
	}

	data, _ := json.Marshal(schema)
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	os.WriteFile(path, data, 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunCheck([]string{path})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "Schema is valid and portable") {
		t.Errorf("expected 'valid and portable' message, got:\n%s", output)
	}
	if !strings.Contains(output, "Kind: object") {
		t.Errorf("expected 'Kind: object', got:\n%s", output)
	}
	if !strings.Contains(output, "Properties: 4") {
		t.Errorf("expected 'Properties: 4', got:\n%s", output)
	}
}

func TestCheckSchemaWithExtensions(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind": "string",
		},
		"extensions": map[string]any{
			"js": map[string]any{
				"customValidator": true,
			},
		},
	}

	data, _ := json.Marshal(schema)
	output, code := CheckSchema(data)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output, "portability issues") {
		t.Errorf("expected portability issues, got:\n%s", output)
	}
	if !strings.Contains(output, "Extension") {
		t.Errorf("expected extension warning, got:\n%s", output)
	}
}

func TestCheckSchemaWithRefine(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"kind":   "string",
					"refine": "customCheck",
				},
			},
			"required":    []any{"name"},
			"unknownKeys": "reject",
		},
	}

	data, _ := json.Marshal(schema)
	output, code := CheckSchema(data)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output, "Custom validator") {
		t.Errorf("expected custom validator warning, got:\n%s", output)
	}
}

func TestCheckSchemaWithTransform(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind":      "string",
			"transform": "toLowerCase",
		},
	}

	data, _ := json.Marshal(schema)
	output, code := CheckSchema(data)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output, "Transform") {
		t.Errorf("expected transform warning, got:\n%s", output)
	}
}

func TestCheckBadFile(t *testing.T) {
	code := RunCheck([]string{"/nonexistent/file.json"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestCheckBadJSON(t *testing.T) {
	data := []byte("not json")
	output, code := CheckSchema(data)

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(output, "invalid JSON") {
		t.Errorf("expected JSON error, got: %s", output)
	}
}

func TestCheckMissingRoot(t *testing.T) {
	data := []byte(`{"anyvaliVersion": "1.0"}`)
	output, code := CheckSchema(data)

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(output, "missing 'root'") {
		t.Errorf("expected missing root error, got: %s", output)
	}
}

func TestCheckMissingKind(t *testing.T) {
	data := []byte(`{"root": {}}`)
	output, code := CheckSchema(data)

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(output, "missing 'kind'") {
		t.Errorf("expected missing kind error, got: %s", output)
	}
}

func TestCheckNoArgs(t *testing.T) {
	code := RunCheck([]string{})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestCheckHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunCheck([]string{"--help"})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("expected help output, got: %s", output)
	}
}

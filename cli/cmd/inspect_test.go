package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectObjectSchema(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"kind":      "string",
					"minLength": 1,
					"maxLength": 100,
				},
				"email": map[string]any{
					"kind":   "string",
					"format": "email",
				},
				"age": map[string]any{
					"kind": "optional",
					"schema": map[string]any{
						"kind": "int64",
						"min":  0,
						"max":  150,
					},
				},
				"tags": map[string]any{
					"kind": "optional",
					"schema": map[string]any{
						"kind":     "array",
						"item":     map[string]any{"kind": "string"},
						"maxItems": 10,
					},
				},
			},
			"required":    []any{"name", "email"},
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

	code := RunInspect([]string{path})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "Schema: object") {
		t.Errorf("expected 'Schema: object' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Properties:") {
		t.Errorf("expected 'Properties:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "name:") {
		t.Errorf("expected 'name:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "email:") {
		t.Errorf("expected 'email:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Unknown keys: reject") {
		t.Errorf("expected 'Unknown keys: reject' in output, got:\n%s", output)
	}
}

func TestInspectStringSchema(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind":      "string",
			"minLength": 1,
			"format":    "email",
		},
	}

	data, _ := json.Marshal(schema)
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	os.WriteFile(path, data, 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunInspect([]string{path})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "Schema: string") {
		t.Errorf("expected 'Schema: string' in output, got:\n%s", output)
	}
}

func TestInspectArraySchema(t *testing.T) {
	schema := map[string]any{
		"anyvaliVersion": "1.0",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind":     "array",
			"item":     map[string]any{"kind": "string"},
			"maxItems": 5,
		},
	}

	data, _ := json.Marshal(schema)
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	os.WriteFile(path, data, 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunInspect([]string{path})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "Schema: array") {
		t.Errorf("expected 'Schema: array' in output, got:\n%s", output)
	}
}

func TestInspectBadFile(t *testing.T) {
	code := RunInspect([]string{"/nonexistent/file.json"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestInspectBadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	code := RunInspect([]string{path})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestInspectNoArgs(t *testing.T) {
	code := RunInspect([]string{})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestInspectHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunInspect([]string{"--help"})

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

func TestInspectNode(t *testing.T) {
	root := map[string]any{
		"kind": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"kind":      "string",
				"minLength": 1,
			},
		},
		"required":    []any{"name"},
		"unknownKeys": "reject",
	}
	doc := map[string]any{"root": root}

	output := InspectNode(root, doc)

	if !strings.Contains(output, "Schema: object") {
		t.Errorf("expected 'Schema: object', got:\n%s", output)
	}
	if !strings.Contains(output, "name:") {
		t.Errorf("expected 'name:' property, got:\n%s", output)
	}
}

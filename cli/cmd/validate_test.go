package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	anyvali "github.com/anyvali/anyvali/sdk/go"
)

// helper to write a temp schema file and return its path
func writeTempSchema(t *testing.T, schema map[string]any) string {
	t.Helper()
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write schema file: %v", err)
	}
	return path
}

// helper to write a temp input file and return its path
func writeTempInput(t *testing.T, input any) string {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "input.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	return path
}

func testSchema() map[string]any {
	return map[string]any{
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
			},
			"required":    []any{"name", "email"},
			"unknownKeys": "reject",
		},
	}
}

func TestValidateValidInput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	inputPath := writeTempInput(t, map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
	})

	code := RunValidate([]string{schemaPath, inputPath, "-f", "json"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestValidateInvalidInput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	// Missing required "name" field
	input := `{"email": "alice@example.com"}`

	code := RunValidate([]string{schemaPath, input, "-f", "quiet"})
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid input, got %d", code)
	}
}

func TestValidateJSONOutput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	inputPath := writeTempInput(t, map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
	})

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunValidate([]string{schemaPath, inputPath, "-f", "json"})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}

	if result["valid"] != true {
		t.Errorf("expected valid=true, got %v", result["valid"])
	}
}

func TestValidateTextOutputValid(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	inputPath := writeTempInput(t, map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
	})

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunValidate([]string{schemaPath, inputPath, "-f", "text"})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "Valid") {
		t.Errorf("expected output to contain 'Valid', got: %s", output)
	}
}

func TestValidateTextOutputInvalid(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	input := `{"email": "not-email"}`

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunValidate([]string{schemaPath, input, "-f", "text"})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(output, "Invalid") {
		t.Errorf("expected output to contain 'Invalid', got: %s", output)
	}
}

func TestValidateQuietModeValid(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	inputPath := writeTempInput(t, map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
	})

	code := RunValidate([]string{schemaPath, inputPath, "-f", "quiet"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestValidateQuietModeInvalid(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	input := `{}`

	code := RunValidate([]string{schemaPath, input, "-f", "quiet"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestValidateFileInput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())
	inputPath := writeTempInput(t, map[string]any{
		"name":  "Bob",
		"email": "bob@example.com",
	})

	code := RunValidate([]string{schemaPath, inputPath, "-f", "quiet"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestValidateStdinInput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())

	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte(`{"name":"Charlie","email":"charlie@example.com"}`))
		w.Close()
	}()

	code := RunValidate([]string{schemaPath, "-", "-f", "quiet"})
	os.Stdin = oldStdin

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestValidateBadSchemaFile(t *testing.T) {
	code := RunValidate([]string{"/nonexistent/schema.json", `{}`, "-f", "quiet"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestValidateBadInput(t *testing.T) {
	schemaPath := writeTempSchema(t, testSchema())

	code := RunValidate([]string{schemaPath, "not-valid-json{{{", "-f", "quiet"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestValidateBadSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte(`{"root": {"kind": "unsupported_xyz"}}`), 0644)

	code := RunValidate([]string{path, `{}`, "-f", "quiet"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestValidateNoArgs(t *testing.T) {
	code := RunValidate([]string{})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestValidateHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunValidate([]string{"-h"})

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

func TestValidateInputHelper(t *testing.T) {
	schemaData, _ := json.Marshal(testSchema())
	schema, err := anyvali.ImportJSON(schemaData)
	if err != nil {
		t.Fatalf("import schema: %v", err)
	}

	t.Run("valid json format", func(t *testing.T) {
		input := map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
		}
		output, code := ValidateInput(schema, input, "json", false)
		if code != 0 {
			t.Errorf("expected code 0, got %d", code)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("invalid JSON output: %v", err)
		}
		if result["valid"] != true {
			t.Errorf("expected valid=true")
		}
	})

	t.Run("invalid text format", func(t *testing.T) {
		input := map[string]any{}
		output, code := ValidateInput(schema, input, "text", true)
		if code != 1 {
			t.Errorf("expected code 1, got %d", code)
		}
		if !strings.Contains(output, "Invalid") {
			t.Errorf("expected 'Invalid' in output, got: %s", output)
		}
	})

	t.Run("quiet valid", func(t *testing.T) {
		input := map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
		}
		_, code := ValidateInput(schema, input, "quiet", false)
		if code != 0 {
			t.Errorf("expected code 0, got %d", code)
		}
	})

	t.Run("quiet invalid", func(t *testing.T) {
		_, code := ValidateInput(schema, map[string]any{}, "quiet", false)
		if code != 1 {
			t.Errorf("expected code 1, got %d", code)
		}
	})
}

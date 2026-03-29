package anyvali

import "testing"

func TestObjectSchemaInvalidType(t *testing.T) {
	s := Object(map[string]Schema{"name": String()})
	tests := []any{"hello", 42, true, nil, []any{1}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
		if r.Issues[0].Code != IssueInvalidType {
			t.Fatalf("expected invalid_type, got %s", r.Issues[0].Code)
		}
	}
}

func TestObjectSchemaRequiredFieldMissing(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
		"age":  Int(),
	})
	r := s.SafeParse(map[string]any{"name": "Alice"})
	if r.Success {
		t.Fatal("expected failure for missing required field")
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Code == IssueRequired {
			found = true
		}
	}
	if !found {
		t.Fatal("expected required issue")
	}
}

func TestObjectSchemaAllFieldsRequired(t *testing.T) {
	s := Object(map[string]Schema{
		"a": String(),
		"b": String(),
	})
	r := s.SafeParse(map[string]any{})
	if r.Success {
		t.Fatal("expected failure for all fields missing")
	}
	if len(r.Issues) < 2 {
		t.Fatalf("expected at least 2 issues, got %d", len(r.Issues))
	}
}

func TestObjectSchemaRequiredOverride(t *testing.T) {
	s := Object(map[string]Schema{
		"a": String(),
		"b": String(),
	}).Required("a") // Only "a" is required
	r := s.SafeParse(map[string]any{"a": "hello"})
	if !r.Success {
		t.Fatalf("expected success, got issues: %v", r.Issues)
	}
}

func TestObjectSchemaUnknownKeysReject(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).UnknownKeys(Reject)
	r := s.SafeParse(map[string]any{"name": "Alice", "extra": "val"})
	if r.Success {
		t.Fatal("expected failure for unknown key")
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Code == IssueUnknownKey {
			found = true
		}
	}
	if !found {
		t.Fatal("expected unknown_key issue")
	}
}

func TestObjectSchemaUnknownKeysStrip(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Required("name").UnknownKeys(Strip)
	r := s.SafeParse(map[string]any{"name": "Alice", "extra": "val", "more": 42})
	if !r.Success {
		t.Fatal("expected success with strip mode")
	}
	obj := r.Data.(map[string]any)
	if _, ok := obj["extra"]; ok {
		t.Fatal("expected extra to be stripped")
	}
	if _, ok := obj["more"]; ok {
		t.Fatal("expected more to be stripped")
	}
	if obj["name"] != "Alice" {
		t.Fatal("expected name to remain")
	}
}

func TestObjectSchemaUnknownKeysAllow(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Required("name").UnknownKeys(Allow)
	r := s.SafeParse(map[string]any{"name": "Alice", "extra": "val"})
	if !r.Success {
		t.Fatal("expected success with allow mode")
	}
	obj := r.Data.(map[string]any)
	if obj["extra"] != "val" {
		t.Fatal("expected extra to be preserved")
	}
}

func TestObjectSchemaNestedValidation(t *testing.T) {
	s := Object(map[string]Schema{
		"inner": Object(map[string]Schema{
			"value": Int(),
		}),
	})
	r := s.SafeParse(map[string]any{
		"inner": map[string]any{
			"value": "not-int",
		},
	})
	if r.Success {
		t.Fatal("expected failure for nested invalid value")
	}
	// Check path includes both "inner" and "value"
	found := false
	for _, issue := range r.Issues {
		if len(issue.Path) >= 2 && issue.Path[0] == "inner" && issue.Path[1] == "value" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected path [inner, value], got: %v", r.Issues)
	}
}

func TestObjectSchemaOptionalField(t *testing.T) {
	s := Object(map[string]Schema{
		"name":  String(),
		"email": Optional(String()),
	})
	// email is optional, should succeed without it
	r := s.SafeParse(map[string]any{"name": "Alice"})
	if !r.Success {
		t.Fatalf("expected success with optional field missing: %v", r.Issues)
	}
}

func TestObjectSchemaPropertyValidationError(t *testing.T) {
	s := Object(map[string]Schema{
		"count": Int().Min(0),
	})
	r := s.SafeParse(map[string]any{"count": int64(-1)})
	if r.Success {
		t.Fatal("expected failure for invalid property value")
	}
	if len(r.Issues) == 0 {
		t.Fatal("expected issues")
	}
	// Path should include "count"
	if r.Issues[0].Path[0] != "count" {
		t.Fatalf("expected path [count], got %v", r.Issues[0].Path)
	}
}

func TestObjectSchemaParse(t *testing.T) {
	s := Object(map[string]Schema{"name": String()}).Required("name")
	v, err := s.Parse(map[string]any{"name": "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	obj := v.(map[string]any)
	if obj["name"] != "Alice" {
		t.Fatal("expected Alice")
	}

	_, err = s.Parse(42)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestObjectSchemaDefault(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Default(map[string]any{"name": "default"})
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success with default, got: %v", r.Issues)
	}
}

func TestObjectSchemaToNode(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
		"age":  Int(),
	}).UnknownKeys(Allow)

	node := s.ToNode()
	if node["kind"] != "object" {
		t.Fatal("expected kind=object")
	}
	if node["unknownKeys"] != "allow" {
		t.Fatalf("expected unknownKeys=allow, got %v", node["unknownKeys"])
	}
	props, ok := node["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be map")
	}
	if _, ok := props["name"]; !ok {
		t.Fatal("expected name property")
	}
	if _, ok := props["age"]; !ok {
		t.Fatal("expected age property")
	}
	reqList, ok := node["required"].([]any)
	if !ok {
		t.Fatal("expected required to be []any")
	}
	if len(reqList) < 1 {
		t.Fatal("expected at least 1 required field")
	}
}

func TestObjectSchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Default(map[string]any{"name": "x"})
	node := s.ToNode()
	if _, ok := node["default"]; !ok {
		t.Fatal("expected default in node")
	}
}

func TestObjectSchemaEmptyProperties(t *testing.T) {
	s := Object(map[string]Schema{})
	r := s.SafeParse(map[string]any{})
	if !r.Success {
		t.Fatal("expected success for empty object")
	}
}

func TestObjectSchemaFieldNotRequired(t *testing.T) {
	s := Object(map[string]Schema{
		"a": String(),
		"b": String(),
	}).Required() // No fields required
	r := s.SafeParse(map[string]any{})
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

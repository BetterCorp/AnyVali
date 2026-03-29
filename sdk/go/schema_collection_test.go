package anyvali

import "testing"

// --- Array tests ---

func TestArraySchemaValid(t *testing.T) {
	s := Array(String())
	r := s.SafeParse([]any{"a", "b", "c"})
	if !r.Success {
		t.Fatal("expected success")
	}
	arr := r.Data.([]any)
	if len(arr) != 3 {
		t.Fatalf("expected 3 items, got %d", len(arr))
	}
}

func TestArraySchemaInvalidType(t *testing.T) {
	s := Array(String())
	tests := []any{"hello", 42, true, nil, map[string]any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestArraySchemaItemValidationFails(t *testing.T) {
	s := Array(Int())
	r := s.SafeParse([]any{int64(1), "not-int", int64(3)})
	if r.Success {
		t.Fatal("expected failure")
	}
	// Check path includes index
	found := false
	for _, issue := range r.Issues {
		if len(issue.Path) > 0 && issue.Path[0] == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected path with index 1, got: %v", r.Issues)
	}
}

func TestArraySchemaMinItems(t *testing.T) {
	s := Array(String()).MinItems(2)
	if r := s.SafeParse([]any{"a"}); r.Success {
		t.Fatal("expected failure for too few items")
	}
	if r := s.SafeParse([]any{"a", "b"}); !r.Success {
		t.Fatal("expected success for exact count")
	}
}

func TestArraySchemaMaxItems(t *testing.T) {
	s := Array(String()).MaxItems(2)
	if r := s.SafeParse([]any{"a", "b", "c"}); r.Success {
		t.Fatal("expected failure for too many items")
	}
	if r := s.SafeParse([]any{"a", "b"}); !r.Success {
		t.Fatal("expected success for exact count")
	}
}

func TestArraySchemaEmpty(t *testing.T) {
	s := Array(String())
	r := s.SafeParse([]any{})
	if !r.Success {
		t.Fatal("expected success for empty array")
	}
}

func TestArraySchemaDefault(t *testing.T) {
	s := Array(String()).Default([]any{"x"})
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestArraySchemaParse(t *testing.T) {
	s := Array(String())
	v, err := s.Parse([]any{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	arr := v.([]any)
	if len(arr) != 2 {
		t.Fatalf("expected 2, got %d", len(arr))
	}
	_, err = s.Parse("not-array")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArraySchemaToNode(t *testing.T) {
	s := Array(String()).MinItems(1).MaxItems(10)
	node := s.ToNode()
	if node["kind"] != "array" {
		t.Fatal("expected kind=array")
	}
	item, ok := node["item"].(map[string]any)
	if !ok {
		t.Fatal("expected item node")
	}
	if item["kind"] != "string" {
		t.Fatal("expected item kind=string")
	}
	if node["minItems"] != 1 {
		t.Fatal("expected minItems=1")
	}
	if node["maxItems"] != 10 {
		t.Fatal("expected maxItems=10")
	}
}

func TestArraySchemaToNodeMinimal(t *testing.T) {
	s := Array(Int())
	node := s.ToNode()
	if _, ok := node["minItems"]; ok {
		t.Fatal("unexpected minItems")
	}
	if _, ok := node["maxItems"]; ok {
		t.Fatal("unexpected maxItems")
	}
}

func TestArraySchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := Array(String()).Default([]any{})
	node := s.ToNode()
	if _, ok := node["default"]; !ok {
		t.Fatal("expected default in node")
	}
}

// --- Tuple tests ---

func TestTupleSchemaValid(t *testing.T) {
	s := Tuple(String(), Int(), Bool())
	r := s.SafeParse([]any{"hello", int64(42), true})
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestTupleSchemaInvalidType(t *testing.T) {
	s := Tuple(String())
	tests := []any{"hello", 42, true, nil, map[string]any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestTupleSchemaWrongLength(t *testing.T) {
	s := Tuple(String(), Int())
	if r := s.SafeParse([]any{"hello"}); r.Success {
		t.Fatal("expected failure for too few items")
	}
	if r := s.SafeParse([]any{"hello", int64(1), true}); r.Success {
		t.Fatal("expected failure for too many items")
	}
}

func TestTupleSchemaItemValidationFails(t *testing.T) {
	s := Tuple(String(), Int())
	r := s.SafeParse([]any{"hello", "not-int"})
	if r.Success {
		t.Fatal("expected failure")
	}
	// Check path includes index
	found := false
	for _, issue := range r.Issues {
		if len(issue.Path) > 0 && issue.Path[0] == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected path with index 1, got: %v", r.Issues)
	}
}

func TestTupleSchemaParse(t *testing.T) {
	s := Tuple(String(), Int())
	v, err := s.Parse([]any{"hello", int64(42)})
	if err != nil {
		t.Fatal(err)
	}
	arr := v.([]any)
	if len(arr) != 2 {
		t.Fatal("expected 2 items")
	}
	_, err = s.Parse("not-array")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTupleSchemaToNode(t *testing.T) {
	s := Tuple(String(), Int(), Bool())
	node := s.ToNode()
	if node["kind"] != "tuple" {
		t.Fatal("expected kind=tuple")
	}
	items, ok := node["items"].([]any)
	if !ok {
		t.Fatal("expected items to be []any")
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

// --- Record tests ---

func TestRecordSchemaValid(t *testing.T) {
	s := Record(Int())
	r := s.SafeParse(map[string]any{"a": int64(1), "b": int64(2)})
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestRecordSchemaInvalidType(t *testing.T) {
	s := Record(String())
	tests := []any{"hello", 42, true, nil, []any{1}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestRecordSchemaValueValidationFails(t *testing.T) {
	s := Record(Int())
	r := s.SafeParse(map[string]any{"a": int64(1), "b": "not-int"})
	if r.Success {
		t.Fatal("expected failure")
	}
	// Check path includes key
	found := false
	for _, issue := range r.Issues {
		if len(issue.Path) > 0 && issue.Path[0] == "b" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected path with key 'b', got: %v", r.Issues)
	}
}

func TestRecordSchemaEmpty(t *testing.T) {
	s := Record(String())
	r := s.SafeParse(map[string]any{})
	if !r.Success {
		t.Fatal("expected success for empty record")
	}
}

func TestRecordSchemaParse(t *testing.T) {
	s := Record(String())
	v, err := s.Parse(map[string]any{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	obj := v.(map[string]any)
	if obj["k"] != "v" {
		t.Fatal("expected k=v")
	}
	_, err = s.Parse("not-map")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRecordSchemaToNode(t *testing.T) {
	s := Record(Int())
	node := s.ToNode()
	if node["kind"] != "record" {
		t.Fatal("expected kind=record")
	}
	val, ok := node["value"].(map[string]any)
	if !ok {
		t.Fatal("expected value node")
	}
	if val["kind"] != "int" {
		t.Fatal("expected value kind=int")
	}
}

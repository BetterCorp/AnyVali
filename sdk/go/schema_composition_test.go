package anyvali

import "testing"

// --- Union tests ---

func TestUnionSchemaFirstMatch(t *testing.T) {
	s := Union(String(), Int(), Bool())
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success for string")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestUnionSchemaSecondMatch(t *testing.T) {
	s := Union(String(), Int())
	r := s.SafeParse(int64(42))
	if !r.Success {
		t.Fatal("expected success for int")
	}
}

func TestUnionSchemaAllFail(t *testing.T) {
	s := Union(String(), Int())
	r := s.SafeParse(true)
	if r.Success {
		t.Fatal("expected failure for bool in string|int union")
	}
	if r.Issues[0].Code != IssueInvalidUnion {
		t.Fatalf("expected invalid_union, got %s", r.Issues[0].Code)
	}
	// Check that meta contains unionIssues
	meta := r.Issues[0].Meta
	if meta == nil {
		t.Fatal("expected meta with unionIssues")
	}
	unionIssues, ok := meta["unionIssues"].([]ValidationIssue)
	if !ok {
		t.Fatal("expected unionIssues to be []ValidationIssue")
	}
	if len(unionIssues) == 0 {
		t.Fatal("expected union issues to be non-empty")
	}
}

func TestUnionSchemaAllFailNilInput(t *testing.T) {
	s := Union(String(), Int())
	r := s.SafeParse(nil)
	if r.Success {
		t.Fatal("expected failure for nil in string|int union")
	}
}

func TestUnionSchemaParse(t *testing.T) {
	s := Union(String(), Int())
	v, err := s.Parse("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
	_, err = s.Parse(3.14)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnionSchemaToNode(t *testing.T) {
	s := Union(String(), Int())
	node := s.ToNode()
	if node["kind"] != "union" {
		t.Fatal("expected kind=union")
	}
	schemas, ok := node["schemas"].([]any)
	if !ok {
		t.Fatal("expected schemas to be []any")
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
}

// --- Intersection tests ---

func TestIntersectionSchemaAllPass(t *testing.T) {
	s := Intersection(
		Object(map[string]Schema{"name": String()}).Required("name").UnknownKeys(Allow),
		Object(map[string]Schema{"age": Int()}).Required("age").UnknownKeys(Allow),
	)
	input := map[string]any{"name": "Alice", "age": int64(30)}
	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	obj := r.Data.(map[string]any)
	if obj["name"] != "Alice" {
		t.Fatal("expected name=Alice")
	}
}

func TestIntersectionSchemaOneFails(t *testing.T) {
	s := Intersection(
		Object(map[string]Schema{"name": String()}).Required("name").UnknownKeys(Allow),
		Object(map[string]Schema{"age": Int()}).Required("age").UnknownKeys(Allow),
	)
	// Missing "age"
	r := s.SafeParse(map[string]any{"name": "Alice"})
	if r.Success {
		t.Fatal("expected failure when one schema fails")
	}
}

func TestIntersectionSchemaNonObject(t *testing.T) {
	// Intersection of non-object schemas
	s := Intersection(
		Number().Min(0),
		Number().Max(100),
	)
	r := s.SafeParse(float64(50))
	if !r.Success {
		t.Fatal("expected success for number in range")
	}

	r = s.SafeParse(float64(150))
	if r.Success {
		t.Fatal("expected failure for number > 100")
	}
}

func TestIntersectionSchemaMergesObjects(t *testing.T) {
	s := Intersection(
		Object(map[string]Schema{"a": String()}).Required("a").UnknownKeys(Allow),
		Object(map[string]Schema{"b": Int()}).Required("b").UnknownKeys(Allow),
	)
	input := map[string]any{"a": "hello", "b": int64(42)}
	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	obj := r.Data.(map[string]any)
	if obj["a"] != "hello" {
		t.Fatal("expected a=hello in merged result")
	}
}

func TestIntersectionSchemaParse(t *testing.T) {
	s := Intersection(Number().Min(0), Number().Max(100))
	v, err := s.Parse(float64(50))
	if err != nil {
		t.Fatal(err)
	}
	if v != float64(50) {
		t.Fatalf("expected 50, got %v", v)
	}
	_, err = s.Parse("string")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIntersectionSchemaToNode(t *testing.T) {
	s := Intersection(String(), Int())
	node := s.ToNode()
	if node["kind"] != "intersection" {
		t.Fatal("expected kind=intersection")
	}
	schemas, ok := node["schemas"].([]any)
	if !ok {
		t.Fatal("expected schemas to be []any")
	}
	if len(schemas) != 2 {
		t.Fatalf("expected 2 schemas, got %d", len(schemas))
	}
}

// --- Optional tests ---

func TestOptionalSchemaPresent(t *testing.T) {
	s := Optional(String())
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success for present value")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestOptionalSchemaNil(t *testing.T) {
	s := Optional(String())
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestOptionalSchemaAbsent(t *testing.T) {
	s := Optional(String())
	r := s.SafeParse(absentValue)
	if !r.Success {
		t.Fatal("expected success for absent value")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestOptionalSchemaWithDefault(t *testing.T) {
	s := Optional(String()).Default("fallback")
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestOptionalSchemaWithDefaultAbsent(t *testing.T) {
	s := Optional(String()).Default("fallback")
	r := s.SafeParse(absentValue)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestOptionalSchemaInvalidInner(t *testing.T) {
	s := Optional(Int())
	r := s.SafeParse("not an int")
	if r.Success {
		t.Fatal("expected failure for invalid inner type")
	}
}

func TestOptionalSchemaParse(t *testing.T) {
	s := Optional(String())
	v, err := s.Parse("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
	v, err = s.Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestOptionalSchemaParseError(t *testing.T) {
	s := Optional(Int())
	_, err := s.Parse("bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOptionalSchemaToNode(t *testing.T) {
	s := Optional(String())
	node := s.ToNode()
	if node["kind"] != "optional" {
		t.Fatal("expected kind=optional")
	}
	inner, ok := node["schema"].(map[string]any)
	if !ok {
		t.Fatal("expected schema to be map")
	}
	if inner["kind"] != "string" {
		t.Fatal("expected inner kind=string")
	}
}

func TestOptionalSchemaToNodeWithDefault(t *testing.T) {
	s := Optional(String()).Default("x")
	node := s.ToNode()
	if node["default"] != "x" {
		t.Fatalf("expected default=x, got %v", node["default"])
	}
}

// --- Nullable tests ---

func TestNullableSchemaPresent(t *testing.T) {
	s := Nullable(String())
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestNullableSchemaNil(t *testing.T) {
	s := Nullable(String())
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestNullableSchemaAbsent(t *testing.T) {
	s := Nullable(String())
	r := s.SafeParse(absentValue)
	if !r.Success {
		t.Fatal("expected success for absent")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestNullableSchemaWithDefault(t *testing.T) {
	s := Nullable(String()).Default("fallback")
	// For absent input, should apply default
	r := s.SafeParse(absentValue)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestNullableSchemaWithDefaultNil(t *testing.T) {
	s := Nullable(String()).Default("fallback")
	// nil should pass through as null, NOT apply default
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestNullableSchemaInvalidInner(t *testing.T) {
	s := Nullable(Int())
	r := s.SafeParse("not an int")
	if r.Success {
		t.Fatal("expected failure for invalid inner type")
	}
}

func TestNullableSchemaParse(t *testing.T) {
	s := Nullable(String())
	v, err := s.Parse("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
	_, err = s.Parse(42)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNullableSchemaToNode(t *testing.T) {
	s := Nullable(String())
	node := s.ToNode()
	if node["kind"] != "nullable" {
		t.Fatal("expected kind=nullable")
	}
	inner, ok := node["schema"].(map[string]any)
	if !ok {
		t.Fatal("expected schema to be map")
	}
	if inner["kind"] != "string" {
		t.Fatal("expected inner kind=string")
	}
}

func TestNullableSchemaToNodeWithDefault(t *testing.T) {
	s := Nullable(Int()).Default(0)
	node := s.ToNode()
	if _, ok := node["default"]; !ok {
		t.Fatal("expected default in node")
	}
}

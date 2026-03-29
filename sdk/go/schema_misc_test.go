package anyvali

import "testing"

// --- Any tests ---

func TestAnySchemaAcceptsEverything(t *testing.T) {
	s := Any()
	tests := []any{"hello", 42, 3.14, true, false, nil, []any{1, 2}, map[string]any{"a": 1}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if !r.Success {
			t.Fatalf("expected any to accept %v (%T)", v, v)
		}
		if r.Data != v {
			t.Fatalf("expected data=%v, got %v", v, r.Data)
		}
	}
}

func TestAnySchemaParse(t *testing.T) {
	s := Any()
	v, err := s.Parse("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestAnySchemaToNode(t *testing.T) {
	s := Any()
	node := s.ToNode()
	if node["kind"] != "any" {
		t.Fatal("expected kind=any")
	}
	if len(node) != 1 {
		t.Fatalf("expected only 'kind' key, got %v", node)
	}
}

// --- Unknown tests ---

func TestUnknownSchemaAcceptsEverything(t *testing.T) {
	s := Unknown()
	tests := []any{"hello", 42, true, nil}
	for _, v := range tests {
		r := s.SafeParse(v)
		if !r.Success {
			t.Fatalf("expected unknown to accept %v", v)
		}
	}
}

func TestUnknownSchemaParse(t *testing.T) {
	s := Unknown()
	v, err := s.Parse(42)
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestUnknownSchemaToNode(t *testing.T) {
	s := Unknown()
	node := s.ToNode()
	if node["kind"] != "unknown" {
		t.Fatal("expected kind=unknown")
	}
}

// --- Never tests ---

func TestNeverSchemaRejectsEverything(t *testing.T) {
	s := Never()
	tests := []any{"hello", 42, 3.14, true, nil, []any{}, map[string]any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected never to reject %v", v)
		}
		if r.Issues[0].Code != IssueInvalidType {
			t.Fatalf("expected invalid_type, got %s", r.Issues[0].Code)
		}
	}
}

func TestNeverSchemaParse(t *testing.T) {
	s := Never()
	_, err := s.Parse("anything")
	if err == nil {
		t.Fatal("expected error from never schema")
	}
}

func TestNeverSchemaToNode(t *testing.T) {
	s := Never()
	node := s.ToNode()
	if node["kind"] != "never" {
		t.Fatal("expected kind=never")
	}
}

// --- Literal tests ---

func TestLiteralSchemaString(t *testing.T) {
	s := Literal("hello")
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse("world")
	if r.Success {
		t.Fatal("expected failure")
	}
	if r.Issues[0].Code != IssueInvalidLiteral {
		t.Fatalf("expected invalid_literal, got %s", r.Issues[0].Code)
	}
}

func TestLiteralSchemaNumber(t *testing.T) {
	s := Literal(42.0)
	if r := s.SafeParse(42.0); !r.Success {
		t.Fatal("expected success for float64 match")
	}
	// Integer should match via numeric normalization
	if r := s.SafeParse(int(42)); !r.Success {
		t.Fatal("expected success for int matching float64 literal")
	}
	if r := s.SafeParse(43.0); r.Success {
		t.Fatal("expected failure for different number")
	}
}

func TestLiteralSchemaBool(t *testing.T) {
	s := Literal(true)
	if r := s.SafeParse(true); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse(false); r.Success {
		t.Fatal("expected failure")
	}
}

func TestLiteralSchemaNil(t *testing.T) {
	s := Literal(nil)
	if r := s.SafeParse(nil); !r.Success {
		t.Fatal("expected success for nil literal")
	}
	if r := s.SafeParse("non-nil"); r.Success {
		t.Fatal("expected failure for non-nil")
	}
}

func TestLiteralSchemaParse(t *testing.T) {
	s := Literal("x")
	v, err := s.Parse("x")
	if err != nil {
		t.Fatal(err)
	}
	if v != "x" {
		t.Fatalf("expected 'x', got %v", v)
	}
	_, err = s.Parse("y")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLiteralSchemaToNode(t *testing.T) {
	s := Literal("hello")
	node := s.ToNode()
	if node["kind"] != "literal" {
		t.Fatal("expected kind=literal")
	}
	if node["value"] != "hello" {
		t.Fatalf("expected value=hello, got %v", node["value"])
	}
}

func TestLiteralEqualBothNil(t *testing.T) {
	if !literalEqual(nil, nil) {
		t.Fatal("expected nil == nil")
	}
}

func TestLiteralEqualOneNil(t *testing.T) {
	if literalEqual(nil, "x") {
		t.Fatal("expected nil != x")
	}
	if literalEqual("x", nil) {
		t.Fatal("expected x != nil")
	}
}

func TestLiteralEqualNumeric(t *testing.T) {
	if !literalEqual(int(42), float64(42)) {
		t.Fatal("expected int(42) == float64(42)")
	}
	if !literalEqual(int64(42), int32(42)) {
		t.Fatal("expected int64(42) == int32(42)")
	}
}

func TestLiteralEqualNonNumeric(t *testing.T) {
	if !literalEqual("hello", "hello") {
		t.Fatal("expected 'hello' == 'hello'")
	}
	if literalEqual("hello", "world") {
		t.Fatal("expected 'hello' != 'world'")
	}
	if !literalEqual(true, true) {
		t.Fatal("expected true == true")
	}
	if literalEqual(true, false) {
		t.Fatal("expected true != false")
	}
}

// --- Enum tests ---

func TestEnumSchemaValid(t *testing.T) {
	s := Enum("red", "green", "blue")
	for _, v := range []string{"red", "green", "blue"} {
		r := s.SafeParse(v)
		if !r.Success {
			t.Fatalf("expected success for %q", v)
		}
	}
}

func TestEnumSchemaInvalid(t *testing.T) {
	s := Enum("red", "green", "blue")
	r := s.SafeParse("yellow")
	if r.Success {
		t.Fatal("expected failure")
	}
	if r.Issues[0].Code != IssueInvalidLiteral {
		t.Fatalf("expected invalid_literal, got %s", r.Issues[0].Code)
	}
}

func TestEnumSchemaNumeric(t *testing.T) {
	s := Enum(1.0, 2.0, 3.0)
	if r := s.SafeParse(2.0); !r.Success {
		t.Fatal("expected success for numeric enum")
	}
	if r := s.SafeParse(4.0); r.Success {
		t.Fatal("expected failure for non-member")
	}
}

func TestEnumSchemaDefault(t *testing.T) {
	s := Enum("a", "b").Default("a")
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success with default")
	}
	if r.Data != "a" {
		t.Fatalf("expected 'a', got %v", r.Data)
	}
}

func TestEnumSchemaParse(t *testing.T) {
	s := Enum("x", "y")
	v, err := s.Parse("x")
	if err != nil {
		t.Fatal(err)
	}
	if v != "x" {
		t.Fatalf("expected 'x', got %v", v)
	}
	_, err = s.Parse("z")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnumSchemaToNode(t *testing.T) {
	s := Enum("a", "b", "c")
	node := s.ToNode()
	if node["kind"] != "enum" {
		t.Fatal("expected kind=enum")
	}
	vals, ok := node["values"].([]any)
	if !ok {
		t.Fatal("expected values to be []any")
	}
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
}

// --- Bool tests ---

func TestBoolSchemaValid(t *testing.T) {
	s := Bool()
	if r := s.SafeParse(true); !r.Success || r.Data != true {
		t.Fatal("expected success for true")
	}
	if r := s.SafeParse(false); !r.Success || r.Data != false {
		t.Fatal("expected success for false")
	}
}

func TestBoolSchemaInvalid(t *testing.T) {
	s := Bool()
	tests := []any{"true", 1, nil, []any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestBoolSchemaDefault(t *testing.T) {
	s := Bool().Default(false)
	r := s.SafeParse(nil)
	if !r.Success || r.Data != false {
		t.Fatal("expected false default")
	}
}

func TestBoolSchemaCoerce(t *testing.T) {
	s := Bool().Coerce(CoerceToBool)
	r := s.SafeParse("true")
	if !r.Success || r.Data != true {
		t.Fatal("expected coercion success")
	}
}

func TestBoolSchemaParse(t *testing.T) {
	s := Bool()
	v, err := s.Parse(true)
	if err != nil || v != true {
		t.Fatal("expected success")
	}
	_, err = s.Parse("true")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolSchemaToNode(t *testing.T) {
	s := Bool()
	node := s.ToNode()
	if node["kind"] != "bool" {
		t.Fatal("expected kind=bool")
	}
}

func TestBoolSchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := Bool().Coerce(CoerceToBool).Default(true)
	node := s.ToNode()
	if _, ok := node["coerce"]; !ok {
		t.Fatal("expected coerce")
	}
	if node["default"] != true {
		t.Fatal("expected default=true")
	}
}

// --- Null tests ---

func TestNullSchemaValid(t *testing.T) {
	s := Null()
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
}

func TestNullSchemaInvalid(t *testing.T) {
	s := Null()
	tests := []any{"null", 0, false, []any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestNullSchemaParse(t *testing.T) {
	s := Null()
	v, err := s.Parse(nil)
	if err != nil || v != nil {
		t.Fatal("expected success")
	}
	_, err = s.Parse("null")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNullSchemaToNode(t *testing.T) {
	s := Null()
	node := s.ToNode()
	if node["kind"] != "null" {
		t.Fatal("expected kind=null")
	}
}

// --- Ref tests ---

func TestRefSchemaResolved(t *testing.T) {
	ref := newRefSchema("#/definitions/Name")
	ref.Resolve(String().MinLength(1))

	r := ref.SafeParse("Alice")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = ref.SafeParse("")
	if r.Success {
		t.Fatal("expected failure for empty string")
	}
}

func TestRefSchemaUnresolved(t *testing.T) {
	ref := newRefSchema("#/definitions/Missing")
	r := ref.SafeParse("anything")
	if r.Success {
		t.Fatal("expected failure for unresolved ref")
	}
	if r.Issues[0].Code != IssueUnsupportedSchemaKind {
		t.Fatalf("expected unsupported_schema_kind, got %s", r.Issues[0].Code)
	}
}

func TestRefSchemaParse(t *testing.T) {
	ref := newRefSchema("#/definitions/Name")
	ref.Resolve(String())
	v, err := ref.Parse("hello")
	if err != nil || v != "hello" {
		t.Fatal("expected success")
	}

	ref2 := newRefSchema("#/definitions/Missing")
	_, err = ref2.Parse("anything")
	if err == nil {
		t.Fatal("expected error for unresolved ref")
	}
}

func TestRefSchemaToNode(t *testing.T) {
	ref := newRefSchema("#/definitions/Name")
	node := ref.ToNode()
	if node["kind"] != "ref" {
		t.Fatal("expected kind=ref")
	}
	if node["ref"] != "#/definitions/Name" {
		t.Fatalf("expected ref=#/definitions/Name, got %v", node["ref"])
	}
}

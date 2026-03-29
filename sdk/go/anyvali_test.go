package anyvali

import (
	"encoding/json"
	"math"
	"testing"
)

func TestStringSchema(t *testing.T) {
	s := String()
	result := s.SafeParse("hello")
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", result.Data)
	}

	result = s.SafeParse(42)
	if result.Success {
		t.Fatal("expected failure for non-string")
	}
	if result.Issues[0].Code != IssueInvalidType {
		t.Fatalf("expected invalid_type, got %s", result.Issues[0].Code)
	}
}

func TestStringConstraints(t *testing.T) {
	s := String().MinLength(2).MaxLength(5)
	if r := s.SafeParse("a"); r.Success {
		t.Fatal("expected failure for too short")
	}
	if r := s.SafeParse("abcdef"); r.Success {
		t.Fatal("expected failure for too long")
	}
	if r := s.SafeParse("abc"); !r.Success {
		t.Fatal("expected success for valid length")
	}
}

func TestStringPattern(t *testing.T) {
	s := String().Pattern(`^\d+$`)
	if r := s.SafeParse("123"); !r.Success {
		t.Fatal("expected success for digits")
	}
	if r := s.SafeParse("abc"); r.Success {
		t.Fatal("expected failure for non-digits")
	}
}

func TestStringStartsEndsWith(t *testing.T) {
	s := String().StartsWith("hello").EndsWith("world")
	if r := s.SafeParse("hello world"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("hi world"); r.Success {
		t.Fatal("expected failure for wrong prefix")
	}
}

func TestStringIncludes(t *testing.T) {
	s := String().Includes("bar")
	if r := s.SafeParse("foobarbaz"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("foobaz"); r.Success {
		t.Fatal("expected failure")
	}
}

func TestStringFormat(t *testing.T) {
	tests := []struct {
		format string
		valid  string
		invalid string
	}{
		{"email", "test@example.com", "not-an-email"},
		{"url", "https://example.com", "not-a-url"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "not-a-uuid"},
		{"ipv4", "192.168.1.1", "999.999.999.999"},
		{"ipv6", "::1", "not-ipv6"},
		{"date", "2024-01-15", "2024-13-01"},
		{"date-time", "2024-01-15T10:30:00Z", "not-a-datetime"},
	}
	for _, tt := range tests {
		s := String().Format(tt.format)
		if r := s.SafeParse(tt.valid); !r.Success {
			t.Errorf("format %s: expected %q to be valid", tt.format, tt.valid)
		}
		if r := s.SafeParse(tt.invalid); r.Success {
			t.Errorf("format %s: expected %q to be invalid", tt.format, tt.invalid)
		}
	}
}

func TestNumberSchema(t *testing.T) {
	s := Number()
	result := s.SafeParse(3.14)
	if !result.Success {
		t.Fatal("expected success")
	}

	result = s.SafeParse("not a number")
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestNumberConstraints(t *testing.T) {
	s := Number().Min(0).Max(100)
	if r := s.SafeParse(-1.0); r.Success {
		t.Fatal("expected failure for < min")
	}
	if r := s.SafeParse(101.0); r.Success {
		t.Fatal("expected failure for > max")
	}
	if r := s.SafeParse(50.0); !r.Success {
		t.Fatal("expected success")
	}
}

func TestNumberExclusiveMinMax(t *testing.T) {
	s := Number().ExclusiveMin(0).ExclusiveMax(10)
	if r := s.SafeParse(0.0); r.Success {
		t.Fatal("expected failure for == exclusiveMin")
	}
	if r := s.SafeParse(10.0); r.Success {
		t.Fatal("expected failure for == exclusiveMax")
	}
	if r := s.SafeParse(5.0); !r.Success {
		t.Fatal("expected success")
	}
}

func TestNumberMultipleOf(t *testing.T) {
	s := Number().MultipleOf(3)
	if r := s.SafeParse(9.0); !r.Success {
		t.Fatal("expected success for 9 % 3")
	}
	if r := s.SafeParse(10.0); r.Success {
		t.Fatal("expected failure for 10 % 3")
	}
}

func TestIntSchema(t *testing.T) {
	s := Int()
	if r := s.SafeParse(int64(42)); !r.Success {
		t.Fatal("expected success for int64")
	}
	if r := s.SafeParse(42); !r.Success {
		t.Fatal("expected success for int")
	}
	if r := s.SafeParse("not int"); r.Success {
		t.Fatal("expected failure for string")
	}
}

func TestInt8Range(t *testing.T) {
	s := Int8()
	if r := s.SafeParse(int64(127)); !r.Success {
		t.Fatal("expected success for max int8")
	}
	if r := s.SafeParse(int64(128)); r.Success {
		t.Fatal("expected failure for overflow int8")
	}
	if r := s.SafeParse(int64(-128)); !r.Success {
		t.Fatal("expected success for min int8")
	}
	if r := s.SafeParse(int64(-129)); r.Success {
		t.Fatal("expected failure for underflow int8")
	}
}

func TestUint8Range(t *testing.T) {
	s := Uint8()
	if r := s.SafeParse(int64(255)); !r.Success {
		t.Fatal("expected success for max uint8")
	}
	if r := s.SafeParse(int64(256)); r.Success {
		t.Fatal("expected failure for overflow uint8")
	}
	if r := s.SafeParse(int64(-1)); r.Success {
		t.Fatal("expected failure for negative uint8")
	}
}

func TestFloat32Range(t *testing.T) {
	s := Float32()
	if r := s.SafeParse(1.5); !r.Success {
		t.Fatal("expected success for valid float32")
	}
}

func TestBoolSchema(t *testing.T) {
	s := Bool()
	if r := s.SafeParse(true); !r.Success {
		t.Fatal("expected success for true")
	}
	if r := s.SafeParse(false); !r.Success {
		t.Fatal("expected success for false")
	}
	if r := s.SafeParse("true"); r.Success {
		t.Fatal("expected failure for string")
	}
}

func TestNullSchema(t *testing.T) {
	s := Null()
	if r := s.SafeParse(nil); !r.Success {
		t.Fatal("expected success for nil")
	}
	if r := s.SafeParse("null"); r.Success {
		t.Fatal("expected failure for string")
	}
}

func TestAnySchema(t *testing.T) {
	s := Any()
	for _, v := range []any{"hello", 42, true, nil, []any{1}, map[string]any{"a": 1}} {
		if r := s.SafeParse(v); !r.Success {
			t.Fatalf("expected any to accept %v", v)
		}
	}
}

func TestNeverSchema(t *testing.T) {
	s := Never()
	for _, v := range []any{"hello", 42, true, nil} {
		if r := s.SafeParse(v); r.Success {
			t.Fatalf("expected never to reject %v", v)
		}
	}
}

func TestLiteralSchema(t *testing.T) {
	s := Literal("hello")
	if r := s.SafeParse("hello"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("world"); r.Success {
		t.Fatal("expected failure")
	}

	s2 := Literal(42.0)
	if r := s2.SafeParse(42.0); !r.Success {
		t.Fatal("expected success for numeric literal")
	}
}

func TestEnumSchema(t *testing.T) {
	s := Enum("red", "green", "blue")
	if r := s.SafeParse("red"); !r.Success {
		t.Fatal("expected success for valid enum")
	}
	if r := s.SafeParse("yellow"); r.Success {
		t.Fatal("expected failure for invalid enum")
	}
}

func TestArraySchema(t *testing.T) {
	s := Array(String())
	if r := s.SafeParse([]any{"a", "b"}); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse([]any{"a", 1}); r.Success {
		t.Fatal("expected failure for mixed types")
	}
	if r := s.SafeParse("not array"); r.Success {
		t.Fatal("expected failure for non-array")
	}
}

func TestArrayConstraints(t *testing.T) {
	s := Array(Int()).MinItems(1).MaxItems(3)
	if r := s.SafeParse([]any{}); r.Success {
		t.Fatal("expected failure for empty array")
	}
	if r := s.SafeParse([]any{int64(1), int64(2), int64(3), int64(4)}); r.Success {
		t.Fatal("expected failure for too many items")
	}
	if r := s.SafeParse([]any{int64(1), int64(2)}); !r.Success {
		t.Fatal("expected success")
	}
}

func TestTupleSchema(t *testing.T) {
	s := Tuple(String(), Int())
	if r := s.SafeParse([]any{"hello", int64(42)}); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse([]any{"hello"}); r.Success {
		t.Fatal("expected failure for wrong length")
	}
	if r := s.SafeParse([]any{42, "hello"}); r.Success {
		t.Fatal("expected failure for wrong types")
	}
}

func TestObjectSchema(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
		"age":  Int(),
	})

	input := map[string]any{"name": "Alice", "age": int64(30)}
	if r := s.SafeParse(input); !r.Success {
		t.Fatalf("expected success, got issues: %v", r.Issues)
	}

	// Missing required field
	input2 := map[string]any{"name": "Alice"}
	if r := s.SafeParse(input2); r.Success {
		t.Fatal("expected failure for missing required field")
	}

	// Unknown key rejected by default
	input3 := map[string]any{"name": "Alice", "age": int64(30), "extra": "val"}
	if r := s.SafeParse(input3); r.Success {
		t.Fatal("expected failure for unknown key")
	}
}

func TestObjectUnknownKeysStrip(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Required("name").UnknownKeys(Strip)

	input := map[string]any{"name": "Alice", "extra": "val"}
	r := s.SafeParse(input)
	if !r.Success {
		t.Fatal("expected success with strip mode")
	}
	obj := r.Data.(map[string]any)
	if _, ok := obj["extra"]; ok {
		t.Fatal("expected extra key to be stripped")
	}
}

func TestObjectUnknownKeysAllow(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Required("name").UnknownKeys(Allow)

	input := map[string]any{"name": "Alice", "extra": "val"}
	r := s.SafeParse(input)
	if !r.Success {
		t.Fatal("expected success with allow mode")
	}
	obj := r.Data.(map[string]any)
	if obj["extra"] != "val" {
		t.Fatal("expected extra key to be preserved")
	}
}

func TestRecordSchema(t *testing.T) {
	s := Record(Int())
	input := map[string]any{"a": int64(1), "b": int64(2)}
	if r := s.SafeParse(input); !r.Success {
		t.Fatal("expected success")
	}

	input2 := map[string]any{"a": int64(1), "b": "not int"}
	if r := s.SafeParse(input2); r.Success {
		t.Fatal("expected failure for non-int value")
	}
}

func TestUnionSchema(t *testing.T) {
	s := Union(String(), Int())
	if r := s.SafeParse("hello"); !r.Success {
		t.Fatal("expected success for string")
	}
	if r := s.SafeParse(int64(42)); !r.Success {
		t.Fatal("expected success for int")
	}
	if r := s.SafeParse(true); r.Success {
		t.Fatal("expected failure for bool")
	}
}

func TestIntersectionSchema(t *testing.T) {
	s := Intersection(
		Object(map[string]Schema{"name": String()}).Required("name").UnknownKeys(Allow),
		Object(map[string]Schema{"age": Int()}).Required("age").UnknownKeys(Allow),
	)
	input := map[string]any{"name": "Alice", "age": int64(30)}
	if r := s.SafeParse(input); !r.Success {
		t.Fatalf("expected success, got issues: %v", r.Issues)
	}
}

func TestOptionalSchema(t *testing.T) {
	s := Optional(String())
	if r := s.SafeParse("hello"); !r.Success {
		t.Fatal("expected success for string")
	}
	if r := s.SafeParse(nil); !r.Success {
		t.Fatal("expected success for nil (optional)")
	}
}

func TestNullableSchema(t *testing.T) {
	s := Nullable(String())
	if r := s.SafeParse("hello"); !r.Success {
		t.Fatal("expected success for string")
	}
	if r := s.SafeParse(nil); !r.Success {
		t.Fatal("expected success for nil (nullable)")
	}
	if r := s.SafeParse(42); r.Success {
		t.Fatal("expected failure for int")
	}
}

func TestDefaults(t *testing.T) {
	s := String().Default("fallback")
	// nil is treated as absent in the pipeline, so default applies
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success with default")
	}
	if r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}

	// Present value should not be overwritten
	r = s.SafeParse("actual")
	if r.Data != "actual" {
		t.Fatalf("expected 'actual', got %v", r.Data)
	}
}

func TestCoercionStringToInt(t *testing.T) {
	s := Int().Coerce(CoerceToInt)
	r := s.SafeParse("42")
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	if r.Data != int64(42) {
		t.Fatalf("expected 42, got %v", r.Data)
	}
}

func TestCoercionStringToNumber(t *testing.T) {
	s := Number().Coerce(CoerceToNumber)
	r := s.SafeParse("3.14")
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	if r.Data != 3.14 {
		t.Fatalf("expected 3.14, got %v", r.Data)
	}
}

func TestCoercionStringToBool(t *testing.T) {
	s := Bool().Coerce(CoerceToBool)
	r := s.SafeParse("true")
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	if r.Data != true {
		t.Fatalf("expected true, got %v", r.Data)
	}
}

func TestCoercionTrim(t *testing.T) {
	s := String().Coerce(CoerceTrim)
	r := s.SafeParse("  hello  ")
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestCoercionLower(t *testing.T) {
	s := String().Coerce(CoerceLower)
	r := s.SafeParse("HELLO")
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestCoercionUpper(t *testing.T) {
	s := String().Coerce(CoerceUpper)
	r := s.SafeParse("hello")
	if r.Data != "HELLO" {
		t.Fatalf("expected 'HELLO', got %v", r.Data)
	}
}

func TestCoercionFailed(t *testing.T) {
	s := Int().Coerce(CoerceToInt)
	r := s.SafeParse("not-a-number")
	if r.Success {
		t.Fatal("expected failure for bad coercion")
	}
	if r.Issues[0].Code != IssueCoercionFailed {
		t.Fatalf("expected coercion_failed, got %s", r.Issues[0].Code)
	}
}

func TestParseThrows(t *testing.T) {
	s := String()
	_, err := s.Parse(42)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("expected ValidationError")
	}
	if len(ve.Issues) == 0 {
		t.Fatal("expected issues")
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	original := Object(map[string]Schema{
		"name": String().MinLength(1),
		"age":  Int().Min(0),
		"tags": Array(String()),
	})

	data, err := ExportJSON(original, Portable)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	imported, err := ImportJSON(data)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	input := map[string]any{
		"name": "Alice",
		"age":  int64(30),
		"tags": []any{"go", "dev"},
	}

	r := imported.SafeParse(input)
	if !r.Success {
		t.Fatalf("imported schema validation failed: %v", r.Issues)
	}
}

func TestExportImportNumber(t *testing.T) {
	s := Number().Min(0).Max(100)
	data, err := ExportJSON(s, Portable)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	imported, err := ImportJSON(data)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if r := imported.SafeParse(50.0); !r.Success {
		t.Fatal("expected success")
	}
	if r := imported.SafeParse(150.0); r.Success {
		t.Fatal("expected failure")
	}
}

func TestJsonNumberParsing(t *testing.T) {
	// Simulate JSON-decoded input with json.Number
	s := Int()
	n := json.Number("42")
	r := s.SafeParse(n)
	if !r.Success {
		t.Fatalf("expected success for json.Number, got: %v", r.Issues)
	}
	if r.Data != int64(42) {
		t.Fatalf("expected 42, got %v", r.Data)
	}
}

func TestNumericWidths(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		valid   any
		invalid any
	}{
		{"int8", Int8(), int64(100), int64(200)},
		{"int16", Int16(), int64(30000), int64(40000)},
		{"int32", Int32(), int64(2000000000), int64(3000000000)},
		{"uint8", Uint8(), int64(200), int64(-1)},
		{"uint16", Uint16(), int64(60000), int64(-1)},
		{"uint32", Uint32(), int64(4000000000), int64(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := tt.schema.SafeParse(tt.valid); !r.Success {
				t.Errorf("expected %v to be valid for %s, got issues: %v", tt.valid, tt.name, r.Issues)
			}
			if r := tt.schema.SafeParse(tt.invalid); r.Success {
				t.Errorf("expected %v to be invalid for %s", tt.invalid, tt.name)
			}
		})
	}
}

func TestObjectOptionalField(t *testing.T) {
	s := Object(map[string]Schema{
		"name":  String(),
		"email": Optional(String()),
	})
	input := map[string]any{"name": "Alice"}
	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success with optional field missing: %v", r.Issues)
	}
}

func TestValidationErrorString(t *testing.T) {
	err := &ValidationError{
		Issues: []ValidationIssue{
			{Code: IssueInvalidType, Message: "expected string"},
		},
	}
	s := err.Error()
	if s == "" {
		t.Fatal("expected non-empty error string")
	}
}

func TestRefSchema(t *testing.T) {
	ref := newRefSchema("#/definitions/Name")
	ref.Resolve(String().MinLength(1))

	if r := ref.SafeParse("Alice"); !r.Success {
		t.Fatal("expected success")
	}
	if r := ref.SafeParse(""); r.Success {
		t.Fatal("expected failure for empty string")
	}
}

func TestUnresolvedRef(t *testing.T) {
	ref := newRefSchema("#/definitions/Missing")
	r := ref.SafeParse("anything")
	if r.Success {
		t.Fatal("expected failure for unresolved ref")
	}
}

func TestIntMultipleOf(t *testing.T) {
	s := Int().MultipleOf(5)
	if r := s.SafeParse(int64(15)); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse(int64(13)); r.Success {
		t.Fatal("expected failure")
	}
}

func TestOptionalWithDefault(t *testing.T) {
	s := Optional(String()).Default("default_value")
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "default_value" {
		t.Fatalf("expected 'default_value', got %v", r.Data)
	}
}

func TestNullableWithDefault(t *testing.T) {
	s := Nullable(String()).Default("default_value")
	// null should pass through as null
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != nil {
		t.Fatalf("expected nil, got %v", r.Data)
	}
}

func TestDocumentStructure(t *testing.T) {
	s := Object(map[string]Schema{
		"id":   Int(),
		"name": String(),
	})
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if doc.AnyvaliVersion != "1.0" {
		t.Fatalf("expected version 1.0, got %s", doc.AnyvaliVersion)
	}
	if doc.SchemaVersion != "1" {
		t.Fatalf("expected schema version 1, got %s", doc.SchemaVersion)
	}
	if doc.Root["kind"] != "object" {
		t.Fatalf("expected root kind object, got %v", doc.Root["kind"])
	}
}

func TestIntMaxValues(t *testing.T) {
	s := Int64()
	if r := s.SafeParse(int64(math.MaxInt64)); !r.Success {
		t.Fatal("expected success for max int64")
	}
	if r := s.SafeParse(int64(math.MinInt64)); !r.Success {
		t.Fatal("expected success for min int64")
	}
}

func TestEnumDefault(t *testing.T) {
	s := Enum("a", "b", "c").Default("b")
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success with default")
	}
	if r.Data != "b" {
		t.Fatalf("expected 'b', got %v", r.Data)
	}
}

func TestNestedObjectValidation(t *testing.T) {
	s := Object(map[string]Schema{
		"user": Object(map[string]Schema{
			"name": String(),
			"age":  Int(),
		}),
	})

	input := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"age":  int64(30),
		},
	}

	if r := s.SafeParse(input); !r.Success {
		t.Fatalf("expected success: %v", r.Issues)
	}

	// Nested validation error should have path
	input2 := map[string]any{
		"user": map[string]any{
			"name": 42,
			"age":  int64(30),
		},
	}
	r := s.SafeParse(input2)
	if r.Success {
		t.Fatal("expected failure")
	}
	// Check that path includes both "user" and "name"
	found := false
	for _, issue := range r.Issues {
		if len(issue.Path) >= 2 && issue.Path[0] == "user" && issue.Path[1] == "name" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected path [user, name] in issues: %v", r.Issues)
	}
}

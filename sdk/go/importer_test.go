package anyvali

import (
	"encoding/json"
	"testing"
)

func TestImportNilDocument(t *testing.T) {
	_, err := Import(nil)
	if err == nil {
		t.Fatal("expected error for nil document")
	}
}

func TestImportNilRoot(t *testing.T) {
	doc := &Document{
		AnyvaliVersion: "1.0",
		SchemaVersion:  "1",
		Root:           nil,
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for nil root")
	}
}

func TestImportMissingKind(t *testing.T) {
	doc := &Document{
		AnyvaliVersion: "1.0",
		SchemaVersion:  "1",
		Root:           map[string]any{"not_kind": "string"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing kind")
	}
}

func TestImportUnsupportedKind(t *testing.T) {
	doc := &Document{
		AnyvaliVersion: "1.0",
		SchemaVersion:  "1",
		Root:           map[string]any{"kind": "foobar"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for unsupported kind")
	}
}

func TestImportStringSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":       "string",
			"minLength":  float64(1),
			"maxLength":  float64(100),
			"pattern":    `^\d+$`,
			"startsWith": "a",
			"endsWith":   "z",
			"includes":   "m",
			"format":     "email",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Verify it validates
	r := s.SafeParse("not-matching")
	// Should fail format or pattern
	if r.Success {
		t.Fatal("expected failure for invalid input")
	}
}

func TestImportStringSchemaWithCoercionsAndDefaults(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "string",
			"coerce":  []any{"trim", "lower"},
			"default": "fallback",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Test coercion
	r := s.SafeParse("  HELLO  ")
	if !r.Success || r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
	// Test default
	r = s.SafeParse(nil)
	if !r.Success || r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestImportNumberSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":         "number",
			"min":          float64(0),
			"max":          float64(100),
			"exclusiveMin": float64(0),
			"exclusiveMax": float64(100),
			"multipleOf":   float64(5),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(float64(50))
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	r = s.SafeParse(float64(0))
	if r.Success {
		t.Fatal("expected failure for == exclusiveMin")
	}
}

func TestImportFloat64Schema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "float64"},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(float64(3.14))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportFloat32Schema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "float32", "min": float64(0)},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(float64(1.5))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportIntSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":         "int",
			"min":          float64(0),
			"max":          float64(100),
			"exclusiveMin": float64(0),
			"exclusiveMax": float64(100),
			"multipleOf":   float64(5),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(50))
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestImportInt8Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "int8"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(127))
	if !r.Success {
		t.Fatal("expected success for max int8")
	}
	r = s.SafeParse(int64(128))
	if r.Success {
		t.Fatal("expected failure for int8 overflow")
	}
}

func TestImportInt16Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "int16"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(32767))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportInt32Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "int32"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(2147483647))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportInt64Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "int64"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(42))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportUint8Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "uint8"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(255))
	if !r.Success {
		t.Fatal("expected success for max uint8")
	}
	r = s.SafeParse(int64(256))
	if r.Success {
		t.Fatal("expected failure for uint8 overflow")
	}
}

func TestImportUint16Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "uint16"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(65535))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportUint32Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "uint32"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(4294967295))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportUint64Schema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "uint64"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(42))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportBoolSchema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "bool"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(true)
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportBoolSchemaWithCoerceAndDefault(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "bool",
			"coerce":  []any{"bool"},
			"default": true,
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("true")
	if !r.Success || r.Data != true {
		t.Fatal("expected coerced true")
	}
}

func TestImportNullSchema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "null"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportAnySchema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "any"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("anything")
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportUnknownSchema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "unknown"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(42)
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportNeverSchema(t *testing.T) {
	doc := &Document{Root: map[string]any{"kind": "never"}}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("anything")
	if r.Success {
		t.Fatal("expected failure from never")
	}
}

func TestImportLiteralSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "literal", "value": "hello"},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse("world")
	if r.Success {
		t.Fatal("expected failure")
	}
}

func TestImportLiteralSchemaMissingValue(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "literal"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestImportEnumSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "enum",
			"values": []any{"a", "b", "c"},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("a")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse("d")
	if r.Success {
		t.Fatal("expected failure")
	}
}

func TestImportEnumSchemaMissingValues(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "enum"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing values")
	}
}

func TestImportEnumSchemaWithDefault(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "enum",
			"values":  []any{"a", "b"},
			"default": "a",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(nil)
	if !r.Success || r.Data != "a" {
		t.Fatal("expected default=a")
	}
}

func TestImportArraySchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":     "array",
			"item":     map[string]any{"kind": "string"},
			"minItems": float64(1),
			"maxItems": float64(10),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse([]any{"hello"})
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportArraySchemaMissingItem(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "array"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing item")
	}
}

func TestImportArraySchemaInvalidItem(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "array",
			"item": map[string]any{"kind": "foobar"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid item kind")
	}
}

func TestImportTupleSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "tuple",
			"items": []any{
				map[string]any{"kind": "string"},
				map[string]any{"kind": "int"},
			},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse([]any{"hello", int64(42)})
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestImportTupleSchemaMissingItems(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "tuple"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing items")
	}
}

func TestImportTupleSchemaInvalidItem(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":  "tuple",
			"items": []any{"not-a-map"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid tuple item")
	}
}

func TestImportTupleSchemaItemError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":  "tuple",
			"items": []any{map[string]any{"kind": "foobar"}},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid item kind")
	}
}

func TestImportObjectSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"name": map[string]any{"kind": "string"},
				"age":  map[string]any{"kind": "int"},
			},
			"required":    []any{"name", "age"},
			"unknownKeys": "allow",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(map[string]any{"name": "Alice", "age": int64(30), "extra": true})
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestImportObjectSchemaMissingProperties(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "object"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing properties")
	}
}

func TestImportObjectSchemaInvalidProperty(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"bad": "not-a-map",
			},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid property")
	}
}

func TestImportObjectSchemaPropertyError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"bad": map[string]any{"kind": "foobar"},
			},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid property kind")
	}
}

func TestImportRecordSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":  "record",
			"value": map[string]any{"kind": "int"},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(map[string]any{"a": int64(1)})
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportRecordSchemaMissingValue(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "record"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestImportRecordSchemaValueError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":  "record",
			"value": map[string]any{"kind": "foobar"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid value kind")
	}
}

func TestImportUnionSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "union",
			"schemas": []any{
				map[string]any{"kind": "string"},
				map[string]any{"kind": "int"},
			},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse(int64(42))
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportUnionSchemaMissingSchemas(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "union"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing schemas")
	}
}

func TestImportUnionSchemaInvalidEntry(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "union",
			"schemas": []any{"not-a-map"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid union entry")
	}
}

func TestImportUnionSchemaEntryError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "union",
			"schemas": []any{map[string]any{"kind": "foobar"}},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid entry kind")
	}
}

func TestImportIntersectionSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "intersection",
			"schemas": []any{
				map[string]any{
					"kind":        "object",
					"properties":  map[string]any{"a": map[string]any{"kind": "string"}},
					"required":    []any{"a"},
					"unknownKeys": "allow",
				},
				map[string]any{
					"kind":        "object",
					"properties":  map[string]any{"b": map[string]any{"kind": "int"}},
					"required":    []any{"b"},
					"unknownKeys": "allow",
				},
			},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(map[string]any{"a": "x", "b": int64(1)})
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestImportIntersectionSchemaMissingSchemas(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "intersection"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing schemas")
	}
}

func TestImportIntersectionSchemaInvalidEntry(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "intersection",
			"schemas": []any{"not-a-map"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid entry")
	}
}

func TestImportIntersectionSchemaEntryError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "intersection",
			"schemas": []any{map[string]any{"kind": "foobar"}},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid entry kind")
	}
}

func TestImportOptionalSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "optional",
			"schema": map[string]any{"kind": "string"},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
	r = s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success for string")
	}
}

func TestImportOptionalSchemaWithDefault(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "optional",
			"schema":  map[string]any{"kind": "string"},
			"default": "fallback",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(nil)
	if !r.Success || r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestImportOptionalSchemaMissingSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "optional"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

func TestImportOptionalSchemaInnerError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "optional",
			"schema": map[string]any{"kind": "foobar"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid inner kind")
	}
}

func TestImportNullableSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "nullable",
			"schema": map[string]any{"kind": "string"},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success for nil")
	}
	r = s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success for string")
	}
}

func TestImportNullableSchemaWithDefault(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "nullable",
			"schema":  map[string]any{"kind": "int"},
			"default": float64(0),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	// absent should trigger default
	r := s.SafeParse(absentValue)
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestImportNullableSchemaMissingSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "nullable"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

func TestImportNullableSchemaInnerError(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "nullable",
			"schema": map[string]any{"kind": "foobar"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid inner kind")
	}
}

func TestImportRefSchema(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "ref",
			"ref":  "#/definitions/Name",
		},
		Definitions: map[string]map[string]any{
			"Name": {"kind": "string", "minLength": float64(1)},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("Alice")
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse("")
	if r.Success {
		t.Fatal("expected failure for empty string")
	}
}

func TestImportRefSchemaMissingRef(t *testing.T) {
	doc := &Document{
		Root: map[string]any{"kind": "ref"},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
}

func TestImportRefSchemaUnresolved(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "ref",
			"ref":  "#/definitions/Missing",
		},
		Definitions: map[string]map[string]any{},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Should be unresolved, so parse should fail
	r := s.SafeParse("anything")
	if r.Success {
		t.Fatal("expected failure for unresolved ref")
	}
}

func TestImportRefSchemaInvalidDef(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind": "ref",
			"ref":  "#/definitions/Bad",
		},
		Definitions: map[string]map[string]any{
			"Bad": {"kind": "foobar"},
		},
	}
	_, err := Import(doc)
	if err == nil {
		t.Fatal("expected error for invalid definition")
	}
}

func TestImportRefSchemaShortRef(t *testing.T) {
	// Ref that doesn't start with #/definitions/
	doc := &Document{
		Root: map[string]any{
			"kind": "ref",
			"ref":  "short",
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	// Should be unresolved
	r := s.SafeParse("anything")
	if r.Success {
		t.Fatal("expected failure for unresolved short ref")
	}
}

func TestImportJSONValid(t *testing.T) {
	jsonData := `{
		"anyvaliVersion": "1.0",
		"schemaVersion": "1",
		"root": {"kind": "string"}
	}`
	s, err := ImportJSON([]byte(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("hello")
	if !r.Success {
		t.Fatal("expected success")
	}
}

func TestImportJSONInvalid(t *testing.T) {
	_, err := ImportJSON([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestImportCoercions(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{
		"coerce": []any{"trim", "lower", "upper"},
	}
	importCoercions(b, node)
	if len(b.coercions) != 3 {
		t.Fatalf("expected 3 coercions, got %d", len(b.coercions))
	}
	if b.coercions[0] != CoerceTrim {
		t.Fatal("expected first coercion to be trim")
	}
}

func TestImportCoercionsNoCoerce(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{}
	importCoercions(b, node)
	if len(b.coercions) != 0 {
		t.Fatal("expected no coercions")
	}
}

func TestImportDefault(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{"default": "fallback"}
	importDefault(b, node)
	if !b.hasDefault {
		t.Fatal("expected hasDefault=true")
	}
	if b.defaultValue != "fallback" {
		t.Fatalf("expected 'fallback', got %v", b.defaultValue)
	}
}

func TestImportDefaultMissing(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{}
	importDefault(b, node)
	if b.hasDefault {
		t.Fatal("expected hasDefault=false")
	}
}

func TestGetFloat(t *testing.T) {
	node := map[string]any{"x": float64(3.14)}
	f, ok := getFloat(node, "x")
	if !ok || f != 3.14 {
		t.Fatal("expected 3.14")
	}
	_, ok = getFloat(node, "missing")
	if ok {
		t.Fatal("expected not ok for missing key")
	}
}

func TestGetInt(t *testing.T) {
	node := map[string]any{"x": float64(42)}
	i, ok := getInt(node, "x")
	if !ok || i != 42 {
		t.Fatal("expected 42")
	}
	_, ok = getInt(node, "missing")
	if ok {
		t.Fatal("expected not ok for missing key")
	}
}

func TestGetIntNonNumeric(t *testing.T) {
	node := map[string]any{"x": "not-a-number"}
	_, ok := getInt(node, "x")
	if ok {
		t.Fatal("expected not ok for non-numeric value")
	}
}

func TestExportImportRoundTripAllTypes(t *testing.T) {
	schemas := []struct {
		name   string
		schema Schema
		input  any
	}{
		{"string", String().MinLength(1), "hello"},
		{"number", Number().Min(0), float64(42)},
		{"float64", Float64(), float64(3.14)},
		{"float32", Float32(), float64(1.5)},
		{"int", Int().Min(0), int64(42)},
		{"int8", Int8(), int64(100)},
		{"int16", Int16(), int64(1000)},
		{"int32", Int32(), int64(100000)},
		{"int64", Int64(), int64(42)},
		{"uint8", Uint8(), int64(200)},
		{"uint16", Uint16(), int64(60000)},
		{"uint32", Uint32(), int64(100000)},
		{"uint64", Uint64(), int64(42)},
		{"bool", Bool(), true},
		{"null", Null(), nil},
		{"any", Any(), "anything"},
		{"unknown", Unknown(), 42},
		{"never-fail", Never(), nil},
		{"literal", Literal("x"), "x"},
		{"enum", Enum("a", "b"), "a"},
		{"array", Array(String()), []any{"a", "b"}},
		{"tuple", Tuple(String(), Int()), []any{"a", int64(1)}},
		{"object", Object(map[string]Schema{"k": String()}).Required("k"), map[string]any{"k": "v"}},
		{"record", Record(Int()), map[string]any{"a": int64(1)}},
		{"union", Union(String(), Int()), "hello"},
		{"optional", Optional(String()), "hello"},
		{"nullable", Nullable(String()), nil},
	}

	for _, tt := range schemas {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ExportJSON(tt.schema, Portable)
			if err != nil {
				t.Fatalf("export failed: %v", err)
			}

			imported, err := ImportJSON(data)
			if err != nil {
				t.Fatalf("import failed: %v", err)
			}

			result := imported.SafeParse(tt.input)
			origResult := tt.schema.SafeParse(tt.input)

			if result.Success != origResult.Success {
				t.Errorf("round-trip mismatch: original=%v, imported=%v", origResult.Success, result.Success)
				if !result.Success {
					t.Logf("issues: %v", result.Issues)
				}
			}

			if result.Success && origResult.Success {
				expectedJSON, _ := json.Marshal(origResult.Data)
				actualJSON, _ := json.Marshal(result.Data)
				if string(expectedJSON) != string(actualJSON) {
					t.Errorf("data mismatch: expected=%s, got=%s", expectedJSON, actualJSON)
				}
			}
		})
	}
}

func TestImportObjectSchemaWithUnknownKeysMode(t *testing.T) {
	modes := []string{"reject", "strip", "allow"}
	for _, mode := range modes {
		doc := &Document{
			Root: map[string]any{
				"kind": "object",
				"properties": map[string]any{
					"name": map[string]any{"kind": "string"},
				},
				"required":    []any{"name"},
				"unknownKeys": mode,
			},
		}
		_, err := Import(doc)
		if err != nil {
			t.Fatalf("failed to import object with unknownKeys=%s: %v", mode, err)
		}
	}
}

func TestImportNumberSchemaWithCoerce(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":   "number",
			"coerce": []any{"number"},
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("3.14")
	if !r.Success || r.Data != 3.14 {
		t.Fatalf("expected 3.14, got %v (success=%v)", r.Data, r.Success)
	}
}

func TestImportIntSchemaWithCoerceAndDefault(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":    "int",
			"coerce":  []any{"int"},
			"default": float64(0),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse("42")
	if !r.Success {
		t.Fatalf("expected success for coercion, got: %v", r.Issues)
	}
}

func TestImportUintSchemaWithConstraints(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":       "uint32",
			"min":        float64(10),
			"max":        float64(200),
			"multipleOf": float64(10),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(int64(100))
	if !r.Success {
		t.Fatal("expected success")
	}
	r = s.SafeParse(int64(5))
	if r.Success {
		t.Fatal("expected failure for < min")
	}
}

func TestImportFloat32SchemaWithConstraints(t *testing.T) {
	doc := &Document{
		Root: map[string]any{
			"kind":         "float32",
			"min":          float64(0),
			"max":          float64(100),
			"exclusiveMin": float64(-1),
			"exclusiveMax": float64(101),
			"multipleOf":   float64(0.5),
		},
	}
	s, err := Import(doc)
	if err != nil {
		t.Fatal(err)
	}
	r := s.SafeParse(float64(50))
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

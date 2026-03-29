package anyvali

import (
	"encoding/json"
	"testing"
)

func TestExportStringSchema(t *testing.T) {
	s := String().MinLength(1).MaxLength(100)
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.AnyvaliVersion != "1.0" {
		t.Fatal("wrong version")
	}
	if doc.SchemaVersion != "1" {
		t.Fatal("wrong schema version")
	}
	if doc.Root["kind"] != "string" {
		t.Fatal("expected root kind=string")
	}
}

func TestExportNumberSchema(t *testing.T) {
	s := Number().Min(0).Max(100)
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "number" {
		t.Fatal("expected kind=number")
	}
	if doc.Root["min"] != float64(0) {
		t.Fatal("expected min=0")
	}
}

func TestExportFloat64Schema(t *testing.T) {
	s := Float64()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "float64" {
		t.Fatal("expected kind=float64")
	}
}

func TestExportFloat32Schema(t *testing.T) {
	s := Float32()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "float32" {
		t.Fatal("expected kind=float32")
	}
}

func TestExportIntSchema(t *testing.T) {
	s := Int().Min(0)
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "int" {
		t.Fatal("expected kind=int")
	}
}

func TestExportInt8Schema(t *testing.T) {
	s := Int8()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "int8" {
		t.Fatal("expected kind=int8")
	}
}

func TestExportInt16Schema(t *testing.T) {
	s := Int16()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "int16" {
		t.Fatal("expected kind=int16")
	}
}

func TestExportInt32Schema(t *testing.T) {
	s := Int32()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "int32" {
		t.Fatal("expected kind=int32")
	}
}

func TestExportInt64Schema(t *testing.T) {
	s := Int64()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "int64" {
		t.Fatal("expected kind=int64")
	}
}

func TestExportUint8Schema(t *testing.T) {
	s := Uint8()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "uint8" {
		t.Fatal("expected kind=uint8")
	}
}

func TestExportUint16Schema(t *testing.T) {
	s := Uint16()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "uint16" {
		t.Fatal("expected kind=uint16")
	}
}

func TestExportUint32Schema(t *testing.T) {
	s := Uint32()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "uint32" {
		t.Fatal("expected kind=uint32")
	}
}

func TestExportUint64Schema(t *testing.T) {
	s := Uint64()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "uint64" {
		t.Fatal("expected kind=uint64")
	}
}

func TestExportBoolSchema(t *testing.T) {
	s := Bool()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "bool" {
		t.Fatal("expected kind=bool")
	}
}

func TestExportNullSchema(t *testing.T) {
	s := Null()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "null" {
		t.Fatal("expected kind=null")
	}
}

func TestExportAnySchema(t *testing.T) {
	s := Any()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "any" {
		t.Fatal("expected kind=any")
	}
}

func TestExportUnknownSchema(t *testing.T) {
	s := Unknown()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "unknown" {
		t.Fatal("expected kind=unknown")
	}
}

func TestExportNeverSchema(t *testing.T) {
	s := Never()
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "never" {
		t.Fatal("expected kind=never")
	}
}

func TestExportLiteralSchema(t *testing.T) {
	s := Literal("hello")
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "literal" {
		t.Fatal("expected kind=literal")
	}
	if doc.Root["value"] != "hello" {
		t.Fatal("expected value=hello")
	}
}

func TestExportEnumSchema(t *testing.T) {
	s := Enum("a", "b", "c")
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "enum" {
		t.Fatal("expected kind=enum")
	}
}

func TestExportArraySchema(t *testing.T) {
	s := Array(String()).MinItems(1).MaxItems(10)
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "array" {
		t.Fatal("expected kind=array")
	}
	if doc.Root["minItems"] != 1 {
		t.Fatal("expected minItems=1")
	}
}

func TestExportTupleSchema(t *testing.T) {
	s := Tuple(String(), Int())
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "tuple" {
		t.Fatal("expected kind=tuple")
	}
}

func TestExportObjectSchema(t *testing.T) {
	s := Object(map[string]Schema{"name": String()})
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "object" {
		t.Fatal("expected kind=object")
	}
}

func TestExportRecordSchema(t *testing.T) {
	s := Record(Int())
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "record" {
		t.Fatal("expected kind=record")
	}
}

func TestExportUnionSchema(t *testing.T) {
	s := Union(String(), Int())
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "union" {
		t.Fatal("expected kind=union")
	}
}

func TestExportIntersectionSchema(t *testing.T) {
	s := Intersection(
		Object(map[string]Schema{"a": String()}).UnknownKeys(Allow),
		Object(map[string]Schema{"b": Int()}).UnknownKeys(Allow),
	)
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "intersection" {
		t.Fatal("expected kind=intersection")
	}
}

func TestExportOptionalSchema(t *testing.T) {
	s := Optional(String())
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "optional" {
		t.Fatal("expected kind=optional")
	}
}

func TestExportNullableSchema(t *testing.T) {
	s := Nullable(String())
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "nullable" {
		t.Fatal("expected kind=nullable")
	}
}

func TestExportRefSchema(t *testing.T) {
	ref := newRefSchema("#/definitions/Name")
	ref.Resolve(String())
	doc, err := Export(ref, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root["kind"] != "ref" {
		t.Fatal("expected kind=ref")
	}
	if doc.Root["ref"] != "#/definitions/Name" {
		t.Fatal("expected ref value")
	}
}

func TestExportJSONAllSchemas(t *testing.T) {
	schemas := []Schema{
		String(),
		Number(),
		Float64(),
		Float32(),
		Int(),
		Int8(),
		Int16(),
		Int32(),
		Int64(),
		Uint8(),
		Uint16(),
		Uint32(),
		Uint64(),
		Bool(),
		Null(),
		Any(),
		Unknown(),
		Never(),
		Literal("x"),
		Enum("a", "b"),
		Array(String()),
		Tuple(String(), Int()),
		Object(map[string]Schema{"k": String()}),
		Record(Int()),
		Union(String(), Int()),
		Intersection(Number(), Number()),
		Optional(String()),
		Nullable(String()),
	}
	for _, s := range schemas {
		data, err := ExportJSON(s, Portable)
		if err != nil {
			t.Fatalf("ExportJSON failed for %v: %v", s.ToNode()["kind"], err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty JSON")
		}
		// Verify it's valid JSON
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("invalid JSON output: %v", err)
		}
	}
}

func TestExportExtendedMode(t *testing.T) {
	s := String()
	doc, err := Export(s, Extended)
	if err != nil {
		t.Fatal(err)
	}
	if doc.AnyvaliVersion != "1.0" {
		t.Fatal("wrong version")
	}
}

func TestExportDocumentStructure(t *testing.T) {
	s := Object(map[string]Schema{"id": Int()})
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Definitions == nil {
		t.Fatal("expected definitions map to be initialized")
	}
	if doc.Extensions == nil {
		t.Fatal("expected extensions map to be initialized")
	}
}

func TestCollectDefinitionsNilNode(t *testing.T) {
	defs := make(map[string]map[string]any)
	collectDefinitions(nil, defs)
	if len(defs) != 0 {
		t.Fatal("expected no definitions for nil node")
	}
}

func TestCollectDefinitionsNestedArray(t *testing.T) {
	// Test that collectDefinitions walks into arrays of nodes
	node := map[string]any{
		"kind": "union",
		"schemas": []any{
			map[string]any{"kind": "string"},
			map[string]any{"kind": "int"},
		},
	}
	defs := make(map[string]map[string]any)
	collectDefinitions(node, defs)
	// Should not panic and should complete
}

func TestExportWithCoercionsAndDefaults(t *testing.T) {
	s := String().Coerce(CoerceTrim).Default("x")
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	coerce, ok := doc.Root["coerce"].([]any)
	if !ok {
		t.Fatal("expected coerce in exported root")
	}
	if len(coerce) != 1 || coerce[0] != "trim" {
		t.Fatalf("expected [trim], got %v", coerce)
	}
	if doc.Root["default"] != "x" {
		t.Fatal("expected default=x")
	}
}

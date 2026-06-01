package anyvali

import (
	"strings"
	"testing"
)

func TestDescribeAndMetadataExportAcrossSchemas(t *testing.T) {
	schemas := []Schema{
		String().Describe("string field").Metadata(map[string]any{"ui": "input"}),
		Number().Describe("number field").Metadata(map[string]any{"ui": "number"}),
		Float32().Describe("float32 field").Metadata(map[string]any{"ui": "float"}),
		Int().Describe("int field").Metadata(map[string]any{"ui": "integer"}),
		Bool().Describe("bool field").Metadata(map[string]any{"ui": "checkbox"}),
		Null().Describe("null field").Metadata(map[string]any{"ui": "none"}),
		Any().Describe("any field").Metadata(map[string]any{"ui": "any"}),
		Unknown().Describe("unknown field").Metadata(map[string]any{"ui": "unknown"}),
		Never().Describe("never field").Metadata(map[string]any{"ui": "never"}),
		Literal("x").Describe("literal field").Metadata(map[string]any{"ui": "literal"}),
		Enum("a", "b").Describe("enum field").Metadata(map[string]any{"ui": "select"}),
		Array(String()).Describe("array field").Metadata(map[string]any{"ui": "list"}),
		Tuple(String(), Int()).Describe("tuple field").Metadata(map[string]any{"ui": "tuple"}),
		Object(map[string]Schema{"name": String()}).Describe("object field").Metadata(map[string]any{"ui": "form"}),
		Record(String()).Describe("record field").Metadata(map[string]any{"ui": "map"}),
		Union(String(), Int()).Describe("union field").Metadata(map[string]any{"ui": "choice"}),
		Intersection(Object(map[string]Schema{"name": String()}), Object(map[string]Schema{"age": Int()})).Describe("intersection field").Metadata(map[string]any{"ui": "merge"}),
		Optional(String()).Describe("optional field").Metadata(map[string]any{"ui": "optional"}),
		Nullable(String()).Describe("nullable field").Metadata(map[string]any{"ui": "nullable"}),
		newRefSchema("#/definitions/User").Describe("ref field").Metadata(map[string]any{"ui": "ref"}),
	}

	for _, schema := range schemas {
		node := schema.ToNode()
		meta, ok := node["metadata"].(map[string]any)
		if !ok {
			t.Fatalf("%v did not export metadata", node["kind"])
		}
		if meta["description"] == "" {
			t.Fatalf("%v did not export description", node["kind"])
		}
		if meta["ui"] == "" {
			t.Fatalf("%v did not export custom metadata", node["kind"])
		}
	}
}

func TestDescribeOptionsAndMetadataReplace(t *testing.T) {
	schema := String().
		Describe("email", DescribeOpts{
			Title:             "Email",
			Deprecated:        true,
			DeprecatedMessage: "use contactEmail",
			NotStable:         true,
			Since:             "0.3.0",
			Sensitive:         true,
			Readonly:          true,
			Examples:          []any{"team@example.com"},
		}).
		Metadata(map[string]any{"group": "contact"}).
		Metadata(map[string]any{"widget": "email"}, MetadataOpts{Replace: true})

	meta := schema.ToNode()["metadata"].(map[string]any)
	if meta["description"] != "email" || meta["title"] != "Email" {
		t.Fatalf("reserved describe metadata was not preserved: %#v", meta)
	}
	if meta["deprecatedMessage"] != "use contactEmail" || meta["widget"] != "email" {
		t.Fatalf("metadata replace did not keep expected fields: %#v", meta)
	}
	if _, ok := meta["group"]; ok {
		t.Fatalf("metadata replace kept non-reserved field: %#v", meta)
	}
}

func TestDescribeRejectsInvalidReservedCombinations(t *testing.T) {
	assertPanicContains(t, "deprecatedMessage requires deprecated", func() {
		String().Describe("x", DescribeOpts{DeprecatedMessage: "old"})
	})
	assertPanicContains(t, "readonly and writeonly", func() {
		String().Describe("x", DescribeOpts{Readonly: true, Writeonly: true})
	})
	assertPanicContains(t, "reserved key", func() {
		String().Metadata(map[string]any{"description": "use Describe"})
	})
}

func TestTypedParseHelpers(t *testing.T) {
	value, err := TypedParse[string](String(), "ok")
	if err != nil || value != "ok" {
		t.Fatalf("TypedParse returned %q, %v", value, err)
	}

	if _, err := TypedParse[string](Int(), 1); err == nil {
		t.Fatal("TypedParse should fail when the parsed value cannot be cast to T")
	}

	if _, err := TypedParse[string](String(), 1); err == nil {
		t.Fatal("TypedParse should return validation errors")
	}

	success := TypedSafeParse[string](String(), "ok")
	if !success.Success || success.Data != "ok" {
		t.Fatalf("TypedSafeParse success mismatch: %#v", success)
	}

	validationFailure := TypedSafeParse[string](String(), 1)
	if validationFailure.Success || len(validationFailure.Issues) == 0 {
		t.Fatalf("TypedSafeParse should return validation issues: %#v", validationFailure)
	}

	typeFailure := TypedSafeParse[string](Int(), 1)
	if typeFailure.Success || len(typeFailure.Issues) != 1 || typeFailure.Issues[0].Code != "type_assertion_failed" {
		t.Fatalf("TypedSafeParse should report type assertion failure: %#v", typeFailure)
	}
}

func assertPanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if !strings.Contains(got.(string), want) {
			t.Fatalf("panic %q does not contain %q", got, want)
		}
	}()
	fn()
}

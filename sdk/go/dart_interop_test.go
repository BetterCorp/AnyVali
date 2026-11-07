package anyvali

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestDartPortableCoercionsImport covers the canonical document emitted by Dart
// for native integer, floating-point, boolean, and string coercion builders.
func TestDartPortableCoercionsImport(t *testing.T) {
	document := []byte(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"object","properties":{"count":{"kind":"int","coerce":["string->int"]},"name":{"kind":"string","coerce":["trim","lower"]},"active":{"kind":"bool","coerce":["string->bool"]},"ratio":{"kind":"number","coerce":["string->number"]}},"required":["count","name","active","ratio"],"unknownKeys":"strip"},"definitions":{},"extensions":{}}`)
	schema, err := ImportJSON(document)
	if err != nil {
		t.Fatal(err)
	}
	result := schema.SafeParse(map[string]any{"count": "7", "name": " Alice ", "active": "true", "ratio": "2.5"})
	if !result.Success {
		t.Fatalf("coercions failed: %+v", result.Issues)
	}
	encoded, err := json.Marshal(result.Data)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"active":true,"count":7,"name":"alice","ratio":2.5}` {
		t.Fatalf("unexpected result: %s", encoded)
	}
}

// TestDartUnknownKeyInheritanceContract anchors the shared precedence used by
// Dart: restrictive parents propagate unchanged; allow defers to each child.
func TestDartUnknownKeyInheritanceContract(t *testing.T) {
	for _, parent := range []UnknownKeyMode{Allow, Strip, Reject} {
		for _, child := range []UnknownKeyMode{Allow, Strip, Reject} {
			schema := Object(map[string]Schema{
				"child": Object(map[string]Schema{"name": String()}).UnknownKeys(child),
			}).UnknownKeys(parent)
			result := schema.SafeParse(map[string]any{"child": map[string]any{"name": "Alice", "admin": true}})
			effective := parent
			if parent == Allow {
				effective = child
			}
			if effective == Reject {
				if result.Success || len(result.Issues) != 1 || result.Issues[0].Code != "unknown_key" {
					t.Fatalf("%s/%s: expected unknown_key, got %+v", parent, child, result)
				}
			} else {
				if !result.Success {
					t.Fatalf("%s/%s: %+v", parent, child, result.Issues)
				}
				_, retained := result.Data.(map[string]any)["child"].(map[string]any)["admin"]
				if retained != (effective == Allow) {
					t.Fatalf("%s/%s: unexpected retention", parent, child)
				}
			}
		}
	}
}

// TestDartUint64CanonicalCoercionBoundaries verifies exact conversion across the
// signed boundary and uint64 maximum without weakening the target range checks.
func TestDartUint64CanonicalCoercionBoundaries(t *testing.T) {
	schema, err := ImportJSON([]byte(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"uint64","coerce":["string->int"]},"definitions":{},"extensions":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"9223372036854775807", "9223372036854775808", "18446744073709551615"} {
		result := schema.SafeParse(" " + input + " ")
		if !result.Success {
			t.Fatalf("%s: %+v", input, result.Issues)
		}
		encoded, err := json.Marshal(result.Data)
		if err != nil || string(encoded) != input {
			t.Fatalf("%s: precision lost: %s (%v)", input, encoded, err)
		}
	}
	for _, input := range []string{"-1", "18446744073709551616", "+1", "1.5"} {
		if result := schema.SafeParse(input); result.Success {
			t.Fatalf("accepted invalid uint64 %s", input)
		}
	}
	for _, target := range []Schema{Int().Coerce("string->int"), Uint8().Coerce("string->int")} {
		if result := target.SafeParse("18446744073709551615"); result.Success {
			t.Fatal("target range check bypassed")
		}
	}
}

// TestDartTransitiveDefaults keeps declaring-node coercion/default order across
// nullable/optional wrappers and resolved references, including false/null values.
func TestDartTransitiveDefaults(t *testing.T) {
	doc := []byte(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"object","required":["text"],"properties":{
 "text":{"kind":"nullable","schema":{"kind":"string","coerce":["trim"],"default":"  fallback  "}},
 "flag":{"kind":"ref","ref":"#/definitions/Flag"},
 "nil":{"kind":"nullable","schema":{"kind":"null","default":null}},
 "bag":{"kind":"nullable","schema":{"kind":"any","default":{"value":"original"}}},
 "empty":{"kind":"optional","schema":{"kind":"nullable","schema":{"kind":"array","items":{"kind":"string"},"default":[]}}}
 }},"definitions":{"Flag":{"kind":"optional","schema":{"kind":"bool","default":false}}},"extensions":{}}`)
	schema, err := ImportJSON(doc)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]any{"text": "  fallback  ", "flag": false, "nil": nil, "empty": []any{}, "bag": map[string]any{"value": "original"}}
	parsed := schema.SafeParse(map[string]any{})
	if !parsed.Success || !reflect.DeepEqual(parsed.Data, expected) {
		t.Fatalf("default values: %+v", parsed)
	}
	parsed.Data.(map[string]any)["bag"].(map[string]any)["value"] = "changed"
	if again := schema.SafeParse(map[string]any{}); !again.Success || !reflect.DeepEqual(again.Data, expected) {
		t.Fatalf("shared defaults: %+v", again)
	}
	present := schema.SafeParse(map[string]any{"text": "  present  "})
	if !present.Success || present.Data.(map[string]any)["text"] != "present" {
		t.Fatalf("present coercion: %+v", present)
	}
	exported, err := ExportJSON(schema, Portable)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := ImportJSON(exported)
	if err != nil {
		t.Fatal(err)
	}
	if roundTrip := restored.SafeParse(map[string]any{}); !roundTrip.Success || !reflect.DeepEqual(roundTrip.Data, expected) {
		t.Fatalf("roundtrip: %+v", roundTrip)
	}
}

// TestDartTransitiveDefaultFailuresAndCycles checks invalid values and bounded recursive lookup.
func TestDartTransitiveDefaultFailuresAndCycles(t *testing.T) {
	for _, doc := range []string{
		`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"object","properties":{"x":{"kind":"nullable","schema":{"kind":"string","default":42}}},"required":[]},"definitions":{},"extensions":{}}`,
		`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"object","properties":{"x":{"kind":"ref","ref":"#/definitions/Cycle"}},"required":[]},"definitions":{"Cycle":{"kind":"ref","ref":"#/definitions/Cycle"}},"extensions":{}}`,
	} {
		schema, err := ImportJSON([]byte(doc))
		if err != nil {
			t.Fatal(err)
		}
		result := schema.SafeParse(map[string]any{})
		if _, cycle := schema.(*ObjectSchema).properties["x"].(*RefSchema); cycle {
			if !result.Success {
				t.Fatalf("cycle lookup: %+v", result)
			}
		} else if result.Success || len(result.Issues) != 1 || result.Issues[0].Code != IssueDefaultInvalid || !reflect.DeepEqual(result.Issues[0].Path, []any{"x"}) {
			t.Fatalf("invalid default: %+v", result)
		}
	}
}

// TestDartCanonicalSignedOverflowDiagnostic distinguishes integer overflow from invalid types.
func TestDartCanonicalSignedOverflowDiagnostic(t *testing.T) {
	for _, target := range []*IntSchema{Int(), Int64(), Int8()} {
		result := target.Coerce("string->int").SafeParse("18446744073709551615")
		if result.Success || len(result.Issues) != 1 || result.Issues[0].Code != IssueTooLarge || result.Issues[0].Received != "18446744073709551615" {
			t.Fatalf("canonical overflow: %+v", result)
		}
	}
	legacy := Int().Coerce().SafeParse("18446744073709551615")
	if legacy.Success || legacy.Issues[0].Code != IssueCoercionFailed {
		t.Fatalf("legacy coercion changed: %+v", legacy)
	}
}

// TestDartPrimitiveDefaultRoundTrip preserves primitive defaults and validates impossible never defaults.
func TestDartPrimitiveDefaultRoundTrip(t *testing.T) {
	for _, test := range []struct {
		kind  string
		value any
	}{
		{"null", nil}, {"any", map[string]any{"items": []any{"x"}}}, {"unknown", false}, {"never", 42},
	} {
		doc := map[string]any{"anyvaliVersion": "1.0", "schemaVersion": "1.1", "root": map[string]any{"kind": test.kind, "default": test.value}, "definitions": map[string]any{}, "extensions": map[string]any{}}
		input, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		schema, err := ImportJSON(input)
		if err != nil {
			t.Fatal(err)
		}
		exported, err := ExportJSON(schema, Portable)
		if err != nil {
			t.Fatal(err)
		}
		restored, err := ImportJSON(exported)
		if err != nil {
			t.Fatal(err)
		}
		for _, candidate := range []Schema{schema, restored} {
			result := candidate.SafeParse(absentValue)
			if test.kind == "never" {
				if result.Success || result.Issues[0].Code != IssueDefaultInvalid {
					t.Fatalf("never default: %+v", result)
				}
			} else if !result.Success || !reflect.DeepEqual(result.Data, test.value) {
				t.Fatalf("%s default: %+v", test.kind, result)
			}
		}
	}
}

// TestDartRefOccurrenceModifiers prevents defaults/coercions from leaking through
// shared definition resolution and preserves target pipeline order across export.
func TestDartRefOccurrenceModifiers(t *testing.T) {
	doc := []byte(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","extensions":{},"root":{"kind":"object","required":["c","d"],"properties":{
 "a":{"kind":"ref","ref":"#/definitions/Value","default":7},
 "b":{"kind":"ref","ref":"#/definitions/Value","default":8},
 "c":{"kind":"ref","ref":"#/definitions/Value","coerce":["string->int"]},
 "d":{"kind":"ref","ref":"#/definitions/Value"},
 "chain":{"kind":"ref","ref":"#/definitions/Upper","default":" x ","coerce":["trim"]},
 "wrapped":{"kind":"optional","schema":{"kind":"nullable","default":" x ","schema":{"kind":"string","coerce":["trim"],"startsWith":"x"}}}
 }},"definitions":{"Value":{"kind":"int"},"Upper":{"kind":"string","coerce":["upper"]}}}`)
	schema, err := ImportJSON(doc)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]any{"a": int64(7), "b": int64(8), "c": int64(9), "d": int64(10), "chain": " X ", "wrapped": "x"}
	check := func(candidate Schema) {
		t.Helper()
		result := candidate.SafeParse(map[string]any{"c": "9", "d": int64(10)})
		if !result.Success || !reflect.DeepEqual(result.Data, expected) {
			t.Fatalf("ref modifiers: %+v", result)
		}
	}
	check(schema)
	encoded, err := ExportJSON(schema, Portable)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := ImportJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	check(restored)
	for _, key := range []string{"a", "d"} {
		input := map[string]any{"c": "9", "d": int64(10)}
		input[key] = "7"
		if result := schema.SafeParse(input); result.Success {
			t.Fatalf("coercion leaked to %s", key)
		}
	}
	present := schema.SafeParse(map[string]any{"c": "9", "d": int64(10), "chain": " y "})
	if !present.Success || present.Data.(map[string]any)["chain"] != "Y" {
		t.Fatalf("ref coercion order: %+v", present)
	}
}

// TestDartCanonicalCoercionTypedInput applies string conversions only to their
// declared source and leaves target validators responsible for other types.
func TestDartCanonicalCoercionTypedInput(t *testing.T) {
	for _, test := range []struct {
		schema   Schema
		input    any
		expected any
	}{
		{Int().Coerce("string->int"), int64(42), int64(42)},
		{Number().Coerce("string->number"), float64(2.5), float64(2.5)},
		{Bool().Coerce("string->bool"), true, true},
	} {
		result := test.schema.SafeParse(test.input)
		if !result.Success || !reflect.DeepEqual(result.Data, test.expected) {
			t.Fatalf("typed canonical input: %+v", result)
		}
		invalid := test.schema.SafeParse(map[string]any{})
		if invalid.Success || invalid.Issues[0].Code != IssueInvalidType {
			t.Fatalf("non-string invalid input bypassed validation: %+v", invalid)
		}
	}
	legacy := Int().Coerce().SafeParse(int64(42))
	if legacy.Success || legacy.Issues[0].Code != IssueCoercionFailed {
		t.Fatalf("legacy coercion changed: %+v", legacy)
	}
}

// TestDartStringPreprocessingTypedTargets leaves non-string sources to validation
// while retaining ordered preprocessing for incoming strings.
func TestDartStringPreprocessingTypedTargets(t *testing.T) {
	for _, operation := range []CoercionType{CoerceTrim, CoerceLower, CoerceUpper} {
		value, err := applyCoercion(int64(42), operation)
		if err != nil || value != int64(42) {
			t.Fatalf("string-source bypass: %v %v", value, err)
		}
		result := String().Coerce(operation).SafeParse(int64(42))
		if result.Success || result.Issues[0].Code != IssueInvalidType {
			t.Fatalf("target validation: %+v", result)
		}
	}
	integer := Int().Coerce(CoerceTrim, "string->int")
	if r := integer.SafeParse(int64(42)); !r.Success || r.Data != int64(42) {
		t.Fatalf("typed preprocessing: %+v", r)
	}
	boolean := Bool().Coerce(CoerceUpper, "string->bool")
	if r := boolean.SafeParse("falſe"); !r.Success || r.Data != false {
		t.Fatalf("ordered bool preprocessing: %+v", r)
	}
}

// TestJavaScriptPortableCoercionFixture consumes a document generated and asserted
// by the JS SDK, including nested schema edges and data keys named coerce.
func TestJavaScriptPortableCoercionFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/js-ref-coercions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Schema json.RawMessage
		Input  any
		Output any
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	schema, err := ImportJSON(fixture.Schema)
	if err != nil {
		t.Fatal(err)
	}
	result := schema.SafeParse(fixture.Input)
	if !result.Success {
		t.Fatalf("JS export: %+v", result)
	}
	got, err := json.Marshal(result.Data)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(fixture.Output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("JS/Go output differs: %s != %s", got, want)
	}
	encoded, err := ExportJSON(schema, Portable)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := ImportJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	again := restored.SafeParse(fixture.Input)
	if !again.Success {
		t.Fatalf("reimport: %+v", again)
	}
	roundTrip, err := json.Marshal(again.Data)
	if err != nil || !bytes.Equal(roundTrip, want) {
		t.Fatalf("roundtrip output: %s (%v)", roundTrip, err)
	}
}

// TestDartWrapperCoercions retains optional/nullable modifiers at their own node
// and distinguishes declaring-node defaults from defaults entering an inner coercion.
func TestDartWrapperCoercions(t *testing.T) {
	for _, kind := range []string{"optional", "nullable"} {
		root := map[string]any{"kind": kind, "schema": map[string]any{"kind": "int"}, "coerce": []any{"string->int"}}
		importRoot := func(node map[string]any) Schema {
			t.Helper()
			doc := map[string]any{"anyvaliVersion": "1.0", "schemaVersion": "1.1", "root": node, "definitions": map[string]any{}, "extensions": map[string]any{}}
			data, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ImportJSON(data)
			if err != nil {
				t.Fatal(err)
			}
			return schema
		}
		schema := importRoot(root)
		for _, input := range []any{"7", int64(7)} {
			result := schema.SafeParse(input)
			if !result.Success || result.Data != int64(7) {
				t.Fatalf("%s wrapper conversion: %+v", kind, result)
			}
		}
		if kind == "nullable" {
			if result := schema.SafeParse(nil); !result.Success || result.Data != nil {
				t.Fatalf("nullable null: %+v", result)
			}
		}
		// Go currently exports the same shape in both modes; this checks
		// export/import consistency, not portable normalization.
		encoded, err := ExportJSON(schema, Portable)
		if err != nil {
			t.Fatal(err)
		}
		restored, err := ImportJSON(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if result := restored.SafeParse("7"); !result.Success || result.Data != int64(7) {
			t.Fatalf("%s roundtrip: %+v", kind, result)
		}
		root["coerce"] = "string->int"
		if result := importRoot(root).SafeParse("7"); !result.Success || result.Data != int64(7) {
			t.Fatalf("single coercion identifier: %+v", result)
		}
		root["default"] = "7"
		if result := importRoot(root).SafeParse(absentValue); result.Success || result.Issues[0].Code != IssueDefaultInvalid {
			t.Fatalf("own default must skip own conversion: %+v", result)
		}
		delete(root, "coerce")
		root["schema"] = map[string]any{"kind": "int", "coerce": []any{"string->int"}}
		if result := importRoot(root).SafeParse(absentValue); !result.Success || result.Data != int64(7) {
			t.Fatalf("default must enter target conversion: %+v", result)
		}
	}
}

// TestDartDefaultsForEveryGoKind ensures every built-in importer retains the same
// common default and metadata fields, including container/composition defaults.
func TestDartDefaultsForEveryGoKind(t *testing.T) {
	cases := []struct {
		node  map[string]any
		value any
	}{
		{map[string]any{"kind": "string"}, "x"}, {map[string]any{"kind": "bool"}, false},
		{map[string]any{"kind": "null"}, nil}, {map[string]any{"kind": "any"}, map[string]any{"value": "x"}},
		{map[string]any{"kind": "unknown"}, false}, {map[string]any{"kind": "never"}, 42},
		{map[string]any{"kind": "literal", "value": "x"}, "x"},
		{map[string]any{"kind": "enum", "values": []any{"x"}}, "x"},
		{map[string]any{"kind": "array", "items": map[string]any{"kind": "string"}}, []any{"x"}},
		{map[string]any{"kind": "tuple", "elements": []any{map[string]any{"kind": "string"}}}, []any{"x"}},
		{map[string]any{"kind": "record", "values": map[string]any{"kind": "string"}}, map[string]any{"key": "x"}},
		{map[string]any{"kind": "object", "properties": map[string]any{"value": map[string]any{"kind": "string"}}, "required": []any{"value"}}, map[string]any{"value": "x"}},
		{map[string]any{"kind": "union", "variants": []any{map[string]any{"kind": "string"}, map[string]any{"kind": "int"}}}, "x"},
		{map[string]any{"kind": "intersection", "allOf": []any{map[string]any{"kind": "string"}, map[string]any{"kind": "string", "minLength": 1}}}, "x"},
		{map[string]any{"kind": "optional", "schema": map[string]any{"kind": "string"}}, "x"},
		{map[string]any{"kind": "nullable", "schema": map[string]any{"kind": "string"}}, nil},
		{map[string]any{"kind": "ref", "ref": "#/definitions/Text"}, "x"},
	}
	for _, kind := range []string{"int", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "number", "float32", "float64"} {
		cases = append(cases, struct {
			node  map[string]any
			value any
		}{map[string]any{"kind": kind}, 7})
	}
	for _, test := range cases {
		t.Run(test.node["kind"].(string), func(t *testing.T) {
			metadata := map[string]any{"title": "Example", "custom": map[string]any{"scope": "dart"}}
			test.node["default"] = test.value
			test.node["metadata"] = metadata
			doc := map[string]any{"anyvaliVersion": "1.0", "schemaVersion": "1.1", "root": test.node, "definitions": map[string]any{"Text": map[string]any{"kind": "string"}}, "extensions": map[string]any{}}
			input, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ImportJSON(input)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := ExportJSON(schema, Portable)
			if err != nil {
				t.Fatal(err)
			}
			var exported Document
			if err := json.Unmarshal(encoded, &exported); err != nil {
				t.Fatal(err)
			}
			if _, present := exported.Root["default"]; !present {
				t.Fatalf("dropped default: %s", encoded)
			}
			if !reflect.DeepEqual(exported.Root["metadata"], metadata) {
				t.Fatalf("dropped metadata: %s", encoded)
			}
			restored, err := ImportJSON(encoded)
			if err != nil {
				t.Fatal(err)
			}
			for _, candidate := range []Schema{schema, restored} {
				result := candidate.SafeParse(absentValue)
				if test.node["kind"] == "never" {
					if result.Success || result.Issues[0].Code != IssueDefaultInvalid {
						t.Fatalf("never default: %+v", result)
					}
				} else {
					if !result.Success {
						t.Fatalf("default: %+v", result)
					}
					got, err := json.Marshal(result.Data)
					if err != nil {
						t.Fatal(err)
					}
					want, err := json.Marshal(test.value)
					if err != nil || !bytes.Equal(got, want) {
						t.Fatalf("default output: %s != %s (%v)", got, want, err)
					}
				}
			}
		})
	}
}

// TestDartImportedSensitiveMetadata enables sensitive helpers on imported nodes.
func TestDartImportedSensitiveMetadata(t *testing.T) {
	schema, err := ImportJSON([]byte(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"int","metadata":{"sensitive":true}},"definitions":{},"extensions":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := Encrypt(schema, int64(7), func(_ []any, value any) (any, error) {
		if value != int64(7) {
			t.Fatalf("unexpected plaintext: %v", value)
		}
		return "encrypted:7", nil
	})
	if err != nil || stored != "encrypted:7" {
		t.Fatalf("imported sensitivity: %v %v", stored, err)
	}
	if result := SafeParseEncrypted(schema, stored); !result.Success {
		t.Fatalf("encrypted validation: %+v", result)
	}
	plain, err := Decrypt(schema, stored, func(_ []any, _ any) (any, error) { return int64(7), nil })
	if err != nil || plain != int64(7) {
		t.Fatalf("decrypt: %v %v", plain, err)
	}
}

// TestDartCoercionTargetValidation rejects documents that would otherwise silently
// change unconstrained values, while accepting matching wrapped numeric targets.
func TestDartCoercionTargetValidation(t *testing.T) {
	invalid := []map[string]any{
		{"kind": "any"}, {"kind": "unknown"}, {"kind": "never"}, {"kind": "null"},
		{"kind": "string"}, {"kind": "bool"}, {"kind": "number"},
		{"kind": "literal", "value": "7"}, {"kind": "enum", "values": []any{"7"}},
		{"kind": "array", "items": map[string]any{"kind": "int"}},
		{"kind": "tuple", "elements": []any{map[string]any{"kind": "int"}}},
		{"kind": "record", "values": map[string]any{"kind": "int"}},
		{"kind": "object", "properties": map[string]any{}},
		{"kind": "union", "variants": []any{map[string]any{"kind": "int"}}},
		{"kind": "intersection", "allOf": []any{map[string]any{"kind": "int"}}},
		{"kind": "optional", "schema": map[string]any{"kind": "any"}},
		{"kind": "nullable", "schema": map[string]any{"kind": "string"}},
		{"kind": "ref", "ref": "#/definitions/Any"},
	}
	defs := map[string]map[string]any{"Any": {"kind": "any"}, "Int": {"kind": "int"}, "Cycle": {"kind": "ref", "ref": "#/definitions/Cycle"}}
	for _, node := range invalid {
		for _, operations := range []any{[]any{"string->int"}, []any{"int"}, []any{"trim"}} {
			if (node["kind"] == "string" || node["kind"] == "nullable") && reflect.DeepEqual(operations, []any{"trim"}) {
				continue
			}
			node["coerce"] = operations
			if schema, err := Import(&Document{Root: node, Definitions: defs}); err == nil || schema != nil {
				t.Fatalf("accepted incompatible target %v", node)
			}
		}
	}
	for _, operations := range []any{true, nil, []any{false}, []any{"unknown"}, []any{"string->int", "trim"}, []any{"string->number"}, []any{"trim"}} {
		node := map[string]any{"kind": "int", "coerce": operations}
		if _, err := Import(&Document{Root: node}); err == nil {
			t.Fatalf("accepted invalid coercion configuration %v", operations)
		}
	}
	for _, ref := range []string{"#/definitions/Missing", "#/definitions/Cycle"} {
		if _, err := Import(&Document{Root: map[string]any{"kind": "ref", "ref": ref, "coerce": "string->int"}, Definitions: defs}); err == nil {
			t.Fatalf("accepted unresolved coercion target %s", ref)
		}
	}
	for _, node := range []map[string]any{
		{"kind": "int"},
		{"kind": "optional", "schema": map[string]any{"kind": "int"}},
		{"kind": "nullable", "schema": map[string]any{"kind": "optional", "schema": map[string]any{"kind": "ref", "ref": "#/definitions/Int"}}},
		{"kind": "ref", "ref": "#/definitions/Int"},
	} {
		node["coerce"] = []any{"trim", "string->int"}
		schema, err := Import(&Document{Root: node, Definitions: defs})
		if err != nil {
			t.Fatal(err)
		}
		if result := schema.SafeParse(" 7 "); !result.Success || result.Data != int64(7) {
			t.Fatalf("matching coercion failed: %+v", result)
		}
	}
	if _, err := Import(&Document{Root: map[string]any{"kind": "any", "coerce": []any{}}}); err != nil {
		t.Fatalf("empty list must remain a no-op: %v", err)
	}
}

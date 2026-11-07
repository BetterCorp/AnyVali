package anyvali

import (
	"encoding/json"
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

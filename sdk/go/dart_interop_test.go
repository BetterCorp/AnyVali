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

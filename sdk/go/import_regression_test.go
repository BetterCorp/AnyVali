package anyvali

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func recursiveDocument(kind string) []byte {
	return []byte(fmt.Sprintf(`{"anyvaliVersion":"1.0","schemaVersion":"1.1","root":{"kind":"ref","ref":"#/definitions/JsonValue"},"definitions":{"JsonValue":{"kind":"union","variants":[{"kind":%q},{"kind":"array","items":{"kind":"ref","ref":"#/definitions/JsonValue"}},{"kind":"record","valueSchema":{"kind":"ref","ref":"#/definitions/JsonValue"}}]}}}`, kind))
}

func TestConcurrentImportIsolation(t *testing.T) {
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			kind, valid, invalid := "string", any("leaf"), any(true)
			if worker%2 == 1 {
				kind, valid, invalid = "bool", true, "leaf"
			}
			for i := 0; i < 50; i++ {
				for _, doc := range [][]byte{[]byte(`{"root":{"kind":"string"}}`), recursiveDocument(kind)} {
					s, err := ImportJSON(doc)
					if err != nil {
						t.Error(err)
						return
					}
					if s.ToNode()["kind"] == "ref" {
						if r := s.SafeParse(map[string]any{"nested": []any{valid}}); !r.Success {
							t.Errorf("recursive parse: %+v", r)
						}
						if s.SafeParse(map[string]any{"nested": []any{invalid}}).Success {
							t.Error("cross-document definition leakage")
						}
					}
				}
				if _, err := ImportJSON([]byte(`{"root":{"kind":"ref","ref":"#/definitions/JsonValue"},"definitions":{"JsonValue":{"kind":"invalid"}}}`)); err == nil {
					t.Error("expected import error")
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestCanonicalRecordRoundTrip(t *testing.T) {
	for _, source := range []string{
		`{"root":{"kind":"record","valueSchema":{"kind":"record","valueSchema":{"kind":"string"}}}}`,
		`{"root":{"kind":"record","value":{"kind":"record","value":{"kind":"string"}}}}`,
		string(recursiveDocument("string")),
	} {
		s, err := ImportJSON([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		for round := 0; round < 2; round++ {
			if !s.SafeParse(map[string]any{"outer": map[string]any{"inner": "leaf"}}).Success {
				t.Fatal("valid nested record rejected")
			}
			if s.SafeParse(map[string]any{"outer": map[string]any{"inner": true}}).Success {
				t.Fatal("invalid leaf accepted")
			}
			data, err := ExportJSON(s, Portable)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatal(err)
			}
			root := doc["root"].(map[string]any)
			if root["kind"] == "record" {
				if _, ok := root["values"]; !ok {
					t.Fatal("missing canonical values")
				}
				if _, ok := root["value"]; ok {
					t.Fatal("exported legacy key")
				}
			}
			s, err = ImportJSON(data)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := ImportJSON([]byte(`{"root":{"kind":"record","valueSchema":null,"value":{"kind":"string"}}}`)); err == nil {
		t.Fatal("invalid canonical key must not fall back to alias")
	}
}

func TestImportedDefaultsDistinguishNull(t *testing.T) {
	for _, field := range []string{`{"kind":"string","default":"fallback"}`, `{"kind":"bool","default":true}`, `{"kind":"number","default":12}`, `{"kind":"int","default":12}`} {
		s, err := ImportJSON([]byte(`{"root":{"kind":"object","required":["value"],"properties":{"value":` + field + `}}}`))
		if err != nil {
			t.Fatal(err)
		}
		for round := 0; round < 2; round++ {
			if !s.SafeParse(map[string]any{}).Success {
				t.Fatal("missing default rejected")
			}
			if s.SafeParse(map[string]any{"value": nil}).Success {
				t.Fatal("null selected default")
			}
			data, err := ExportJSON(s, Portable)
			if err != nil {
				t.Fatal(err)
			}
			s, err = ImportJSON(data)
			if err != nil {
				t.Fatal(err)
			}
		}
		scalar, err := ImportJSON([]byte(`{"root":` + field + `}`))
		if err != nil {
			t.Fatal(err)
		}
		if scalar.SafeParse(nil).Success {
			t.Fatal("root null selected default")
		}
	}
	for _, value := range []any{nil, "fallback"} {
		s := Object(map[string]Schema{"value": Nullable(String()).Default(value)})
		for _, input := range []map[string]any{{}, {"value": nil}} {
			r := s.SafeParse(input)
			if !r.Success {
				t.Fatalf("nullable default: %+v", r)
			}
			if _, present := input["value"]; present && r.Data.(map[string]any)["value"] != nil {
				t.Fatal("nullable null replaced")
			}
		}
	}
}

func TestImportedDefinitionsInNativeParents(t *testing.T) {
	s, err := ImportJSON(recursiveDocument("string"))
	if err != nil {
		t.Fatal(err)
	}
	parent := Object(map[string]Schema{"value": s})
	data, err := ExportJSON(parent, Portable)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := ImportJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if !restored.SafeParse(map[string]any{"value": []any{"leaf"}}).Success {
		t.Fatal("native parent lost recursive definitions")
	}
	other, err := ImportJSON(recursiveDocument("bool"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Export(Object(map[string]Schema{"first": s, "second": other}), Portable); err == nil {
		t.Fatal("conflicting definitions silently merged")
	}
}

func TestRecursiveExportUsesCanonicalChildKeys(t *testing.T) {
	s, err := ImportJSON(recursiveDocument("string"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Export(s, Portable)
	if err != nil {
		t.Fatal(err)
	}
	variants, ok := doc.Definitions["JsonValue"]["variants"].([]any)
	if !ok || len(variants) != 3 {
		t.Fatal("union must export canonical variants")
	}
	if variants[1].(map[string]any)["items"] == nil {
		t.Fatal("array must export canonical items")
	}
	if variants[2].(map[string]any)["values"] == nil {
		t.Fatal("record must export canonical values")
	}
}

func TestExportIncludesEmptyDocumentMaps(t *testing.T) {
	data, err := ExportJSON(String(), Portable)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"definitions", "extensions"} {
		if value, ok := doc[key].(map[string]any); !ok || len(value) != 0 {
			t.Fatalf("%s must be an empty object: %s", key, data)
		}
	}
}

func TestSpecRecordInputAndAliasPrecedence(t *testing.T) {
	s, err := ImportJSON([]byte(`{"root":{"kind":"record","values":{"kind":"string"},"valueSchema":{"kind":"bool"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !s.SafeParse(map[string]any{"key": "leaf"}).Success || s.SafeParse(map[string]any{"key": true}).Success {
		t.Fatal("spec values must take precedence")
	}
}

func TestRecursiveValidationDepth(t *testing.T) {
	s, err := ImportJSON([]byte(`{"root":{"kind":"ref","ref":"#/definitions/Node"},"definitions":{"Node":{"kind":"object","required":[],"properties":{"child":{"kind":"ref","ref":"#/definitions/Node"}}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	var deep any = map[string]any{}
	for i := 0; i < 1000; i++ {
		deep = map[string]any{"child": deep}
	}
	result := s.SafeParse(deep)
	if result.Success || result.Issues[0].Message != "maximum validation depth exceeded" {
		t.Fatalf("unexpected deep result: %+v", result)
	}
	cycle := map[string]any{}
	cycle["child"] = cycle
	if s.SafeParse(cycle).Success {
		t.Fatal("cyclic input accepted")
	}
	if !s.SafeParse(map[string]any{"child": map[string]any{}}).Success {
		t.Fatal("valid input rejected after failure")
	}
	ref := newRefSchema("#/definitions/Tuple")
	ref.Resolve(Tuple(ref))
	array := make([]any, 1)
	array[0] = array
	if ref.SafeParse(array).Success {
		t.Fatal("cyclic tuple accepted")
	}
}

func TestEquivalentObjectDefinitionsExportDeterministically(t *testing.T) {
	data := []byte(`{"root":{"kind":"ref","ref":"#/definitions/Value"},"definitions":{"Value":{"kind":"object","properties":{"a":{"kind":"string"},"b":{"kind":"bool"},"c":{"kind":"number"}}}}}`)
	first, err := ImportJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ImportJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	parent := Object(map[string]Schema{"first": first, "second": second})
	for i := 0; i < 20; i++ {
		if _, err := Export(parent, Portable); err != nil {
			t.Fatal(err)
		}
	}
}

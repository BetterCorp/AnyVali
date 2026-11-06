package anyvali

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImportJSONRequiresOneCompleteDocument(t *testing.T) {
	document := `{"root":{"kind":"string"}}`
	for _, suffix := range []string{` {}`, ` true`, ` invalid`} {
		_, err := ImportJSON([]byte(document + suffix))
		if err == nil {
			t.Fatalf("accepted trailing content %q", suffix)
		}
	}
	schema, err := ImportJSON([]byte(document + " \n\t"))
	if err != nil || !schema.SafeParse("ok").Success {
		t.Fatalf("valid document with trailing whitespace was rejected: %v", err)
	}
}

func TestImportRejectsInvalidCountConstraints(t *testing.T) {
	for _, tc := range []struct {
		kind, key string
		empty     any
	}{
		{"string", "minLength", ""},
		{"string", "maxLength", ""},
		{"array", "minItems", []any{}},
		{"array", "maxItems", []any{}},
	} {
		t.Run(tc.key, func(t *testing.T) {
			for _, value := range []any{
				json.Number("-1"),
				json.Number("0.5"),
				json.Number("9223372036854775808"),
				json.Number("1e1001"),
				"2",
			} {
				root := map[string]any{"kind": tc.kind, tc.key: value}
				if tc.kind == "array" {
					root["items"] = map[string]any{"kind": "any"}
				}
				_, err := ImportJSON(mustJSON(t, map[string]any{"root": root}))
				if err == nil || !strings.Contains(err.Error(), tc.key) {
					t.Fatalf("%s=%v: expected constraint import error, got %v", tc.key, value, err)
				}
			}

			root := map[string]any{"kind": tc.kind, tc.key: 0}
			if tc.kind == "array" {
				root["items"] = map[string]any{"kind": "any"}
			}
			schema, err := ImportJSON(mustJSON(t, map[string]any{"root": root}))
			if err != nil {
				t.Fatal(err)
			}
			if result := schema.SafeParse(tc.empty); !result.Success {
				t.Fatalf("zero constraint rejected empty input: %v", result.Issues)
			}
			if got := schema.ToNode()[tc.key]; got != 0 {
				t.Fatalf("zero constraint was not preserved on export: %v", got)
			}
		})
	}
}

func TestImportRejectsInvalidUnsignedNumericConstraints(t *testing.T) {
	for _, key := range []string{"min", "max", "exclusiveMin", "exclusiveMax", "multipleOf"} {
		for _, value := range []any{"10", json.Number("1e1001")} {
			root := map[string]any{"kind": "uint64", key: value}
			_, err := ImportJSON(mustJSON(t, map[string]any{"root": root}))
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("%s=%v: invalid unsigned constraint was accepted", key, value)
			}
		}
	}
}

func TestImportedIntegerExclusiveBounds(t *testing.T) {
	for _, kind := range []string{"int64", "uint64"} {
		for _, tc := range []struct {
			key, bound, code   string
			accepted, rejected int64
		}{
			{"exclusiveMin", "2", IssueTooSmall, 3, 2},
			{"exclusiveMax", "2", IssueTooLarge, 1, 2},
			{"exclusiveMin", "2.5", IssueTooSmall, 3, 2},
			{"exclusiveMax", "2.5", IssueTooLarge, 2, 3},
		} {
			t.Run(kind+"/"+tc.key+"/"+tc.bound, func(t *testing.T) {
				root := map[string]any{"kind": kind, tc.key: json.Number(tc.bound)}
				schema, err := ImportJSON(mustJSON(t, map[string]any{"root": root}))
				if err != nil {
					t.Fatal(err)
				}
				if result := schema.SafeParse(tc.accepted); !result.Success {
					t.Fatalf("value on valid side of exclusive bound was rejected: %v", result.Issues)
				}
				result := schema.SafeParse(tc.rejected)
				if result.Success || len(result.Issues) != 1 || result.Issues[0].Code != tc.code {
					t.Fatalf("exclusive bound was not enforced: %+v", result)
				}
			})
		}
	}
}

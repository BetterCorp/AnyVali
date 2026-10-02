package anyvali

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestLiteralAndEnumRejectCrossTypeLookalikes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema Schema
		input  any
	}{
		{"boolean as string", Literal(true), "true"},
		{"number as string", Literal(1), "1"},
		{"enum number as string", Enum(1, 2), "1"},
		{"distinct large integers", Literal(int64(9007199254740993)), int64(9007199254740992)},
		{"distinct large unsigned integers", Enum(uint64(9007199254740993)), uint64(9007199254740992)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.schema.SafeParse(tc.input); result.Success {
				t.Fatalf("accepted lookalike input %T(%v)", tc.input, tc.input)
			}
		})
	}
	if result := Literal(json.Number("1e0")).SafeParse(int64(1)); !result.Success {
		t.Fatalf("equivalent numeric values should still match: %v", result.Issues)
	}
}

func TestIntegerConversionRejectsOverflowWithoutWrapping(t *testing.T) {
	if result := Int64().SafeParse(uint64(math.MaxUint64)); result.Success {
		t.Fatal("uint above MaxInt64 wrapped into an accepted signed value")
	}
	if result := Uint64().SafeParse(math.Exp2(64)); result.Success {
		t.Fatal("float64 2^64 overflowed into an accepted uint64")
	}
	if result := Uint64().SafeParse(json.Number("18446744073709551615")); !result.Success {
		t.Fatalf("exact MaxUint64 JSON number should be accepted: %v", result.Issues)
	}
	if result := Int64().SafeParse(json.Number("1e0")); !result.Success {
		t.Fatalf("integral exponent should be accepted: %v", result.Issues)
	}
}

func TestImportJSONPreservesWideIntegerConstraints(t *testing.T) {
	for _, tc := range []struct {
		name, schema, input string
		wantValid           bool
	}{
		{"signed maximum", `{"root":{"kind":"int64","max":9007199254740992}}`, `9007199254740993`, false},
		{"signed minimum", `{"root":{"kind":"int64","min":9007199254740993}}`, `9007199254740992`, false},
		{"unsigned maximum", `{"root":{"kind":"uint64","max":18446744073709551614}}`, `18446744073709551615`, false},
		{"unsigned maximum accepted", `{"root":{"kind":"uint64","max":18446744073709551614}}`, `18446744073709551614`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := ImportJSON([]byte(tc.schema))
			if err != nil {
				t.Fatal(err)
			}
			result := schema.SafeParse(json.Number(tc.input))
			if result.Success != tc.wantValid {
				t.Fatalf("got success=%v, issues=%v", result.Success, result.Issues)
			}
			if tc.name == "unsigned maximum" && !strings.Contains(string(mustJSON(t, schema.ToNode())), "18446744073709551614") {
				t.Fatal("wide uint64 constraint lost on export")
			}
		})
	}
	for _, tc := range []struct {
		schema, input string
		valid         bool
	}{
		{`{"root":{"kind":"int64","min":1.5}}`, `1`, false},
		{`{"root":{"kind":"int64","min":1.5}}`, `2`, true},
		{`{"root":{"kind":"uint64","min":-1}}`, `0`, true},
		{`{"root":{"kind":"uint64","max":-1}}`, `0`, false},
		{`{"root":{"kind":"int64","multipleOf":1.5}}`, `3`, true},
		{`{"root":{"kind":"int64","multipleOf":1.5}}`, `2`, false},
		{`{"root":{"kind":"int64","min":1e400}}`, `0`, false},
	} {
		schema, err := ImportJSON([]byte(tc.schema))
		if err != nil {
			t.Fatalf("valid numeric constraint rejected: %s: %v", tc.schema, err)
		}
		if result := schema.SafeParse(json.Number(tc.input)); result.Success != tc.valid {
			t.Fatalf("schema=%s input=%s: success=%v issues=%v", tc.schema, tc.input, result.Success, result.Issues)
		}
	}
	for _, raw := range []string{
		`{"root":{"kind":"int64","max":"10"}}`,
		`{"root":{"kind":"number","min":1e400}}`,
		`{"root":{"kind":"float32","max":1e400}}`,
	} {
		if _, err := ImportJSON([]byte(raw)); err == nil {
			t.Fatalf("accepted invalid constraint: %s", raw)
		}
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

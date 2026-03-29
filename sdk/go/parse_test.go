package anyvali

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToFloat64AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected float64
		ok       bool
	}{
		{"float64", float64(3.14), 3.14, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", int(42), 42.0, true},
		{"int8", int8(127), 127.0, true},
		{"int16", int16(1000), 1000.0, true},
		{"int32", int32(100000), 100000.0, true},
		{"int64", int64(999999), 999999.0, true},
		{"uint", uint(42), 42.0, true},
		{"uint8", uint8(255), 255.0, true},
		{"uint16", uint16(65535), 65535.0, true},
		{"uint32", uint32(100000), 100000.0, true},
		{"uint64", uint64(100000), 100000.0, true},
		{"json.Number valid", json.Number("3.14"), 3.14, true},
		{"json.Number invalid", json.Number("not-a-number"), 0, false},
		{"string", "hello", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
		{"slice", []any{1}, 0, false},
		{"map", map[string]any{"a": 1}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := toFloat64(tt.input)
			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got ok=%v", tt.ok, ok)
			}
			if ok && f != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, f)
			}
		})
	}
}

func TestToInt64AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int64
		ok       bool
	}{
		{"int", int(42), 42, true},
		{"int8", int8(-128), -128, true},
		{"int16", int16(1000), 1000, true},
		{"int32", int32(-100000), -100000, true},
		{"int64", int64(math.MaxInt64), math.MaxInt64, true},
		{"uint", uint(42), 42, true},
		{"uint8", uint8(255), 255, true},
		{"uint16", uint16(65535), 65535, true},
		{"uint32", uint32(100000), 100000, true},
		{"uint64 in range", uint64(100), 100, true},
		{"uint64 overflow", uint64(math.MaxUint64), 0, false},
		{"float32 integral", float32(42), 42, true},
		{"float32 fractional", float32(3.14), 0, false},
		{"float64 integral", float64(42), 42, true},
		{"float64 fractional", float64(3.14), 0, false},
		{"json.Number valid", json.Number("42"), 42, true},
		{"json.Number float", json.Number("3.14"), 0, false},
		{"json.Number invalid", json.Number("abc"), 0, false},
		{"string", "hello", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, ok := toInt64(tt.input)
			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got ok=%v", tt.ok, ok)
			}
			if ok && i != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, i)
			}
		})
	}
}

func TestToUint64AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected uint64
		ok       bool
	}{
		{"int positive", int(42), 42, true},
		{"int negative", int(-1), 0, false},
		{"int8 positive", int8(127), 127, true},
		{"int8 negative", int8(-1), 0, false},
		{"int16 positive", int16(1000), 1000, true},
		{"int16 negative", int16(-1), 0, false},
		{"int32 positive", int32(100000), 100000, true},
		{"int32 negative", int32(-1), 0, false},
		{"int64 positive", int64(100), 100, true},
		{"int64 negative", int64(-1), 0, false},
		{"uint", uint(42), 42, true},
		{"uint8", uint8(255), 255, true},
		{"uint16", uint16(65535), 65535, true},
		{"uint32", uint32(100000), 100000, true},
		{"uint64", uint64(math.MaxUint64), math.MaxUint64, true},
		{"float32 integral", float32(42), 42, true},
		{"float32 fractional", float32(3.14), 0, false},
		{"float32 negative", float32(-1), 0, false},
		{"float64 integral", float64(42), 42, true},
		{"float64 fractional", float64(3.14), 0, false},
		{"float64 negative", float64(-1), 0, false},
		{"json.Number valid", json.Number("42"), 42, true},
		{"json.Number negative", json.Number("-1"), 0, false},
		{"json.Number float", json.Number("3.14"), 0, false},
		{"string", "hello", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, ok := toUint64(tt.input)
			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got ok=%v", tt.ok, ok)
			}
			if ok && u != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, u)
			}
		})
	}
}

func TestTypeNameAllTypes(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, "null"},
		{true, "bool"},
		{false, "bool"},
		{"hello", "string"},
		{float64(3.14), "number"},
		{float32(1.5), "number"},
		{int(42), "int"},
		{int8(1), "int"},
		{int16(1), "int"},
		{int32(1), "int"},
		{int64(1), "int"},
		{uint(1), "uint"},
		{uint8(1), "uint"},
		{uint16(1), "uint"},
		{uint32(1), "uint"},
		{uint64(1), "uint"},
		{json.Number("42"), "number"},
		{[]any{1, 2}, "array"},
		{map[string]any{"a": 1}, "object"},
	}
	for _, tt := range tests {
		result := typeName(tt.input)
		if result != tt.expected {
			t.Errorf("typeName(%v) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestTypeNameCustomType(t *testing.T) {
	type myStruct struct{}
	result := typeName(myStruct{})
	if result != "anyvali.myStruct" {
		t.Fatalf("expected type name for custom struct, got %q", result)
	}
}

func TestRunPipelineCoercionFailed(t *testing.T) {
	b := &baseSchema{}
	b.addCoercion(CoerceToInt)
	result := b.runPipeline("not-a-number", func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if result.Success {
		t.Fatal("expected failure from coercion")
	}
	if result.Issues[0].Code != IssueCoercionFailed {
		t.Fatalf("expected coercion_failed, got %s", result.Issues[0].Code)
	}
}

func TestRunPipelineCoercionSuccess(t *testing.T) {
	b := &baseSchema{}
	b.addCoercion(CoerceTrim)
	result := b.runPipeline("  hello  ", func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", result.Data)
	}
}

func TestRunPipelineDefaultApplied(t *testing.T) {
	b := &baseSchema{}
	b.setDefault("fallback")
	result := b.runPipeline(nil, func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", result.Data)
	}
}

func TestRunPipelineAbsentNoDefault(t *testing.T) {
	b := &baseSchema{}
	result := b.runPipeline(absentValue, func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != nil {
		t.Fatalf("expected nil, got %v", result.Data)
	}
}

func TestRunPipelineAbsentWithDefault(t *testing.T) {
	b := &baseSchema{}
	b.setDefault("default_val")
	result := b.runPipeline(absentValue, func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != "default_val" {
		t.Fatalf("expected 'default_val', got %v", result.Data)
	}
}

func TestRunPipelineValidationFails(t *testing.T) {
	b := &baseSchema{}
	result := b.runPipeline("hello", func(v any) (any, []ValidationIssue) {
		return nil, []ValidationIssue{{Code: IssueInvalidType, Message: "bad"}}
	})
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestAddCoercionNode(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{"kind": "string"}
	b.addCoercionNode(node)
	if _, ok := node["coerce"]; ok {
		t.Fatal("expected no coerce key when no coercions")
	}

	b.addCoercion(CoerceTrim)
	b.addCoercion(CoerceLower)
	b.addCoercionNode(node)
	cs, ok := node["coerce"].([]any)
	if !ok {
		t.Fatal("expected coerce to be []any")
	}
	if len(cs) != 2 {
		t.Fatalf("expected 2 coercions, got %d", len(cs))
	}
	if cs[0] != "trim" || cs[1] != "lower" {
		t.Fatalf("expected [trim, lower], got %v", cs)
	}
}

func TestAddDefaultNode(t *testing.T) {
	b := &baseSchema{}
	node := map[string]any{"kind": "string"}
	b.addDefaultNode(node)
	if _, ok := node["default"]; ok {
		t.Fatal("expected no default key when no default set")
	}

	b.setDefault("val")
	b.addDefaultNode(node)
	if node["default"] != "val" {
		t.Fatalf("expected 'val', got %v", node["default"])
	}
}

func TestRunPipelineMultipleCoercions(t *testing.T) {
	b := &baseSchema{}
	b.addCoercion(CoerceTrim)
	b.addCoercion(CoerceLower)
	result := b.runPipeline("  HELLO  ", func(v any) (any, []ValidationIssue) {
		return v, nil
	})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", result.Data)
	}
}

package anyvali

import (
	"math"
	"testing"
)

func TestNumberSchemaValidTypes(t *testing.T) {
	s := Number()
	tests := []struct {
		name  string
		input any
	}{
		{"float64", float64(3.14)},
		{"float32", float32(2.5)},
		{"int", int(42)},
		{"int8", int8(1)},
		{"int16", int16(1)},
		{"int32", int32(1)},
		{"int64", int64(1)},
		{"uint", uint(1)},
		{"uint8", uint8(1)},
		{"uint16", uint16(1)},
		{"uint32", uint32(1)},
		{"uint64", uint64(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := s.SafeParse(tt.input)
			if !r.Success {
				t.Fatalf("expected success for %T, got issues: %v", tt.input, r.Issues)
			}
		})
	}
}

func TestNumberSchemaInvalidTypes(t *testing.T) {
	s := Number()
	tests := []any{"hello", true, nil, []any{}, map[string]any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
		if r.Issues[0].Code != IssueInvalidType {
			t.Fatalf("expected invalid_type, got %s", r.Issues[0].Code)
		}
	}
}

func TestNumberSchemaMin(t *testing.T) {
	s := Number().Min(0)
	if r := s.SafeParse(float64(-1)); r.Success {
		t.Fatal("expected failure for < min")
	}
	if r := s.SafeParse(float64(0)); !r.Success {
		t.Fatal("expected success for == min")
	}
	if r := s.SafeParse(float64(1)); !r.Success {
		t.Fatal("expected success for > min")
	}
}

func TestNumberSchemaMax(t *testing.T) {
	s := Number().Max(100)
	if r := s.SafeParse(float64(101)); r.Success {
		t.Fatal("expected failure for > max")
	}
	if r := s.SafeParse(float64(100)); !r.Success {
		t.Fatal("expected success for == max")
	}
}

func TestNumberSchemaExclusiveMin(t *testing.T) {
	s := Number().ExclusiveMin(0)
	if r := s.SafeParse(float64(0)); r.Success {
		t.Fatal("expected failure for == exclusiveMin")
	}
	if r := s.SafeParse(float64(-1)); r.Success {
		t.Fatal("expected failure for < exclusiveMin")
	}
	if r := s.SafeParse(float64(0.001)); !r.Success {
		t.Fatal("expected success for > exclusiveMin")
	}
}

func TestNumberSchemaExclusiveMax(t *testing.T) {
	s := Number().ExclusiveMax(10)
	if r := s.SafeParse(float64(10)); r.Success {
		t.Fatal("expected failure for == exclusiveMax")
	}
	if r := s.SafeParse(float64(11)); r.Success {
		t.Fatal("expected failure for > exclusiveMax")
	}
	if r := s.SafeParse(float64(9.999)); !r.Success {
		t.Fatal("expected success for < exclusiveMax")
	}
}

func TestNumberSchemaMultipleOf(t *testing.T) {
	s := Number().MultipleOf(0.5)
	if r := s.SafeParse(float64(1.5)); !r.Success {
		t.Fatal("expected success for 1.5 multipleOf 0.5")
	}
	if r := s.SafeParse(float64(1.3)); r.Success {
		t.Fatal("expected failure for 1.3 multipleOf 0.5")
	}
}

func TestNumberSchemaMultipleOfZero(t *testing.T) {
	s := Number().MultipleOf(0)
	// multipleOf 0 should be skipped (no division by zero)
	if r := s.SafeParse(float64(42)); !r.Success {
		t.Fatal("expected success when multipleOf is 0")
	}
}

func TestNumberSchemaNaN(t *testing.T) {
	s := Number()
	r := s.SafeParse(math.NaN())
	if r.Success {
		t.Fatal("expected failure for NaN")
	}
	if r.Issues[0].Code != IssueInvalidNumber {
		t.Fatalf("expected invalid_number, got %s", r.Issues[0].Code)
	}
}

func TestNumberSchemaInf(t *testing.T) {
	s := Number()
	r := s.SafeParse(math.Inf(1))
	if r.Success {
		t.Fatal("expected failure for +Inf")
	}
	r = s.SafeParse(math.Inf(-1))
	if r.Success {
		t.Fatal("expected failure for -Inf")
	}
}

func TestFloat32Schema(t *testing.T) {
	s := Float32()
	if r := s.SafeParse(float64(1.5)); !r.Success {
		t.Fatal("expected success for valid float32 value")
	}
}

func TestFloat32SchemaOutOfRange(t *testing.T) {
	s := Float32()
	r := s.SafeParse(float64(math.MaxFloat64))
	if r.Success {
		t.Fatal("expected failure for value out of float32 range")
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Code == IssueInvalidNumber {
			found = true
		}
	}
	if !found {
		t.Fatal("expected invalid_number issue for out-of-range")
	}
}

func TestFloat64Schema(t *testing.T) {
	s := Float64()
	if r := s.SafeParse(float64(math.MaxFloat64)); !r.Success {
		t.Fatal("expected success for max float64")
	}
}

func TestNumberSchemaDefault(t *testing.T) {
	s := Number().Default(42.0)
	r := s.SafeParse(nil)
	if !r.Success || r.Data != 42.0 {
		t.Fatalf("expected 42.0, got %v", r.Data)
	}
}

func TestNumberSchemaCoerce(t *testing.T) {
	s := Number().Coerce(CoerceToNumber)
	r := s.SafeParse("3.14")
	if !r.Success || r.Data != 3.14 {
		t.Fatalf("expected 3.14, got %v", r.Data)
	}
}

func TestNumberSchemaParse(t *testing.T) {
	s := Number()
	v, err := s.Parse(3.14)
	if err != nil {
		t.Fatal(err)
	}
	if v != 3.14 {
		t.Fatalf("expected 3.14, got %v", v)
	}
	_, err = s.Parse("string")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNumberSchemaToNode(t *testing.T) {
	s := Number().Min(0).Max(100).ExclusiveMin(0).ExclusiveMax(100).MultipleOf(5)
	node := s.ToNode()
	if node["kind"] != "number" {
		t.Fatal("expected kind=number")
	}
	if node["min"] != float64(0) {
		t.Fatalf("expected min=0, got %v", node["min"])
	}
	if node["max"] != float64(100) {
		t.Fatalf("expected max=100, got %v", node["max"])
	}
	if node["exclusiveMin"] != float64(0) {
		t.Fatal("missing exclusiveMin")
	}
	if node["exclusiveMax"] != float64(100) {
		t.Fatal("missing exclusiveMax")
	}
	if node["multipleOf"] != float64(5) {
		t.Fatal("missing multipleOf")
	}
}

func TestNumberSchemaToNodeMinimal(t *testing.T) {
	s := Number()
	node := s.ToNode()
	if node["kind"] != "number" {
		t.Fatal("expected kind=number")
	}
	if _, ok := node["min"]; ok {
		t.Fatal("unexpected min")
	}
}

func TestFloat64SchemaToNode(t *testing.T) {
	s := Float64()
	node := s.ToNode()
	if node["kind"] != "float64" {
		t.Fatalf("expected kind=float64, got %v", node["kind"])
	}
}

func TestFloat32SchemaToNode(t *testing.T) {
	s := Float32()
	node := s.ToNode()
	if node["kind"] != "float32" {
		t.Fatalf("expected kind=float32, got %v", node["kind"])
	}
}

func TestNumberSchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := Number().Coerce(CoerceToNumber).Default(0)
	node := s.ToNode()
	if _, ok := node["coerce"]; !ok {
		t.Fatal("expected coerce in node")
	}
	if node["default"] != float64(0) {
		t.Fatal("expected default in node")
	}
}

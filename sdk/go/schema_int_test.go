package anyvali

import (
	"math"
	"testing"
)

func TestIntSchemaSignedAllWidths(t *testing.T) {
	tests := []struct {
		name   string
		schema *IntSchema
		low    int64
		high   int64
	}{
		{"int", Int(), math.MinInt64, math.MaxInt64},
		{"int8", Int8(), math.MinInt8, math.MaxInt8},
		{"int16", Int16(), math.MinInt16, math.MaxInt16},
		{"int32", Int32(), math.MinInt32, math.MaxInt32},
		{"int64", Int64(), math.MinInt64, math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name+" valid low", func(t *testing.T) {
			r := tt.schema.SafeParse(tt.low)
			if !r.Success {
				t.Fatalf("expected success for min %s, got: %v", tt.name, r.Issues)
			}
		})
		t.Run(tt.name+" valid high", func(t *testing.T) {
			r := tt.schema.SafeParse(tt.high)
			if !r.Success {
				t.Fatalf("expected success for max %s, got: %v", tt.name, r.Issues)
			}
		})
	}
}

func TestIntSchemaSignedOverflow(t *testing.T) {
	tests := []struct {
		name   string
		schema *IntSchema
		over   int64
		under  int64
	}{
		{"int8", Int8(), 128, -129},
		{"int16", Int16(), 32768, -32769},
		{"int32", Int32(), 2147483648, -2147483649},
	}
	for _, tt := range tests {
		t.Run(tt.name+" overflow", func(t *testing.T) {
			r := tt.schema.SafeParse(tt.over)
			if r.Success {
				t.Fatalf("expected failure for overflow %s", tt.name)
			}
		})
		t.Run(tt.name+" underflow", func(t *testing.T) {
			r := tt.schema.SafeParse(tt.under)
			if r.Success {
				t.Fatalf("expected failure for underflow %s", tt.name)
			}
		})
	}
}

func TestIntSchemaUnsignedAllWidths(t *testing.T) {
	tests := []struct {
		name   string
		schema *IntSchema
		high   uint64
	}{
		{"uint8", Uint8(), math.MaxUint8},
		{"uint16", Uint16(), math.MaxUint16},
		{"uint32", Uint32(), math.MaxUint32},
		{"uint64", Uint64(), math.MaxUint64},
	}
	for _, tt := range tests {
		t.Run(tt.name+" valid 0", func(t *testing.T) {
			r := tt.schema.SafeParse(int64(0))
			if !r.Success {
				t.Fatalf("expected success for 0 %s, got: %v", tt.name, r.Issues)
			}
		})
	}
}

func TestIntSchemaUnsignedOverflow(t *testing.T) {
	// uint8 overflow
	s := Uint8()
	r := s.SafeParse(int64(256))
	if r.Success {
		t.Fatal("expected failure for uint8 overflow")
	}

	// uint16 overflow
	s2 := Uint16()
	r2 := s2.SafeParse(int64(65536))
	if r2.Success {
		t.Fatal("expected failure for uint16 overflow")
	}
}

func TestIntSchemaUnsignedNegative(t *testing.T) {
	s := Uint8()
	r := s.SafeParse(int64(-1))
	if r.Success {
		t.Fatal("expected failure for negative uint")
	}
}

func TestIntSchemaInvalidType(t *testing.T) {
	s := Int()
	tests := []any{"hello", true, nil, []any{}, map[string]any{}}
	for _, v := range tests {
		r := s.SafeParse(v)
		if r.Success {
			t.Fatalf("expected failure for %v (%T)", v, v)
		}
	}
}

func TestIntSchemaFromFloat(t *testing.T) {
	s := Int()
	// Integer-valued float64 should work
	r := s.SafeParse(float64(42))
	if !r.Success {
		t.Fatal("expected success for integer-valued float64")
	}
	// Fractional float64 should fail
	r = s.SafeParse(float64(3.14))
	if r.Success {
		t.Fatal("expected failure for fractional float64")
	}
}

func TestIntSchemaMin(t *testing.T) {
	s := Int().Min(5)
	if r := s.SafeParse(int64(4)); r.Success {
		t.Fatal("expected failure for < min")
	}
	if r := s.SafeParse(int64(5)); !r.Success {
		t.Fatal("expected success for == min")
	}
}

func TestIntSchemaMax(t *testing.T) {
	s := Int().Max(10)
	if r := s.SafeParse(int64(11)); r.Success {
		t.Fatal("expected failure for > max")
	}
	if r := s.SafeParse(int64(10)); !r.Success {
		t.Fatal("expected success for == max")
	}
}

func TestIntSchemaExclusiveMin(t *testing.T) {
	s := Int().ExclusiveMin(5)
	if r := s.SafeParse(int64(5)); r.Success {
		t.Fatal("expected failure for == exclusiveMin")
	}
	if r := s.SafeParse(int64(6)); !r.Success {
		t.Fatal("expected success for > exclusiveMin")
	}
}

func TestIntSchemaExclusiveMax(t *testing.T) {
	s := Int().ExclusiveMax(10)
	if r := s.SafeParse(int64(10)); r.Success {
		t.Fatal("expected failure for == exclusiveMax")
	}
	if r := s.SafeParse(int64(9)); !r.Success {
		t.Fatal("expected success for < exclusiveMax")
	}
}

func TestIntSchemaMultipleOf(t *testing.T) {
	s := Int().MultipleOf(3)
	if r := s.SafeParse(int64(9)); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse(int64(10)); r.Success {
		t.Fatal("expected failure")
	}
}

func TestIntSchemaMultipleOfZero(t *testing.T) {
	s := Int().MultipleOf(0)
	if r := s.SafeParse(int64(42)); !r.Success {
		t.Fatal("expected success when multipleOf is 0")
	}
}

func TestUintSchemaMin(t *testing.T) {
	s := Uint64().Min(5)
	if r := s.SafeParse(int64(4)); r.Success {
		t.Fatal("expected failure for < min")
	}
	if r := s.SafeParse(int64(5)); !r.Success {
		t.Fatal("expected success for == min")
	}
}

func TestUintSchemaMax(t *testing.T) {
	s := Uint64().Max(10)
	if r := s.SafeParse(int64(11)); r.Success {
		t.Fatal("expected failure for > max")
	}
	if r := s.SafeParse(int64(10)); !r.Success {
		t.Fatal("expected success for == max")
	}
}

func TestUintSchemaExclusiveMin(t *testing.T) {
	s := Uint64().ExclusiveMin(5)
	if r := s.SafeParse(int64(5)); r.Success {
		t.Fatal("expected failure for == exclusiveMin")
	}
	if r := s.SafeParse(int64(6)); !r.Success {
		t.Fatal("expected success for > exclusiveMin")
	}
}

func TestUintSchemaExclusiveMax(t *testing.T) {
	s := Uint64().ExclusiveMax(10)
	if r := s.SafeParse(int64(10)); r.Success {
		t.Fatal("expected failure for == exclusiveMax")
	}
	if r := s.SafeParse(int64(9)); !r.Success {
		t.Fatal("expected success for < exclusiveMax")
	}
}

func TestUintSchemaMultipleOf(t *testing.T) {
	s := Uint64().MultipleOf(3)
	if r := s.SafeParse(int64(9)); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse(int64(10)); r.Success {
		t.Fatal("expected failure")
	}
}

func TestUintSchemaMultipleOfZero(t *testing.T) {
	s := Uint64().MultipleOf(0)
	if r := s.SafeParse(int64(42)); !r.Success {
		t.Fatal("expected success when multipleOf is 0")
	}
}

func TestIntSchemaReturnAsInt64WhenFits(t *testing.T) {
	s := Uint64()
	r := s.SafeParse(int64(42))
	if !r.Success {
		t.Fatal("expected success")
	}
	if _, ok := r.Data.(int64); !ok {
		t.Fatalf("expected int64 result, got %T", r.Data)
	}
}

func TestIntSchemaDefault(t *testing.T) {
	s := Int().Default(42)
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
}

func TestIntSchemaCoerce(t *testing.T) {
	s := Int().Coerce(CoerceToInt)
	r := s.SafeParse("99")
	if !r.Success {
		t.Fatalf("expected success, got: %v", r.Issues)
	}
	if r.Data != int64(99) {
		t.Fatalf("expected 99, got %v", r.Data)
	}
}

func TestIntSchemaParse(t *testing.T) {
	s := Int()
	v, err := s.Parse(int64(42))
	if err != nil {
		t.Fatal(err)
	}
	if v != int64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
	_, err = s.Parse("string")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIntSchemaToNode(t *testing.T) {
	s := Int().Min(0).Max(100).ExclusiveMin(0).ExclusiveMax(100).MultipleOf(5)
	node := s.ToNode()
	if node["kind"] != "int" {
		t.Fatal("expected kind=int")
	}
	if node["min"] != int64(0) {
		t.Fatal("missing min")
	}
	if node["max"] != int64(100) {
		t.Fatal("missing max")
	}
	if node["exclusiveMin"] != int64(0) {
		t.Fatal("missing exclusiveMin")
	}
	if node["exclusiveMax"] != int64(100) {
		t.Fatal("missing exclusiveMax")
	}
	if node["multipleOf"] != int64(5) {
		t.Fatal("missing multipleOf")
	}
}

func TestIntSchemaToNodeMinimal(t *testing.T) {
	s := Int()
	node := s.ToNode()
	if _, ok := node["min"]; ok {
		t.Fatal("unexpected min")
	}
}

func TestIntSchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := Int().Coerce(CoerceToInt).Default(0)
	node := s.ToNode()
	if _, ok := node["coerce"]; !ok {
		t.Fatal("expected coerce in node")
	}
	if _, ok := node["default"]; !ok {
		t.Fatal("expected default in node")
	}
}

func TestUint8ToNode(t *testing.T) {
	s := Uint8()
	node := s.ToNode()
	if node["kind"] != "uint8" {
		t.Fatalf("expected kind=uint8, got %v", node["kind"])
	}
}

func TestUint16ToNode(t *testing.T) {
	s := Uint16()
	node := s.ToNode()
	if node["kind"] != "uint16" {
		t.Fatalf("expected kind=uint16, got %v", node["kind"])
	}
}

func TestUint32ToNode(t *testing.T) {
	s := Uint32()
	node := s.ToNode()
	if node["kind"] != "uint32" {
		t.Fatalf("expected kind=uint32, got %v", node["kind"])
	}
}

func TestUint64ToNode(t *testing.T) {
	s := Uint64()
	node := s.ToNode()
	if node["kind"] != "uint64" {
		t.Fatalf("expected kind=uint64, got %v", node["kind"])
	}
}

func TestInt8ToNode(t *testing.T) {
	s := Int8()
	node := s.ToNode()
	if node["kind"] != "int8" {
		t.Fatalf("expected kind=int8, got %v", node["kind"])
	}
}

func TestInt16ToNode(t *testing.T) {
	s := Int16()
	node := s.ToNode()
	if node["kind"] != "int16" {
		t.Fatalf("expected kind=int16, got %v", node["kind"])
	}
}

func TestInt32ToNode(t *testing.T) {
	s := Int32()
	node := s.ToNode()
	if node["kind"] != "int32" {
		t.Fatalf("expected kind=int32, got %v", node["kind"])
	}
}

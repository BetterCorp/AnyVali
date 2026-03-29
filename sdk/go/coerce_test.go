package anyvali

import "testing"

func TestCoerceToIntFromString(t *testing.T) {
	v, err := coerceToInt("42")
	if err != nil {
		t.Fatal(err)
	}
	if v != int64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestCoerceToIntFromStringNegative(t *testing.T) {
	v, err := coerceToInt("-10")
	if err != nil {
		t.Fatal(err)
	}
	if v != int64(-10) {
		t.Fatalf("expected -10, got %v", v)
	}
}

func TestCoerceToIntFromStringInvalid(t *testing.T) {
	_, err := coerceToInt("abc")
	if err == nil {
		t.Fatal("expected error for non-numeric string")
	}
}

func TestCoerceToIntFromStringFloat(t *testing.T) {
	_, err := coerceToInt("3.14")
	if err == nil {
		t.Fatal("expected error for float string")
	}
}

func TestCoerceToIntFromNonString(t *testing.T) {
	_, err := coerceToInt(42)
	if err == nil {
		t.Fatal("expected error for int input")
	}
	_, err = coerceToInt(true)
	if err == nil {
		t.Fatal("expected error for bool input")
	}
	_, err = coerceToInt(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestCoerceToNumberFromString(t *testing.T) {
	v, err := coerceToNumber("3.14")
	if err != nil {
		t.Fatal(err)
	}
	if v != 3.14 {
		t.Fatalf("expected 3.14, got %v", v)
	}
}

func TestCoerceToNumberFromStringInt(t *testing.T) {
	v, err := coerceToNumber("42")
	if err != nil {
		t.Fatal(err)
	}
	if v != 42.0 {
		t.Fatalf("expected 42.0, got %v", v)
	}
}

func TestCoerceToNumberFromStringInvalid(t *testing.T) {
	_, err := coerceToNumber("abc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCoerceToNumberFromNonString(t *testing.T) {
	_, err := coerceToNumber(42)
	if err == nil {
		t.Fatal("expected error for non-string input")
	}
}

func TestCoerceToBoolFromStringTrue(t *testing.T) {
	v, err := coerceToBool("true")
	if err != nil {
		t.Fatal(err)
	}
	if v != true {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestCoerceToBoolFromStringFalse(t *testing.T) {
	v, err := coerceToBool("false")
	if err != nil {
		t.Fatal(err)
	}
	if v != false {
		t.Fatalf("expected false, got %v", v)
	}
}

func TestCoerceToBoolFromStringOnes(t *testing.T) {
	v, err := coerceToBool("1")
	if err != nil {
		t.Fatal(err)
	}
	if v != true {
		t.Fatalf("expected true for '1', got %v", v)
	}
	v, err = coerceToBool("0")
	if err != nil {
		t.Fatal(err)
	}
	if v != false {
		t.Fatalf("expected false for '0', got %v", v)
	}
}

func TestCoerceToBoolFromStringInvalid(t *testing.T) {
	_, err := coerceToBool("maybe")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCoerceToBoolFromNonString(t *testing.T) {
	_, err := coerceToBool(42)
	if err == nil {
		t.Fatal("expected error for non-string")
	}
}

func TestCoerceTrimString(t *testing.T) {
	v, err := coerceTrimString("  hello  ")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestCoerceTrimStringNoSpaces(t *testing.T) {
	v, err := coerceTrimString("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestCoerceTrimStringEmpty(t *testing.T) {
	v, err := coerceTrimString("   ")
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Fatalf("expected empty string, got %q", v)
	}
}

func TestCoerceTrimStringNonString(t *testing.T) {
	_, err := coerceTrimString(42)
	if err == nil {
		t.Fatal("expected error for non-string")
	}
}

func TestCoerceLowerString(t *testing.T) {
	v, err := coerceLowerString("HELLO")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestCoerceLowerStringAlreadyLower(t *testing.T) {
	v, err := coerceLowerString("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestCoerceLowerStringNonString(t *testing.T) {
	_, err := coerceLowerString(42)
	if err == nil {
		t.Fatal("expected error for non-string")
	}
}

func TestCoerceUpperString(t *testing.T) {
	v, err := coerceUpperString("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "HELLO" {
		t.Fatalf("expected 'HELLO', got %v", v)
	}
}

func TestCoerceUpperStringNonString(t *testing.T) {
	_, err := coerceUpperString(42)
	if err == nil {
		t.Fatal("expected error for non-string")
	}
}

func TestApplyCoercionAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		coercion CoercionType
		input    any
		expected any
	}{
		{"int", CoerceToInt, "42", int64(42)},
		{"number", CoerceToNumber, "3.14", 3.14},
		{"bool", CoerceToBool, "true", true},
		{"trim", CoerceTrim, "  x  ", "x"},
		{"lower", CoerceLower, "ABC", "abc"},
		{"upper", CoerceUpper, "abc", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := applyCoercion(tt.input, tt.coercion)
			if err != nil {
				t.Fatal(err)
			}
			if v != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, v)
			}
		})
	}
}

func TestApplyCoercionUnknownType(t *testing.T) {
	_, err := applyCoercion("hello", CoercionType("unknown_coercion"))
	if err == nil {
		t.Fatal("expected error for unknown coercion type")
	}
}

func TestApplyCoercionFailurePaths(t *testing.T) {
	tests := []struct {
		name     string
		coercion CoercionType
		input    any
	}{
		{"int from non-string", CoerceToInt, 42},
		{"number from non-string", CoerceToNumber, true},
		{"bool from non-string", CoerceToBool, 3.14},
		{"trim from non-string", CoerceTrim, 42},
		{"lower from non-string", CoerceLower, true},
		{"upper from non-string", CoerceUpper, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := applyCoercion(tt.input, tt.coercion)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

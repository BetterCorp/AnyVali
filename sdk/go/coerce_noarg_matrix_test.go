package anyvali

import "testing"

// Full FROM-STRING coercion matrix exercised through the no-arg / default
// Coerce() ergonomic added for cross-SDK parity (like Zod's z.coerce.number()).
//
//   Int().Coerce()    -> string coerced to int, target inferred from kind
//   Number().Coerce() -> string coerced to number, target inferred from kind
//   Bool().Coerce()   -> string coerced to bool, target inferred from kind
//
// The matrix below is the canonical cross-language contract. A REJECT row must
// surface a coercion_failed issue (not silently pass nor a different code).

func assertCoerceFails(t *testing.T, label string, r ParseResult) {
	t.Helper()
	if r.Success {
		t.Fatalf("%s: expected coercion failure, got success with data %v (%T)", label, r.Data, r.Data)
	}
	found := false
	for _, iss := range r.Issues {
		if iss.Code == IssueCoercionFailed {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("%s: expected a %q issue, got issues: %v", label, IssueCoercionFailed, r.Issues)
	}
}

func TestNoArgCoerceIntFromStringMatrix(t *testing.T) {
	accept := []struct {
		in   string
		want int64
	}{
		{"42", 42},
		{"  42  ", 42},
		{"-7", -7},
	}
	for _, c := range accept {
		s := Int().Coerce()
		r := s.SafeParse(c.in)
		if !r.Success {
			t.Fatalf("int ACCEPT %q: expected success, got issues: %v", c.in, r.Issues)
		}
		if r.Data != c.want {
			t.Fatalf("int ACCEPT %q: expected %d, got %v (%T)", c.in, c.want, r.Data, r.Data)
		}
	}

	reject := []string{"3.14", "0x10", "1_000", "+5", "Infinity", "", "abc"}
	for _, in := range reject {
		s := Int().Coerce()
		assertCoerceFails(t, "int REJECT "+in, s.SafeParse(in))
	}
}

func TestNoArgCoerceNumberFromStringMatrix(t *testing.T) {
	accept := []struct {
		in   string
		want float64
	}{
		{"3.14", 3.14},
		{"-1.5e3", -1500},
		{"  2  ", 2},
		{"0", 0},
	}
	for _, c := range accept {
		s := Number().Coerce()
		r := s.SafeParse(c.in)
		if !r.Success {
			t.Fatalf("number ACCEPT %q: expected success, got issues: %v", c.in, r.Issues)
		}
		if r.Data != c.want {
			t.Fatalf("number ACCEPT %q: expected %v, got %v (%T)", c.in, c.want, r.Data, r.Data)
		}
	}

	reject := []string{"0x10", "Infinity", "NaN", "", "1_000", "abc"}
	for _, in := range reject {
		s := Number().Coerce()
		assertCoerceFails(t, "number REJECT "+in, s.SafeParse(in))
	}
}

func TestNoArgCoerceBoolFromStringMatrix(t *testing.T) {
	accept := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"  TrUe  ", true},
	}
	for _, c := range accept {
		s := Bool().Coerce()
		r := s.SafeParse(c.in)
		if !r.Success {
			t.Fatalf("bool ACCEPT %q: expected success, got issues: %v", c.in, r.Issues)
		}
		if r.Data != c.want {
			t.Fatalf("bool ACCEPT %q: expected %v, got %v (%T)", c.in, c.want, r.Data, r.Data)
		}
	}

	reject := []string{"yes", "no", "on", "off", "t", "f", "2", ""}
	for _, in := range reject {
		s := Bool().Coerce()
		assertCoerceFails(t, "bool REJECT "+in, s.SafeParse(in))
	}
}

// The no-arg form on the sized integer/float schemas must infer the right
// portable coercion (int vs number) from the schema kind.
func TestNoArgCoerceInfersKind(t *testing.T) {
	if r := Int8().Coerce().SafeParse("12"); !r.Success || r.Data != int64(12) {
		t.Fatalf("Int8().Coerce() of \"12\": expected 12, got success=%v data=%v", r.Success, r.Data)
	}
	if r := Uint16().Coerce().SafeParse("300"); !r.Success || r.Data != int64(300) {
		t.Fatalf("Uint16().Coerce() of \"300\": expected 300, got success=%v data=%v", r.Success, r.Data)
	}
	if r := Float64().Coerce().SafeParse("1.25"); !r.Success || r.Data != 1.25 {
		t.Fatalf("Float64().Coerce() of \"1.25\": expected 1.25, got success=%v data=%v", r.Success, r.Data)
	}
	if r := Float32().Coerce().SafeParse("1.5"); !r.Success || r.Data != 1.5 {
		t.Fatalf("Float32().Coerce() of \"1.5\": expected 1.5, got success=%v data=%v", r.Success, r.Data)
	}
	// A float kind must accept a decimal that the int inference would reject,
	// proving the kind is what drives target selection.
	if r := Int().Coerce().SafeParse("1.25"); r.Success {
		t.Fatalf("Int().Coerce() of \"1.25\": expected failure, got data %v", r.Data)
	}
}

// String transform coercions (string kind): trim, lower, upper; chainable.
// These remain explicit (no implicit string-to-self target).
func TestNoArgCoerceStringTransformsExplicit(t *testing.T) {
	if r := String().Coerce(CoerceTrim).SafeParse("  hi  "); !r.Success || r.Data != "hi" {
		t.Fatalf("trim: expected \"hi\", got success=%v data=%v", r.Success, r.Data)
	}
	if r := String().Coerce(CoerceLower).SafeParse("HI"); !r.Success || r.Data != "hi" {
		t.Fatalf("lower: expected \"hi\", got success=%v data=%v", r.Success, r.Data)
	}
	if r := String().Coerce(CoerceUpper).SafeParse("hi"); !r.Success || r.Data != "HI" {
		t.Fatalf("upper: expected \"HI\", got success=%v data=%v", r.Success, r.Data)
	}
	// Chainable, applied left to right: trim then upper.
	if r := String().Coerce(CoerceTrim, CoerceUpper).SafeParse("  hi  "); !r.Success || r.Data != "HI" {
		t.Fatalf("trim+upper variadic: expected \"HI\", got success=%v data=%v", r.Success, r.Data)
	}
	if r := String().Coerce(CoerceTrim).Coerce(CoerceLower).SafeParse("  HI  "); !r.Success || r.Data != "hi" {
		t.Fatalf("trim then lower chained: expected \"hi\", got success=%v data=%v", r.Success, r.Data)
	}
	// No-arg on a string schema is a no-op (no implicit target).
	if r := String().Coerce().SafeParse("plain"); !r.Success || r.Data != "plain" {
		t.Fatalf("string no-arg no-op: expected \"plain\", got success=%v data=%v", r.Success, r.Data)
	}
}

// Variadic typed form must still append in order (parity with prior call sites).
func TestVariadicCoerceTypedStillWorks(t *testing.T) {
	if r := Int().Coerce(CoerceToInt).SafeParse("5"); !r.Success || r.Data != int64(5) {
		t.Fatalf("Int().Coerce(CoerceToInt): expected 5, got success=%v data=%v", r.Success, r.Data)
	}
	if r := Number().Coerce(CoerceToNumber).SafeParse("2.5"); !r.Success || r.Data != 2.5 {
		t.Fatalf("Number().Coerce(CoerceToNumber): expected 2.5, got success=%v data=%v", r.Success, r.Data)
	}
	if r := Bool().Coerce(CoerceToBool).SafeParse("true"); !r.Success || r.Data != true {
		t.Fatalf("Bool().Coerce(CoerceToBool): expected true, got success=%v data=%v", r.Success, r.Data)
	}
}

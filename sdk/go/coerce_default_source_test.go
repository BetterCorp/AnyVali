package anyvali

import "testing"

// Regression tests for the "default / no-explicit-source" form of enabling
// coercion. The only portable coercion source is "string", so enabling
// coercion on a numeric/bool schema with the kind that matches the target type
// (and no separate `from`/source) MUST coerce a string input rather than
// silently no-op into an invalid_type failure.
//
// In the Go SDK the idiomatic default form is Coerce(CoerceToNumber) /
// Coerce(CoerceToInt) / Coerce(CoerceToBool): the source is implicitly the
// only portable one ("string"). These mirror the JS `.coerce({})` regression
// cases across SDKs.

func TestNumberDefaultCoerceFromString(t *testing.T) {
	s := Number().Coerce(CoerceToNumber)
	r := s.SafeParse("3.14")
	if !r.Success {
		t.Fatalf("expected success coercing \"3.14\" -> 3.14, got issues: %v", r.Issues)
	}
	if r.Data != 3.14 {
		t.Fatalf("expected 3.14, got %v (%T)", r.Data, r.Data)
	}
}

func TestIntDefaultCoerceFromString(t *testing.T) {
	s := Int().Coerce(CoerceToInt)
	r := s.SafeParse("42")
	if !r.Success {
		t.Fatalf("expected success coercing \"42\" -> 42, got issues: %v", r.Issues)
	}
	if r.Data != int64(42) {
		t.Fatalf("expected 42, got %v (%T)", r.Data, r.Data)
	}
}

func TestBoolDefaultCoerceFromString(t *testing.T) {
	s := Bool().Coerce(CoerceToBool)

	rTrue := s.SafeParse("true")
	if !rTrue.Success {
		t.Fatalf("expected success coercing \"true\" -> true, got issues: %v", rTrue.Issues)
	}
	if rTrue.Data != true {
		t.Fatalf("expected true, got %v (%T)", rTrue.Data, rTrue.Data)
	}

	rFalse := s.SafeParse("false")
	if !rFalse.Success {
		t.Fatalf("expected success coercing \"false\" -> false, got issues: %v", rFalse.Issues)
	}
	if rFalse.Data != false {
		t.Fatalf("expected false, got %v (%T)", rFalse.Data, rFalse.Data)
	}
}

func TestObjectNumericFieldsDefaultCoerce(t *testing.T) {
	s := Object(map[string]Schema{
		"lumpSum":              Number().Coerce(CoerceToNumber),
		"monthlyContributions": Number().Coerce(CoerceToNumber),
		"investmentTerm":       Int().Coerce(CoerceToInt),
	})

	input := map[string]any{
		"lumpSum":              "1000000",
		"monthlyContributions": "1000",
		"investmentTerm":       "20",
	}

	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success coercing string numeric fields, got issues: %v", r.Issues)
	}

	out, ok := r.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", r.Data)
	}
	if out["lumpSum"] != 1000000.0 {
		t.Fatalf("lumpSum: expected 1000000.0, got %v (%T)", out["lumpSum"], out["lumpSum"])
	}
	if out["monthlyContributions"] != 1000.0 {
		t.Fatalf("monthlyContributions: expected 1000.0, got %v (%T)", out["monthlyContributions"], out["monthlyContributions"])
	}
	if out["investmentTerm"] != int64(20) {
		t.Fatalf("investmentTerm: expected 20, got %v (%T)", out["investmentTerm"], out["investmentTerm"])
	}
}

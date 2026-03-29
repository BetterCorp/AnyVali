package anyvali

import "testing"

func TestIsAbsent(t *testing.T) {
	if !isAbsent(absentValue) {
		t.Fatal("expected absent to be absent")
	}
	if isAbsent(nil) {
		t.Fatal("nil should not be absent")
	}
	if isAbsent("hello") {
		t.Fatal("string should not be absent")
	}
	if isAbsent(42) {
		t.Fatal("int should not be absent")
	}
}

func TestApplyDefaultAbsentWithDefault(t *testing.T) {
	result := applyDefault(absentValue, "fallback", true)
	if result != "fallback" {
		t.Fatalf("expected 'fallback', got %v", result)
	}
}

func TestApplyDefaultAbsentWithoutDefault(t *testing.T) {
	result := applyDefault(absentValue, nil, false)
	if result != absentValue {
		t.Fatalf("expected absent value to pass through, got %v", result)
	}
}

func TestApplyDefaultPresentValue(t *testing.T) {
	result := applyDefault("hello", "fallback", true)
	if result != "hello" {
		t.Fatalf("expected 'hello', got %v", result)
	}
}

func TestApplyDefaultNilNotAbsent(t *testing.T) {
	result := applyDefault(nil, "fallback", true)
	if result != nil {
		t.Fatalf("expected nil (not absent), got %v", result)
	}
}

func TestDefaultInPipeline(t *testing.T) {
	// Test that defaults work through the full pipeline
	s := String().Default("default_val")

	// nil triggers default
	r := s.SafeParse(nil)
	if !r.Success || r.Data != "default_val" {
		t.Fatalf("expected default_val for nil input, got %v (success=%v)", r.Data, r.Success)
	}

	// Present value should not trigger default
	r = s.SafeParse("real")
	if !r.Success || r.Data != "real" {
		t.Fatalf("expected 'real', got %v", r.Data)
	}
}

func TestDefaultNumberSchema(t *testing.T) {
	s := Number().Default(42.0)
	r := s.SafeParse(nil)
	if !r.Success || r.Data != 42.0 {
		t.Fatalf("expected 42.0, got %v (success=%v)", r.Data, r.Success)
	}
}

func TestDefaultBoolSchema(t *testing.T) {
	s := Bool().Default(true)
	r := s.SafeParse(nil)
	if !r.Success || r.Data != true {
		t.Fatalf("expected true, got %v (success=%v)", r.Data, r.Success)
	}
}

func TestDefaultIntSchema(t *testing.T) {
	s := Int().Default(7)
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success with default, got issues: %v", r.Issues)
	}
}

func TestDefaultArraySchema(t *testing.T) {
	s := Array(String()).Default([]any{"a", "b"})
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success with default, got issues: %v", r.Issues)
	}
}

func TestDefaultObjectSchema(t *testing.T) {
	s := Object(map[string]Schema{
		"name": String(),
	}).Default(map[string]any{"name": "default"})
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatalf("expected success with default, got issues: %v", r.Issues)
	}
}

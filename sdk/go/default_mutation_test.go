package anyvali

import "testing"

// Repro for AVV-007: mutable default values shared between parses.
// The default value is stored by reference and handed to validation. For
// schemas that pass values through (Any/Unknown items, allowed unknown keys),
// the materialized result aliases the stored default, so mutating one parse
// result corrupts the default for the next parse.

func TestDefaultNotSharedArrayOfAny(t *testing.T) {
	s := Array(Any()).Default([]any{map[string]any{"n": float64(0)}})

	r1, err := s.Parse(nil)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}
	m1 := r1.([]any)[0].(map[string]any)
	m1["n"] = float64(99)

	r2, err := s.Parse(nil)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}
	m2 := r2.([]any)[0].(map[string]any)
	if m2["n"] != float64(0) {
		t.Errorf("default mutated across parses: got n=%v, want 0", m2["n"])
	}
}

func TestDefaultNotSharedObjectAllowedKey(t *testing.T) {
	s := Object(map[string]Schema{}).
		UnknownKeys(Allow).
		Default(map[string]any{"tags": []any{"orig"}})

	r1, err := s.Parse(nil)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}
	// Mutate the slice element in place: if the slice aliases the stored
	// default's backing array, this corrupts the default for later parses.
	r1.(map[string]any)["tags"].([]any)[0] = "MUTATED"

	r2, err := s.Parse(nil)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}
	tags2 := r2.(map[string]any)["tags"].([]any)
	if tags2[0] != "orig" {
		t.Errorf("default mutated across parses: got tags[0]=%v, want orig", tags2[0])
	}
}

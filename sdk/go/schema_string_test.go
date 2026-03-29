package anyvali

import (
	"testing"
)

func TestStringSchemaValidateNonString(t *testing.T) {
	s := String()
	tests := []any{42, 3.14, true, nil, []any{}, map[string]any{}}
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

func TestStringSchemaMinLength(t *testing.T) {
	s := String().MinLength(3)
	if r := s.SafeParse("ab"); r.Success {
		t.Fatal("expected failure for too short")
	}
	if r := s.SafeParse("abc"); !r.Success {
		t.Fatal("expected success for exact length")
	}
	if r := s.SafeParse("abcd"); !r.Success {
		t.Fatal("expected success for longer")
	}
}

func TestStringSchemaMaxLength(t *testing.T) {
	s := String().MaxLength(3)
	if r := s.SafeParse("abcd"); r.Success {
		t.Fatal("expected failure for too long")
	}
	if r := s.SafeParse("abc"); !r.Success {
		t.Fatal("expected success for exact length")
	}
	if r := s.SafeParse("ab"); !r.Success {
		t.Fatal("expected success for shorter")
	}
}

func TestStringSchemaMinAndMaxLength(t *testing.T) {
	s := String().MinLength(2).MaxLength(4)
	if r := s.SafeParse("a"); r.Success {
		t.Fatal("expected failure")
	}
	if r := s.SafeParse("abcde"); r.Success {
		t.Fatal("expected failure")
	}
	if r := s.SafeParse("abc"); !r.Success {
		t.Fatal("expected success")
	}
}

func TestStringSchemaUnicodeLength(t *testing.T) {
	s := String().MinLength(2).MaxLength(3)
	// Unicode characters - each emoji is 1 rune
	if r := s.SafeParse("\u00e9\u00e9"); !r.Success {
		t.Fatal("expected success for 2 unicode runes")
	}
}

func TestStringSchemaPattern(t *testing.T) {
	s := String().Pattern(`^[a-z]+$`)
	if r := s.SafeParse("hello"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("Hello"); r.Success {
		t.Fatal("expected failure for uppercase")
	}
	if r := s.SafeParse(""); r.Success {
		t.Fatal("expected failure for empty string")
	}
}

func TestStringSchemaStartsWith(t *testing.T) {
	s := String().StartsWith("pre")
	if r := s.SafeParse("prefix"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("other"); r.Success {
		t.Fatal("expected failure")
	}
	r := s.SafeParse("other")
	if r.Issues[0].Code != IssueInvalidString {
		t.Fatalf("expected invalid_string, got %s", r.Issues[0].Code)
	}
}

func TestStringSchemaEndsWith(t *testing.T) {
	s := String().EndsWith("fix")
	if r := s.SafeParse("suffix"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("other"); r.Success {
		t.Fatal("expected failure")
	}
}

func TestStringSchemaIncludes(t *testing.T) {
	s := String().Includes("mid")
	if r := s.SafeParse("amidst"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("other"); r.Success {
		t.Fatal("expected failure")
	}
}

func TestStringSchemaFormat(t *testing.T) {
	s := String().Format("email")
	if r := s.SafeParse("user@example.com"); !r.Success {
		t.Fatal("expected success")
	}
	if r := s.SafeParse("not-email"); r.Success {
		t.Fatal("expected failure")
	}
	r := s.SafeParse("not-email")
	if r.Issues[0].Code != IssueInvalidString {
		t.Fatalf("expected invalid_string, got %s", r.Issues[0].Code)
	}
}

func TestStringSchemaMultipleConstraintsFail(t *testing.T) {
	// Both minLength and pattern should fail - we get multiple issues
	s := String().MinLength(10).Pattern(`^\d+$`)
	r := s.SafeParse("abc")
	if r.Success {
		t.Fatal("expected failure")
	}
	if len(r.Issues) < 2 {
		t.Fatalf("expected at least 2 issues, got %d", len(r.Issues))
	}
}

func TestStringSchemaDefault(t *testing.T) {
	s := String().Default("fallback")
	r := s.SafeParse(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "fallback" {
		t.Fatalf("expected 'fallback', got %v", r.Data)
	}
}

func TestStringSchemaCoerce(t *testing.T) {
	s := String().Coerce(CoerceTrim).Coerce(CoerceLower)
	r := s.SafeParse("  HELLO  ")
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestStringSchemaParse(t *testing.T) {
	s := String()
	v, err := s.Parse("hello")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}

	_, err = s.Parse(42)
	if err == nil {
		t.Fatal("expected error for non-string")
	}
}

func TestStringSchemaToNode(t *testing.T) {
	s := String().MinLength(1).MaxLength(10).Pattern(`\d+`).StartsWith("a").EndsWith("z").Includes("m").Format("email")
	node := s.ToNode()

	if node["kind"] != "string" {
		t.Fatal("expected kind=string")
	}
	if node["minLength"] != 1 {
		t.Fatalf("expected minLength=1, got %v", node["minLength"])
	}
	if node["maxLength"] != 10 {
		t.Fatalf("expected maxLength=10, got %v", node["maxLength"])
	}
	if node["pattern"] != `\d+` {
		t.Fatalf("expected pattern, got %v", node["pattern"])
	}
	if node["startsWith"] != "a" {
		t.Fatalf("expected startsWith=a, got %v", node["startsWith"])
	}
	if node["endsWith"] != "z" {
		t.Fatalf("expected endsWith=z, got %v", node["endsWith"])
	}
	if node["includes"] != "m" {
		t.Fatalf("expected includes=m, got %v", node["includes"])
	}
	if node["format"] != "email" {
		t.Fatalf("expected format=email, got %v", node["format"])
	}
}

func TestStringSchemaToNodeMinimal(t *testing.T) {
	s := String()
	node := s.ToNode()
	if node["kind"] != "string" {
		t.Fatal("expected kind=string")
	}
	if _, ok := node["minLength"]; ok {
		t.Fatal("unexpected minLength")
	}
	if _, ok := node["maxLength"]; ok {
		t.Fatal("unexpected maxLength")
	}
}

func TestStringSchemaToNodeWithCoerceAndDefault(t *testing.T) {
	s := String().Coerce(CoerceTrim).Default("x")
	node := s.ToNode()
	if _, ok := node["coerce"]; !ok {
		t.Fatal("expected coerce in node")
	}
	if node["default"] != "x" {
		t.Fatalf("expected default='x', got %v", node["default"])
	}
}

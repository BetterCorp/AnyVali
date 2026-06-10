package anyvali

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. Invalid Regex Panic (CWE-20) - regexp.MustCompile panics on invalid patterns
// ---------------------------------------------------------------------------

func TestSecurity_CWE20_InvalidRegexHandledGracefully(t *testing.T) {
	// Invalid regex patterns should not panic -- they should produce
	// a validation failure instead.
	s := String().Pattern("(")
	r := s.SafeParse("abc")
	if r.Success {
		t.Fatal("expected failure for invalid regex pattern")
	}
	if len(r.Issues) == 0 || r.Issues[0].Code != IssueInvalidString {
		t.Fatalf("expected invalid_string issue, got %#v", r.Issues)
	}
}

func TestSecurity_CWE20_InvalidRegexBracketHandledGracefully(t *testing.T) {
	s := String().Pattern("[")
	r := s.SafeParse("abc")
	if r.Success {
		t.Fatal("expected failure for invalid regex pattern '['")
	}
	if len(r.Issues) == 0 || r.Issues[0].Code != IssueInvalidString {
		t.Fatalf("expected invalid_string issue, got %#v", r.Issues)
	}
}

func TestSecurity_CVE_2016_4055_ReDoS_RE2_Immune(t *testing.T) {
	// Go's RE2-based engine is immune to catastrophic backtracking.
	// This pattern is dangerous in PCRE engines but safe in RE2.
	s := String().Pattern("(a+)+$")

	// Should compile without panic and validate normally
	r := s.SafeParse("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if !r.Success {
		t.Fatal("expected success for matching string")
	}

	// Non-matching input should fail quickly (no exponential backtracking)
	r = s.SafeParse("aaaaaaaaaaaaaaaaaaaaaaaaaaaaab")
	if r.Success {
		t.Fatal("expected failure for non-matching string")
	}
}

func TestSecurity_CWE20_ValidRegexPatterns(t *testing.T) {
	// Ensure well-formed patterns do not panic and work correctly
	patterns := []struct {
		pattern string
		match   string
		noMatch string
	}{
		{`^\d+$`, "12345", "abc"},
		{`^[a-z]+$`, "hello", "HELLO"},
		{`^.{1,10}$`, "short", "this string is too long for the pattern"},
	}

	for _, p := range patterns {
		t.Run(p.pattern, func(t *testing.T) {
			s := String().Pattern(p.pattern)
			if r := s.SafeParse(p.match); !r.Success {
				t.Errorf("expected %q to match pattern %q", p.match, p.pattern)
			}
			if r := s.SafeParse(p.noMatch); r.Success {
				t.Errorf("expected %q to NOT match pattern %q", p.noMatch, p.pattern)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Recursive $ref DoS (CVE-2003-1564 class)
// ---------------------------------------------------------------------------

func TestSecurity_CVE2003_1564_CircularRefSelfReference(t *testing.T) {
	// A schema that references itself: root is a ref to its own definition,
	// and the definition is also a ref to itself. The importer should handle
	// this without infinite recursion because importRefSchema resolves one
	// level of indirection only.
	doc := &Document{
		AnyvaliVersion: "1",
		SchemaVersion:  "1",
		Root: map[string]any{
			"kind": "ref",
			"ref":  "#/definitions/Self",
		},
		Definitions: map[string]map[string]any{
			"Self": {
				"kind": "ref",
				"ref":  "#/definitions/Self",
			},
		},
	}

	// Import should not hang or stack overflow
	schema, err := Import(doc)
	if err != nil {
		// An error is acceptable -- it means the library detected the cycle
		t.Logf("import returned error (acceptable): %v", err)
		return
	}

	// If import succeeds, validation with the self-referencing schema
	// should fail gracefully (unresolved ref or similar)
	r := schema.SafeParse("test")
	t.Logf("self-referencing schema parse result: success=%v, issues=%v", r.Success, r.Issues)
}

func TestSecurity_CVE2003_1564_DeeplyNestedSchema(t *testing.T) {
	// Create a deeply nested schema via JSON import:
	// array(array(array(...(string)...)))
	const depth = 100
	inner := map[string]any{"kind": "string"}
	for i := 0; i < depth; i++ {
		inner = map[string]any{
			"kind": "array",
			"item": inner,
		}
	}

	doc := &Document{
		AnyvaliVersion: "1",
		SchemaVersion:  "1",
		Root:           inner,
	}

	schema, err := Import(doc)
	if err != nil {
		t.Fatalf("import failed for depth=%d: %v", depth, err)
	}

	// Should reject a non-array at the top level
	r := schema.SafeParse("not an array")
	if r.Success {
		t.Fatal("expected failure when passing string to deeply nested array schema")
	}
}

func TestSecurity_RecursiveTreeValidation(t *testing.T) {
	// Build a recursive tree schema manually using RefSchema:
	// TreeNode = { value: string, children: array(ref(TreeNode)) }
	ref := newRefSchema("#/definitions/TreeNode")

	tree := Object(map[string]Schema{
		"value":    String(),
		"children": Array(ref),
	})

	// Resolve the ref to point back to tree
	ref.Resolve(tree)

	// Validate a reasonable-depth tree
	input := map[string]any{
		"value": "root",
		"children": []any{
			map[string]any{
				"value": "child1",
				"children": []any{
					map[string]any{
						"value":    "grandchild",
						"children": []any{},
					},
				},
			},
			map[string]any{
				"value":    "child2",
				"children": []any{},
			},
		},
	}

	r := tree.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success for valid tree, got issues: %v", r.Issues)
	}
}

// ---------------------------------------------------------------------------
// 3. Integer Overflow (CWE-190)
// ---------------------------------------------------------------------------

func TestSecurity_CWE190_Int8Boundaries(t *testing.T) {
	s := Int8()

	// Valid boundaries
	if r := s.SafeParse(int64(math.MinInt8)); !r.Success {
		t.Fatal("expected success for MinInt8")
	}
	if r := s.SafeParse(int64(math.MaxInt8)); !r.Success {
		t.Fatal("expected success for MaxInt8")
	}

	// Overflow: MAX+1
	if r := s.SafeParse(int64(math.MaxInt8 + 1)); r.Success {
		t.Fatal("expected failure for MaxInt8+1")
	}
	// Underflow: MIN-1
	if r := s.SafeParse(int64(math.MinInt8 - 1)); r.Success {
		t.Fatal("expected failure for MinInt8-1")
	}
}

func TestSecurity_CWE190_Int16Boundaries(t *testing.T) {
	s := Int16()

	if r := s.SafeParse(int64(math.MinInt16)); !r.Success {
		t.Fatal("expected success for MinInt16")
	}
	if r := s.SafeParse(int64(math.MaxInt16)); !r.Success {
		t.Fatal("expected success for MaxInt16")
	}
	if r := s.SafeParse(int64(math.MaxInt16 + 1)); r.Success {
		t.Fatal("expected failure for MaxInt16+1")
	}
	if r := s.SafeParse(int64(math.MinInt16 - 1)); r.Success {
		t.Fatal("expected failure for MinInt16-1")
	}
}

func TestSecurity_CWE190_Int32Boundaries(t *testing.T) {
	s := Int32()

	if r := s.SafeParse(int64(math.MinInt32)); !r.Success {
		t.Fatal("expected success for MinInt32")
	}
	if r := s.SafeParse(int64(math.MaxInt32)); !r.Success {
		t.Fatal("expected success for MaxInt32")
	}
	if r := s.SafeParse(int64(math.MaxInt32 + 1)); r.Success {
		t.Fatal("expected failure for MaxInt32+1")
	}
	if r := s.SafeParse(int64(math.MinInt32 - 1)); r.Success {
		t.Fatal("expected failure for MinInt32-1")
	}
}

func TestSecurity_CWE190_Int64Boundaries(t *testing.T) {
	s := Int64()

	if r := s.SafeParse(int64(math.MinInt64)); !r.Success {
		t.Fatal("expected success for MinInt64")
	}
	if r := s.SafeParse(int64(math.MaxInt64)); !r.Success {
		t.Fatal("expected success for MaxInt64")
	}

	// int64 cannot overflow in Go's type system (no int65), but we can test
	// that a uint64 beyond MaxInt64 is rejected by toInt64.
	if r := s.SafeParse(uint64(math.MaxInt64) + 1); r.Success {
		t.Fatal("expected failure for uint64 beyond MaxInt64")
	}
}

func TestSecurity_CWE190_Uint8Boundaries(t *testing.T) {
	s := Uint8()

	if r := s.SafeParse(int64(0)); !r.Success {
		t.Fatal("expected success for 0")
	}
	if r := s.SafeParse(int64(math.MaxUint8)); !r.Success {
		t.Fatal("expected success for MaxUint8")
	}
	if r := s.SafeParse(int64(math.MaxUint8 + 1)); r.Success {
		t.Fatal("expected failure for MaxUint8+1")
	}
	if r := s.SafeParse(int64(-1)); r.Success {
		t.Fatal("expected failure for -1 on uint8")
	}
}

func TestSecurity_CWE190_Uint16Boundaries(t *testing.T) {
	s := Uint16()

	if r := s.SafeParse(int64(0)); !r.Success {
		t.Fatal("expected success for 0")
	}
	if r := s.SafeParse(int64(math.MaxUint16)); !r.Success {
		t.Fatal("expected success for MaxUint16")
	}
	if r := s.SafeParse(int64(math.MaxUint16 + 1)); r.Success {
		t.Fatal("expected failure for MaxUint16+1")
	}
	if r := s.SafeParse(int64(-1)); r.Success {
		t.Fatal("expected failure for -1 on uint16")
	}
}

func TestSecurity_CWE190_Uint32Boundaries(t *testing.T) {
	s := Uint32()

	if r := s.SafeParse(int64(0)); !r.Success {
		t.Fatal("expected success for 0")
	}
	if r := s.SafeParse(int64(math.MaxUint32)); !r.Success {
		t.Fatal("expected success for MaxUint32")
	}
	if r := s.SafeParse(int64(math.MaxUint32 + 1)); r.Success {
		t.Fatal("expected failure for MaxUint32+1")
	}
	if r := s.SafeParse(int64(-1)); r.Success {
		t.Fatal("expected failure for -1 on uint32")
	}
}

func TestSecurity_CWE190_Uint64Boundaries(t *testing.T) {
	s := Uint64()

	if r := s.SafeParse(int64(0)); !r.Success {
		t.Fatal("expected success for 0")
	}
	if r := s.SafeParse(uint64(math.MaxUint64)); !r.Success {
		t.Fatal("expected success for MaxUint64")
	}
	if r := s.SafeParse(int64(-1)); r.Success {
		t.Fatal("expected failure for -1 on uint64")
	}
}

// ---------------------------------------------------------------------------
// 4. NaN / Infinity (CWE-20)
// ---------------------------------------------------------------------------

func TestSecurity_CWE20_NaN_Number(t *testing.T) {
	s := Number()
	r := s.SafeParse(math.NaN())
	if r.Success {
		t.Fatal("expected Number() to reject NaN")
	}
	if len(r.Issues) == 0 || r.Issues[0].Code != IssueInvalidNumber {
		t.Fatalf("expected issue code %s, got %v", IssueInvalidNumber, r.Issues)
	}
}

func TestSecurity_CWE20_NaN_Float64(t *testing.T) {
	s := Float64()
	r := s.SafeParse(math.NaN())
	if r.Success {
		t.Fatal("expected Float64() to reject NaN")
	}
}

func TestSecurity_CWE20_NaN_Float32(t *testing.T) {
	s := Float32()
	r := s.SafeParse(math.NaN())
	if r.Success {
		t.Fatal("expected Float32() to reject NaN")
	}
}

func TestSecurity_CWE20_NaN_Int(t *testing.T) {
	s := Int()
	r := s.SafeParse(math.NaN())
	if r.Success {
		t.Fatal("expected Int() to reject NaN (non-integer float)")
	}
}

func TestSecurity_CWE20_PositiveInfinity(t *testing.T) {
	schemas := []struct {
		name   string
		schema Schema
	}{
		{"Number", Number()},
		{"Float64", Float64()},
		{"Float32", Float32()},
	}
	for _, tt := range schemas {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.schema.SafeParse(math.Inf(1))
			if r.Success {
				t.Fatalf("expected %s to reject +Inf", tt.name)
			}
		})
	}
}

func TestSecurity_CWE20_NegativeInfinity(t *testing.T) {
	schemas := []struct {
		name   string
		schema Schema
	}{
		{"Number", Number()},
		{"Float64", Float64()},
		{"Float32", Float32()},
	}
	for _, tt := range schemas {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.schema.SafeParse(math.Inf(-1))
			if r.Success {
				t.Fatalf("expected %s to reject -Inf", tt.name)
			}
		})
	}
}

func TestSecurity_CWE20_Infinity_Int(t *testing.T) {
	s := Int()
	// +Inf is not an integer; toInt64 should reject it
	r := s.SafeParse(math.Inf(1))
	if r.Success {
		t.Fatal("expected Int() to reject +Inf")
	}

	r = s.SafeParse(math.Inf(-1))
	if r.Success {
		t.Fatal("expected Int() to reject -Inf")
	}
}

// ---------------------------------------------------------------------------
// 5. Format Validation Bypass (CWE-20)
// ---------------------------------------------------------------------------

func TestSecurity_CWE20_EmailEdgeCases(t *testing.T) {
	s := String().Format("email")

	// Standard valid email
	r := s.SafeParse("user@example.com")
	if !r.Success {
		t.Fatal("expected valid email to pass")
	}

	// Edge cases that should be rejected
	invalid := []string{
		"",                          // empty
		"@example.com",              // no local part
		"user@",                     // no domain
		"user@@example.com",         // double @
		"user@.example.com",         // domain starts with dot
		// ".user@example.com",      // local starts with dot (debatable, some validators allow it)
		"user @example.com",         // space in local part
		"user\t@example.com",        // tab in local part
		"user\n@example.com",        // newline in local part
		"user\x00@example.com",      // null byte
		string([]byte{0x80}) + "@x", // invalid UTF-8 byte
	}

	for _, e := range invalid {
		t.Run(fmt.Sprintf("reject_%q", e), func(t *testing.T) {
			r := s.SafeParse(e)
			if r.Success {
				t.Errorf("expected email %q to be rejected", e)
			}
		})
	}
}

func TestSecurity_CWE20_TamperedEmailFormatNameNotIgnored(t *testing.T) {
	s := String().Format("email\x00")
	r := s.SafeParse("not-an-email")
	if r.Success {
		t.Fatal("tampered format name bypassed email validation")
	}
}

func TestSecurity_CWE20_ImportedTamperedEmailFormatNameNotIgnored(t *testing.T) {
	doc := &Document{
		AnyvaliVersion: "1",
		SchemaVersion:  "1",
		Root: map[string]any{
			"kind":   "string",
			"format": "email\x00",
		},
		Definitions: map[string]map[string]any{},
	}
	schema, err := Import(doc)
	if err != nil {
		return
	}
	if r := schema.SafeParse("not-an-email"); r.Success {
		t.Fatal("imported tampered format name bypassed email validation")
	}
}

func TestSecurity_CWE20_URLJavascriptProtocol(t *testing.T) {
	s := String().Format("url")

	// javascript: should be rejected (only http/https allowed)
	dangerous := []string{
		"javascript:alert(1)",
		"javascript:void(0)",
		"data:text/html,<script>alert(1)</script>",
		"file:///etc/passwd",
		"ftp://example.com",
	}

	for _, u := range dangerous {
		t.Run(fmt.Sprintf("reject_%s", u), func(t *testing.T) {
			r := s.SafeParse(u)
			if r.Success {
				t.Errorf("expected URL %q to be rejected (non-http/https scheme)", u)
			}
		})
	}

	// Valid HTTP/HTTPS should pass
	if r := s.SafeParse("https://example.com"); !r.Success {
		t.Fatal("expected https URL to pass")
	}
	if r := s.SafeParse("http://example.com"); !r.Success {
		t.Fatal("expected http URL to pass")
	}
}

func TestSecurity_CWE20_IPv4OctalNotation(t *testing.T) {
	s := String().Format("ipv4")

	// Standard dotted-decimal
	if r := s.SafeParse("192.168.1.1"); !r.Success {
		t.Fatal("expected valid IPv4 to pass")
	}

	// Octal notation (e.g., 0177.0.0.1 = 127.0.0.1 in some parsers)
	// Go's net.ParseIP rejects octal notation, which is the safe behavior
	octal := []string{
		"0177.0.0.1",   // 127.0.0.1 in octal
		"0300.0250.1.1", // octal for 192.168.1.1
		"010.010.010.010",
	}

	for _, ip := range octal {
		t.Run(fmt.Sprintf("octal_%s", ip), func(t *testing.T) {
			r := s.SafeParse(ip)
			if r.Success {
				t.Errorf("expected octal IPv4 %q to be rejected", ip)
			}
		})
	}
}

func TestSecurity_CWE20_IPv6EdgeCases(t *testing.T) {
	s := String().Format("ipv6")

	// Valid IPv6
	valid := []string{
		"::1",
		"fe80::1",
		"2001:db8::1",
		"::",
	}
	for _, ip := range valid {
		t.Run(fmt.Sprintf("valid_%s", ip), func(t *testing.T) {
			if r := s.SafeParse(ip); !r.Success {
				t.Errorf("expected %q to be valid IPv6", ip)
			}
		})
	}

	// Edge cases that should be rejected
	invalid := []string{
		"",
		"not-ipv6",
		":::1",     // triple colon
		"12345::1", // segment too large
	}
	for _, ip := range invalid {
		t.Run(fmt.Sprintf("invalid_%s", ip), func(t *testing.T) {
			if r := s.SafeParse(ip); r.Success {
				t.Errorf("expected %q to be rejected as IPv6", ip)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5b. Unicode length constraints
// ---------------------------------------------------------------------------

func TestSecurity_UnicodeLength_AstralCodePointCountsAsOneCharacter(t *testing.T) {
	emoji := "😀"
	if r := String().MaxLength(1).SafeParse(emoji); !r.Success {
		t.Fatalf("expected maxLength(1) to accept one astral code point, got %v", r.Issues)
	}
	if r := String().MinLength(2).SafeParse(emoji); r.Success {
		t.Fatal("expected minLength(2) to reject one astral code point")
	}
}

func TestSecurity_UnicodeLength_ImportedMaxLengthUsesCodePoints(t *testing.T) {
	doc := &Document{
		AnyvaliVersion: "1",
		SchemaVersion:  "1",
		Root: map[string]any{
			"kind":      "string",
			"maxLength": 1,
		},
		Definitions: map[string]map[string]any{},
	}
	schema, err := Import(doc)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if r := schema.SafeParse("😀"); !r.Success {
		t.Fatalf("expected imported maxLength(1) to accept one astral code point, got %v", r.Issues)
	}
}

// ---------------------------------------------------------------------------
// 6. Object Key Safety (Prototype pollution class -- Go safe by design)
// ---------------------------------------------------------------------------

func TestSecurity_ObjectProtoKeys(t *testing.T) {
	// In JavaScript, __proto__ and constructor are dangerous object keys.
	// Go maps have no prototype chain, so these are safe -- but we verify
	// they work as normal property names without any special treatment.
	s := Object(map[string]Schema{
		"__proto__":   String(),
		"constructor": String(),
		"toString":    String(),
	})

	input := map[string]any{
		"__proto__":   "safe",
		"constructor": "safe",
		"toString":    "safe",
	}

	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success for proto-named keys, got issues: %v", r.Issues)
	}

	// Verify the data was parsed correctly
	data, ok := r.Data.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	if data["__proto__"] != "safe" {
		t.Fatal("__proto__ key value mismatch")
	}
	if data["constructor"] != "safe" {
		t.Fatal("constructor key value mismatch")
	}
}

func TestSecurity_ObjectEmptyKeyName(t *testing.T) {
	// Empty string as a property key
	s := Object(map[string]Schema{
		"": String(),
	})

	r := s.SafeParse(map[string]any{"": "value"})
	if !r.Success {
		t.Fatalf("expected success for empty key, got: %v", r.Issues)
	}
}

func TestSecurity_ObjectKeyWithSpecialChars(t *testing.T) {
	// Keys with null bytes, newlines, etc.
	s := Object(map[string]Schema{
		"key\x00": String(),
		"key\n":   String(),
	}).UnknownKeys(Allow)

	input := map[string]any{
		"key\x00": "val1",
		"key\n":   "val2",
	}

	r := s.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success for special char keys, got: %v", r.Issues)
	}
}

// ---------------------------------------------------------------------------
// 7. Large Input DoS (CWE-400)
// ---------------------------------------------------------------------------

func TestSecurity_CWE400_VeryLongString(t *testing.T) {
	s := String()

	// 1MB string
	long := strings.Repeat("a", 1_000_000)
	r := s.SafeParse(long)
	if !r.Success {
		t.Fatal("expected success for 1MB string (no maxLength set)")
	}

	// With MaxLength, it should be rejected
	bounded := String().MaxLength(100)
	r = bounded.SafeParse(long)
	if r.Success {
		t.Fatal("expected failure for 1MB string with maxLength=100")
	}
}

func TestSecurity_CWE400_DeeplyNestedObject(t *testing.T) {
	// Build a deeply nested object as raw JSON then import it
	const depth = 50

	// Build nested object schema: {a: {a: {a: ... string ...}}}
	inner := map[string]any{"kind": "string"}
	for i := 0; i < depth; i++ {
		inner = map[string]any{
			"kind": "object",
			"properties": map[string]any{
				"a": inner,
			},
			"required":    []any{"a"},
			"unknownKeys": "reject",
		}
	}

	doc := &Document{
		AnyvaliVersion: "1",
		SchemaVersion:  "1",
		Root:           inner,
	}

	schema, err := Import(doc)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Build matching input
	var input any = "leaf"
	for i := 0; i < depth; i++ {
		input = map[string]any{"a": input}
	}

	r := schema.SafeParse(input)
	if !r.Success {
		t.Fatalf("expected success for depth=%d nested object, got: %v", depth, r.Issues)
	}
}

func TestSecurity_CWE400_LargeArray(t *testing.T) {
	s := Array(String())

	// 10,000-element array
	arr := make([]any, 10_000)
	for i := range arr {
		arr[i] = fmt.Sprintf("item-%d", i)
	}

	r := s.SafeParse(arr)
	if !r.Success {
		t.Fatal("expected success for 10k-element array")
	}

	// With MaxItems, it should be rejected
	bounded := Array(String()).MaxItems(100)
	r = bounded.SafeParse(arr)
	if r.Success {
		t.Fatal("expected failure for 10k-element array with maxItems=100")
	}
}

func TestSecurity_CWE400_LargeObjectManyKeys(t *testing.T) {
	// Object with many properties -- test that validation handles it
	props := make(map[string]Schema)
	for i := 0; i < 1000; i++ {
		props[fmt.Sprintf("key_%d", i)] = String()
	}
	s := Object(props)

	input := make(map[string]any)
	for i := 0; i < 1000; i++ {
		input[fmt.Sprintf("key_%d", i)] = "value"
	}

	r := s.SafeParse(input)
	if !r.Success {
		t.Fatal("expected success for 1000-key object")
	}
}

func TestSecurity_CWE400_ImportLargeJSON(t *testing.T) {
	// Build a large but valid schema document as JSON
	props := make(map[string]any)
	for i := 0; i < 500; i++ {
		props[fmt.Sprintf("field_%d", i)] = map[string]any{"kind": "string"}
	}

	reqList := make([]any, 500)
	for i := 0; i < 500; i++ {
		reqList[i] = fmt.Sprintf("field_%d", i)
	}

	doc := map[string]any{
		"anyvaliVersion": "1",
		"schemaVersion":  "1",
		"root": map[string]any{
			"kind":        "object",
			"properties":  props,
			"required":    reqList,
			"unknownKeys": "reject",
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	schema, err := ImportJSON(data)
	if err != nil {
		t.Fatalf("ImportJSON failed for large schema: %v", err)
	}

	// Validate with matching input
	input := make(map[string]any)
	for i := 0; i < 500; i++ {
		input[fmt.Sprintf("field_%d", i)] = "value"
	}

	r := schema.SafeParse(input)
	if !r.Success {
		t.Fatal("expected success for 500-field object input")
	}
}

func TestSecurity_CWE400_UnionWithManyVariants(t *testing.T) {
	// Union with many schemas -- validation must try each one
	schemas := make([]Schema, 100)
	for i := 0; i < 100; i++ {
		schemas[i] = Literal(fmt.Sprintf("option_%d", i))
	}
	s := Union(schemas...)

	// Last variant should match
	r := s.SafeParse("option_99")
	if !r.Success {
		t.Fatal("expected success for last union variant")
	}

	// Non-matching should fail
	r = s.SafeParse("nonexistent")
	if r.Success {
		t.Fatal("expected failure for non-matching union")
	}
}

// ---------------------------------------------------------------------------
// CWE-20 / spec 5.1: non-portable coercion bypass
//
// Go's strconv parsers are more permissive than the ECMA-262 reference (JS):
// ParseInt accepts a leading "+", ParseFloat accepts hex floats / underscores,
// and ParseBool accepts "t"/"T"/"f"/"F". Each let a string that every other SDK
// rejects coerce into a number/bool. Coercion must accept ASCII decimals (and
// only true/1/false/0) so behaviour is identical across SDKs.
// ---------------------------------------------------------------------------

func TestSecurity_CWE20_CoerceIntRejectsNonDecimal(t *testing.T) {
	s := Int().Coerce(CoerceToInt)
	for _, bad := range []string{"+5", "1_000", "1.0", "1e3", "0x10", "０x", "１２３"} {
		if s.SafeParse(bad).Success {
			t.Fatalf("string->int must reject %q", bad)
		}
	}
	for _, good := range []string{"42", "-7", "  42  "} {
		if !s.SafeParse(good).Success {
			t.Fatalf("string->int must accept %q", good)
		}
	}
}

func TestSecurity_CWE20_CoerceNumberRejectsNonDecimal(t *testing.T) {
	s := Number().Coerce(CoerceToNumber)
	for _, bad := range []string{"0x1p4", "0x10", "1_000.5", "inf", "Infinity", "nan"} {
		if s.SafeParse(bad).Success {
			t.Fatalf("string->number must reject %q", bad)
		}
	}
	for _, good := range []string{"3.14", "+5", ".5", "1e3", "  3.5  "} {
		if !s.SafeParse(good).Success {
			t.Fatalf("string->number must accept %q", good)
		}
	}
}

func TestSecurity_CWE20_CoerceBoolRejectsNonPortableWords(t *testing.T) {
	s := Bool().Coerce(CoerceToBool)
	for _, bad := range []string{"t", "T", "f", "F", "yes", "no"} {
		if s.SafeParse(bad).Success {
			t.Fatalf("string->bool must reject %q", bad)
		}
	}
	for _, good := range []string{"true", "TRUE", "1", "false", "0"} {
		if !s.SafeParse(good).Success {
			t.Fatalf("string->bool must accept %q", good)
		}
	}
}

package anyvali

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// LiteralSchema validates that the value matches a specific literal value.
type LiteralSchema struct {
	baseSchema
	value any
}

func newLiteralSchema(value any) *LiteralSchema {
	return &LiteralSchema{value: value}
}

func (s *LiteralSchema) Describe(description string, opts ...DescribeOpts) *LiteralSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *LiteralSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *LiteralSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *LiteralSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *LiteralSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *LiteralSchema) validate(value any) (any, []ValidationIssue) {
	if !literalEqual(value, s.value) {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidLiteral,
			Message:  fmt.Sprintf("expected %v, received %v", s.value, value),
			Expected: fmt.Sprintf("%v", s.value),
			Received: fmt.Sprintf("%v", value),
		}}
	}
	return value, nil
}

// primitiveLiteralValue preserves primitive value semantics for defined Go types.
func primitiveLiteralValue(value any) any {
	if value == nil {
		return nil
	}
	if _, ok := value.(json.Number); ok {
		return value
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint()
	case reflect.Float32:
		return float32(v.Float())
	case reflect.Float64:
		return v.Float()
	}
	return value
}

func literalEqual(a, b any) bool {
	a, b = primitiveLiteralValue(a), primitiveLiteralValue(b)
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Numeric values compare by their exact JSON value. String and boolean
	// literals must never match a different type with the same printed text.
	ar, aIsNum := exactNumberRat(a)
	br, bIsNum := exactNumberRat(b)
	if aIsNum || bIsNum {
		return aIsNum && bIsNum && ar.Cmp(br) == 0
	}
	switch value := a.(type) {
	case string:
		other, ok := b.(string)
		return ok && value == other
	case bool:
		other, ok := b.(bool)
		return ok && value == other
	}
	return false
}

func (s *LiteralSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":  "literal",
		"value": s.value,
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

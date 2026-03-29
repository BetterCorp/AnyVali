package anyvali

import "fmt"

// LiteralSchema validates that the value matches a specific literal value.
type LiteralSchema struct {
	baseSchema
	value any
}

func newLiteralSchema(value any) *LiteralSchema {
	return &LiteralSchema{value: value}
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

func literalEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Normalize numeric types for comparison
	af, aIsNum := toFloat64(a)
	bf, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		return af == bf
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (s *LiteralSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":  "literal",
		"value": s.value,
	}
	s.addDefaultNode(node)
	return node
}

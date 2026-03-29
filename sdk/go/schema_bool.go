package anyvali

import "fmt"

// BoolSchema validates boolean values.
type BoolSchema struct {
	baseSchema
}

func newBoolSchema() *BoolSchema {
	return &BoolSchema{}
}

func (s *BoolSchema) Default(value bool) *BoolSchema {
	s.setDefault(value)
	return s
}

func (s *BoolSchema) Coerce(c CoercionType) *BoolSchema {
	s.addCoercion(c)
	return s
}

func (s *BoolSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *BoolSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *BoolSchema) validate(value any) (any, []ValidationIssue) {
	b, ok := value.(bool)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected bool, received %s", typeName(value)),
			Expected: "bool",
			Received: typeName(value),
		}}
	}
	return b, nil
}

func (s *BoolSchema) ToNode() map[string]any {
	node := map[string]any{"kind": "bool"}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	return node
}

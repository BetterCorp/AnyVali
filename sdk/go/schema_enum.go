package anyvali

import "fmt"

// EnumSchema validates that the value is one of a set of allowed values.
type EnumSchema struct {
	baseSchema
	values []any
}

func newEnumSchema(values []any) *EnumSchema {
	return &EnumSchema{values: values}
}

func (s *EnumSchema) Describe(description string, opts ...DescribeOpts) *EnumSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *EnumSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *EnumSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *EnumSchema) Default(value any) *EnumSchema {
	s.setDefault(value)
	return s
}

func (s *EnumSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *EnumSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *EnumSchema) validate(value any) (any, []ValidationIssue) {
	for _, v := range s.values {
		if literalEqual(value, v) {
			return value, nil
		}
	}
	return nil, []ValidationIssue{{
		Code:     IssueInvalidLiteral,
		Message:  fmt.Sprintf("value %v is not one of the allowed values %v", value, s.values),
		Expected: fmt.Sprintf("%v", s.values),
		Received: fmt.Sprintf("%v", value),
	}}
}

func (s *EnumSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":   "enum",
		"values": s.values,
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

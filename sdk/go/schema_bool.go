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

// Coerce enables coercion on the schema.
//
// Called with no arguments, it enables the idiomatic default form: a string
// input is coerced to bool, inferred from the schema kind. The only portable
// coercion source is "string". This mirrors the no-arg ergonomic in the other
// SDKs (e.g. Zod's z.coerce.boolean()).
//
// Called with explicit CoercionType values, it appends each in order
// (e.g. Coerce(CoerceToBool)).
func (s *BoolSchema) Coerce(c ...CoercionType) *BoolSchema {
	if len(c) == 0 {
		s.addCoercion(CoerceToBool)
		return s
	}
	for _, ct := range c {
		s.addCoercion(ct)
	}
	return s
}

func (s *BoolSchema) Describe(description string, opts ...DescribeOpts) *BoolSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *BoolSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *BoolSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
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
	s.addMetadataNode(node)
	return node
}

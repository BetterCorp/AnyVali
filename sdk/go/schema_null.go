package anyvali

import "fmt"

// NullSchema validates that the value is null (nil).
type NullSchema struct {
	baseSchema
}

func newNullSchema() *NullSchema {
	return &NullSchema{}
}

func (s *NullSchema) Describe(description string, opts ...DescribeOpts) *NullSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *NullSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *NullSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *NullSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *NullSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *NullSchema) validate(value any) (any, []ValidationIssue) {
	if value != nil {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected null, received %s", typeName(value)),
			Expected: "null",
			Received: typeName(value),
		}}
	}
	return nil, nil
}

func (s *NullSchema) ToNode() map[string]any {
	node := map[string]any{"kind": "null"}
	s.addMetadataNode(node)
	return node
}

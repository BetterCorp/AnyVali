package anyvali

import "fmt"

// NeverSchema rejects all values.
type NeverSchema struct {
	baseSchema
}

func newNeverSchema() *NeverSchema {
	return &NeverSchema{}
}

func (s *NeverSchema) Describe(description string, opts ...DescribeOpts) *NeverSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *NeverSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *NeverSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *NeverSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *NeverSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *NeverSchema) validate(value any) (any, []ValidationIssue) {
	return nil, []ValidationIssue{{
		Code:     IssueInvalidType,
		Message:  fmt.Sprintf("expected never, received %s", typeName(value)),
		Expected: "never",
		Received: typeName(value),
	}}
}

func (s *NeverSchema) ToNode() map[string]any {
	node := map[string]any{"kind": "never"}
	s.addMetadataNode(node)
	return node
}

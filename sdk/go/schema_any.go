package anyvali

// AnySchema accepts any value.
type AnySchema struct {
	baseSchema
}

func newAnySchema() *AnySchema {
	return &AnySchema{}
}

func (s *AnySchema) Describe(description string, opts ...DescribeOpts) *AnySchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *AnySchema) Metadata(meta map[string]any, opts ...MetadataOpts) *AnySchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *AnySchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *AnySchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *AnySchema) validate(value any) (any, []ValidationIssue) {
	return value, nil
}

func (s *AnySchema) ToNode() map[string]any {
	node := map[string]any{"kind": "any"}
	s.addMetadataNode(node)
	return node
}

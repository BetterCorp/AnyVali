package anyvali

// UnknownSchema accepts any value but signals that the type is unknown.
type UnknownSchema struct {
	baseSchema
}

func newUnknownSchema() *UnknownSchema {
	return &UnknownSchema{}
}

func (s *UnknownSchema) Describe(description string, opts ...DescribeOpts) *UnknownSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *UnknownSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *UnknownSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *UnknownSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *UnknownSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *UnknownSchema) validate(value any) (any, []ValidationIssue) {
	return value, nil
}

func (s *UnknownSchema) ToNode() map[string]any {
	node := map[string]any{"kind": "unknown"}
	s.addMetadataNode(node)
	return node
}

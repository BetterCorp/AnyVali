package anyvali

// OptionalSchema wraps a schema to make it optional (absent values are allowed).
type OptionalSchema struct {
	baseSchema
	inner Schema
}

func newOptionalSchema(inner Schema) *OptionalSchema {
	return &OptionalSchema{inner: inner}
}

func (s *OptionalSchema) Describe(description string, opts ...DescribeOpts) *OptionalSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *OptionalSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *OptionalSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *OptionalSchema) Default(value any) *OptionalSchema {
	s.setDefault(value)
	return s
}

func (s *OptionalSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *OptionalSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *OptionalSchema) validate(value any) (any, []ValidationIssue) {
	if value == nil {
		return nil, nil
	}
	result := s.inner.SafeParse(value)
	if result.Success {
		return result.Data, nil
	}
	return nil, result.Issues
}

func (s *OptionalSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":   "optional",
		"schema": s.inner.ToNode(),
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

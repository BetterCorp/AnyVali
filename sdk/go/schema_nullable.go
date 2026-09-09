package anyvali

// NullableSchema wraps a schema to allow null values.
type NullableSchema struct {
	baseSchema
	inner Schema
}

func newNullableSchema(inner Schema) *NullableSchema {
	return &NullableSchema{inner: inner}
}

func (s *NullableSchema) Describe(description string, opts ...DescribeOpts) *NullableSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *NullableSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *NullableSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *NullableSchema) Default(value any) *NullableSchema {
	s.setDefault(value)
	return s
}

func (s *NullableSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *NullableSchema) SafeParse(input any) ParseResult {
	return parseAtDepth(s, input, 0)
}

func (s *NullableSchema) safeParseAtDepth(input any, depth int) ParseResult {
	if isAbsent(input) {
		if s.hasDefault {
			value := deepCopyDefault(s.defaultValue)
			if value == nil {
				return ParseResult{Success: true, Data: nil}
			}
			return parseAtDepth(s.inner, value, depth+1)
		}
		return ParseResult{Success: true, Data: nil}
	}
	if input == nil {
		return ParseResult{Success: true, Data: nil}
	}
	return parseAtDepth(s.inner, input, depth+1)
}

func (s *NullableSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":   "nullable",
		"schema": s.inner.ToNode(),
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

package anyvali

// OptionalSchema wraps a schema to make it optional (absent values are allowed).
type OptionalSchema struct {
	baseSchema
	inner Schema
}

func newOptionalSchema(inner Schema) *OptionalSchema {
	return &OptionalSchema{inner: inner}
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
	// If absent and has default, use the default
	if input == nil || isAbsent(input) {
		if s.hasDefault {
			// Validate the default through the inner schema
			return s.inner.SafeParse(s.defaultValue)
		}
		return ParseResult{Success: true, Data: nil}
	}
	return s.inner.SafeParse(input)
}

func (s *OptionalSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":   "optional",
		"schema": s.inner.ToNode(),
	}
	s.addDefaultNode(node)
	return node
}

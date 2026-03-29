package anyvali

// AnySchema accepts any value.
type AnySchema struct {
	baseSchema
}

func newAnySchema() *AnySchema {
	return &AnySchema{}
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
	return map[string]any{"kind": "any"}
}

package anyvali

// UnknownSchema accepts any value but signals that the type is unknown.
type UnknownSchema struct {
	baseSchema
}

func newUnknownSchema() *UnknownSchema {
	return &UnknownSchema{}
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
	return map[string]any{"kind": "unknown"}
}

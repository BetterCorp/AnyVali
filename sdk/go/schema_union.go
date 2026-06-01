package anyvali

import "fmt"

// UnionSchema validates that the value matches at least one of the given schemas.
type UnionSchema struct {
	baseSchema
	schemas []Schema
}

func newUnionSchema(schemas []Schema) *UnionSchema {
	return &UnionSchema{schemas: schemas}
}

func (s *UnionSchema) Describe(description string, opts ...DescribeOpts) *UnionSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *UnionSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *UnionSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *UnionSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *UnionSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *UnionSchema) validate(value any) (any, []ValidationIssue) {
	var allIssues []ValidationIssue

	for _, schema := range s.schemas {
		result := schema.SafeParse(value)
		if result.Success {
			return result.Data, nil
		}
		allIssues = append(allIssues, result.Issues...)
	}

	return nil, []ValidationIssue{{
		Code:     IssueInvalidUnion,
		Message:  fmt.Sprintf("value does not match any schema in union"),
		Expected: "union match",
		Received: typeName(value),
		Meta: map[string]any{
			"unionIssues": allIssues,
		},
	}}
}

func (s *UnionSchema) ToNode() map[string]any {
	schemas := make([]any, len(s.schemas))
	for i, schema := range s.schemas {
		schemas[i] = schema.ToNode()
	}
	node := map[string]any{
		"kind":    "union",
		"schemas": schemas,
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

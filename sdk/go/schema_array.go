package anyvali

import "fmt"

// ArraySchema validates arrays where each item matches a given schema.
type ArraySchema struct {
	baseSchema
	item     Schema
	minItems *int
	maxItems *int
}

func newArraySchema(item Schema) *ArraySchema {
	return &ArraySchema{item: item}
}

func (s *ArraySchema) MinItems(n int) *ArraySchema {
	s.minItems = &n
	return s
}

func (s *ArraySchema) MaxItems(n int) *ArraySchema {
	s.maxItems = &n
	return s
}

func (s *ArraySchema) Default(value []any) *ArraySchema {
	s.setDefault(value)
	return s
}

func (s *ArraySchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *ArraySchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *ArraySchema) validate(value any) (any, []ValidationIssue) {
	arr, ok := value.([]any)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected array, received %s", typeName(value)),
			Expected: "array",
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue

	if s.minItems != nil && len(arr) < *s.minItems {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("array must have at least %d items", *s.minItems),
			Expected: fmt.Sprintf(">= %d items", *s.minItems),
			Received: fmt.Sprintf("%d items", len(arr)),
		})
	}

	if s.maxItems != nil && len(arr) > *s.maxItems {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("array must have at most %d items", *s.maxItems),
			Expected: fmt.Sprintf("<= %d items", *s.maxItems),
			Received: fmt.Sprintf("%d items", len(arr)),
		})
	}

	parsed := make([]any, len(arr))
	for i, item := range arr {
		result := s.item.SafeParse(item)
		if !result.Success {
			for _, issue := range result.Issues {
				issue.Path = append([]any{i}, issue.Path...)
				issues = append(issues, issue)
			}
		} else {
			parsed[i] = result.Data
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return parsed, nil
}

func (s *ArraySchema) ToNode() map[string]any {
	node := map[string]any{
		"kind": "array",
		"item": s.item.ToNode(),
	}
	if s.minItems != nil {
		node["minItems"] = *s.minItems
	}
	if s.maxItems != nil {
		node["maxItems"] = *s.maxItems
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	return node
}

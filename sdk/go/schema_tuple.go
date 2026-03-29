package anyvali

import "fmt"

// TupleSchema validates fixed-length arrays where each position has its own schema.
type TupleSchema struct {
	baseSchema
	items []Schema
}

func newTupleSchema(items []Schema) *TupleSchema {
	return &TupleSchema{items: items}
}

func (s *TupleSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *TupleSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *TupleSchema) validate(value any) (any, []ValidationIssue) {
	arr, ok := value.([]any)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected array (tuple), received %s", typeName(value)),
			Expected: "tuple",
			Received: typeName(value),
		}}
	}

	if len(arr) != len(s.items) {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected tuple of length %d, received length %d", len(s.items), len(arr)),
			Expected: fmt.Sprintf("%d items", len(s.items)),
			Received: fmt.Sprintf("%d items", len(arr)),
		}}
	}

	var issues []ValidationIssue
	parsed := make([]any, len(arr))

	for i, item := range arr {
		result := s.items[i].SafeParse(item)
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

func (s *TupleSchema) ToNode() map[string]any {
	items := make([]any, len(s.items))
	for i, item := range s.items {
		items[i] = item.ToNode()
	}
	node := map[string]any{
		"kind":  "tuple",
		"items": items,
	}
	s.addDefaultNode(node)
	return node
}

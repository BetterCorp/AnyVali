package anyvali

import "fmt"

// RecordSchema validates a map where all values match a given schema.
type RecordSchema struct {
	baseSchema
	valueSchema Schema
}

func newRecordSchema(valueSchema Schema) *RecordSchema {
	return &RecordSchema{valueSchema: valueSchema}
}

func (s *RecordSchema) Describe(description string, opts ...DescribeOpts) *RecordSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *RecordSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *RecordSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *RecordSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *RecordSchema) SafeParse(input any) ParseResult {
	return parseAtDepth(s, input, newParseContext())
}

func (s *RecordSchema) safeParseAtDepth(input any, ctx parseContext) ParseResult {
	return s.runPipeline(input, func(value any) (any, []ValidationIssue) { return s.validateAtDepth(value, ctx) })
}

func (s *RecordSchema) validateAtDepth(value any, ctx parseContext) (any, []ValidationIssue) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected object (record), received %s", typeName(value)),
			Expected: "record",
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue
	parsed := make(map[string]any)

	for key, val := range obj {
		result := parseAtDepth(s.valueSchema, val, ctx.child())
		if ctx.budget.calls > maxValidationCalls {
			return nil, result.Issues
		}
		if !result.Success {
			for _, issue := range result.Issues {
				issue.Path = append([]any{key}, issue.Path...)
				issues = append(issues, issue)
			}
		} else {
			parsed[key] = result.Data
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return parsed, nil
}

func (s *RecordSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind":   "record",
		"values": s.valueSchema.ToNode(),
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

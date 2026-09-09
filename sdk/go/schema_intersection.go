package anyvali

// IntersectionSchema validates that the value matches all of the given schemas.
type IntersectionSchema struct {
	baseSchema
	schemas []Schema
}

func newIntersectionSchema(schemas []Schema) *IntersectionSchema {
	return &IntersectionSchema{schemas: schemas}
}

func (s *IntersectionSchema) Describe(description string, opts ...DescribeOpts) *IntersectionSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *IntersectionSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *IntersectionSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

func (s *IntersectionSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *IntersectionSchema) SafeParse(input any) ParseResult {
	return parseAtDepth(s, input, 0)
}

func (s *IntersectionSchema) safeParseAtDepth(input any, depth int) ParseResult {
	return s.runPipeline(input, func(value any) (any, []ValidationIssue) { return s.validateAtDepth(value, depth) })
}

func (s *IntersectionSchema) validateAtDepth(value any, depth int) (any, []ValidationIssue) {
	var issues []ValidationIssue
	var lastData any = value

	for _, schema := range s.schemas {
		result := parseAtDepth(schema, value, depth+1)
		if !result.Success {
			issues = append(issues, result.Issues...)
		} else {
			// For objects, merge results
			if obj, ok := result.Data.(map[string]any); ok {
				if merged, ok := lastData.(map[string]any); ok {
					for k, v := range obj {
						merged[k] = v
					}
					lastData = merged
				} else {
					lastData = result.Data
				}
			} else {
				lastData = result.Data
			}
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}

	// If input is an object, return the merged result
	if _, ok := value.(map[string]any); ok {
		return lastData, nil
	}
	return lastData, nil
}

func (s *IntersectionSchema) ToNode() map[string]any {
	schemas := make([]any, len(s.schemas))
	for i, schema := range s.schemas {
		schemas[i] = schema.ToNode()
	}
	node := map[string]any{
		"kind":    "intersection",
		"schemas": schemas,
	}
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

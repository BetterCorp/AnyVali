package anyvali

import "fmt"

// RefSchema references a named definition for recursion and reuse.
type RefSchema struct {
	baseSchema
	ref      string
	resolved Schema
}

func newRefSchema(ref string) *RefSchema {
	return &RefSchema{ref: ref}
}

func (s *RefSchema) Describe(description string, opts ...DescribeOpts) *RefSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *RefSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *RefSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
	return s
}

// Resolve sets the actual schema this reference points to.
func (s *RefSchema) Resolve(schema Schema) *RefSchema {
	s.resolved = schema
	return s
}

func (s *RefSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *RefSchema) SafeParse(input any) ParseResult {
	return parseAtDepth(s, input, newParseContext())
}

// safeParseAtDepth applies local modifiers before delegating without losing absence.
func (s *RefSchema) safeParseAtDepth(input any, ctx parseContext) ParseResult {
	if s.resolved == nil {
		return ParseResult{
			Success: false,
			Issues: []ValidationIssue{{
				Code:    IssueUnsupportedSchemaKind,
				Message: fmt.Sprintf("unresolved reference: %s", s.ref),
			}},
		}
	}
	if isAbsent(input) && !s.hasDefault {
		return parseAtDepth(s.resolved, input, ctx.child())
	}
	return s.runPipeline(input, func(value any) (any, []ValidationIssue) {
		result := parseAtDepth(s.resolved, value, ctx.child())
		return result.Data, result.Issues
	})
}

// ToNode exports this reference occurrence and its local modifiers.
func (s *RefSchema) ToNode() map[string]any {
	node := map[string]any{
		"kind": "ref",
		"ref":  s.ref,
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

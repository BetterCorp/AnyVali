package anyvali

import "fmt"

// ObjectSchema validates objects (maps) with defined properties.
type ObjectSchema struct {
	baseSchema
	properties  map[string]Schema
	required    map[string]bool
	unknownKeys UnknownKeyMode
}

func newObjectSchema(props map[string]Schema) *ObjectSchema {
	required := make(map[string]bool)
	for k := range props {
		required[k] = true
	}
	return &ObjectSchema{
		properties:  props,
		required:    required,
		unknownKeys: Reject,
	}
}

// Required sets which fields are required. By default, all properties are required.
func (s *ObjectSchema) Required(fields ...string) *ObjectSchema {
	s.required = make(map[string]bool)
	for _, f := range fields {
		s.required[f] = true
	}
	return s
}

// UnknownKeys sets the unknown key handling mode.
func (s *ObjectSchema) UnknownKeys(mode UnknownKeyMode) *ObjectSchema {
	s.unknownKeys = mode
	return s
}

func (s *ObjectSchema) Default(value map[string]any) *ObjectSchema {
	s.setDefault(value)
	return s
}

func (s *ObjectSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *ObjectSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *ObjectSchema) validate(value any) (any, []ValidationIssue) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected object, received %s", typeName(value)),
			Expected: "object",
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue
	parsed := make(map[string]any)

	// Check required fields and validate known properties
	for key, schema := range s.properties {
		val, exists := obj[key]
		if !exists {
			if s.required[key] {
				// Check if the schema is optional
				if _, isOpt := schema.(*OptionalSchema); isOpt {
					continue
				}
				issues = append(issues, ValidationIssue{
					Code:    IssueRequired,
					Message: fmt.Sprintf("required field %q is missing", key),
					Path:    []any{key},
				})
			}
			continue
		}

		result := schema.SafeParse(val)
		if !result.Success {
			for _, issue := range result.Issues {
				issue.Path = append([]any{key}, issue.Path...)
				issues = append(issues, issue)
			}
		} else {
			parsed[key] = result.Data
		}
	}

	// Handle unknown keys
	for key, val := range obj {
		if _, known := s.properties[key]; known {
			continue
		}
		switch s.unknownKeys {
		case Reject:
			issues = append(issues, ValidationIssue{
				Code:    IssueUnknownKey,
				Message: fmt.Sprintf("unknown key %q", key),
				Path:    []any{key},
			})
		case Allow:
			parsed[key] = val
		case Strip:
			// Just skip it
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return parsed, nil
}

func (s *ObjectSchema) ToNode() map[string]any {
	props := make(map[string]any)
	for k, v := range s.properties {
		props[k] = v.ToNode()
	}

	requiredList := make([]any, 0)
	for k := range s.required {
		if s.required[k] {
			requiredList = append(requiredList, k)
		}
	}

	node := map[string]any{
		"kind":        "object",
		"properties":  props,
		"required":    requiredList,
		"unknownKeys": string(s.unknownKeys),
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	return node
}

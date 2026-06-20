package anyvali

import "fmt"

// ObjectSchema validates objects (maps) with defined properties.
type ObjectSchema struct {
	baseSchema
	properties          map[string]Schema
	required            map[string]bool
	unknownKeys         UnknownKeyMode
	unknownKeysExplicit bool
}

func newObjectSchema(props map[string]Schema) *ObjectSchema {
	required := make(map[string]bool)
	for k := range props {
		required[k] = true
	}
	return &ObjectSchema{
		properties:          props,
		required:            required,
		unknownKeys:         Strip,
		unknownKeysExplicit: false,
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
	s.unknownKeysExplicit = true
	return s
}

func (s *ObjectSchema) effectiveUnknownKeys() UnknownKeyMode {
	if s.unknownKeysExplicit {
		return s.unknownKeys
	}
	return Strip
}

func (s *ObjectSchema) exportUnknownKeys() UnknownKeyMode {
	if s.unknownKeysExplicit {
		return s.unknownKeys
	}
	return Strip
}

func (s *ObjectSchema) Describe(description string, opts ...DescribeOpts) *ObjectSchema {
	var o *DescribeOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setDescribe(description, o)
	return s
}

func (s *ObjectSchema) Metadata(meta map[string]any, opts ...MetadataOpts) *ObjectSchema {
	var o *MetadataOpts
	if len(opts) > 0 {
		o = &opts[0]
	}
	s.setMetadata(meta, o)
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
	return s.validateWithInherited(value, "", false)
}

func (s *ObjectSchema) safeParseWithInherited(input any, inherited UnknownKeyMode) ParseResult {
	return s.runPipeline(input, func(value any) (any, []ValidationIssue) {
		return s.validateWithInherited(value, inherited, true)
	})
}

func (s *ObjectSchema) validateWithInherited(value any, inherited UnknownKeyMode, hasInherited bool) (any, []ValidationIssue) {
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
	mode := s.effectiveUnknownKeys()
	if hasInherited {
		mode = inherited
	}
	propagate := mode == Strip || mode == Reject

	// Check required fields and validate known properties
	for key, schema := range s.properties {
		val, exists := obj[key]
		if !exists {
			defaultInfo, canHaveDefault := schema.(interface{ defaultInfo() (any, bool) })
			hasDefault := false
			if canHaveDefault {
				_, hasDefault = defaultInfo.defaultInfo()
			}
			if canHaveDefault && hasDefault {
				result := safeParseChild(schema, absentValue, mode, propagate)
				if !result.Success {
					for _, issue := range result.Issues {
						issue.Path = append([]any{key}, issue.Path...)
						issues = append(issues, issue)
					}
				} else {
					parsed[key] = result.Data
				}
			} else if s.required[key] {
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

		result := safeParseChild(schema, val, mode, propagate)
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
		switch mode {
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

func safeParseChild(schema Schema, val any, mode UnknownKeyMode, propagate bool) ParseResult {
	if propagate {
		if obj, ok := schema.(*ObjectSchema); ok {
			return obj.safeParseWithInherited(val, mode)
		}
	}
	return schema.SafeParse(val)
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
		"unknownKeys": string(s.exportUnknownKeys()),
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	s.addMetadataNode(node)
	return node
}

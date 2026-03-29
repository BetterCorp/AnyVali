package anyvali

import (
	"fmt"
	"regexp"
	"strings"
)

// StringSchema validates string values with optional constraints.
type StringSchema struct {
	baseSchema
	minLength  *int
	maxLength  *int
	pattern    *string
	patternRe  *regexp.Regexp
	startsWith *string
	endsWith   *string
	includes   *string
	format     *string
}

func newStringSchema() *StringSchema {
	return &StringSchema{}
}

func (s *StringSchema) MinLength(n int) *StringSchema {
	s.minLength = &n
	return s
}

func (s *StringSchema) MaxLength(n int) *StringSchema {
	s.maxLength = &n
	return s
}

func (s *StringSchema) Pattern(p string) *StringSchema {
	s.pattern = &p
	s.patternRe = regexp.MustCompile(p)
	return s
}

func (s *StringSchema) StartsWith(prefix string) *StringSchema {
	s.startsWith = &prefix
	return s
}

func (s *StringSchema) EndsWith(suffix string) *StringSchema {
	s.endsWith = &suffix
	return s
}

func (s *StringSchema) Includes(substr string) *StringSchema {
	s.includes = &substr
	return s
}

func (s *StringSchema) Format(f string) *StringSchema {
	s.format = &f
	return s
}

func (s *StringSchema) Default(value string) *StringSchema {
	s.setDefault(value)
	return s
}

func (s *StringSchema) Coerce(c CoercionType) *StringSchema {
	s.addCoercion(c)
	return s
}

func (s *StringSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *StringSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *StringSchema) validate(value any) (any, []ValidationIssue) {
	str, ok := value.(string)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected string, received %s", typeName(value)),
			Expected: "string",
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue
	runeLen := len([]rune(str))

	if s.minLength != nil && runeLen < *s.minLength {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("string must have at least %d characters", *s.minLength),
			Expected: fmt.Sprintf(">= %d", *s.minLength),
			Received: fmt.Sprintf("%d", runeLen),
		})
	}

	if s.maxLength != nil && runeLen > *s.maxLength {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("string must have at most %d characters", *s.maxLength),
			Expected: fmt.Sprintf("<= %d", *s.maxLength),
			Received: fmt.Sprintf("%d", runeLen),
		})
	}

	if s.pattern != nil && s.patternRe != nil && !s.patternRe.MatchString(str) {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidString,
			Message:  fmt.Sprintf("string does not match pattern %q", *s.pattern),
			Expected: *s.pattern,
			Received: str,
		})
	}

	if s.startsWith != nil {
		if !strings.HasPrefix(str, *s.startsWith) {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidString,
				Message:  fmt.Sprintf("string must start with %q", *s.startsWith),
				Expected: *s.startsWith,
				Received: str,
			})
		}
	}

	if s.endsWith != nil {
		if !strings.HasSuffix(str, *s.endsWith) {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidString,
				Message:  fmt.Sprintf("string must end with %q", *s.endsWith),
				Expected: *s.endsWith,
				Received: str,
			})
		}
	}

	if s.includes != nil {
		if !strings.Contains(str, *s.includes) {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidString,
				Message:  fmt.Sprintf("string must include %q", *s.includes),
				Expected: *s.includes,
				Received: str,
			})
		}
	}

	if s.format != nil {
		if !validateFormat(str, *s.format) {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidString,
				Message:  fmt.Sprintf("invalid %s format", *s.format),
				Expected: *s.format,
				Received: str,
			})
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return str, nil
}

func (s *StringSchema) ToNode() map[string]any {
	node := map[string]any{"kind": "string"}
	if s.minLength != nil {
		node["minLength"] = *s.minLength
	}
	if s.maxLength != nil {
		node["maxLength"] = *s.maxLength
	}
	if s.pattern != nil {
		node["pattern"] = *s.pattern
	}
	if s.startsWith != nil {
		node["startsWith"] = *s.startsWith
	}
	if s.endsWith != nil {
		node["endsWith"] = *s.endsWith
	}
	if s.includes != nil {
		node["includes"] = *s.includes
	}
	if s.format != nil {
		node["format"] = *s.format
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	return node
}

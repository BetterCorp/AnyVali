package anyvali

import (
	"fmt"
	"math"
)

// IntSchema validates integer values with optional constraints and range checks.
type IntSchema struct {
	baseSchema
	kind         string
	min          *int64
	max          *int64
	exclusiveMin *int64
	exclusiveMax *int64
	multipleOf   *int64
	rangeLow     int64
	rangeHigh    int64
	unsigned     bool
	uRangeLow    uint64
	uRangeHigh   uint64
}

func newIntSchema(kind string, low, high int64) *IntSchema {
	return &IntSchema{
		kind:      kind,
		rangeLow:  low,
		rangeHigh: high,
	}
}

func newUintSchema(kind string, low, high uint64) *IntSchema {
	return &IntSchema{
		kind:       kind,
		unsigned:   true,
		uRangeLow:  low,
		uRangeHigh: high,
	}
}

func (s *IntSchema) Min(n int64) *IntSchema {
	s.min = &n
	return s
}

func (s *IntSchema) Max(n int64) *IntSchema {
	s.max = &n
	return s
}

func (s *IntSchema) ExclusiveMin(n int64) *IntSchema {
	s.exclusiveMin = &n
	return s
}

func (s *IntSchema) ExclusiveMax(n int64) *IntSchema {
	s.exclusiveMax = &n
	return s
}

func (s *IntSchema) MultipleOf(n int64) *IntSchema {
	s.multipleOf = &n
	return s
}

func (s *IntSchema) Default(value int64) *IntSchema {
	s.setDefault(value)
	return s
}

func (s *IntSchema) Coerce(c CoercionType) *IntSchema {
	s.addCoercion(c)
	return s
}

func (s *IntSchema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *IntSchema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *IntSchema) validate(value any) (any, []ValidationIssue) {
	if s.unsigned {
		return s.validateUnsigned(value)
	}
	return s.validateSigned(value)
}

func (s *IntSchema) validateSigned(value any) (any, []ValidationIssue) {
	i, ok := toInt64(value)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected integer, received %s", typeName(value)),
			Expected: s.kind,
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue

	// Range check for specific int width
	if i < s.rangeLow || i > s.rangeHigh {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidNumber,
			Message:  fmt.Sprintf("value %d out of range for %s (%d to %d)", i, s.kind, s.rangeLow, s.rangeHigh),
			Expected: s.kind,
			Received: fmt.Sprintf("%d", i),
		})
	}

	if s.min != nil && i < *s.min {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be >= %d", *s.min),
			Expected: fmt.Sprintf(">= %d", *s.min),
			Received: fmt.Sprintf("%d", i),
		})
	}

	if s.max != nil && i > *s.max {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be <= %d", *s.max),
			Expected: fmt.Sprintf("<= %d", *s.max),
			Received: fmt.Sprintf("%d", i),
		})
	}

	if s.exclusiveMin != nil && i <= *s.exclusiveMin {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be > %d", *s.exclusiveMin),
			Expected: fmt.Sprintf("> %d", *s.exclusiveMin),
			Received: fmt.Sprintf("%d", i),
		})
	}

	if s.exclusiveMax != nil && i >= *s.exclusiveMax {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be < %d", *s.exclusiveMax),
			Expected: fmt.Sprintf("< %d", *s.exclusiveMax),
			Received: fmt.Sprintf("%d", i),
		})
	}

	if s.multipleOf != nil && *s.multipleOf != 0 && i%*s.multipleOf != 0 {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidNumber,
			Message:  fmt.Sprintf("value must be a multiple of %d", *s.multipleOf),
			Expected: fmt.Sprintf("multiple of %d", *s.multipleOf),
			Received: fmt.Sprintf("%d", i),
		})
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return i, nil
}

func (s *IntSchema) validateUnsigned(value any) (any, []ValidationIssue) {
	u, ok := toUint64(value)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected unsigned integer, received %s", typeName(value)),
			Expected: s.kind,
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue

	if u < s.uRangeLow || u > s.uRangeHigh {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidNumber,
			Message:  fmt.Sprintf("value %d out of range for %s (%d to %d)", u, s.kind, s.uRangeLow, s.uRangeHigh),
			Expected: s.kind,
			Received: fmt.Sprintf("%d", u),
		})
	}

	if s.min != nil && *s.min >= 0 && u < uint64(*s.min) {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be >= %d", *s.min),
			Expected: fmt.Sprintf(">= %d", *s.min),
			Received: fmt.Sprintf("%d", u),
		})
	}

	if s.max != nil && *s.max >= 0 && u > uint64(*s.max) {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be <= %d", *s.max),
			Expected: fmt.Sprintf("<= %d", *s.max),
			Received: fmt.Sprintf("%d", u),
		})
	}

	if s.exclusiveMin != nil && *s.exclusiveMin >= 0 && u <= uint64(*s.exclusiveMin) {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be > %d", *s.exclusiveMin),
			Expected: fmt.Sprintf("> %d", *s.exclusiveMin),
			Received: fmt.Sprintf("%d", u),
		})
	}

	if s.exclusiveMax != nil && *s.exclusiveMax >= 0 && u >= uint64(*s.exclusiveMax) {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be < %d", *s.exclusiveMax),
			Expected: fmt.Sprintf("< %d", *s.exclusiveMax),
			Received: fmt.Sprintf("%d", u),
		})
	}

	if s.multipleOf != nil && *s.multipleOf > 0 && u%uint64(*s.multipleOf) != 0 {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidNumber,
			Message:  fmt.Sprintf("value must be a multiple of %d", *s.multipleOf),
			Expected: fmt.Sprintf("multiple of %d", *s.multipleOf),
			Received: fmt.Sprintf("%d", u),
		})
	}

	if len(issues) > 0 {
		return nil, issues
	}
	// Return as int64 if it fits, otherwise uint64
	if u <= math.MaxInt64 {
		return int64(u), nil
	}
	return u, nil
}

func (s *IntSchema) ToNode() map[string]any {
	node := map[string]any{"kind": s.kind}
	if s.min != nil {
		node["min"] = *s.min
	}
	if s.max != nil {
		node["max"] = *s.max
	}
	if s.exclusiveMin != nil {
		node["exclusiveMin"] = *s.exclusiveMin
	}
	if s.exclusiveMax != nil {
		node["exclusiveMax"] = *s.exclusiveMax
	}
	if s.multipleOf != nil {
		node["multipleOf"] = *s.multipleOf
	}
	s.addCoercionNode(node)
	s.addDefaultNode(node)
	return node
}

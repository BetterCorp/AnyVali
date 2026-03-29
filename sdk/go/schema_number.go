package anyvali

import (
	"fmt"
	"math"
)

// Float64Schema validates float64 values with optional constraints.
type Float64Schema struct {
	baseSchema
	kind         string
	min          *float64
	max          *float64
	exclusiveMin *float64
	exclusiveMax *float64
	multipleOf   *float64
	rangeLow     float64
	rangeHigh    float64
	hasRange     bool
}

func newFloat64Schema(kind string) *Float64Schema {
	return &Float64Schema{kind: kind}
}

func newFloat32Schema() *Float64Schema {
	s := &Float64Schema{
		kind:      "float32",
		rangeLow:  -math.MaxFloat32,
		rangeHigh: math.MaxFloat32,
		hasRange:  true,
	}
	return s
}

// NumberSchema is an alias for Float64Schema.
type NumberSchema = Float64Schema

// Float32Schema is an alias for Float64Schema with range constraints.
type Float32Schema = Float64Schema

func (s *Float64Schema) Min(n float64) *Float64Schema {
	s.min = &n
	return s
}

func (s *Float64Schema) Max(n float64) *Float64Schema {
	s.max = &n
	return s
}

func (s *Float64Schema) ExclusiveMin(n float64) *Float64Schema {
	s.exclusiveMin = &n
	return s
}

func (s *Float64Schema) ExclusiveMax(n float64) *Float64Schema {
	s.exclusiveMax = &n
	return s
}

func (s *Float64Schema) MultipleOf(n float64) *Float64Schema {
	s.multipleOf = &n
	return s
}

func (s *Float64Schema) Default(value float64) *Float64Schema {
	s.setDefault(value)
	return s
}

func (s *Float64Schema) Coerce(c CoercionType) *Float64Schema {
	s.addCoercion(c)
	return s
}

func (s *Float64Schema) Parse(input any) (any, error) {
	result := s.SafeParse(input)
	if !result.Success {
		return nil, &ValidationError{Issues: result.Issues}
	}
	return result.Data, nil
}

func (s *Float64Schema) SafeParse(input any) ParseResult {
	return s.runPipeline(input, s.validate)
}

func (s *Float64Schema) validate(value any) (any, []ValidationIssue) {
	f, ok := toFloat64(value)
	if !ok {
		return nil, []ValidationIssue{{
			Code:     IssueInvalidType,
			Message:  fmt.Sprintf("expected number, received %s", typeName(value)),
			Expected: s.kind,
			Received: typeName(value),
		}}
	}

	var issues []ValidationIssue

	// Range check for float32
	if s.hasRange {
		if f < s.rangeLow || f > s.rangeHigh {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidNumber,
				Message:  fmt.Sprintf("value %v out of range for %s", f, s.kind),
				Expected: s.kind,
				Received: fmt.Sprintf("%v", f),
			})
		}
	}

	if math.IsNaN(f) || math.IsInf(f, 0) {
		issues = append(issues, ValidationIssue{
			Code:     IssueInvalidNumber,
			Message:  "value must be a finite number",
			Expected: "finite number",
			Received: fmt.Sprintf("%v", f),
		})
		return nil, issues
	}

	if s.min != nil && f < *s.min {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be >= %v", *s.min),
			Expected: fmt.Sprintf(">= %v", *s.min),
			Received: fmt.Sprintf("%v", f),
		})
	}

	if s.max != nil && f > *s.max {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be <= %v", *s.max),
			Expected: fmt.Sprintf("<= %v", *s.max),
			Received: fmt.Sprintf("%v", f),
		})
	}

	if s.exclusiveMin != nil && f <= *s.exclusiveMin {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooSmall,
			Message:  fmt.Sprintf("value must be > %v", *s.exclusiveMin),
			Expected: fmt.Sprintf("> %v", *s.exclusiveMin),
			Received: fmt.Sprintf("%v", f),
		})
	}

	if s.exclusiveMax != nil && f >= *s.exclusiveMax {
		issues = append(issues, ValidationIssue{
			Code:     IssueTooLarge,
			Message:  fmt.Sprintf("value must be < %v", *s.exclusiveMax),
			Expected: fmt.Sprintf("< %v", *s.exclusiveMax),
			Received: fmt.Sprintf("%v", f),
		})
	}

	if s.multipleOf != nil && *s.multipleOf != 0 {
		remainder := math.Mod(f, *s.multipleOf)
		if math.Abs(remainder) > 1e-10 {
			issues = append(issues, ValidationIssue{
				Code:     IssueInvalidNumber,
				Message:  fmt.Sprintf("value must be a multiple of %v", *s.multipleOf),
				Expected: fmt.Sprintf("multiple of %v", *s.multipleOf),
				Received: fmt.Sprintf("%v", f),
			})
		}
	}

	if len(issues) > 0 {
		return nil, issues
	}
	return f, nil
}

func (s *Float64Schema) ToNode() map[string]any {
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

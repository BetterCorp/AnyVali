package anyvali

import "fmt"

// TypedParseResult is a generic parse result with typed data.
type TypedParseResult[T any] struct {
	Success bool
	Data    T
	Issues  []ValidationIssue
}

// TypedParse parses input using the given schema and returns a typed result.
// Returns an error if validation fails or the result cannot be cast to T.
func TypedParse[T any](s Schema, input any) (T, error) {
	result, err := s.Parse(input)
	if err != nil {
		var zero T
		return zero, err
	}
	typed, ok := result.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("anyvali: type assertion failed: expected %T, got %T", zero, result)
	}
	return typed, nil
}

// TypedSafeParse parses input and returns a TypedParseResult with typed data.
func TypedSafeParse[T any](s Schema, input any) TypedParseResult[T] {
	result := s.SafeParse(input)
	if !result.Success {
		return TypedParseResult[T]{Success: false, Issues: result.Issues}
	}
	typed, ok := result.Data.(T)
	if !ok {
		var zero T
		return TypedParseResult[T]{Success: false, Data: zero, Issues: []ValidationIssue{{
			Code:    "type_assertion_failed",
			Message: fmt.Sprintf("Cannot convert %T to target type", result.Data),
			Path:    []any{},
		}}}
	}
	return TypedParseResult[T]{Success: true, Data: typed}
}

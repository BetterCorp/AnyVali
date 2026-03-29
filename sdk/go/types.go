package anyvali

import (
	"fmt"
	"strings"
)

// Schema is the core interface that all schema types implement.
type Schema interface {
	Parse(input any) (any, error)
	SafeParse(input any) ParseResult
	ToNode() map[string]any
}

// ParseResult is the result of a non-throwing parse operation.
type ParseResult struct {
	Success bool
	Data    any
	Issues  []ValidationIssue
}

// ValidationIssue represents a single validation failure.
type ValidationIssue struct {
	Code     string
	Message  string
	Path     []any
	Expected string
	Received string
	Meta     map[string]any
}

// ValidationError implements the error interface and wraps a list of ValidationIssues.
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	msgs := make([]string, len(e.Issues))
	for i, issue := range e.Issues {
		msgs[i] = fmt.Sprintf("[%s] %s", issue.Code, issue.Message)
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// Document is the portable JSON interchange document.
type Document struct {
	AnyvaliVersion string                    `json:"anyvaliVersion"`
	SchemaVersion  string                    `json:"schemaVersion"`
	Root           map[string]any            `json:"root"`
	Definitions    map[string]map[string]any `json:"definitions,omitempty"`
	Extensions     map[string]any            `json:"extensions,omitempty"`
}

// ExportMode controls what is included in exported documents.
type ExportMode string

const (
	Portable ExportMode = "portable"
	Extended ExportMode = "extended"
)

// UnknownKeyMode controls how unknown keys are handled in objects.
type UnknownKeyMode string

const (
	Reject UnknownKeyMode = "reject"
	Strip  UnknownKeyMode = "strip"
	Allow  UnknownKeyMode = "allow"
)

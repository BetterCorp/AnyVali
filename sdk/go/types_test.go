package anyvali

import (
	"strings"
	"testing"
)

func TestValidationErrorSingleIssue(t *testing.T) {
	err := &ValidationError{
		Issues: []ValidationIssue{
			{Code: IssueInvalidType, Message: "expected string"},
		},
	}
	s := err.Error()
	if !strings.Contains(s, "validation failed:") {
		t.Fatalf("expected 'validation failed:' prefix, got %q", s)
	}
	if !strings.Contains(s, "[invalid_type]") {
		t.Fatalf("expected issue code in error string, got %q", s)
	}
	if !strings.Contains(s, "expected string") {
		t.Fatalf("expected issue message in error string, got %q", s)
	}
}

func TestValidationErrorMultipleIssues(t *testing.T) {
	err := &ValidationError{
		Issues: []ValidationIssue{
			{Code: IssueInvalidType, Message: "expected string"},
			{Code: IssueTooSmall, Message: "too small"},
			{Code: IssueTooLarge, Message: "too large"},
		},
	}
	s := err.Error()
	if !strings.Contains(s, "[invalid_type] expected string") {
		t.Fatal("missing first issue")
	}
	if !strings.Contains(s, "[too_small] too small") {
		t.Fatal("missing second issue")
	}
	if !strings.Contains(s, "[too_large] too large") {
		t.Fatal("missing third issue")
	}
	// Issues should be joined by "; "
	if strings.Count(s, "; ") != 2 {
		t.Fatalf("expected 2 semicolons, got %d in %q", strings.Count(s, "; "), s)
	}
}

func TestValidationErrorNoIssues(t *testing.T) {
	err := &ValidationError{
		Issues: []ValidationIssue{},
	}
	s := err.Error()
	if !strings.HasPrefix(s, "validation failed: ") {
		t.Fatalf("unexpected error string: %q", s)
	}
}

func TestParseResultSuccess(t *testing.T) {
	r := ParseResult{Success: true, Data: "hello"}
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
	if len(r.Issues) != 0 {
		t.Fatalf("expected no issues, got %v", r.Issues)
	}
}

func TestParseResultFailure(t *testing.T) {
	r := ParseResult{
		Success: false,
		Issues: []ValidationIssue{
			{Code: IssueInvalidType, Message: "bad type"},
		},
	}
	if r.Success {
		t.Fatal("expected failure")
	}
	if r.Data != nil {
		t.Fatalf("expected nil data, got %v", r.Data)
	}
	if len(r.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(r.Issues))
	}
}

func TestValidationIssueFields(t *testing.T) {
	issue := ValidationIssue{
		Code:     IssueInvalidType,
		Message:  "expected string",
		Path:     []any{"user", "name"},
		Expected: "string",
		Received: "int",
		Meta:     map[string]any{"extra": true},
	}
	if issue.Code != IssueInvalidType {
		t.Fatal("wrong code")
	}
	if issue.Message != "expected string" {
		t.Fatal("wrong message")
	}
	if len(issue.Path) != 2 {
		t.Fatal("wrong path length")
	}
	if issue.Expected != "string" {
		t.Fatal("wrong expected")
	}
	if issue.Received != "int" {
		t.Fatal("wrong received")
	}
	if issue.Meta["extra"] != true {
		t.Fatal("wrong meta")
	}
}

func TestExportModeConstants(t *testing.T) {
	if Portable != "portable" {
		t.Fatal("unexpected Portable value")
	}
	if Extended != "extended" {
		t.Fatal("unexpected Extended value")
	}
}

func TestUnknownKeyModeConstants(t *testing.T) {
	if Reject != "reject" {
		t.Fatal("unexpected Reject value")
	}
	if Strip != "strip" {
		t.Fatal("unexpected Strip value")
	}
	if Allow != "allow" {
		t.Fatal("unexpected Allow value")
	}
}

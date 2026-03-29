package com.anyvali;

import java.util.List;
import java.util.stream.Collectors;

/**
 * Exception thrown by parse() when validation fails.
 */
public class ValidationError extends RuntimeException {
    private final List<ValidationIssue> issues;

    public ValidationError(List<ValidationIssue> issues) {
        super("Validation failed: " + issues.stream()
                .map(ValidationIssue::message)
                .collect(Collectors.joining("; ")));
        this.issues = List.copyOf(issues);
    }

    public List<ValidationIssue> getIssues() {
        return issues;
    }
}

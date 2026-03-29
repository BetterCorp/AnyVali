package com.anyvali;

import java.util.List;

/**
 * Result of a safeParse() call.
 */
public record ParseResult(
        boolean success,
        Object data,
        List<ValidationIssue> issues
) {
    public ParseResult {
        issues = issues != null ? List.copyOf(issues) : List.of();
    }

    public static ParseResult success(Object data) {
        return new ParseResult(true, data, List.of());
    }

    public static ParseResult failure(List<ValidationIssue> issues) {
        return new ParseResult(false, null, issues);
    }
}

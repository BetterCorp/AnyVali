package com.anyvali;

import java.util.List;

/**
 * Result of a safeParse() call.
 */
public record ParseResult<T>(
        boolean success,
        T data,
        List<ValidationIssue> issues
) {
    public ParseResult {
        issues = issues != null ? List.copyOf(issues) : List.of();
    }

    public static <T> ParseResult<T> success(T data) {
        return new ParseResult<>(true, data, List.of());
    }

    public static <T> ParseResult<T> failure(List<ValidationIssue> issues) {
        return new ParseResult<>(false, null, issues);
    }
}

package com.anyvali;

import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * A single validation issue.
 */
public record ValidationIssue(
        String code,
        String message,
        List<Object> path,
        Object expected,
        Object received,
        Map<String, Object> meta
) {
    public ValidationIssue {
        path = path != null ? List.copyOf(path) : List.of();
        meta = meta != null ? Map.copyOf(meta) : null;
    }

    public ValidationIssue(String code, String message, List<Object> path,
                           Object expected, Object received) {
        this(code, message, path, expected, received, null);
    }

    public ValidationIssue(String code, String message, List<Object> path) {
        this(code, message, path, null, null, null);
    }
}

package com.anyvali.parse;

import com.anyvali.IssueCodes;
import com.anyvali.ValidationContext;

/**
 * Portable coercion functions for the parse pipeline.
 */
public final class Coercion {
    private Coercion() {}

    /**
     * Apply configured coercions to a value.
     * Coercion order: type coercions first, then string transformations.
     */
    public static Object applyCoercion(Object value, CoercionConfig config,
                                       ValidationContext ctx) {
        Object result = value;

        // Type coercions (string -> target type)
        if (config.toInt() && result instanceof String s) {
            result = coerceToInt(s, ctx);
            if (ctx.issueCount() > 0) return result;
        }

        if (config.toNumber() && result instanceof String s) {
            result = coerceToNumber(s, ctx);
            if (ctx.issueCount() > 0) return result;
        }

        if (config.toBool() && result instanceof String s) {
            result = coerceToBool(s, ctx);
            if (ctx.issueCount() > 0) return result;
        }

        // String transformations (only apply to strings)
        if (result instanceof String s) {
            if (config.trim()) {
                s = s.strip();
            }
            if (config.lower()) {
                s = s.toLowerCase();
            }
            if (config.upper()) {
                s = s.toUpperCase();
            }
            result = s;
        }

        return result;
    }

    private static Object coerceToInt(String value, ValidationContext ctx) {
        String stripped = value.strip();
        try {
            return Long.parseLong(stripped);
        } catch (NumberFormatException e) {
            // Try parsing as double, then check if it's a whole number
            try {
                double d = Double.parseDouble(stripped);
                if (d == Math.floor(d) && !Double.isInfinite(d)) {
                    return (long) d;
                }
            } catch (NumberFormatException e2) {
                // fall through
            }
        }
        ctx.addIssue(IssueCodes.COERCION_FAILED,
                "Failed to coerce '" + value + "' to integer",
                "integer", value);
        return value;
    }

    private static Object coerceToNumber(String value, ValidationContext ctx) {
        String stripped = value.strip();
        try {
            return Double.parseDouble(stripped);
        } catch (NumberFormatException e) {
            ctx.addIssue(IssueCodes.COERCION_FAILED,
                    "Failed to coerce '" + value + "' to number",
                    "number", value);
            return value;
        }
    }

    private static Object coerceToBool(String value, ValidationContext ctx) {
        String lower = value.strip().toLowerCase();
        return switch (lower) {
            case "true", "1", "yes" -> Boolean.TRUE;
            case "false", "0", "no" -> Boolean.FALSE;
            default -> {
                ctx.addIssue(IssueCodes.COERCION_FAILED,
                        "Failed to coerce '" + value + "' to boolean",
                        "boolean", value);
                yield value;
            }
        };
    }
}

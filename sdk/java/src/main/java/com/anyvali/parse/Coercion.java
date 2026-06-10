package com.anyvali.parse;

import com.anyvali.IssueCodes;
import com.anyvali.ValidationContext;

import java.util.regex.Pattern;

/**
 * Portable coercion functions for the parse pipeline.
 */
public final class Coercion {
    private Coercion() {}

    // Strict ASCII decimal grammars. Java's Long/Double.parseDouble are more
    // permissive than the ECMA-262 reference (JS): they accept a leading "+",
    // a float fallback for int, hex floats ("0x1p4"), type suffixes ("1.0f",
    // "1d") and "Infinity"/"NaN". Each let a string that every other SDK
    // rejects coerce into a number -- a cross-language validation bypass.
    // Gate on these before parsing (spec 5.1: decimal only).
    private static final Pattern DECIMAL_INT =
            Pattern.compile("^-?[0-9]+$");
    private static final Pattern DECIMAL_FLOAT =
            Pattern.compile("^[+-]?([0-9]+\\.?[0-9]*|\\.[0-9]+)([eE][+-]?[0-9]+)?$");

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
        if (DECIMAL_INT.matcher(stripped).matches()) {
            try {
                return Long.parseLong(stripped);
            } catch (NumberFormatException e) {
                // out of long range -- fall through to failure
            }
        }
        ctx.addIssue(IssueCodes.COERCION_FAILED,
                "Failed to coerce '" + value + "' to integer",
                "integer", value);
        return value;
    }

    private static Object coerceToNumber(String value, ValidationContext ctx) {
        String stripped = value.strip();
        if (DECIMAL_FLOAT.matcher(stripped).matches()) {
            try {
                double d = Double.parseDouble(stripped);
                if (!Double.isInfinite(d) && !Double.isNaN(d)) {
                    return d;
                }
            } catch (NumberFormatException e) {
                // fall through to failure
            }
        }
        ctx.addIssue(IssueCodes.COERCION_FAILED,
                "Failed to coerce '" + value + "' to number",
                "number", value);
        return value;
    }

    private static Object coerceToBool(String value, ValidationContext ctx) {
        // Spec 5.1: only "true"/"1" and "false"/"0" (case-insensitive).
        // "yes"/"no" are non-portable and diverge from the JS reference.
        String lower = value.strip().toLowerCase();
        return switch (lower) {
            case "true", "1" -> Boolean.TRUE;
            case "false", "0" -> Boolean.FALSE;
            default -> {
                ctx.addIssue(IssueCodes.COERCION_FAILED,
                        "Failed to coerce '" + value + "' to boolean",
                        "boolean", value);
                yield value;
            }
        };
    }
}

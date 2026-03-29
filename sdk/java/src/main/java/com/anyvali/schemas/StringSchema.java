package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;
import com.anyvali.format.FormatValidators;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.regex.Pattern;

/**
 * Schema for string values with optional constraints.
 */
public class StringSchema extends Schema {
    private Integer minLength;
    private Integer maxLength;
    private String pattern;
    private String startsWith;
    private String endsWith;
    private String includes;
    private String format;

    public StringSchema() {}

    private StringSchema(StringSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.minLength = other.minLength;
        this.maxLength = other.maxLength;
        this.pattern = other.pattern;
        this.startsWith = other.startsWith;
        this.endsWith = other.endsWith;
        this.includes = other.includes;
        this.format = other.format;
    }

    public StringSchema minLength(int n) {
        var copy = new StringSchema(this);
        copy.minLength = n;
        return copy;
    }

    public StringSchema maxLength(int n) {
        var copy = new StringSchema(this);
        copy.maxLength = n;
        return copy;
    }

    public StringSchema pattern(String p) {
        var copy = new StringSchema(this);
        copy.pattern = p;
        return copy;
    }

    public StringSchema startsWith(String s) {
        var copy = new StringSchema(this);
        copy.startsWith = s;
        return copy;
    }

    public StringSchema endsWith(String s) {
        var copy = new StringSchema(this);
        copy.endsWith = s;
        return copy;
    }

    public StringSchema includes(String s) {
        var copy = new StringSchema(this);
        copy.includes = s;
        return copy;
    }

    public StringSchema format(String f) {
        var copy = new StringSchema(this);
        copy.format = f;
        return copy;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (!(input instanceof String s)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected string, received " + received, "string", received);
            return null;
        }

        if (minLength != null && s.length() < minLength) {
            ctx.addIssue(IssueCodes.TOO_SMALL,
                    "String must have at least " + minLength + " character(s)",
                    minLength, s.length());
        }

        if (maxLength != null && s.length() > maxLength) {
            ctx.addIssue(IssueCodes.TOO_LARGE,
                    "String must have at most " + maxLength + " character(s)",
                    maxLength, s.length());
        }

        if (pattern != null && !Pattern.compile(pattern).matcher(s).find()) {
            ctx.addIssue(IssueCodes.INVALID_STRING,
                    "String does not match pattern '" + pattern + "'",
                    pattern, s);
        }

        if (startsWith != null && !s.startsWith(startsWith)) {
            ctx.addIssue(IssueCodes.INVALID_STRING,
                    "String must start with '" + startsWith + "'",
                    startsWith, s);
        }

        if (endsWith != null && !s.endsWith(endsWith)) {
            ctx.addIssue(IssueCodes.INVALID_STRING,
                    "String must end with '" + endsWith + "'",
                    endsWith, s);
        }

        if (includes != null && !s.contains(includes)) {
            ctx.addIssue(IssueCodes.INVALID_STRING,
                    "String must include '" + includes + "'",
                    includes, s);
        }

        if (format != null && !FormatValidators.validate(format, s)) {
            ctx.addIssue(IssueCodes.INVALID_STRING,
                    "Invalid " + format, format, s);
        }

        return s;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "string");
        if (minLength != null) node.put("minLength", minLength);
        if (maxLength != null) node.put("maxLength", maxLength);
        if (pattern != null) node.put("pattern", pattern);
        if (startsWith != null) node.put("startsWith", startsWith);
        if (endsWith != null) node.put("endsWith", endsWith);
        if (includes != null) node.put("includes", includes);
        if (format != null) node.put("format", format);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new StringSchema(this);
    }
}

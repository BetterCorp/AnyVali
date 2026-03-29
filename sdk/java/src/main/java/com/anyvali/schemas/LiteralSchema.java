package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Objects;

/**
 * Schema that accepts only a specific literal value.
 */
public class LiteralSchema extends Schema {
    private final Object value;

    public LiteralSchema(Object value) {
        this.value = value;
    }

    private LiteralSchema(LiteralSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.value = other.value;
    }

    @Override
    protected boolean acceptsNull() {
        return value == null;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        // Strict type + value check
        boolean match;
        if (value == null) {
            match = (input == null);
        } else if (input == null) {
            match = false;
        } else {
            // Check both type and value match
            // Don't allow Boolean matching Integer, etc.
            match = value.getClass().equals(input.getClass()) && Objects.equals(value, input);
            // Also allow Long/Integer cross-matching for numeric types
            if (!match && value instanceof Number && input instanceof Number) {
                if (!(value instanceof Boolean) && !(input instanceof Boolean)) {
                    match = ((Number) value).longValue() == ((Number) input).longValue()
                            && ((Number) value).doubleValue() == ((Number) input).doubleValue();
                }
            }
        }

        if (!match) {
            ctx.addIssue(IssueCodes.INVALID_LITERAL,
                    "Expected literal " + value + ", received " + input,
                    value, input);
            return null;
        }
        return input;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "literal");
        node.put("value", value);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new LiteralSchema(this);
    }

    public Object getValue() {
        return value;
    }
}

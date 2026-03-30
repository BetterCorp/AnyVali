package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Schema that accepts one of a fixed set of values.
 */
public class EnumSchema extends Schema<Object> {
    private final List<Object> values;

    public EnumSchema(List<Object> values) {
        this.values = new ArrayList<>(values);
    }

    private EnumSchema(EnumSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.values = new ArrayList<>(other.values);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        // Check for type-strict match
        for (Object v : values) {
            if (v == null && input == null) return input;
            if (v != null && input != null) {
                // Allow Number cross-type matching (Long/Integer/etc)
                if (v instanceof Number && input instanceof Number
                        && !(v instanceof Boolean) && !(input instanceof Boolean)) {
                    if (((Number) v).longValue() == ((Number) input).longValue()
                            && ((Number) v).doubleValue() == ((Number) input).doubleValue()) {
                        return input;
                    }
                } else if (v.getClass().equals(input.getClass()) && v.equals(input)) {
                    return input;
                }
            }
        }

        String valuesStr = values.stream()
                .map(String::valueOf)
                .collect(Collectors.joining(","));
        ctx.addIssue(IssueCodes.INVALID_TYPE,
                "Expected one of enum(" + valuesStr + "), received " + input,
                "enum(" + valuesStr + ")", String.valueOf(input));
        return null;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "enum");
        node.put("values", new ArrayList<>(values));
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema<Object> copy() {
        return new EnumSchema(this);
    }
}

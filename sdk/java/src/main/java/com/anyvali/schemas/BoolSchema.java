package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema for boolean values.
 */
public class BoolSchema extends Schema<Boolean> {

    public BoolSchema() {}

    private BoolSchema(BoolSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (!(input instanceof Boolean)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected boolean, received " + received, "boolean", received);
            return null;
        }
        return input;
    }

    @Override
    protected Map<String, Object> toNode() {
        return addCommonNodeFields(new LinkedHashMap<>(Map.of("kind", "bool")));
    }

    @Override
    protected Schema<Boolean> copy() {
        return new BoolSchema(this);
    }
}

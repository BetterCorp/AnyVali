package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema that only accepts null.
 */
public class NullSchema extends Schema<Object> {

    public NullSchema() {}

    private NullSchema(NullSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
    }

    @Override
    protected boolean acceptsNull() {
        return true;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (input != null) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected null, received " + received, "null", received);
            return null;
        }
        return null;
    }

    @Override
    protected Map<String, Object> toNode() {
        return addCommonNodeFields(new LinkedHashMap<>(Map.of("kind", "null")));
    }

    @Override
    protected Schema<Object> copy() {
        return new NullSchema(this);
    }
}

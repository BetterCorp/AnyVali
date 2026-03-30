package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema that rejects all values.
 */
public class NeverSchema extends Schema<Object> {

    public NeverSchema() {}

    private NeverSchema(NeverSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        String received = input != null ? input.getClass().getSimpleName() : "null";
        ctx.addIssue(IssueCodes.INVALID_TYPE,
                "No value is allowed", "never", received);
        return null;
    }

    @Override
    protected Map<String, Object> toNode() {
        return addCommonNodeFields(new LinkedHashMap<>(Map.of("kind", "never")));
    }

    @Override
    protected Schema<Object> copy() {
        return new NeverSchema(this);
    }
}

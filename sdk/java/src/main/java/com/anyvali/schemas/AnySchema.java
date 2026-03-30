package com.anyvali.schemas;

import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema that accepts any value.
 */
public class AnySchema extends Schema<Object> {

    public AnySchema() {}

    private AnySchema(AnySchema other) {
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
        return input;
    }

    @Override
    protected Map<String, Object> toNode() {
        return addCommonNodeFields(new LinkedHashMap<>(Map.of("kind", "any")));
    }

    @Override
    protected Schema<Object> copy() {
        return new AnySchema(this);
    }
}

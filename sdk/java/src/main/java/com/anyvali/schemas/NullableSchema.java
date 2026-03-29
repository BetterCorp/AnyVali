package com.anyvali.schemas;

import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Wraps a schema to allow null values.
 */
public class NullableSchema extends Schema {
    private final Schema inner;

    public NullableSchema(Schema inner) {
        this.inner = inner;
    }

    private NullableSchema(NullableSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.inner = other.inner;
    }

    @Override
    protected boolean acceptsNull() {
        return true;
    }

    @Override
    public Object runPipeline(Object input, ValidationContext ctx) {
        if (input == null) {
            return null;
        }
        return inner.runPipeline(input, ctx);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (input == null) {
            return null;
        }
        return inner.validate(input, ctx);
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "nullable");
        node.put("schema", inner.toNode());
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new NullableSchema(this);
    }

    public Schema getInner() {
        return inner;
    }
}

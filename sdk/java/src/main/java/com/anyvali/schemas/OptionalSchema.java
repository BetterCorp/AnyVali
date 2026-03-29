package com.anyvali.schemas;

import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Wraps a schema to make it optional (accepts absent/undefined values).
 */
public class OptionalSchema extends Schema {
    private final Schema inner;

    public OptionalSchema(Schema inner) {
        this.inner = inner;
        // Inherit defaults from inner
        if (inner.hasDefault) {
            this.hasDefault = true;
            this.defaultValue = inner.defaultValue;
        }
    }

    private OptionalSchema(OptionalSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.inner = other.inner;
    }

    @Override
    protected boolean acceptsNull() {
        return inner.acceptsNull();
    }

    @Override
    public Object runPipeline(Object input, ValidationContext ctx) {
        // If absent (sentinel), return null without error
        if (input == ABSENT) {
            if (hasDefault) {
                return deepCopyValue(defaultValue);
            }
            return null;
        }
        return inner.runPipeline(input, ctx);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        return inner.validate(input, ctx);
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "optional");
        node.put("schema", inner.toNode());
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new OptionalSchema(this);
    }

    public Schema getInner() {
        return inner;
    }
}

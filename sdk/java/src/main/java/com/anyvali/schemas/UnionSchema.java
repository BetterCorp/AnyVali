package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Schema that accepts any of the given schemas (first match wins).
 */
public class UnionSchema extends Schema<Object> {
    private final List<Schema<?>> schemas;

    public UnionSchema(List<? extends Schema<?>> schemas) {
        this.schemas = new ArrayList<>(schemas);
    }

    private UnionSchema(UnionSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.schemas = new ArrayList<>(other.schemas);
    }

    @Override
    protected boolean acceptsNull() {
        return schemas.stream().anyMatch(Schema::acceptsNullValue);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        for (Schema<?> schema : schemas) {
            var trialCtx = new ValidationContext(
                    new ArrayList<>(ctx.getPath()),
                    new ArrayList<>(),
                    ctx.getDefinitions()
            );
            Object result = schema.runPipeline(input, trialCtx);
            if (!trialCtx.hasIssues()) {
                return result;
            }
        }

        String received = input != null ? input.getClass().getSimpleName() : "null";
        ctx.addIssue(IssueCodes.INVALID_UNION,
                "Input does not match any schema in the union",
                null, received);
        return null;
    }

    @Override
    protected Map<String, Object> toNode() {
        var variants = new ArrayList<Map<String, Object>>();
        for (Schema<?> s : schemas) {
            variants.add(s.toPortableNode());
        }
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "union");
        node.put("variants", variants);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema<Object> copy() {
        return new UnionSchema(this);
    }

    public List<Schema<?>> getSchemas() {
        return schemas;
    }
}

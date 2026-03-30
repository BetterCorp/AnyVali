package com.anyvali.schemas;

import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Schema that requires input to match all given schemas.
 */
public class IntersectionSchema extends Schema {
    private final List<Schema> schemas;

    public IntersectionSchema(List<Schema> schemas) {
        this.schemas = new ArrayList<>(schemas);
    }

    private IntersectionSchema(IntersectionSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.schemas = new ArrayList<>(other.schemas);
    }

    @Override
    protected boolean acceptsNull() {
        return schemas.stream().allMatch(Schema::acceptsNullValue);
    }

    @Override
    @SuppressWarnings("unchecked")
    protected Object validate(Object input, ValidationContext ctx) {
        List<Object> results = new ArrayList<>();
        boolean hasFailure = false;

        for (Schema schema : schemas) {
            var trialCtx = new ValidationContext(
                    new ArrayList<>(ctx.getPath()),
                    new ArrayList<>(),
                    ctx.getDefinitions()
            );
            Object result = schema.runPipeline(input, trialCtx);
            results.add(result);
            if (trialCtx.hasIssues()) {
                hasFailure = true;
                ctx.getIssues().addAll(trialCtx.getIssues());
            }
        }

        if (hasFailure) {
            return null;
        }

        // Merge map results
        if (results.stream().allMatch(r -> r instanceof Map)) {
            var merged = new LinkedHashMap<String, Object>();
            for (Object r : results) {
                merged.putAll((Map<String, Object>) r);
            }
            return merged;
        }

        // For non-map, return the last result
        return results.isEmpty() ? input : results.get(results.size() - 1);
    }

    @Override
    protected Map<String, Object> toNode() {
        var allOf = new ArrayList<Map<String, Object>>();
        for (Schema s : schemas) {
            allOf.add(s.toPortableNode());
        }
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "intersection");
        node.put("allOf", allOf);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new IntersectionSchema(this);
    }

    public List<Schema> getSchemas() {
        return schemas;
    }
}

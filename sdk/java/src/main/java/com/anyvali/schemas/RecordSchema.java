package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema for record/map with string keys and uniform value schema.
 */
public class RecordSchema extends Schema {
    private final Schema valueSchema;

    public RecordSchema(Schema valueSchema) {
        this.valueSchema = valueSchema;
    }

    private RecordSchema(RecordSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.valueSchema = other.valueSchema;
    }

    @Override
    @SuppressWarnings("unchecked")
    protected Object validate(Object input, ValidationContext ctx) {
        if (!(input instanceof Map<?, ?> rawMap)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected object, received " + received, "object", received);
            return null;
        }

        var result = new LinkedHashMap<String, Object>();
        for (var entry : rawMap.entrySet()) {
            Object key = entry.getKey();
            if (!(key instanceof String sKey)) {
                ctx.addIssue(IssueCodes.INVALID_TYPE,
                        "Record keys must be strings, received " + key.getClass().getSimpleName(),
                        "string", key.getClass().getSimpleName());
                continue;
            }
            var childCtx = ctx.child(sKey);
            Object parsed = valueSchema.runPipeline(entry.getValue(), childCtx);
            result.put(sKey, parsed);
        }
        return result;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "record");
        node.put("values", valueSchema.toNode());
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new RecordSchema(this);
    }

    public Schema getValueSchema() {
        return valueSchema;
    }
}

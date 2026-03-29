package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Schema for arrays/lists with element validation.
 */
public class ArraySchema extends Schema {
    private final Schema items;
    private Integer minItems;
    private Integer maxItems;

    public ArraySchema(Schema items) {
        this.items = items;
    }

    public ArraySchema(Schema items, Integer minItems, Integer maxItems) {
        this.items = items;
        this.minItems = minItems;
        this.maxItems = maxItems;
    }

    private ArraySchema(ArraySchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.items = other.items;
        this.minItems = other.minItems;
        this.maxItems = other.maxItems;
    }

    public ArraySchema minItems(int n) {
        var copy = new ArraySchema(this);
        copy.minItems = n;
        return copy;
    }

    public ArraySchema maxItems(int n) {
        var copy = new ArraySchema(this);
        copy.maxItems = n;
        return copy;
    }

    @Override
    @SuppressWarnings("unchecked")
    protected Object validate(Object input, ValidationContext ctx) {
        if (!(input instanceof List<?> list)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected array, received " + received, "array", received);
            return null;
        }

        if (minItems != null && list.size() < minItems) {
            ctx.addIssue(IssueCodes.TOO_SMALL,
                    "Array must have at least " + minItems + " item(s)",
                    minItems, list.size());
        }

        if (maxItems != null && list.size() > maxItems) {
            ctx.addIssue(IssueCodes.TOO_LARGE,
                    "Array must have at most " + maxItems + " item(s)",
                    maxItems, list.size());
        }

        var result = new ArrayList<>();
        for (int i = 0; i < list.size(); i++) {
            var childCtx = ctx.child(i);
            Object parsed = items.runPipeline(list.get(i), childCtx);
            result.add(parsed);
        }
        return result;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "array");
        node.put("items", items.toNode());
        if (minItems != null) node.put("minItems", minItems);
        if (maxItems != null) node.put("maxItems", maxItems);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new ArraySchema(this);
    }

    public Schema getItems() {
        return items;
    }
}

package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Schema for fixed-length tuples with positional element schemas.
 */
public class TupleSchema extends Schema {
    private final List<Schema> items;

    public TupleSchema(List<Schema> items) {
        this.items = new ArrayList<>(items);
    }

    private TupleSchema(TupleSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.items = new ArrayList<>(other.items);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (!(input instanceof List<?> list)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected tuple, received " + received, "tuple", received);
            return null;
        }

        if (list.size() != items.size()) {
            String code = list.size() < items.size() ? IssueCodes.TOO_SMALL : IssueCodes.TOO_LARGE;
            ctx.addIssue(code,
                    "Expected exactly " + items.size() + " items, received " + list.size(),
                    items.size(), list.size());
            return null;
        }

        var result = new ArrayList<>();
        for (int i = 0; i < items.size(); i++) {
            var childCtx = ctx.child(i);
            Object parsed = items.get(i).runPipeline(list.get(i), childCtx);
            result.add(parsed);
        }
        return result;
    }

    @Override
    protected Map<String, Object> toNode() {
        var elements = new ArrayList<Map<String, Object>>();
        for (Schema item : items) {
            elements.add(item.toNode());
        }
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "tuple");
        node.put("elements", elements);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new TupleSchema(this);
    }

    public List<Schema> getItems() {
        return items;
    }
}

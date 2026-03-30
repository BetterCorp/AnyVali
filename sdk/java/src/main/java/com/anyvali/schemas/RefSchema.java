package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Schema that references a named definition.
 */
public class RefSchema extends Schema {
    private final String ref;
    private Schema resolved;
    private Map<String, Schema> definitions;

    public RefSchema(String ref) {
        this.ref = ref;
    }

    private RefSchema(RefSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.ref = other.ref;
        this.resolved = other.resolved;
        this.definitions = other.definitions;
    }

    public void resolve(Schema schema) {
        this.resolved = schema;
    }

    public void setDefinitions(Map<String, Schema> definitions) {
        this.definitions = definitions;
    }

    private Schema getResolved() {
        if (resolved != null) return resolved;
        if (definitions != null) {
            String refName = ref;
            if (refName.startsWith("#/definitions/")) {
                refName = refName.substring("#/definitions/".length());
            }
            if (definitions.containsKey(refName)) {
                return definitions.get(refName);
            }
        }
        return null;
    }

    @Override
    protected boolean acceptsNull() {
        Schema r = getResolved();
        return r != null && r.acceptsNullValue();
    }

    @Override
    public Object runPipeline(Object input, ValidationContext ctx) {
        Schema r = getResolved();
        if (r != null) {
            return r.runPipeline(input, ctx);
        }
        // Try context definitions
        String refName = ref;
        if (refName.startsWith("#/definitions/")) {
            refName = refName.substring("#/definitions/".length());
        }
        if (ctx.getDefinitions().containsKey(refName)) {
            return ctx.getDefinitions().get(refName).runPipeline(input, ctx);
        }
        ctx.addIssue(IssueCodes.INVALID_TYPE,
                "Unresolved reference: " + ref, ref, null);
        return null;
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        Schema r = getResolved();
        if (r != null) {
            return r.validateValue(input, ctx);
        }
        ctx.addIssue(IssueCodes.INVALID_TYPE,
                "Unresolved reference: " + ref, ref, null);
        return null;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "ref");
        node.put("ref", ref);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new RefSchema(this);
    }

    public String getRef() {
        return ref;
    }
}

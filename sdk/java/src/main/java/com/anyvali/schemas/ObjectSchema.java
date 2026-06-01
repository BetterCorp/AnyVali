package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.UnknownKeyMode;
import com.anyvali.ValidationContext;

import java.util.*;

/**
 * Schema for objects/maps with named properties.
 */
public class ObjectSchema extends Schema<Map<String, Object>> {
    private final Map<String, Schema<?>> properties;
    private final Set<String> required;
    private final UnknownKeyMode unknownKeys;
    private final boolean unknownKeysExplicit;

    /**
     * Create an object schema. If required is null, all properties are required by default.
     */
    public ObjectSchema(Map<String, ? extends Schema<?>> properties, Set<String> required,
                        UnknownKeyMode unknownKeys) {
        this(properties, required, unknownKeys, true);
    }

    private ObjectSchema(Map<String, ? extends Schema<?>> properties, Set<String> required,
                         UnknownKeyMode unknownKeys, boolean unknownKeysExplicit) {
        this.properties = new LinkedHashMap<>(properties);
        this.required = required != null ? new LinkedHashSet<>(required) : new LinkedHashSet<>(properties.keySet());
        this.unknownKeys = unknownKeys != null ? unknownKeys : UnknownKeyMode.STRIP;
        this.unknownKeysExplicit = unknownKeysExplicit;
    }

    public ObjectSchema(Map<String, ? extends Schema<?>> properties) {
        this(properties, null, UnknownKeyMode.STRIP, false);
    }

    public ObjectSchema(Map<String, ? extends Schema<?>> properties, Set<String> required) {
        this(properties, required, UnknownKeyMode.STRIP, false);
    }

    private ObjectSchema(ObjectSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.metadata = other.metadata;
        this.properties = new LinkedHashMap<>(other.properties);
        this.required = new LinkedHashSet<>(other.required);
        this.unknownKeys = other.unknownKeys;
        this.unknownKeysExplicit = other.unknownKeysExplicit;
    }

    private UnknownKeyMode effectiveUnknownKeys() {
        return unknownKeysExplicit ? unknownKeys : UnknownKeyMode.STRIP;
    }

    private UnknownKeyMode exportUnknownKeys() {
        return unknownKeysExplicit ? unknownKeys : UnknownKeyMode.STRIP;
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

        Map<String, Object> map = (Map<String, Object>) rawMap;
        var result = new LinkedHashMap<String, Object>();

        // Validate known properties
        for (var entry : properties.entrySet()) {
            String key = entry.getKey();
            Schema<?> schema = entry.getValue();

            if (map.containsKey(key)) {
                var childCtx = ctx.child(key);
                Object parsed = schema.runPipeline(map.get(key), childCtx);
                result.put(key, parsed);
            } else if (required.contains(key)) {
                if (schema.hasDefaultValue()) {
                    // Apply default via pipeline
                    var childCtx = ctx.child(key);
                    Object parsed = schema.runPipeline(ABSENT, childCtx);
                    result.put(key, parsed);
                } else {
                    var childCtx = ctx.child(key);
                    childCtx.addIssue(IssueCodes.REQUIRED,
                            "Required field '" + key + "' is missing", key, null);
                }
            } else {
                // Optional and not present - check for defaults
                if (schema.hasDefaultValue()) {
                    var childCtx = ctx.child(key);
                    Object parsed = schema.runPipeline(ABSENT, childCtx);
                    result.put(key, parsed);
                }
            }
        }

        // Handle unknown keys
        Set<String> known = properties.keySet();
        for (String key : map.keySet()) {
            if (!known.contains(key)) {
                switch (effectiveUnknownKeys()) {
                    case REJECT -> {
                        var childCtx = ctx.child(key);
                        childCtx.addIssue(IssueCodes.UNKNOWN_KEY,
                                "Unknown key '" + key + "'", null, key);
                    }
                    case STRIP -> {
                        // Just don't include
                    }
                    case ALLOW -> result.put(key, map.get(key));
                }
            }
        }

        return result;
    }

    @Override
    protected Map<String, Object> toNode() {
        var props = new LinkedHashMap<String, Object>();
        for (var entry : properties.entrySet()) {
            props.put(entry.getKey(), entry.getValue().toPortableNode());
        }
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", "object");
        node.put("properties", props);
        node.put("required", new ArrayList<>(required).stream().sorted().toList());
        node.put("unknownKeys", exportUnknownKeys().getValue());
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema<Map<String, Object>> copy() {
        return new ObjectSchema(this);
    }

    public Map<String, Schema<?>> getProperties() {
        return properties;
    }

    public Set<String> getRequired() {
        return required;
    }

    public UnknownKeyMode getUnknownKeys() {
        return effectiveUnknownKeys();
    }
}

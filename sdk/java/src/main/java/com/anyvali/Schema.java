package com.anyvali;

import com.anyvali.parse.Coercion;
import com.anyvali.parse.CoercionConfig;
import com.anyvali.schemas.NullableSchema;
import com.anyvali.schemas.OptionalSchema;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Abstract base class for all AnyVali schema types.
 * <p>
 * Schemas are immutable: modifier methods return new instances.
 * parse() throws ValidationError; safeParse() returns ParseResult.
 *
 * @param <T> the Java type that this schema parses to
 */
public abstract class Schema<T> {

    /** Sentinel value representing an absent field. */
    public static final Object ABSENT = new Object() {
        @Override
        public String toString() {
            return "<absent>";
        }
    };

    protected CoercionConfig coercion;
    protected Object defaultValue = ABSENT;
    protected boolean hasDefault = false;

    protected Schema() {
    }

    // ---- Public parse API ----

    /**
     * Parse input, throwing ValidationError on failure.
     */
    @SuppressWarnings("unchecked")
    public T parse(Object input) {
        ParseResult<T> result = safeParse(input);
        if (!result.success()) {
            throw new ValidationError(result.issues());
        }
        return result.data();
    }

    /**
     * Parse input, returning a ParseResult.
     */
    @SuppressWarnings("unchecked")
    public ParseResult<T> safeParse(Object input) {
        ValidationContext ctx = new ValidationContext();
        Object value = runPipeline(input, ctx);
        if (ctx.hasIssues()) {
            return ParseResult.failure(ctx.getIssues());
        }
        return ParseResult.success((T) value);
    }

    // ---- 5-step pipeline ----

    /**
     * Run the full parse pipeline: presence check, coercion, defaults, validation.
     */
    public Object runPipeline(Object input, ValidationContext ctx) {
        // Step 1: presence check
        boolean isAbsent = (input == ABSENT) ||
                (input == null && !acceptsNull());

        if (input == null && acceptsNull()) {
            isAbsent = false;
        }

        // Step 2: coercion (only if present)
        Object value = input;
        if (!isAbsent && coercion != null) {
            int issuesBefore = ctx.issueCount();
            value = Coercion.applyCoercion(value, coercion, ctx);
            if (ctx.issueCount() > issuesBefore) {
                return null;
            }
        }

        // Step 3: defaults (only if absent)
        boolean usedDefault = false;
        if (isAbsent && hasDefault) {
            value = deepCopyValue(defaultValue);
            isAbsent = false;
            usedDefault = true;
        }

        // Step 4: validate
        if (usedDefault) {
            int issuesBefore = ctx.issueCount();
            Object result = validate(value, ctx);
            if (ctx.issueCount() > issuesBefore) {
                // Replace new issues with a single default_invalid
                var allIssues = ctx.getIssues();
                var newIssues = allIssues.subList(issuesBefore, allIssues.size());
                String expectedStr = !newIssues.isEmpty() && newIssues.get(0).expected() != null
                        ? String.valueOf(newIssues.get(0).expected()) : null;
                newIssues.clear();
                ctx.addIssue(IssueCodes.DEFAULT_INVALID,
                        "Default value " + defaultValue + " is invalid",
                        expectedStr, String.valueOf(defaultValue));
                return null;
            }
            return result;
        } else {
            return validate(value, ctx);
        }
    }

    /**
     * Whether this schema accepts null as a valid value.
     */
    protected boolean acceptsNull() {
        return false;
    }

    public boolean acceptsNullValue() {
        return acceptsNull();
    }

    // ---- Abstract methods ----

    /**
     * Validate and return parsed value, adding issues to ctx.
     */
    protected abstract Object validate(Object input, ValidationContext ctx);

    public Object validateValue(Object input, ValidationContext ctx) {
        return validate(input, ctx);
    }

    /**
     * Return the schema node map for interchange export.
     */
    protected abstract Map<String, Object> toNode();

    public Map<String, Object> toPortableNode() {
        return toNode();
    }

    // ---- Modifier methods (return new instances) ----

    /**
     * Wrap this schema as optional.
     */
    public OptionalSchema optional() {
        return new OptionalSchema(this);
    }

    /**
     * Wrap this schema as nullable.
     */
    public NullableSchema nullable() {
        return new NullableSchema(this);
    }

    /**
     * Set a default value, returning a new schema instance.
     */
    public Schema<T> withDefault(Object value) {
        Schema<T> copy = copy();
        copy.defaultValue = value;
        copy.hasDefault = true;
        return copy;
    }

    /**
     * Enable coercion with the given config, returning a new schema instance.
     */
    public Schema<T> coerce(CoercionConfig config) {
        Schema<T> copy = copy();
        copy.coercion = config;
        return copy;
    }

    // ---- Export ----

    /**
     * Export this schema as an AnyVali document map.
     */
    public Map<String, Object> export() {
        return export(ExportMode.PORTABLE);
    }

    /**
     * Export this schema as an AnyVali document map.
     */
    public Map<String, Object> export(ExportMode mode) {
        Map<String, Object> node = toNode();
        AnyValiDocument doc = new AnyValiDocument(node);
        return doc.toMap();
    }

    /**
     * Add common node fields (default, coercion) to a node map.
     */
    protected Map<String, Object> addCommonNodeFields(Map<String, Object> node) {
        if (hasDefault) {
            node.put("default", defaultValue);
        }
        if (coercion != null) {
            var coerceMap = new LinkedHashMap<String, Object>();
            if (coercion.toInt()) coerceMap.put("toInt", true);
            if (coercion.toNumber()) coerceMap.put("toNumber", true);
            if (coercion.toBool()) coerceMap.put("toBool", true);
            if (coercion.trim()) coerceMap.put("trim", true);
            if (coercion.lower()) coerceMap.put("lower", true);
            if (coercion.upper()) coerceMap.put("upper", true);
            if (!coerceMap.isEmpty()) {
                node.put("coerce", coerceMap);
            }
        }
        return node;
    }

    /**
     * Create a deep copy of this schema (subclasses must implement).
     */
    protected abstract Schema<T> copy();

    public boolean hasDefaultValue() {
        return hasDefault;
    }

    public Object getDefaultValue() {
        return defaultValue;
    }

    public CoercionConfig getCoercionConfig() {
        return coercion;
    }

    public void applyImportedDefault(Object value) {
        defaultValue = value;
        hasDefault = true;
    }

    public void applyImportedCoercion(CoercionConfig config) {
        coercion = config;
    }

    // ---- Utility ----

    /**
     * Map a Java value to its AnyVali portable type name.
     */
    public static String anyvaliTypeName(Object value) {
        if (value == null) return "null";
        if (value instanceof Boolean) return "boolean";
        if (value instanceof Integer || value instanceof Long
                || value instanceof Short || value instanceof Byte) return "integer";
        if (value instanceof Float || value instanceof Double) return "number";
        if (value instanceof String) return "string";
        if (value instanceof java.util.List) return "array";
        if (value instanceof java.util.Map) return "object";
        return value.getClass().getSimpleName();
    }

    /**
     * Deep copy a value (for defaults).
     */
    @SuppressWarnings("unchecked")
    public static Object deepCopyValue(Object value) {
        if (value == null || value instanceof String || value instanceof Number
                || value instanceof Boolean) {
            return value;
        }
        if (value instanceof java.util.List<?> list) {
            var copy = new java.util.ArrayList<>();
            for (Object item : list) {
                copy.add(deepCopyValue(item));
            }
            return copy;
        }
        if (value instanceof Map<?, ?> map) {
            var copy = new LinkedHashMap<>();
            for (var entry : map.entrySet()) {
                copy.put(entry.getKey(), deepCopyValue(entry.getValue()));
            }
            return copy;
        }
        return value;
    }
}

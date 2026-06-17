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

    private static final java.util.Set<String> RESERVED_METADATA_KEYS = java.util.Set.of(
        "title", "description", "deprecated", "deprecatedMessage",
        "notStable", "since", "sensitive", "readonly", "writeonly", "examples"
    );

    protected CoercionConfig coercion;
    protected Object defaultValue = ABSENT;
    protected boolean hasDefault = false;
    protected Map<String, Object> metadata;

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

    /**
     * Enable coercion from string with the target type inferred from this
     * schema's kind, returning a new schema instance.
     * <p>
     * "string" is the only portable coercion source, so this ergonomic overload
     * enables string-to-target coercion: numeric schemas coerce ASCII decimal
     * strings; boolean schemas coerce boolean strings. The accepted/rejected
     * grammars are identical to the explicit {@code coerce(...)} form.
     * <p>
     * Schema kinds that have no portable string source (e.g. string itself,
     * objects, arrays) cannot infer a target and must use the explicit
     * {@link #coerce(CoercionConfig)} overload instead.
     *
     * @throws IllegalStateException if this schema kind cannot infer a coercion
     *         target from string
     */
    public Schema<T> coerce() {
        CoercionConfig config = defaultCoercionConfig();
        if (config == null) {
            throw new IllegalStateException(
                    "coerce() with no arguments cannot infer a coercion target for this "
                            + "schema kind; use coerce(CoercionConfig) to configure string "
                            + "transformations or an explicit target");
        }
        return coerce(config);
    }

    /**
     * The CoercionConfig used by the no-arg {@link #coerce()} ergonomic, with
     * the string-to-target enabled based on this schema's kind.
     * <p>
     * Returns {@code null} for kinds that cannot infer a portable string source.
     * Numeric and boolean schemas override this.
     */
    protected CoercionConfig defaultCoercionConfig() {
        return null;
    }

    // ---- Describe & Metadata ----

    /**
     * Options for the describe() method.
     */
    public static class DescribeOptions {
        private String title;
        private Boolean deprecated;
        private String deprecatedMessage;
        private Boolean notStable;
        private String since;
        private Boolean sensitive;
        private Boolean readonly;
        private Boolean writeonly;
        private java.util.List<Object> examples;

        public DescribeOptions title(String t) { this.title = t; return this; }
        public DescribeOptions deprecated(boolean d) { this.deprecated = d; return this; }
        public DescribeOptions deprecatedMessage(String m) { this.deprecatedMessage = m; return this; }
        public DescribeOptions notStable(boolean n) { this.notStable = n; return this; }
        public DescribeOptions since(String s) { this.since = s; return this; }
        public DescribeOptions sensitive(boolean s) { this.sensitive = s; return this; }
        public DescribeOptions readonly(boolean r) { this.readonly = r; return this; }
        public DescribeOptions writeonly(boolean w) { this.writeonly = w; return this; }
        public DescribeOptions examples(java.util.List<Object> e) { this.examples = e; return this; }
    }

    /**
     * Add documentation metadata. Returns a new schema instance.
     */
    public Schema<T> describe(String description) {
        return describe(description, null);
    }

    /**
     * Add documentation metadata with options. Returns a new schema instance.
     */
    public Schema<T> describe(String description, DescribeOptions opts) {
        if (description == null) throw new IllegalArgumentException("describe(): description must not be null");
        Schema<T> copy = copy();
        if (copy.metadata == null) copy.metadata = new LinkedHashMap<>();
        copy.metadata.put("description", description);
        if (opts != null) {
            if (opts.title != null) copy.metadata.put("title", opts.title);
            if (opts.deprecated != null) copy.metadata.put("deprecated", opts.deprecated);
            if (opts.deprecatedMessage != null) {
                if (opts.deprecated == null || !opts.deprecated) {
                    throw new IllegalArgumentException("describe(): deprecatedMessage requires deprecated to be true");
                }
                copy.metadata.put("deprecatedMessage", opts.deprecatedMessage);
            }
            if (opts.notStable != null) copy.metadata.put("notStable", opts.notStable);
            if (opts.since != null) copy.metadata.put("since", opts.since);
            if (opts.sensitive != null) copy.metadata.put("sensitive", opts.sensitive);
            if (opts.readonly != null) copy.metadata.put("readonly", opts.readonly);
            if (opts.writeonly != null) copy.metadata.put("writeonly", opts.writeonly);
            if (Boolean.TRUE.equals(opts.readonly) && Boolean.TRUE.equals(opts.writeonly)) {
                throw new IllegalArgumentException("describe(): readonly and writeonly cannot both be true");
            }
            if (opts.examples != null) copy.metadata.put("examples", new java.util.ArrayList<>(opts.examples));
        }
        return copy;
    }

    /**
     * Attach arbitrary metadata. Reserved keys must use describe().
     */
    public Schema<T> metadata(Map<String, Object> meta) {
        return metadata(meta, false);
    }

    /**
     * Attach arbitrary metadata. If replace is true, replaces non-reserved metadata.
     */
    public Schema<T> metadata(Map<String, Object> meta, boolean replace) {
        for (String key : meta.keySet()) {
            if (RESERVED_METADATA_KEYS.contains(key)) {
                throw new IllegalArgumentException(
                    "metadata(): \"" + key + "\" is a reserved key. Use describe() instead."
                );
            }
        }
        Schema<T> copy = copy();
        if (replace) {
            Map<String, Object> preserved = new LinkedHashMap<>();
            if (copy.metadata != null) {
                for (var entry : copy.metadata.entrySet()) {
                    if (RESERVED_METADATA_KEYS.contains(entry.getKey())) {
                        preserved.put(entry.getKey(), entry.getValue());
                    }
                }
            }
            preserved.putAll(meta);
            copy.metadata = preserved;
        } else {
            if (copy.metadata == null) copy.metadata = new LinkedHashMap<>();
            copy.metadata.putAll(meta);
        }
        return copy;
    }

    /**
     * Get the metadata map.
     */
    public Map<String, Object> getMetadata() {
        return metadata;
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
        if (metadata != null && !metadata.isEmpty()) {
            node.put("metadata", new LinkedHashMap<>(metadata));
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

    public void applyImportedMetadata(Map<String, Object> meta) {
        this.metadata = meta;
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

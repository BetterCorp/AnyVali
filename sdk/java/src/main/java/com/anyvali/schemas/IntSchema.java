package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Base class for integer schemas with explicit bit-width bounds.
 * IntSchema (int) is an alias for int64.
 * <p>
 * Subclasses: Int8Schema, Int16Schema, Int32Schema, Int64Schema,
 *             Uint8Schema, Uint16Schema, Uint32Schema, Uint64Schema.
 */
public class IntSchema extends Schema {
    protected final String kind;
    protected final long rangeMin;
    protected final long rangeMax;
    // For uint64, the max exceeds Long.MAX_VALUE so we need special handling
    protected final boolean isUint64;

    protected Long min;
    protected Long max;
    protected Long exclusiveMin;
    protected Long exclusiveMax;
    protected Long multipleOf;

    // Ranges for each integer type
    private static final long INT8_MIN = -128;
    private static final long INT8_MAX = 127;
    private static final long INT16_MIN = -32768;
    private static final long INT16_MAX = 32767;
    private static final long INT32_MIN = -2147483648L;
    private static final long INT32_MAX = 2147483647L;
    private static final long INT64_MIN = Long.MIN_VALUE;
    private static final long INT64_MAX = Long.MAX_VALUE;
    private static final long UINT8_MIN = 0;
    private static final long UINT8_MAX = 255;
    private static final long UINT16_MIN = 0;
    private static final long UINT16_MAX = 65535;
    private static final long UINT32_MIN = 0;
    private static final long UINT32_MAX = 4294967295L;
    // uint64 max is 2^64 - 1, which exceeds Long.MAX_VALUE
    // We use Long.MAX_VALUE as a proxy and set isUint64 flag

    public IntSchema() {
        this("int", INT64_MIN, INT64_MAX, false);
    }

    protected IntSchema(String kind, long rangeMin, long rangeMax, boolean isUint64) {
        this.kind = kind;
        this.rangeMin = rangeMin;
        this.rangeMax = rangeMax;
        this.isUint64 = isUint64;
    }

    protected IntSchema(IntSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.kind = other.kind;
        this.rangeMin = other.rangeMin;
        this.rangeMax = other.rangeMax;
        this.isUint64 = other.isUint64;
        this.min = other.min;
        this.max = other.max;
        this.exclusiveMin = other.exclusiveMin;
        this.exclusiveMax = other.exclusiveMax;
        this.multipleOf = other.multipleOf;
    }

    public IntSchema min(long v) {
        var copy = copyAs();
        copy.min = v;
        return copy;
    }

    public IntSchema max(long v) {
        var copy = copyAs();
        copy.max = v;
        return copy;
    }

    public IntSchema exclusiveMin(long v) {
        var copy = copyAs();
        copy.exclusiveMin = v;
        return copy;
    }

    public IntSchema exclusiveMax(long v) {
        var copy = copyAs();
        copy.exclusiveMax = v;
        return copy;
    }

    public IntSchema multipleOf(long v) {
        var copy = copyAs();
        copy.multipleOf = v;
        return copy;
    }

    protected IntSchema copyAs() {
        return new IntSchema(this);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        // Reject booleans
        if (input instanceof Boolean) {
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected integer, received boolean", kind, "boolean");
            return null;
        }

        // Accept Number types, convert floats that are whole numbers
        long value;
        if (input instanceof Double d || input instanceof Float) {
            double dv = ((Number) input).doubleValue();
            if (Double.isNaN(dv) || Double.isInfinite(dv) || dv != Math.floor(dv)) {
                String received = anyvaliTypeName(input);
                ctx.addIssue(IssueCodes.INVALID_TYPE,
                        "Expected integer, received " + received, kind, received);
                return null;
            }
            value = (long) dv;
        } else if (input instanceof Number n) {
            value = n.longValue();
        } else {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected integer, received " + received, kind, received);
            return null;
        }

        // Range check for the specific integer width
        if (!isUint64) {
            if (value > rangeMax) {
                ctx.addIssue(IssueCodes.TOO_LARGE,
                        "Value " + value + " is out of " + kind + " range [" + rangeMin + ", " + rangeMax + "]",
                        kind, String.valueOf(value));
                return null;
            }
            if (value < rangeMin) {
                ctx.addIssue(IssueCodes.TOO_SMALL,
                        "Value " + value + " is out of " + kind + " range [" + rangeMin + ", " + rangeMax + "]",
                        kind, String.valueOf(value));
                return null;
            }
        } else {
            // uint64: value must be >= 0 (since Java long is signed, negative means out of range)
            if (value < 0) {
                ctx.addIssue(IssueCodes.TOO_SMALL,
                        "Value " + value + " is out of uint64 range [0, 18446744073709551615]",
                        kind, String.valueOf(value));
                return null;
            }
        }

        // User constraints
        if (min != null && value < min) {
            ctx.addIssue(IssueCodes.TOO_SMALL,
                    "Number must be >= " + min, min, value);
        }
        if (max != null && value > max) {
            ctx.addIssue(IssueCodes.TOO_LARGE,
                    "Number must be <= " + max, max, value);
        }
        if (exclusiveMin != null && value <= exclusiveMin) {
            ctx.addIssue(IssueCodes.TOO_SMALL,
                    "Number must be > " + exclusiveMin, exclusiveMin, value);
        }
        if (exclusiveMax != null && value >= exclusiveMax) {
            ctx.addIssue(IssueCodes.TOO_LARGE,
                    "Number must be < " + exclusiveMax, exclusiveMax, value);
        }
        if (multipleOf != null && multipleOf != 0) {
            if (value % multipleOf != 0) {
                ctx.addIssue(IssueCodes.INVALID_NUMBER,
                        "Number must be a multiple of " + multipleOf,
                        multipleOf, value);
            }
        }

        return value;
    }

    @Override
    protected Map<String, Object> toNode() {
        var node = new LinkedHashMap<String, Object>();
        node.put("kind", kind);
        if (min != null) node.put("min", min);
        if (max != null) node.put("max", max);
        if (exclusiveMin != null) node.put("exclusiveMin", exclusiveMin);
        if (exclusiveMax != null) node.put("exclusiveMax", exclusiveMax);
        if (multipleOf != null) node.put("multipleOf", multipleOf);
        return addCommonNodeFields(node);
    }

    @Override
    protected Schema copy() {
        return new IntSchema(this);
    }

    // ---- Static factory methods for specific integer widths ----

    public static IntSchema int8() {
        return new Int8Schema();
    }

    public static IntSchema int16() {
        return new Int16Schema();
    }

    public static IntSchema int32() {
        return new Int32Schema();
    }

    public static IntSchema int64() {
        return new Int64Schema();
    }

    public static IntSchema uint8() {
        return new Uint8Schema();
    }

    public static IntSchema uint16() {
        return new Uint16Schema();
    }

    public static IntSchema uint32() {
        return new Uint32Schema();
    }

    public static IntSchema uint64() {
        return new Uint64Schema();
    }

    // ---- Inner classes for each width ----

    public static class Int8Schema extends IntSchema {
        public Int8Schema() { super("int8", INT8_MIN, INT8_MAX, false); }
        private Int8Schema(Int8Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Int8Schema(this); }
        @Override protected Schema copy() { return new Int8Schema(this); }
    }

    public static class Int16Schema extends IntSchema {
        public Int16Schema() { super("int16", INT16_MIN, INT16_MAX, false); }
        private Int16Schema(Int16Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Int16Schema(this); }
        @Override protected Schema copy() { return new Int16Schema(this); }
    }

    public static class Int32Schema extends IntSchema {
        public Int32Schema() { super("int32", INT32_MIN, INT32_MAX, false); }
        private Int32Schema(Int32Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Int32Schema(this); }
        @Override protected Schema copy() { return new Int32Schema(this); }
    }

    public static class Int64Schema extends IntSchema {
        public Int64Schema() { super("int64", INT64_MIN, INT64_MAX, false); }
        private Int64Schema(Int64Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Int64Schema(this); }
        @Override protected Schema copy() { return new Int64Schema(this); }
    }

    public static class Uint8Schema extends IntSchema {
        public Uint8Schema() { super("uint8", UINT8_MIN, UINT8_MAX, false); }
        private Uint8Schema(Uint8Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Uint8Schema(this); }
        @Override protected Schema copy() { return new Uint8Schema(this); }
    }

    public static class Uint16Schema extends IntSchema {
        public Uint16Schema() { super("uint16", UINT16_MIN, UINT16_MAX, false); }
        private Uint16Schema(Uint16Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Uint16Schema(this); }
        @Override protected Schema copy() { return new Uint16Schema(this); }
    }

    public static class Uint32Schema extends IntSchema {
        public Uint32Schema() { super("uint32", UINT32_MIN, UINT32_MAX, false); }
        private Uint32Schema(Uint32Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Uint32Schema(this); }
        @Override protected Schema copy() { return new Uint32Schema(this); }
    }

    public static class Uint64Schema extends IntSchema {
        public Uint64Schema() { super("uint64", 0, Long.MAX_VALUE, true); }
        private Uint64Schema(Uint64Schema other) { super(other); }
        @Override protected IntSchema copyAs() { return new Uint64Schema(this); }
        @Override protected Schema copy() { return new Uint64Schema(this); }
    }
}

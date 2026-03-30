package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Base class for number/float schemas with numeric constraints.
 * NumberSchema is the number alias (= float64).
 */
public class NumberSchema extends Schema<Double> {
    protected String kind;
    protected Double min;
    protected Double max;
    protected Double exclusiveMin;
    protected Double exclusiveMax;
    protected Double multipleOf;

    public NumberSchema() {
        this.kind = "number";
    }

    protected NumberSchema(String kind) {
        this.kind = kind;
    }

    protected NumberSchema(NumberSchema other) {
        this.coercion = other.coercion;
        this.defaultValue = other.defaultValue;
        this.hasDefault = other.hasDefault;
        this.kind = other.kind;
        this.min = other.min;
        this.max = other.max;
        this.exclusiveMin = other.exclusiveMin;
        this.exclusiveMax = other.exclusiveMax;
        this.multipleOf = other.multipleOf;
    }

    public NumberSchema min(double v) {
        var copy = copyAs();
        copy.min = v;
        return copy;
    }

    public NumberSchema max(double v) {
        var copy = copyAs();
        copy.max = v;
        return copy;
    }

    public NumberSchema exclusiveMin(double v) {
        var copy = copyAs();
        copy.exclusiveMin = v;
        return copy;
    }

    public NumberSchema exclusiveMax(double v) {
        var copy = copyAs();
        copy.exclusiveMax = v;
        return copy;
    }

    public NumberSchema multipleOf(double v) {
        var copy = copyAs();
        copy.multipleOf = v;
        return copy;
    }

    protected NumberSchema copyAs() {
        return new NumberSchema(this);
    }

    @Override
    protected Object validate(Object input, ValidationContext ctx) {
        if (!checkIsNumber(input, ctx)) {
            return null;
        }

        double value = toDouble(input);
        checkRange(value, ctx);
        checkConstraints(value, ctx);
        return value;
    }

    protected boolean checkIsNumber(Object input, ValidationContext ctx) {
        if (input instanceof Boolean) {
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected " + kind + ", received boolean", kind, "boolean");
            return false;
        }
        if (!(input instanceof Number)) {
            String received = anyvaliTypeName(input);
            ctx.addIssue(IssueCodes.INVALID_TYPE,
                    "Expected " + kind + ", received " + received, kind, received);
            return false;
        }
        double d = ((Number) input).doubleValue();
        if (Double.isNaN(d) || Double.isInfinite(d)) {
            ctx.addIssue(IssueCodes.INVALID_NUMBER,
                    "Number must be finite", "finite number", input);
            return false;
        }
        return true;
    }

    protected void checkRange(double value, ValidationContext ctx) {
        // Overridden in Float32Schema
    }

    protected void checkConstraints(double value, ValidationContext ctx) {
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
    protected Schema<Double> copy() {
        return new NumberSchema(this);
    }

    protected static double toDouble(Object input) {
        return ((Number) input).doubleValue();
    }
}

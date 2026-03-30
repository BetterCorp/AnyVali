package com.anyvali.schemas;

import com.anyvali.IssueCodes;
import com.anyvali.Schema;
import com.anyvali.ValidationContext;

/**
 * Schema for float32 values with range checking.
 */
public class Float32Schema extends NumberSchema {
    private static final double FLOAT32_MAX = 3.4028235e+38;

    public Float32Schema() {
        super("float32");
    }

    private Float32Schema(Float32Schema other) {
        super(other);
    }

    @Override
    protected void checkRange(double value, ValidationContext ctx) {
        if (Math.abs(value) > FLOAT32_MAX && value != 0.0) {
            ctx.addIssue(IssueCodes.INVALID_NUMBER,
                    "Value " + value + " is out of float32 range",
                    "float32", value);
        }
    }

    @Override
    protected NumberSchema copyAs() {
        return new Float32Schema(this);
    }

    @Override
    protected Schema<Double> copy() {
        return new Float32Schema(this);
    }
}

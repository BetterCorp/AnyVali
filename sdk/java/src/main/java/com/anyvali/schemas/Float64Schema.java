package com.anyvali.schemas;

import com.anyvali.Schema;

/**
 * Schema for float64 values.
 */
public class Float64Schema extends NumberSchema {

    public Float64Schema() {
        super("float64");
    }

    private Float64Schema(Float64Schema other) {
        super(other);
    }

    @Override
    protected NumberSchema copyAs() {
        return new Float64Schema(this);
    }

    @Override
    protected Schema copy() {
        return new Float64Schema(this);
    }
}

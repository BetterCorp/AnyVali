package com.anyvali.parse;

import com.anyvali.Schema;

/**
 * Default value application utilities.
 */
public final class Defaults {
    private Defaults() {}

    /**
     * Deep copy a default value before applying it.
     */
    public static Object applyDefault(Object defaultValue) {
        return Schema.deepCopyValue(defaultValue);
    }
}

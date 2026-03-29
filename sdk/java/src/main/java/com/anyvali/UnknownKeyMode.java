package com.anyvali;

/**
 * How to handle unknown keys in object schemas.
 */
public enum UnknownKeyMode {
    REJECT("reject"),
    STRIP("strip"),
    ALLOW("allow");

    private final String value;

    UnknownKeyMode(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    public static UnknownKeyMode fromString(String s) {
        for (UnknownKeyMode m : values()) {
            if (m.value.equals(s)) {
                return m;
            }
        }
        throw new IllegalArgumentException("Unknown key mode: " + s);
    }
}

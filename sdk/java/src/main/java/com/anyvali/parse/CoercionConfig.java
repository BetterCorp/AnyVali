package com.anyvali.parse;

/**
 * Configuration for coercion behavior.
 */
public record CoercionConfig(
        boolean toInt,
        boolean toNumber,
        boolean toBool,
        boolean trim,
        boolean lower,
        boolean upper
) {
    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private boolean toInt;
        private boolean toNumber;
        private boolean toBool;
        private boolean trim;
        private boolean lower;
        private boolean upper;

        public Builder toInt(boolean v) { this.toInt = v; return this; }
        public Builder toNumber(boolean v) { this.toNumber = v; return this; }
        public Builder toBool(boolean v) { this.toBool = v; return this; }
        public Builder trim(boolean v) { this.trim = v; return this; }
        public Builder lower(boolean v) { this.lower = v; return this; }
        public Builder upper(boolean v) { this.upper = v; return this; }

        public CoercionConfig build() {
            return new CoercionConfig(toInt, toNumber, toBool, trim, lower, upper);
        }
    }
}

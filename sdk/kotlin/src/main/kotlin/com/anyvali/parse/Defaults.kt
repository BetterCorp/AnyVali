package com.anyvali.parse

object Defaults {
    /**
     * Apply a default value if the input is absent (not present).
     * null is NOT treated as absent.
     */
    fun applyDefault(present: Boolean, inputValue: Any?, defaultValue: Any?): Any? {
        return if (present) inputValue else defaultValue
    }
}

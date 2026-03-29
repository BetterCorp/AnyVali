package com.anyvali.parse

data class CoercionConfig(
    val type: String // e.g. "string->int", "string->number", "string->bool", "trim", "lower", "upper"
) {
    companion object {
        val STRING_TO_INT = CoercionConfig("string->int")
        val STRING_TO_NUMBER = CoercionConfig("string->number")
        val STRING_TO_BOOL = CoercionConfig("string->bool")
        val TRIM = CoercionConfig("trim")
        val LOWER = CoercionConfig("lower")
        val UPPER = CoercionConfig("upper")
    }
}

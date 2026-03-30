package com.anyvali

import kotlinx.serialization.json.JsonObject

abstract class Schema<out T> {
    abstract val kind: String
    open val hasCustomValidators: Boolean get() = false

    abstract fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue>

    abstract fun exportNode(): JsonObject

    fun parse(input: Any?): T {
        val result = safeParse(input)
        return result.getOrThrow()
    }

    fun parse(input: Any?, definitions: Map<String, Schema<*>>): T {
        val result = safeParse(input, definitions)
        return result.getOrThrow()
    }

    fun safeParse(input: Any?): ParseResult<T> {
        return safeParse(input, emptyMap())
    }

    fun safeParse(input: Any?, definitions: Map<String, Schema<*>>): ParseResult<T> {
        val ctx = ValidationContext(definitions = definitions)
        return safeParseWithContext(input, ctx)
    }

    @Suppress("UNCHECKED_CAST")
    open fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<T> {
        val issues = validateValue(input, ctx)
        return if (issues.isEmpty()) {
            ParseResult.Success(input as T)
        } else {
            ParseResult.Failure(issues)
        }
    }

    fun export(mode: ExportMode = ExportMode.PORTABLE): AnyValiDocument {
        if (mode == ExportMode.PORTABLE && hasCustomValidators) {
            throw ValidationError(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.CUSTOM_VALIDATION_NOT_PORTABLE,
                        message = "Schema uses custom validators that are not portable"
                    )
                )
            )
        }
        return AnyValiDocument(root = exportNode())
    }

    companion object {
        fun getJsonTypeName(value: Any?): String = when (value) {
            null -> "null"
            is Boolean -> "boolean"
            is String -> "string"
            is Long, is Int, is Short, is Byte -> "number"
            is Double, is Float -> "number"
            is Number -> "number"
            is List<*> -> "array"
            is Map<*, *> -> "object"
            else -> value::class.simpleName ?: "unknown"
        }

        fun isIntegerValue(value: Any?): Boolean = when (value) {
            is Long, is Int, is Short, is Byte -> true
            is Double -> value == value.toLong().toDouble() && !value.isNaN() && !value.isInfinite()
            is Float -> value == value.toLong().toFloat() && !value.isNaN() && !value.isInfinite()
            else -> false
        }

        fun toDouble(value: Any?): Double = when (value) {
            is Double -> value
            is Float -> value.toDouble()
            is Long -> value.toDouble()
            is Int -> value.toDouble()
            is Short -> value.toDouble()
            is Byte -> value.toDouble()
            is Number -> value.toDouble()
            else -> throw IllegalArgumentException("Cannot convert to double: $value")
        }

        fun toLong(value: Any?): Long = when (value) {
            is Long -> value
            is Int -> value.toLong()
            is Short -> value.toLong()
            is Byte -> value.toLong()
            is Double -> value.toLong()
            is Float -> value.toLong()
            is Number -> value.toLong()
            else -> throw IllegalArgumentException("Cannot convert to long: $value")
        }
    }
}

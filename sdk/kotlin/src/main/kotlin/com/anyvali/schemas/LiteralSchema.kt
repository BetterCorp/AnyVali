package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class LiteralSchema(
    val value: Any?
) : Schema<Any?>() {
    override val kind: String = "literal"

    override fun validateValue(input: Any?, ctx: ValidationContext): List<ValidationIssue> {
        if (!valuesEqual(input, value)) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_LITERAL,
                    path = ctx.path,
                    expected = value?.toString() ?: "null",
                    received = input?.toString() ?: "null"
                )
            )
        }
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("literal"))
        when (value) {
            null -> put("value", JsonNull)
            is String -> put("value", JsonPrimitive(value))
            is Boolean -> put("value", JsonPrimitive(value))
            is Long -> put("value", JsonPrimitive(value))
            is Int -> put("value", JsonPrimitive(value))
            is Double -> put("value", JsonPrimitive(value))
            is Number -> put("value", JsonPrimitive(value.toDouble()))
        }
    }

    private fun valuesEqual(a: Any?, b: Any?): Boolean {
        if (a == null && b == null) return true
        if (a == null || b == null) return false
        if (a is Boolean && b is Boolean) return a == b
        if (a is Boolean || b is Boolean) return false
        if (a is String && b is String) return a == b
        if (a is String || b is String) return false
        if (a is Number && b is Number) {
            return toDouble(a) == toDouble(b)
        }
        return a == b
    }
}

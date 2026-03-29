package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class EnumSchema(
    val values: List<Any?>
) : Schema() {
    override val kind: String = "enum"

    override fun validateValue(input: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val matched = values.any { v -> valuesEqual(input, v) }
        if (!matched) {
            val enumStr = values.joinToString(",") { it?.toString() ?: "null" }
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = "enum($enumStr)",
                    received = input?.toString() ?: "null"
                )
            )
        }
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("enum"))
        put("values", JsonArray(values.map { v ->
            when (v) {
                null -> JsonNull
                is String -> JsonPrimitive(v)
                is Boolean -> JsonPrimitive(v)
                is Long -> JsonPrimitive(v)
                is Int -> JsonPrimitive(v)
                is Double -> JsonPrimitive(v)
                is Number -> JsonPrimitive(v.toDouble())
                else -> JsonPrimitive(v.toString())
            }
        }))
    }

    private fun valuesEqual(a: Any?, b: Any?): Boolean {
        if (a == null && b == null) return true
        if (a == null || b == null) return false
        if (a is Boolean && b is Boolean) return a == b
        if (a is Boolean || b is Boolean) return false
        if (a is String && b is String) return a == b
        if (a is String || b is String) return false
        if (a is Number && b is Number) {
            return Schema.toDouble(a) == Schema.toDouble(b)
        }
        return a == b
    }
}

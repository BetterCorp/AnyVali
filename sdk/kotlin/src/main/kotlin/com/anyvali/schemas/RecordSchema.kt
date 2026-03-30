package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class RecordSchema(
    val values: Schema<*>
) : Schema<Map<String, Any?>>() {
    override val kind: String = "record"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Map<String, Any?>> {
        if (input !is Map<*, *>) {
            return ParseResult.Failure(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.INVALID_TYPE,
                        path = ctx.path,
                        expected = "record",
                        received = getJsonTypeName(input)
                    )
                )
            )
        }

        @Suppress("UNCHECKED_CAST")
        val inputMap = input as Map<String, Any?>
        val issues = mutableListOf<ValidationIssue>()
        val result = mutableMapOf<String, Any?>()

        for ((key, value) in inputMap) {
            val childCtx = ctx.child(key)
            val childResult = values.safeParseWithContext(value, childCtx)
            when (childResult) {
                is ParseResult.Success -> result[key] = childResult.value
                is ParseResult.Failure -> issues.addAll(childResult.issues)
            }
        }

        return if (issues.isEmpty()) ParseResult.Success(result) else ParseResult.Failure(issues)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val result = safeParseWithContext(value, ctx)
        return if (result is ParseResult.Failure) result.issues else emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("record"))
        put("values", values.exportNode())
    }
}

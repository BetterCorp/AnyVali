package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class IntersectionSchema(
    val allOf: List<Schema<*>>
) : Schema<Any?>() {
    override val kind: String = "intersection"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Any?> {
        val allIssues = mutableListOf<ValidationIssue>()
        var mergedResult: Any? = input

        for (schema in allOf) {
            val result = schema.safeParseWithContext(input, ctx)
            when (result) {
                is ParseResult.Success -> {
                    // For objects, merge results
                    if (mergedResult is Map<*, *> && result.value is Map<*, *>) {
                        @Suppress("UNCHECKED_CAST")
                        val current = (mergedResult as Map<String, Any?>).toMutableMap()
                        @Suppress("UNCHECKED_CAST")
                        current.putAll(result.value as Map<String, Any?>)
                        mergedResult = current
                    } else {
                        mergedResult = result.value
                    }
                }
                is ParseResult.Failure -> allIssues.addAll(result.issues)
            }
        }

        return if (allIssues.isEmpty()) {
            ParseResult.Success(mergedResult)
        } else {
            ParseResult.Failure(allIssues)
        }
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val result = safeParseWithContext(value, ctx)
        return if (result is ParseResult.Failure) result.issues else emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("intersection"))
        put("allOf", JsonArray(allOf.map { it.exportNode() }))
    }
}

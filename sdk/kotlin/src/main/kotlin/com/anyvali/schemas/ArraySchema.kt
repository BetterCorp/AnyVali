package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class ArraySchema(
    val items: Schema,
    val minItems: Int? = null,
    val maxItems: Int? = null
) : Schema() {
    override val kind: String = "array"

    fun minItems(n: Int) = copy(minItems = n)
    fun maxItems(n: Int) = copy(maxItems = n)

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        if (input !is List<*>) {
            return ParseResult.Failure(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.INVALID_TYPE,
                        path = ctx.path,
                        expected = "array",
                        received = getJsonTypeName(input)
                    )
                )
            )
        }

        val issues = mutableListOf<ValidationIssue>()

        minItems?.let {
            if (input.size < it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = it.toString(),
                        received = input.size.toString()
                    )
                )
            }
        }
        maxItems?.let {
            if (input.size > it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = it.toString(),
                        received = input.size.toString()
                    )
                )
            }
        }

        if (issues.isNotEmpty()) return ParseResult.Failure(issues)

        val result = mutableListOf<Any?>()
        for ((index, element) in input.withIndex()) {
            val childCtx = ctx.child(index)
            val childResult = items.safeParseWithContext(element, childCtx)
            when (childResult) {
                is ParseResult.Success -> result.add(childResult.value)
                is ParseResult.Failure -> issues.addAll(childResult.issues)
            }
        }

        return if (issues.isEmpty()) ParseResult.Success(result) else ParseResult.Failure(issues)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        // Delegate to safeParseWithContext
        val result = safeParseWithContext(value, ctx)
        return if (result is ParseResult.Failure) result.issues else emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("array"))
        put("items", items.exportNode())
        minItems?.let { put("minItems", JsonPrimitive(it)) }
        maxItems?.let { put("maxItems", JsonPrimitive(it)) }
    }
}

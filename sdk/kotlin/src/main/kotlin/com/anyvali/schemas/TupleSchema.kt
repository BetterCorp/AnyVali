package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class TupleSchema(
    val elements: List<Schema>
) : Schema() {
    override val kind: String = "tuple"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        if (input !is List<*>) {
            return ParseResult.Failure(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.INVALID_TYPE,
                        path = ctx.path,
                        expected = "tuple",
                        received = getJsonTypeName(input)
                    )
                )
            )
        }

        val issues = mutableListOf<ValidationIssue>()

        if (input.size < elements.size) {
            issues.add(
                ValidationIssue(
                    code = IssueCodes.TOO_SMALL,
                    path = ctx.path,
                    expected = elements.size.toString(),
                    received = input.size.toString()
                )
            )
            return ParseResult.Failure(issues)
        }
        if (input.size > elements.size) {
            issues.add(
                ValidationIssue(
                    code = IssueCodes.TOO_LARGE,
                    path = ctx.path,
                    expected = elements.size.toString(),
                    received = input.size.toString()
                )
            )
            return ParseResult.Failure(issues)
        }

        val result = mutableListOf<Any?>()
        for ((index, schema) in elements.withIndex()) {
            val childCtx = ctx.child(index)
            val childResult = schema.safeParseWithContext(input[index], childCtx)
            when (childResult) {
                is ParseResult.Success -> result.add(childResult.value)
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
        put("kind", JsonPrimitive("tuple"))
        put("elements", JsonArray(elements.map { it.exportNode() }))
    }
}

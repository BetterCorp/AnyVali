package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class UnionSchema(
    val variants: List<Schema>
) : Schema() {
    override val kind: String = "union"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        for (variant in variants) {
            val result = variant.safeParseWithContext(input, ctx)
            if (result.isSuccess) return result
        }
        val expected = variants.joinToString(" | ") { it.kind }
        return ParseResult.Failure(
            listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_UNION,
                    path = ctx.path,
                    expected = expected,
                    received = getJsonTypeName(input)
                )
            )
        )
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val result = safeParseWithContext(value, ctx)
        return if (result is ParseResult.Failure) result.issues else emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("union"))
        put("variants", JsonArray(variants.map { it.exportNode() }))
    }
}

package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class NullableSchema(
    val inner: Schema<*>
) : Schema<Any?>() {
    override val kind: String = "nullable"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Any?> {
        if (input == null) return ParseResult.Success(null)
        return inner.safeParseWithContext(input, ctx)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        if (value == null) return emptyList()
        return inner.validateValue(value, ctx)
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("nullable"))
        put("schema", inner.exportNode())
    }
}

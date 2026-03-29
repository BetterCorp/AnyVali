package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class OptionalSchema(
    val inner: Schema
) : Schema() {
    override val kind: String = "optional"

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        // OptionalSchema should not be called directly with absent values;
        // absent handling is done in ObjectSchema.
        // If called, it delegates to inner.
        return inner.safeParseWithContext(input, ctx)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return inner.validateValue(value, ctx)
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("optional"))
        put("schema", inner.exportNode())
    }
}

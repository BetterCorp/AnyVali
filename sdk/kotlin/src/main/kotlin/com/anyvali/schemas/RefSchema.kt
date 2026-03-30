package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class RefSchema(
    val ref: String
) : Schema<Any?>() {
    override val kind: String = "ref"

    private fun resolveRef(ctx: ValidationContext): Schema<*> {
        // ref format: "#/definitions/Name"
        val defName = ref.removePrefix("#/definitions/")
        return ctx.definitions[defName]
            ?: throw IllegalStateException("Unresolved ref: $ref")
    }

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Any?> {
        val resolved = resolveRef(ctx)
        return resolved.safeParseWithContext(input, ctx)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val resolved = resolveRef(ctx)
        return resolved.validateValue(value, ctx)
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("ref"))
        put("ref", JsonPrimitive(ref))
    }
}

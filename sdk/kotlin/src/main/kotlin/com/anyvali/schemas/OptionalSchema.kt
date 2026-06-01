package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class OptionalSchema(
    val inner: Schema<*>
) : Schema<Any?>() {
    override val kind: String = "optional"

    fun describe(description: String, opts: DescribeOptions? = null): OptionalSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): OptionalSchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Any?> {
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
        addMetadataToNode(this)
    }
}

package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class NullableSchema(
    val inner: Schema<*>
) : Schema<Any?>() {
    override val kind: String = "nullable"

    fun describe(description: String, opts: DescribeOptions? = null): NullableSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): NullableSchema {
        applyMetadata(meta, replace)
        return this
    }

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
        addMetadataToNode(this)
    }
}

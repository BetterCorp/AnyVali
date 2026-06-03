package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class OptionalSchema(
    val inner: Schema<*>,
    val defaultValue: Any? = UNSET
) : Schema<Any?>() {
    override val kind: String = "optional"

    fun default(v: Any?) = copy(defaultValue = v)

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
        if (defaultValue !== UNSET) {
            when (defaultValue) {
                null -> put("default", JsonNull)
                is String -> put("default", JsonPrimitive(defaultValue))
                is Boolean -> put("default", JsonPrimitive(defaultValue))
                is Number -> put("default", JsonPrimitive(defaultValue))
                else -> put("default", JsonPrimitive(defaultValue.toString()))
            }
        }
        addMetadataToNode(this)
    }

    companion object {
        val UNSET = Any()
    }
}

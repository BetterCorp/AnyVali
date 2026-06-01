package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class NullSchema : Schema<Nothing?>() {
    override val kind: String = "null"

    fun describe(description: String, opts: DescribeOptions? = null): NullSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): NullSchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        if (value != null) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = "null",
                    received = getJsonTypeName(value)
                )
            )
        }
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("null"))
        addMetadataToNode(this)
    }
}

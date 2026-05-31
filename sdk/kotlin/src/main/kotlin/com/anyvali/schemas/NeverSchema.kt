package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class NeverSchema : Schema<Nothing>() {
    override val kind: String = "never"

    fun describe(description: String, opts: DescribeOptions? = null): NeverSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): NeverSchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return listOf(
            ValidationIssue(
                code = IssueCodes.INVALID_TYPE,
                path = ctx.path,
                expected = "never",
                received = getJsonTypeName(value)
            )
        )
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("never"))
        addMetadataToNode(this)
    }
}

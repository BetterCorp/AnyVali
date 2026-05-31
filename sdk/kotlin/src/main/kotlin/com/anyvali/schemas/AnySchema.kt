package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class AnySchema : Schema<Any?>() {
    override val kind: String = "any"

    fun describe(description: String, opts: DescribeOptions? = null): AnySchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): AnySchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("any"))
        addMetadataToNode(this)
    }
}

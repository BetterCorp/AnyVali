package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class UnknownSchema : Schema<Any?>() {
    override val kind: String = "unknown"

    fun describe(description: String, opts: DescribeOptions? = null): UnknownSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): UnknownSchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("unknown"))
        addMetadataToNode(this)
    }
}

package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class UnknownSchema : Schema() {
    override val kind: String = "unknown"

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("unknown"))
    }
}

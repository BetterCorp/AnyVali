package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class AnySchema : Schema() {
    override val kind: String = "any"

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("any"))
    }
}

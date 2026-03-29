package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class NullSchema : Schema() {
    override val kind: String = "null"

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
    }
}

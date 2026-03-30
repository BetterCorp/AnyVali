package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

class NeverSchema : Schema<Nothing>() {
    override val kind: String = "never"

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
    }
}

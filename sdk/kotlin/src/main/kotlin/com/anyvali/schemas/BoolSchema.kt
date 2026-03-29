package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class BoolSchema(
    val defaultValue: Any? = UNSET,
    val coerce: String? = null
) : Schema() {
    override val kind: String = "bool"

    fun default(v: Boolean) = copy(defaultValue = v)
    fun coerce(c: String) = copy(coerce = c)

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        var value = input

        if (coerce == "string->bool" && value is String) {
            val lower = value.trim().lowercase()
            value = when (lower) {
                "true", "1" -> true
                "false", "0" -> false
                else -> {
                    return ParseResult.Failure(
                        listOf(
                            ValidationIssue(
                                code = IssueCodes.COERCION_FAILED,
                                path = ctx.path,
                                expected = "bool",
                                received = input as String
                            )
                        )
                    )
                }
            }
        }

        val issues = validateValue(value, ctx)
        return if (issues.isEmpty()) {
            ParseResult.Success(value)
        } else {
            ParseResult.Failure(issues)
        }
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        if (value !is Boolean) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = "bool",
                    received = getJsonTypeName(value)
                )
            )
        }
        return emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("bool"))
        if (defaultValue !== UNSET) put("default", JsonPrimitive(defaultValue as Boolean))
        coerce?.let { put("coerce", JsonPrimitive(it)) }
    }

    companion object {
        val UNSET = Any()
    }
}

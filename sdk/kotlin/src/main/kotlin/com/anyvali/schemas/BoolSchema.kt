package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class BoolSchema(
    val defaultValue: Any? = UNSET,
    val coerce: String? = null
) : Schema<Boolean>() {
    override val kind: String = "bool"

    fun default(v: Boolean) = copy(defaultValue = v)
    fun coerce(c: String) = copy(coerce = c)
    /**
     * No-arg ergonomic: enable string coercion with the target inferred from the
     * schema kind. Equivalent to the portable generic "string" source (spec 5.1)
     * and to JS `.coerce()` / `.coerce({ from: "string" })`.
     */
    fun coerce() = copy(coerce = "string")

    fun describe(description: String, opts: DescribeOptions? = null): BoolSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): BoolSchema {
        applyMetadata(meta, replace)
        return this
    }

    @Suppress("UNCHECKED_CAST")
    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Boolean> {
        var value = input

        // Both the typed token ("string->bool") and the generic portable "string"
        // source (set via the no-arg coerce()) enable string->bool coercion.
        if ((coerce == "string->bool" || coerce == "string") && value is String) {
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
            ParseResult.Success(value as Boolean)
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
        addMetadataToNode(this)
    }

    companion object {
        val UNSET = Any()
    }
}

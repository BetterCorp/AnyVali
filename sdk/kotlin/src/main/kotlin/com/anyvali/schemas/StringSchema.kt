package com.anyvali.schemas

import com.anyvali.*
import com.anyvali.format.FormatValidators
import com.anyvali.parse.Coercion
import kotlinx.serialization.json.*

data class StringSchema(
    val minLength: Int? = null,
    val maxLength: Int? = null,
    val pattern: String? = null,
    val startsWith: String? = null,
    val endsWith: String? = null,
    val includes: String? = null,
    val format: String? = null,
    val defaultValue: Any? = UNSET,
    val coerce: List<String> = emptyList()
) : Schema() {
    override val kind: String = "string"

    fun minLength(n: Int) = copy(minLength = n)
    fun maxLength(n: Int) = copy(maxLength = n)
    fun pattern(p: String) = copy(pattern = p)
    fun startsWith(s: String) = copy(startsWith = s)
    fun endsWith(s: String) = copy(endsWith = s)
    fun includes(s: String) = copy(includes = s)
    fun format(f: String) = copy(format = f)
    fun default(v: String) = copy(defaultValue = v)
    fun coerce(vararg c: String) = copy(coerce = c.toList())

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        var value = input

        // Apply coercions
        if (coerce.isNotEmpty() && value is String) {
            for (c in coerce) {
                value = when (c) {
                    "trim" -> (value as String).trim()
                    "lower" -> (value as String).lowercase()
                    "upper" -> (value as String).uppercase()
                    else -> value
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
        if (value !is String) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = "string",
                    received = getJsonTypeName(value)
                )
            )
        }
        val issues = mutableListOf<ValidationIssue>()

        minLength?.let {
            if (value.length < it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = it.toString(),
                        received = value.length.toString()
                    )
                )
            }
        }
        maxLength?.let {
            if (value.length > it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = it.toString(),
                        received = value.length.toString()
                    )
                )
            }
        }
        pattern?.let {
            if (!Regex(it).containsMatchIn(value)) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_STRING,
                        path = ctx.path,
                        expected = it,
                        received = value
                    )
                )
            }
        }
        startsWith?.let {
            if (!value.startsWith(it)) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_STRING,
                        path = ctx.path,
                        expected = it,
                        received = value
                    )
                )
            }
        }
        endsWith?.let {
            if (!value.endsWith(it)) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_STRING,
                        path = ctx.path,
                        expected = it,
                        received = value
                    )
                )
            }
        }
        includes?.let {
            if (!value.contains(it)) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_STRING,
                        path = ctx.path,
                        expected = it,
                        received = value
                    )
                )
            }
        }
        format?.let {
            if (!FormatValidators.validate(it, value)) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_STRING,
                        path = ctx.path,
                        expected = it,
                        received = value
                    )
                )
            }
        }

        return issues
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("string"))
        minLength?.let { put("minLength", JsonPrimitive(it)) }
        maxLength?.let { put("maxLength", JsonPrimitive(it)) }
        pattern?.let { put("pattern", JsonPrimitive(it)) }
        startsWith?.let { put("startsWith", JsonPrimitive(it)) }
        endsWith?.let { put("endsWith", JsonPrimitive(it)) }
        includes?.let { put("includes", JsonPrimitive(it)) }
        format?.let { put("format", JsonPrimitive(it)) }
        if (defaultValue !== UNSET) put("default", JsonPrimitive(defaultValue as? String))
        if (coerce.isNotEmpty()) {
            if (coerce.size == 1) put("coerce", JsonPrimitive(coerce[0]))
            else put("coerce", JsonArray(coerce.map { JsonPrimitive(it) }))
        }
    }

    companion object {
        val UNSET = Any()
    }
}

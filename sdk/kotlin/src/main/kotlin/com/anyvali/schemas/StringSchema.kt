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
) : Schema<String>() {
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

    fun describe(description: String, opts: DescribeOptions? = null): StringSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): StringSchema {
        applyMetadata(meta, replace)
        return this
    }

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<String> {
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
            ParseResult.Success(value as String)
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
        val length = value.codePointCount(0, value.length)

        minLength?.let {
            if (length < it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = it.toString(),
                        received = length.toString()
                    )
                )
            }
        }
        maxLength?.let {
            if (length > it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = it.toString(),
                        received = length.toString()
                    )
                )
            }
        }
        pattern?.let {
            try {
                if (!Regex(toEcmaAnchors(it)).containsMatchIn(value)) {
                    issues.add(
                        ValidationIssue(
                            code = IssueCodes.INVALID_STRING,
                            path = ctx.path,
                            expected = it,
                            received = value
                        )
                    )
                }
            } catch (_: Exception) {
                // Invalid regex pattern - treat as validation failure
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
        addMetadataToNode(this)
    }

    companion object {
        val UNSET = Any()

        /**
         * Rewrite ECMA-262 anchors to absolute anchors. In ECMA without the
         * multiline flag, "^"/"$" match only the start/end of the whole string.
         * Kotlin/Java "$" also matches just before a final line terminator, so
         * an anchored whitelist like ^[a-z]+$ would accept "abc\n" -- a
         * newline-injection bypass that diverges from the JS reference.
         * Translate unescaped, top-level "^" -> "\A" and "$" -> "\z". Anchors
         * inside character classes and escaped "\^"/"\$" are left untouched.
         */
        fun toEcmaAnchors(p: String): String {
            val sb = StringBuilder(p.length + 4)
            var escaped = false
            var inClass = false
            for (ch in p) {
                when {
                    escaped -> { sb.append(ch); escaped = false }
                    ch == '\\' -> { sb.append(ch); escaped = true }
                    ch == '[' -> { inClass = true; sb.append(ch) }
                    ch == ']' && inClass -> { inClass = false; sb.append(ch) }
                    ch == '^' && !inClass -> sb.append("\\A")
                    ch == '$' && !inClass -> sb.append("\\z")
                    else -> sb.append(ch)
                }
            }
            return sb.toString()
        }
    }
}

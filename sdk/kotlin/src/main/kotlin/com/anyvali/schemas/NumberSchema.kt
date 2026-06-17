package com.anyvali.schemas

import com.anyvali.*
import com.anyvali.parse.Coercion
import kotlinx.serialization.json.*

data class NumberSchema(
    val schemaKind: String = "number",
    val min: Double? = null,
    val max: Double? = null,
    val exclusiveMin: Double? = null,
    val exclusiveMax: Double? = null,
    val multipleOf: Double? = null,
    val defaultValue: Any? = UNSET,
    val coerce: String? = null
) : Schema<Double>() {
    override val kind: String = schemaKind

    fun min(n: Number) = copy(min = n.toDouble())
    fun max(n: Number) = copy(max = n.toDouble())
    fun exclusiveMin(n: Number) = copy(exclusiveMin = n.toDouble())
    fun exclusiveMax(n: Number) = copy(exclusiveMax = n.toDouble())
    fun multipleOf(n: Number) = copy(multipleOf = n.toDouble())
    fun default(v: Number) = copy(defaultValue = v)
    fun coerce(c: String) = copy(coerce = c)
    /**
     * No-arg ergonomic: enable string coercion with the target inferred from the
     * schema kind. Equivalent to the portable generic "string" source (spec 5.1)
     * and to JS `.coerce()` / `.coerce({ from: "string" })`.
     */
    fun coerce() = copy(coerce = "string")

    fun describe(description: String, opts: DescribeOptions? = null): NumberSchema {
        applyDescribe(description, opts)
        return this
    }

    fun metadata(meta: Map<String, Any?>, replace: Boolean = false): NumberSchema {
        applyMetadata(meta, replace)
        return this
    }

    @Suppress("UNCHECKED_CAST")
    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Double> {
        var value = input

        // Apply coercion. Both the typed token ("string->number") and the generic
        // portable "string" source (set via the no-arg coerce()) enable
        // string->number coercion, with the target inferred from this schema's kind.
        if ((coerce == "string->number" || coerce == "string") && value is String) {
            val parsed = Coercion.coerceStringToNumber(value)
            if (parsed == null) {
                return ParseResult.Failure(
                    listOf(
                        ValidationIssue(
                            code = IssueCodes.COERCION_FAILED,
                            path = ctx.path,
                            expected = kind,
                            received = value
                        )
                    )
                )
            }
            value = parsed
        }

        val issues = validateValue(value, ctx)
        return if (issues.isEmpty()) {
            ParseResult.Success((value as Number).toDouble())
        } else {
            ParseResult.Failure(issues)
        }
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        if (value == null || value is Boolean || value is String || value is List<*> || value is Map<*, *>) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = kind,
                    received = getJsonTypeName(value)
                )
            )
        }
        if (value !is Number) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = kind,
                    received = getJsonTypeName(value)
                )
            )
        }

        val dbl = toDouble(value)
        val issues = mutableListOf<ValidationIssue>()

        if (dbl.isNaN() || dbl.isInfinite()) {
            issues.add(
                ValidationIssue(
                    code = IssueCodes.INVALID_NUMBER,
                    path = ctx.path,
                    expected = kind,
                    received = dbl.toString()
                )
            )
            return issues
        }

        // Float32 range check
        if (schemaKind == "float32") {
            if (dbl != 0.0 && (dbl > Float.MAX_VALUE.toDouble() || dbl < -Float.MAX_VALUE.toDouble())) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = "float32",
                        received = dbl.toString()
                    )
                )
            }
        }

        min?.let {
            if (dbl < it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = formatNumber(it),
                        received = formatNumber(dbl)
                    )
                )
            }
        }
        max?.let {
            if (dbl > it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = formatNumber(it),
                        received = formatNumber(dbl)
                    )
                )
            }
        }
        exclusiveMin?.let {
            if (dbl <= it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = formatNumber(it),
                        received = formatNumber(dbl)
                    )
                )
            }
        }
        exclusiveMax?.let {
            if (dbl >= it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = formatNumber(it),
                        received = formatNumber(dbl)
                    )
                )
            }
        }
        multipleOf?.let {
            val remainder = dbl % it
            if (Math.abs(remainder) > 1e-10 && Math.abs(remainder - it) > 1e-10) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_NUMBER,
                        path = ctx.path,
                        expected = formatNumber(it),
                        received = formatNumber(dbl)
                    )
                )
            }
        }

        return issues
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive(schemaKind))
        min?.let { put("min", JsonPrimitive(it)) }
        max?.let { put("max", JsonPrimitive(it)) }
        exclusiveMin?.let { put("exclusiveMin", JsonPrimitive(it)) }
        exclusiveMax?.let { put("exclusiveMax", JsonPrimitive(it)) }
        multipleOf?.let { put("multipleOf", JsonPrimitive(it)) }
        if (defaultValue !== UNSET) {
            val dv = defaultValue
            when (dv) {
                is Double -> put("default", JsonPrimitive(dv))
                is Int -> put("default", JsonPrimitive(dv))
                is Long -> put("default", JsonPrimitive(dv))
                is Float -> put("default", JsonPrimitive(dv))
                else -> {}
            }
        }
        coerce?.let { put("coerce", JsonPrimitive(it)) }
        addMetadataToNode(this)
    }

    companion object {
        val UNSET = Any()

        fun formatNumber(d: Double): String {
            return if (d == d.toLong().toDouble() && !d.isNaN() && !d.isInfinite()) {
                d.toLong().toString()
            } else {
                d.toString()
            }
        }
    }
}

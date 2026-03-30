package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class IntSchema(
    val schemaKind: String = "int",
    val min: Long? = null,
    val max: Long? = null,
    val exclusiveMin: Long? = null,
    val exclusiveMax: Long? = null,
    val multipleOf: Long? = null,
    val defaultValue: Any? = UNSET,
    val coerce: String? = null
) : Schema<Long>() {
    override val kind: String = schemaKind

    fun min(n: Long) = copy(min = n)
    fun max(n: Long) = copy(max = n)
    fun exclusiveMin(n: Long) = copy(exclusiveMin = n)
    fun exclusiveMax(n: Long) = copy(exclusiveMax = n)
    fun multipleOf(n: Long) = copy(multipleOf = n)
    fun default(v: Long) = copy(defaultValue = v)
    fun coerce(c: String) = copy(coerce = c)

    private val rangeMin: Long get() = when (schemaKind) {
        "int8" -> -128L
        "int16" -> -32768L
        "int32" -> -2147483648L
        "int64", "int" -> Long.MIN_VALUE
        "uint8" -> 0L
        "uint16" -> 0L
        "uint32" -> 0L
        "uint64" -> 0L
        else -> Long.MIN_VALUE
    }

    private val rangeMax: Long get() = when (schemaKind) {
        "int8" -> 127L
        "int16" -> 32767L
        "int32" -> 2147483647L
        "int64", "int" -> Long.MAX_VALUE
        "uint8" -> 255L
        "uint16" -> 65535L
        "uint32" -> 4294967295L
        "uint64" -> Long.MAX_VALUE // approximate; full uint64 max not representable in Long
        else -> Long.MAX_VALUE
    }

    @Suppress("UNCHECKED_CAST")
    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<Long> {
        var value = input

        // Apply coercion
        if (coerce == "string->int" && value is String) {
            val trimmed = value.trim()
            val parsed = trimmed.toLongOrNull()
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
            // Normalize to Long
            val longVal: Long = when (value) {
                is Long -> value
                is Int -> value.toLong()
                is Short -> value.toLong()
                is Byte -> value.toLong()
                is Double -> value.toLong()
                is Float -> value.toLong()
                else -> value as Long
            }
            ParseResult.Success(longVal)
        } else {
            ParseResult.Failure(issues)
        }
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        // Must be a number
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

        // Must be integer
        if (!isIntegerValue(value)) {
            return listOf(
                ValidationIssue(
                    code = IssueCodes.INVALID_TYPE,
                    path = ctx.path,
                    expected = kind,
                    received = "number"
                )
            )
        }

        val longVal = toLong(value)
        val issues = mutableListOf<ValidationIssue>()

        // Range checks for typed int schemas
        if (schemaKind != "int" && schemaKind != "int64") {
            if (longVal < rangeMin) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = schemaKind,
                        received = longVal.toString()
                    )
                )
            }
            if (longVal > rangeMax) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = schemaKind,
                        received = longVal.toString()
                    )
                )
            }
        }

        // User-specified constraints
        min?.let {
            if (longVal < it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = it.toString(),
                        received = longVal.toString()
                    )
                )
            }
        }
        max?.let {
            if (longVal > it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = it.toString(),
                        received = longVal.toString()
                    )
                )
            }
        }
        exclusiveMin?.let {
            if (longVal <= it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_SMALL,
                        path = ctx.path,
                        expected = it.toString(),
                        received = longVal.toString()
                    )
                )
            }
        }
        exclusiveMax?.let {
            if (longVal >= it) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.TOO_LARGE,
                        path = ctx.path,
                        expected = it.toString(),
                        received = longVal.toString()
                    )
                )
            }
        }
        multipleOf?.let {
            if (longVal % it != 0L) {
                issues.add(
                    ValidationIssue(
                        code = IssueCodes.INVALID_NUMBER,
                        path = ctx.path,
                        expected = it.toString(),
                        received = longVal.toString()
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
                is Long -> put("default", JsonPrimitive(dv))
                is Int -> put("default", JsonPrimitive(dv))
                else -> {}
            }
        }
        coerce?.let { put("coerce", JsonPrimitive(it)) }
    }

    companion object {
        val UNSET = Any()
    }
}

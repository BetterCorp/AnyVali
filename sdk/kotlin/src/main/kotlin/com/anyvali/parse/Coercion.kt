package com.anyvali.parse

object Coercion {
    // ASCII-only integer grammar: optional leading minus, then one or more
    // ASCII digits. Excludes a leading '+', underscores, unicode digits,
    // hex/octal/binary literals, Infinity/NaN and any fractional/exponent part.
    private val INT_RE = Regex("^-?[0-9]+$")

    // Decimal floating-point grammar (spec 5.1): optional sign, digits with an
    // optional fraction (or a bare fraction), and an optional decimal exponent.
    // Excludes hex/octal/binary literals, Infinity, NaN and underscores that a
    // native parser (Kotlin String.toDouble) would otherwise accept.
    private val DECIMAL_FLOAT_RE = Regex("^[+-]?(?:[0-9]+\\.?[0-9]*|\\.[0-9]+)(?:[eE][+-]?[0-9]+)?$")

    fun coerceStringToInt(value: String): Long? {
        val trimmed = value.trim()
        if (!INT_RE.matches(trimmed)) return null
        return trimmed.toLongOrNull()
    }

    fun coerceStringToNumber(value: String): Double? {
        val trimmed = value.trim()
        if (!DECIMAL_FLOAT_RE.matches(trimmed)) return null
        val parsed = trimmed.toDoubleOrNull() ?: return null
        if (parsed.isNaN() || parsed.isInfinite()) return null
        return parsed
    }

    fun coerceStringToBool(value: String): Boolean? {
        return when (value.trim().lowercase()) {
            "true", "1" -> true
            "false", "0" -> false
            else -> null
        }
    }

    fun trimString(value: String): String = value.trim()

    fun lowerString(value: String): String = value.lowercase()

    fun upperString(value: String): String = value.uppercase()
}

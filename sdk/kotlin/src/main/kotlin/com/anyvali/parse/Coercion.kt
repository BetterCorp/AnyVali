package com.anyvali.parse

object Coercion {
    fun coerceStringToInt(value: String): Long? {
        return value.trim().toLongOrNull()
    }

    fun coerceStringToNumber(value: String): Double? {
        return value.trim().toDoubleOrNull()
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

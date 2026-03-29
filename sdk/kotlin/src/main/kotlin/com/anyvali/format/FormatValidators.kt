package com.anyvali.format

import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException

object FormatValidators {
    private val EMAIL_REGEX = Regex("^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$")
    private val UUID_REGEX = Regex("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
    private val DATE_REGEX = Regex("^\\d{4}-\\d{2}-\\d{2}$")
    private val DATETIME_REGEX = Regex("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(Z|[+-]\\d{2}:\\d{2})$")

    fun validate(format: String, value: String): Boolean {
        return when (format) {
            "email" -> isValidEmail(value)
            "url" -> isValidUrl(value)
            "uuid" -> isValidUuid(value)
            "ipv4" -> isValidIpv4(value)
            "ipv6" -> isValidIpv6(value)
            "date" -> isValidDate(value)
            "date-time" -> isValidDateTime(value)
            else -> true // unknown formats pass
        }
    }

    private fun isValidEmail(value: String): Boolean {
        return EMAIL_REGEX.matches(value)
    }

    private fun isValidUrl(value: String): Boolean {
        return try {
            val uri = java.net.URI(value)
            val scheme = uri.scheme?.lowercase()
            scheme == "http" || scheme == "https"
        } catch (_: Exception) {
            false
        }
    }

    private fun isValidUuid(value: String): Boolean {
        return UUID_REGEX.matches(value)
    }

    private fun isValidIpv4(value: String): Boolean {
        val parts = value.split(".")
        if (parts.size != 4) return false
        return parts.all { part ->
            if (part.isEmpty()) return false
            if (part.length > 1 && part.startsWith("0")) return false
            val num = part.toIntOrNull() ?: return false
            num in 0..255
        }
    }

    private fun isValidIpv6(value: String): Boolean {
        // Handle :: notation
        if (value.isEmpty()) return false

        // Count ::
        val doubleColonCount = Regex("::").findAll(value).count()
        if (doubleColonCount > 1) return false

        if (doubleColonCount == 1) {
            val parts = value.split("::")
            val left = if (parts[0].isEmpty()) emptyList() else parts[0].split(":")
            val right = if (parts.size < 2 || parts[1].isEmpty()) emptyList() else parts[1].split(":")
            val totalGroups = left.size + right.size
            if (totalGroups > 7) return false
            return (left + right).all { isValidHexGroup(it) }
        } else {
            val groups = value.split(":")
            if (groups.size != 8) return false
            return groups.all { isValidHexGroup(it) }
        }
    }

    private fun isValidHexGroup(group: String): Boolean {
        if (group.isEmpty() || group.length > 4) return false
        return group.all { it in '0'..'9' || it in 'a'..'f' || it in 'A'..'F' }
    }

    private fun isValidDate(value: String): Boolean {
        if (!DATE_REGEX.matches(value)) return false
        return try {
            val parts = value.split("-")
            val year = parts[0].toInt()
            val month = parts[1].toInt()
            val day = parts[2].toInt()
            val date = LocalDate.of(year, month, day)
            // Verify round-trip to catch invalid dates like Feb 30
            date.year == year && date.monthValue == month && date.dayOfMonth == day
        } catch (_: Exception) {
            false
        }
    }

    private fun isValidDateTime(value: String): Boolean {
        return DATETIME_REGEX.matches(value) && isValidDate(value.substringBefore("T"))
    }
}

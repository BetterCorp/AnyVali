package com.anyvali.format;

import java.net.URI;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.Map;
import java.util.UUID;
import java.util.function.Predicate;
import java.util.regex.Pattern;

/**
 * Format validators for portable string formats.
 */
public final class FormatValidators {
    private FormatValidators() {}

    // Basic email regex covering common cases
    private static final Pattern EMAIL_RE = Pattern.compile(
            "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9]" +
            "(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?" +
            "(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$"
    );

    // URL regex - accepts http/https with basic structure
    private static final Pattern URL_RE = Pattern.compile(
            "^https?://" +
            "[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?" +
            "(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*" +
            "(?::\\d{1,5})?" +
            "(?:/[^\\s]*)?$"
    );

    // ISO 8601 date: YYYY-MM-DD
    private static final Pattern DATE_RE = Pattern.compile("^\\d{4}-\\d{2}-\\d{2}$");

    // ISO 8601 date-time with required timezone
    private static final Pattern DATETIME_RE = Pattern.compile(
            "^\\d{4}-\\d{2}-\\d{2}[T ]\\d{2}:\\d{2}:\\d{2}" +
            "(?:\\.\\d+)?" +
            "(?:Z|[+-]\\d{2}:\\d{2})$"
    );

    // IPv4 pattern
    private static final Pattern IPV4_RE = Pattern.compile(
            "^(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})$"
    );

    private static final Map<String, Predicate<String>> VALIDATORS = Map.of(
            "email", FormatValidators::validateEmail,
            "url", FormatValidators::validateUrl,
            "uuid", FormatValidators::validateUuid,
            "ipv4", FormatValidators::validateIpv4,
            "ipv6", FormatValidators::validateIpv6,
            "date", FormatValidators::validateDate,
            "date-time", FormatValidators::validateDateTime
    );

    /**
     * Validate a string value against a named format.
     * Returns true if valid, false otherwise.
     * Unknown formats pass (lenient).
     */
    public static boolean validate(String format, String value) {
        var validator = VALIDATORS.get(format);
        if (validator == null) {
            return true; // Unknown formats pass
        }
        return validator.test(value);
    }

    private static boolean validateEmail(String value) {
        return EMAIL_RE.matcher(value).matches();
    }

    private static boolean validateUrl(String value) {
        return URL_RE.matcher(value).matches();
    }

    private static boolean validateUuid(String value) {
        try {
            UUID.fromString(value);
            return true;
        } catch (IllegalArgumentException e) {
            return false;
        }
    }

    private static boolean validateIpv4(String value) {
        var matcher = IPV4_RE.matcher(value);
        if (!matcher.matches()) return false;
        for (int i = 1; i <= 4; i++) {
            String part = matcher.group(i);
            // Reject leading zeros (e.g. "01", "001")
            if (part.length() > 1 && part.charAt(0) == '0') return false;
            int octet = Integer.parseInt(part);
            if (octet < 0 || octet > 255) return false;
        }
        return true;
    }

    private static boolean validateIpv6(String value) {
        // Use Java's built-in validation
        // Wrap in brackets for URI parsing trick
        if (value.isEmpty()) return false;
        try {
            // Quick structural check: must contain at least one colon
            if (!value.contains(":")) return false;
            // Use InetAddress-like validation
            return isValidIpv6(value);
        } catch (Exception e) {
            return false;
        }
    }

    private static boolean isValidIpv6(String address) {
        // Handle :: expansion
        if (address.startsWith(":") && !address.startsWith("::")) return false;
        if (address.endsWith(":") && !address.endsWith("::")) return false;

        String[] parts;
        int expectedGroups = 8;

        if (address.contains("::")) {
            // Only one :: allowed
            int firstDouble = address.indexOf("::");
            if (address.indexOf("::", firstDouble + 2) != -1) return false;

            String[] halves = address.split("::", -1);
            String[] left = halves[0].isEmpty() ? new String[0] : halves[0].split(":");
            String[] right = halves[1].isEmpty() ? new String[0] : halves[1].split(":");

            if (left.length + right.length > 7) return false;

            parts = new String[left.length + right.length];
            System.arraycopy(left, 0, parts, 0, left.length);
            System.arraycopy(right, 0, parts, left.length, right.length);
        } else {
            parts = address.split(":");
            if (parts.length != 8) return false;
        }

        for (String part : parts) {
            if (part.isEmpty() || part.length() > 4) return false;
            try {
                Integer.parseInt(part, 16);
            } catch (NumberFormatException e) {
                return false;
            }
        }
        return true;
    }

    private static boolean validateDate(String value) {
        if (!DATE_RE.matcher(value).matches()) return false;
        try {
            LocalDate.parse(value);
            return true;
        } catch (DateTimeParseException e) {
            return false;
        }
    }

    private static boolean validateDateTime(String value) {
        if (!DATETIME_RE.matcher(value).matches()) return false;
        try {
            String normalized = value.replace(" ", "T");
            if (normalized.endsWith("Z")) {
                normalized = normalized.substring(0, normalized.length() - 1) + "+00:00";
            }
            OffsetDateTime.parse(normalized);
            return true;
        } catch (DateTimeParseException e) {
            return false;
        }
    }
}

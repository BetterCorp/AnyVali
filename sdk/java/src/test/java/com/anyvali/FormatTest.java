package com.anyvali;

import com.anyvali.format.FormatValidators;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Tests for format validators.
 */
class FormatTest {

    @Nested
    class EmailTests {
        @Test
        void validEmail() {
            assertTrue(FormatValidators.validate("email", "user@example.com"));
            assertTrue(FormatValidators.validate("email", "user.name+tag@example.co.uk"));
        }

        @Test
        void invalidEmail() {
            assertFalse(FormatValidators.validate("email", "notanemail"));
            assertFalse(FormatValidators.validate("email", "@example.com"));
            assertFalse(FormatValidators.validate("email", "user@"));
            assertFalse(FormatValidators.validate("email", ""));
        }
    }

    @Nested
    class UrlTests {
        @Test
        void validUrl() {
            assertTrue(FormatValidators.validate("url", "http://example.com"));
            assertTrue(FormatValidators.validate("url", "https://example.com/path"));
            assertTrue(FormatValidators.validate("url", "https://example.com:8080/path"));
        }

        @Test
        void invalidUrl() {
            assertFalse(FormatValidators.validate("url", "ftp://example.com"));
            assertFalse(FormatValidators.validate("url", "not a url"));
            assertFalse(FormatValidators.validate("url", ""));
        }
    }

    @Nested
    class UuidTests {
        @Test
        void validUuid() {
            assertTrue(FormatValidators.validate("uuid", "550e8400-e29b-41d4-a716-446655440000"));
            assertTrue(FormatValidators.validate("uuid", "00000000-0000-0000-0000-000000000000"));
        }

        @Test
        void invalidUuid() {
            assertFalse(FormatValidators.validate("uuid", "not-a-uuid"));
            assertFalse(FormatValidators.validate("uuid", "550e8400-e29b-41d4-a716"));
            assertFalse(FormatValidators.validate("uuid", ""));
        }
    }

    @Nested
    class Ipv4Tests {
        @Test
        void validIpv4() {
            assertTrue(FormatValidators.validate("ipv4", "192.168.1.1"));
            assertTrue(FormatValidators.validate("ipv4", "0.0.0.0"));
            assertTrue(FormatValidators.validate("ipv4", "255.255.255.255"));
        }

        @Test
        void invalidIpv4() {
            assertFalse(FormatValidators.validate("ipv4", "256.1.1.1"));
            assertFalse(FormatValidators.validate("ipv4", "1.2.3"));
            assertFalse(FormatValidators.validate("ipv4", "not an ip"));
            assertFalse(FormatValidators.validate("ipv4", ""));
        }
    }

    @Nested
    class Ipv6Tests {
        @Test
        void validIpv6() {
            assertTrue(FormatValidators.validate("ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"));
            assertTrue(FormatValidators.validate("ipv6", "::1"));
            assertTrue(FormatValidators.validate("ipv6", "fe80::1"));
            assertTrue(FormatValidators.validate("ipv6", "::"));
        }

        @Test
        void invalidIpv6() {
            assertFalse(FormatValidators.validate("ipv6", "not an ipv6"));
            assertFalse(FormatValidators.validate("ipv6", ""));
            assertFalse(FormatValidators.validate("ipv6", "192.168.1.1"));
            assertFalse(FormatValidators.validate("ipv6", ":::1"));
        }
    }

    @Nested
    class DateTests {
        @Test
        void validDate() {
            assertTrue(FormatValidators.validate("date", "2024-01-15"));
            assertTrue(FormatValidators.validate("date", "2000-12-31"));
        }

        @Test
        void invalidDate() {
            assertFalse(FormatValidators.validate("date", "2024-13-01"));
            assertFalse(FormatValidators.validate("date", "2024-02-30"));
            assertFalse(FormatValidators.validate("date", "not a date"));
            assertFalse(FormatValidators.validate("date", ""));
            assertFalse(FormatValidators.validate("date", "24-01-15"));
        }
    }

    @Nested
    class DateTimeTests {
        @Test
        void validDateTime() {
            assertTrue(FormatValidators.validate("date-time", "2024-01-15T10:30:00Z"));
            assertTrue(FormatValidators.validate("date-time", "2024-01-15T10:30:00+05:30"));
            assertTrue(FormatValidators.validate("date-time", "2024-01-15T10:30:00.123Z"));
        }

        @Test
        void invalidDateTime() {
            // No timezone
            assertFalse(FormatValidators.validate("date-time", "2024-01-15T10:30:00"));
            assertFalse(FormatValidators.validate("date-time", "not a datetime"));
            assertFalse(FormatValidators.validate("date-time", ""));
        }
    }

    @Nested
    class UnknownFormatTests {
        @Test
        void unknownFormatsPass() {
            assertTrue(FormatValidators.validate("custom-format", "anything"));
            assertTrue(FormatValidators.validate("some-new-format", ""));
        }
    }

    // ---- Integration with StringSchema ----
    @Nested
    class FormatInSchemaTests {
        @Test
        void emailFormatInString() {
            var s = AnyVali.string().format("email");
            assertTrue(s.safeParse("test@example.com").success());
            assertFalse(s.safeParse("notanemail").success());
        }

        @Test
        void urlFormatInString() {
            var s = AnyVali.string().format("url");
            assertTrue(s.safeParse("https://example.com").success());
            assertFalse(s.safeParse("not a url").success());
        }

        @Test
        void dateFormatInString() {
            var s = AnyVali.string().format("date");
            assertTrue(s.safeParse("2024-01-15").success());
            assertFalse(s.safeParse("not a date").success());
        }

        @Test
        void uuidFormatInString() {
            var s = AnyVali.string().format("uuid");
            assertTrue(s.safeParse("550e8400-e29b-41d4-a716-446655440000").success());
            assertFalse(s.safeParse("bad").success());
        }
    }
}

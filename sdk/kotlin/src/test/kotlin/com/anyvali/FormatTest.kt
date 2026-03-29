package com.anyvali

import com.anyvali.format.FormatValidators
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue
import kotlin.test.assertFalse

class FormatTest {

    // ---- Email ----

    @Test
    fun `email accepts valid email`() {
        val s = string().format("email")
        assertEquals("user@example.com", s.parse("user@example.com"))
    }

    @Test
    fun `email rejects without at`() {
        val s = string().format("email")
        val result = s.safeParse("not-an-email")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_STRING, result.issues[0].code)
        assertEquals("email", result.issues[0].expected)
    }

    @Test
    fun `email rejects without dot after at`() {
        val s = string().format("email")
        val result = s.safeParse("user@localhost")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- URL ----

    @Test
    fun `url accepts https`() {
        val s = string().format("url")
        assertEquals("https://example.com", s.parse("https://example.com"))
    }

    @Test
    fun `url accepts http with path and query`() {
        val s = string().format("url")
        assertEquals("http://example.com/path?q=1", s.parse("http://example.com/path?q=1"))
    }

    @Test
    fun `url rejects ftp`() {
        val s = string().format("url")
        val result = s.safeParse("ftp://files.example.com")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- UUID ----

    @Test
    fun `uuid accepts valid`() {
        val s = string().format("uuid")
        assertEquals("550e8400-e29b-41d4-a716-446655440000", s.parse("550e8400-e29b-41d4-a716-446655440000"))
    }

    @Test
    fun `uuid rejects invalid`() {
        val s = string().format("uuid")
        val result = s.safeParse("not-a-uuid")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- IPv4 ----

    @Test
    fun `ipv4 accepts valid`() {
        val s = string().format("ipv4")
        assertEquals("192.168.1.1", s.parse("192.168.1.1"))
    }

    @Test
    fun `ipv4 rejects leading zeros`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.01.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `ipv4 rejects out of range octet`() {
        val s = string().format("ipv4")
        val result = s.safeParse("256.1.1.1")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- IPv6 ----

    @Test
    fun `ipv6 accepts full address`() {
        val s = string().format("ipv6")
        assertEquals("2001:0db8:85a3:0000:0000:8a2e:0370:7334", s.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))
    }

    @Test
    fun `ipv6 accepts compressed`() {
        val s = string().format("ipv6")
        assertEquals("::1", s.parse("::1"))
    }

    @Test
    fun `ipv6 rejects invalid`() {
        val s = string().format("ipv6")
        val result = s.safeParse("not:an:ipv6")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Date ----

    @Test
    fun `date accepts valid date`() {
        val s = string().format("date")
        assertEquals("2024-02-29", s.parse("2024-02-29"))
    }

    @Test
    fun `date rejects invalid leap day`() {
        val s = string().format("date")
        val result = s.safeParse("2023-02-29")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- DateTime ----

    @Test
    fun `datetime accepts with Z`() {
        val s = string().format("date-time")
        assertEquals("2024-01-15T10:30:00Z", s.parse("2024-01-15T10:30:00Z"))
    }

    @Test
    fun `datetime accepts with offset`() {
        val s = string().format("date-time")
        assertEquals("2024-01-15T10:30:00+05:30", s.parse("2024-01-15T10:30:00+05:30"))
    }

    @Test
    fun `datetime rejects without timezone`() {
        val s = string().format("date-time")
        val result = s.safeParse("2024-01-15T10:30:00")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Direct FormatValidators tests ----

    @Test
    fun `FormatValidators unknown format passes`() {
        assertTrue(FormatValidators.validate("custom-format", "anything"))
    }

    @Test
    fun `FormatValidators ipv4 edge cases`() {
        assertTrue(FormatValidators.validate("ipv4", "0.0.0.0"))
        assertTrue(FormatValidators.validate("ipv4", "255.255.255.255"))
        assertFalse(FormatValidators.validate("ipv4", "1.2.3"))
        assertFalse(FormatValidators.validate("ipv4", "1.2.3.4.5"))
        assertFalse(FormatValidators.validate("ipv4", ""))
    }

    @Test
    fun `FormatValidators ipv6 edge cases`() {
        assertTrue(FormatValidators.validate("ipv6", "fe80::1"))
        assertTrue(FormatValidators.validate("ipv6", "::"))
        assertFalse(FormatValidators.validate("ipv6", ""))
        assertFalse(FormatValidators.validate("ipv6", ":::"))
    }

    @Test
    fun `FormatValidators date edge cases`() {
        assertTrue(FormatValidators.validate("date", "2024-01-01"))
        assertFalse(FormatValidators.validate("date", "2024-13-01"))
        assertFalse(FormatValidators.validate("date", "not-a-date"))
    }

    @Test
    fun `FormatValidators url edge cases`() {
        assertFalse(FormatValidators.validate("url", "not a url"))
        assertFalse(FormatValidators.validate("url", ""))
    }

    @Test
    fun `FormatValidators email edge cases`() {
        assertFalse(FormatValidators.validate("email", ""))
        assertFalse(FormatValidators.validate("email", "@"))
        assertFalse(FormatValidators.validate("email", "a@"))
        assertTrue(FormatValidators.validate("email", "a@b.c"))
    }

    @Test
    fun `FormatValidators uuid edge cases`() {
        assertFalse(FormatValidators.validate("uuid", ""))
        assertFalse(FormatValidators.validate("uuid", "550e8400-e29b-41d4-a716"))
        assertTrue(FormatValidators.validate("uuid", "550E8400-E29B-41D4-A716-446655440000"))
    }
}

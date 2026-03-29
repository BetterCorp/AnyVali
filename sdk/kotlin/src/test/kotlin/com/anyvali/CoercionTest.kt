package com.anyvali

import com.anyvali.schemas.*
import com.anyvali.parse.Coercion
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull

class CoercionTest {

    @Test
    fun `string to int coerces valid string`() {
        val i = int_().coerce("string->int")
        assertEquals(42L, i.parse("42"))
    }

    @Test
    fun `string to int trims whitespace`() {
        val i = int_().coerce("string->int")
        assertEquals(42L, i.parse("  42  "))
    }

    @Test
    fun `string to number coerces valid float string`() {
        val n = number().coerce("string->number")
        assertEquals(3.14, n.parse("3.14"))
    }

    @Test
    fun `string to bool coerces true`() {
        val b = bool().coerce("string->bool")
        assertEquals(true, b.parse("true"))
    }

    @Test
    fun `string to bool coerces false`() {
        val b = bool().coerce("string->bool")
        assertEquals(false, b.parse("false"))
    }

    @Test
    fun `string to bool coerces 1`() {
        val b = bool().coerce("string->bool")
        assertEquals(true, b.parse("1"))
    }

    @Test
    fun `string to bool coerces 0`() {
        val b = bool().coerce("string->bool")
        assertEquals(false, b.parse("0"))
    }

    @Test
    fun `string to bool is case insensitive`() {
        val b = bool().coerce("string->bool")
        assertEquals(true, b.parse("TRUE"))
    }

    @Test
    fun `trim coercion removes whitespace`() {
        val s = string().coerce("trim")
        assertEquals("hello", s.parse("  hello  "))
    }

    @Test
    fun `lower coercion converts to lowercase`() {
        val s = string().coerce("lower")
        assertEquals("hello world", s.parse("HELLO World"))
    }

    @Test
    fun `upper coercion converts to uppercase`() {
        val s = string().coerce("upper")
        assertEquals("HELLO WORLD", s.parse("hello world"))
    }

    @Test
    fun `coercion failure produces coercion_failed`() {
        val i = int_().coerce("string->int")
        val result = i.safeParse("not-a-number")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.COERCION_FAILED, result.issues[0].code)
        assertEquals("int", result.issues[0].expected)
        assertEquals("not-a-number", result.issues[0].received)
    }

    @Test
    fun `coercion happens before validation`() {
        val i = int_().min(10).coerce("string->int")
        val result = i.safeParse("5")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `coercion then validation success`() {
        val i = int_().min(1).max(100).coerce("string->int")
        assertEquals(50L, i.parse("50"))
    }

    @Test
    fun `chained coercions applied left to right`() {
        val s = string().coerce("trim", "lower")
        assertEquals("hello", s.parse("  HELLO  "))
    }

    // ---- Direct Coercion utility tests ----

    @Test
    fun `Coercion coerceStringToInt works`() {
        assertEquals(42L, Coercion.coerceStringToInt("42"))
        assertEquals(-7L, Coercion.coerceStringToInt(" -7 "))
        assertNull(Coercion.coerceStringToInt("abc"))
    }

    @Test
    fun `Coercion coerceStringToNumber works`() {
        assertEquals(3.14, Coercion.coerceStringToNumber("3.14"))
        assertNull(Coercion.coerceStringToNumber("abc"))
    }

    @Test
    fun `Coercion coerceStringToBool works`() {
        assertEquals(true, Coercion.coerceStringToBool("true"))
        assertEquals(false, Coercion.coerceStringToBool("false"))
        assertEquals(true, Coercion.coerceStringToBool("1"))
        assertEquals(false, Coercion.coerceStringToBool("0"))
        assertNull(Coercion.coerceStringToBool("maybe"))
    }

    @Test
    fun `Coercion trim works`() {
        assertEquals("hello", Coercion.trimString("  hello  "))
    }

    @Test
    fun `Coercion lower works`() {
        assertEquals("hello", Coercion.lowerString("HELLO"))
    }

    @Test
    fun `Coercion upper works`() {
        assertEquals("HELLO", Coercion.upperString("hello"))
    }

    @Test
    fun `string to bool coercion failure`() {
        val b = bool().coerce("string->bool")
        val result = b.safeParse("maybe")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.COERCION_FAILED, result.issues[0].code)
    }

    @Test
    fun `number coercion failure on invalid string`() {
        val n = number().coerce("string->number")
        val result = n.safeParse("not-a-number")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.COERCION_FAILED, result.issues[0].code)
    }

    @Test
    fun `int coercion does not apply to non-string`() {
        val i = int_().coerce("string->int")
        // If input is already a number, coercion is skipped
        assertEquals(42L, i.parse(42))
    }

    @Test
    fun `number coercion does not apply to non-string`() {
        val n = number().coerce("string->number")
        assertEquals(3.14, n.parse(3.14))
    }

    @Test
    fun `bool coercion does not apply to non-string`() {
        val b = bool().coerce("string->bool")
        assertEquals(true, b.parse(true))
    }
}

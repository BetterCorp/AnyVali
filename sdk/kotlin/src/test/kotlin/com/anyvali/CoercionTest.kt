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

    // ---- No-arg coerce() ergonomic (regression) ----
    // Enabling coercion with the no-arg coerce() form (the portable generic
    // "string" source, equivalent to JS `.coerce()` / `.coerce({ from: "string" })`)
    // must coerce string input on a numeric/bool target, not silently no-op.

    @Test
    fun `number with no-arg coerce coerces float string`() {
        val n = number().coerce()
        assertEquals(3.14, n.parse("3.14"))
    }

    @Test
    fun `int with no-arg coerce coerces integer string`() {
        val i = int_().coerce()
        assertEquals(42L, i.parse("42"))
    }

    @Test
    fun `bool with no-arg coerce coerces true`() {
        val b = bool().coerce()
        assertEquals(true, b.parse("true"))
    }

    @Test
    fun `bool with no-arg coerce coerces false`() {
        val b = bool().coerce()
        assertEquals(false, b.parse("false"))
    }

    @Test
    fun `object with numeric fields using no-arg coerce coerces all string inputs`() {
        val schema = obj(
            properties = mapOf(
                "lumpSum" to number().coerce().min(0.0),
                "monthlyContributions" to number().coerce().min(0.0),
                "investmentTerm" to number().coerce().min(1.0)
            ),
            required = setOf("lumpSum", "monthlyContributions", "investmentTerm")
        )
        val result = schema.parse(
            mapOf(
                "lumpSum" to "1000",
                "monthlyContributions" to "50.5",
                "investmentTerm" to "10"
            )
        ) as Map<*, *>
        assertEquals(1000.0, result["lumpSum"])
        assertEquals(50.5, result["monthlyContributions"])
        assertEquals(10.0, result["investmentTerm"])
    }

    // ---- Canonical coercion matrix (all FROM STRING via no-arg coerce()) ----
    // Source of truth: spec 5.1 / JS reference (sdk/js/src/parse/coerce.ts).

    private fun assertCoercionFailed(result: ParseResult<*>) {
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.COERCION_FAILED, result.issues[0].code)
    }

    // string->int: ASCII ^-?\d+$ trimmed.
    @Test
    fun `matrix string to int accepts`() {
        val i = int_().coerce()
        assertEquals(42L, i.parse("42"))
        assertEquals(42L, i.parse("  42  "))
        assertEquals(-7L, i.parse("-7"))
    }

    @Test
    fun `matrix string to int rejects`() {
        for (bad in listOf("3.14", "0x10", "1_000", "+5", "Infinity", "", "abc")) {
            assertCoercionFailed(int_().coerce().safeParse(bad))
        }
    }

    // string->number: ASCII decimal float incl exponent, trimmed.
    @Test
    fun `matrix string to number accepts`() {
        val n = number().coerce()
        assertEquals(3.14, n.parse("3.14"))
        assertEquals(-1.5e3, n.parse("-1.5e3"))
        assertEquals(2.0, n.parse("  2  "))
        assertEquals(0.0, n.parse("0"))
    }

    @Test
    fun `matrix string to number rejects`() {
        for (bad in listOf("0x10", "Infinity", "NaN", "", "1_000", "abc")) {
            assertCoercionFailed(number().coerce().safeParse(bad))
        }
    }

    // string->bool: trim + case-insensitive. true<-true/TRUE/1; false<-false/0.
    @Test
    fun `matrix string to bool accepts`() {
        assertEquals(true, bool().coerce().parse("true"))
        assertEquals(true, bool().coerce().parse("TRUE"))
        assertEquals(true, bool().coerce().parse("1"))
        assertEquals(false, bool().coerce().parse("false"))
        assertEquals(false, bool().coerce().parse("0"))
    }

    @Test
    fun `matrix string to bool rejects`() {
        for (bad in listOf("yes", "no", "on", "off", "t", "f", "2", "")) {
            assertCoercionFailed(bool().coerce().safeParse(bad))
        }
    }

    // string transforms (string kind): trim, lower, upper; chainable.
    @Test
    fun `matrix string transforms`() {
        assertEquals("hello", string().coerce("trim").parse("  hello  "))
        assertEquals("hello", string().coerce("lower").parse("HELLO"))
        assertEquals("HELLO", string().coerce("upper").parse("hello"))
        // chainable, applied left to right
        assertEquals("hello", string().coerce("trim", "lower").parse("  HELLO  "))
    }

    // The strict ASCII grammars are also reflected in the Coercion helper itself.
    @Test
    fun `matrix Coercion helper rejects non-ASCII int forms`() {
        assertNull(Coercion.coerceStringToInt("+5"))
        assertNull(Coercion.coerceStringToInt("1_000"))
        assertNull(Coercion.coerceStringToInt("0x10"))
        assertNull(Coercion.coerceStringToInt("3.14"))
        assertNull(Coercion.coerceStringToInt(""))
        assertEquals(42L, Coercion.coerceStringToInt("  42  "))
        assertEquals(-7L, Coercion.coerceStringToInt("-7"))
    }

    @Test
    fun `matrix Coercion helper rejects non-decimal number forms`() {
        assertNull(Coercion.coerceStringToNumber("0x10"))
        assertNull(Coercion.coerceStringToNumber("Infinity"))
        assertNull(Coercion.coerceStringToNumber("NaN"))
        assertNull(Coercion.coerceStringToNumber("1_000"))
        assertNull(Coercion.coerceStringToNumber(""))
        assertEquals(3.14, Coercion.coerceStringToNumber("3.14"))
        assertEquals(-1.5e3, Coercion.coerceStringToNumber("-1.5e3"))
    }
}

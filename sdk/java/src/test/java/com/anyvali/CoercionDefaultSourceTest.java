package com.anyvali;

import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.Map;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Tests for the no-arg {@code coerce()} ergonomic.
 * <p>
 * "string" is the only portable coercion source, so a "just coerce" request on
 * a numeric/bool schema must coerce portable string input with the target type
 * inferred from the schema kind. {@code coerce()} (no args) enables exactly
 * this: int/number schemas coerce ASCII decimal strings, bool schemas coerce
 * boolean strings. The accept/reject grammars are identical to the explicit
 * {@code coerce(CoercionConfig)} form.
 * <p>
 * This file also pins the full FROM-STRING coercion matrix through the no-arg
 * form so a divergence in any native parser is caught at the schema boundary.
 */
class CoercionDefaultSourceTest {

    // =====================================================================
    // Smoke tests: no-arg coerce() enables string coercion per schema kind
    // =====================================================================

    /** No-arg coerce() on a number schema: "3.14" -> 3.14. */
    @Test
    void numberWithDefaultCoerceParsesDecimalString() {
        var s = number().coerce();
        var r = s.safeParse("3.14");
        assertTrue(r.success(),
                "number().coerce() should coerce \"3.14\"; issues="
                        + (r.success() ? "" : r.issues()));
        assertEquals(3.14, s.parse("3.14"));
    }

    /** No-arg coerce() on an int schema: "42" -> 42. */
    @Test
    void intWithDefaultCoerceParsesIntString() {
        var s = int_().coerce();
        var r = s.safeParse("42");
        assertTrue(r.success(),
                "int_().coerce() should coerce \"42\"; issues="
                        + (r.success() ? "" : r.issues()));
        assertEquals(42L, s.parse("42"));
    }

    /** No-arg coerce() on a bool schema: "true" -> true. */
    @Test
    void boolWithDefaultCoerceParsesTrueString() {
        var s = bool_().coerce();
        var r = s.safeParse("true");
        assertTrue(r.success(),
                "bool_().coerce() should coerce \"true\"; issues="
                        + (r.success() ? "" : r.issues()));
        assertEquals(true, s.parse("true"));
    }

    /** No-arg coerce() on a bool schema: "false" -> false. */
    @Test
    void boolWithDefaultCoerceParsesFalseString() {
        var s = bool_().coerce();
        var r = s.safeParse("false");
        assertTrue(r.success(),
                "bool_().coerce() should coerce \"false\"; issues="
                        + (r.success() ? "" : r.issues()));
        assertEquals(false, s.parse("false"));
    }

    /**
     * Object with numeric fields, each using the no-arg coerce(). All string
     * values should coerce and the parse should succeed.
     */
    @Test
    void objectWithDefaultCoerceNumericFields() {
        var schema = object_(Map.of(
                "lumpSum", number().coerce(),
                "monthlyContributions", number().coerce(),
                "investmentTerm", int_().coerce()
        ));

        var input = new LinkedHashMap<String, Object>();
        input.put("lumpSum", "1000000");
        input.put("monthlyContributions", "1000");
        input.put("investmentTerm", "20");

        var r = schema.safeParse(input);
        assertTrue(r.success(),
                "object with no-arg-coerce numeric fields should parse; issues="
                        + (r.success() ? "" : r.issues()));

        var out = r.data();
        assertEquals(1000000.0, out.get("lumpSum"));
        assertEquals(1000.0, out.get("monthlyContributions"));
        assertEquals(20L, out.get("investmentTerm"));
    }

    // =====================================================================
    // Matrix: string -> int  (ASCII ^-?\d+$, trimmed)
    // =====================================================================

    @Test
    void stringToIntAccepts() {
        var s = int_().coerce();
        assertEquals(42L, s.parse("42"));
        assertEquals(42L, s.parse("  42  "));
        assertEquals(-7L, s.parse("-7"));
    }

    @Test
    void stringToIntRejects() {
        var s = int_().coerce();
        for (String bad : new String[]{"3.14", "0x10", "1_000", "+5", "Infinity", "", "abc"}) {
            var r = s.safeParse(bad);
            assertFalse(r.success(), "string->int must reject '" + bad + "'");
            assertEquals(IssueCodes.COERCION_FAILED, r.issues().get(0).code(),
                    "string->int reject of '" + bad + "' should be coercion_failed");
        }
    }

    // =====================================================================
    // Matrix: string -> number  (ASCII decimal float incl exponent, trimmed)
    // =====================================================================

    @Test
    void stringToNumberAccepts() {
        var s = number().coerce();
        assertEquals(3.14, s.parse("3.14"));
        assertEquals(-1.5e3, s.parse("-1.5e3"));
        assertEquals(2.0, s.parse("  2  "));
        assertEquals(0.0, s.parse("0"));
    }

    @Test
    void stringToNumberRejects() {
        var s = number().coerce();
        for (String bad : new String[]{"0x10", "Infinity", "NaN", "", "1_000", "abc"}) {
            var r = s.safeParse(bad);
            assertFalse(r.success(), "string->number must reject '" + bad + "'");
            assertEquals(IssueCodes.COERCION_FAILED, r.issues().get(0).code(),
                    "string->number reject of '" + bad + "' should be coercion_failed");
        }
    }

    // =====================================================================
    // Matrix: string -> bool  (trim + case-insensitive; true/1, false/0)
    // =====================================================================

    @Test
    void stringToBoolAcceptsTrue() {
        var s = bool_().coerce();
        assertEquals(true, s.parse("true"));
        assertEquals(true, s.parse("TRUE"));
        assertEquals(true, s.parse("1"));
    }

    @Test
    void stringToBoolAcceptsFalse() {
        var s = bool_().coerce();
        assertEquals(false, s.parse("false"));
        assertEquals(false, s.parse("0"));
    }

    @Test
    void stringToBoolRejects() {
        var s = bool_().coerce();
        for (String bad : new String[]{"yes", "no", "on", "off", "t", "f", "2", ""}) {
            var r = s.safeParse(bad);
            assertFalse(r.success(), "string->bool must reject '" + bad + "'");
            assertEquals(IssueCodes.COERCION_FAILED, r.issues().get(0).code(),
                    "string->bool reject of '" + bad + "' should be coercion_failed");
        }
    }

    // =====================================================================
    // String transforms remain string-kind only and chainable.
    // The no-arg coerce() is for type inference; string transforms still use
    // the explicit CoercionConfig form (no portable target to infer).
    // =====================================================================

    @Test
    void stringTransformsTrimLowerUpperChain() {
        // trim + lower
        var lower = string().coerce(
                com.anyvali.parse.CoercionConfig.builder().trim(true).lower(true).build());
        assertEquals("hello", lower.parse("  HELLO  "));

        // trim + upper
        var upper = string().coerce(
                com.anyvali.parse.CoercionConfig.builder().trim(true).upper(true).build());
        assertEquals("HELLO", upper.parse("  hello  "));
    }

    /**
     * No-arg coerce() has no inferable target for a string schema, so it must
     * fail fast rather than silently no-op.
     */
    @Test
    void stringNoArgCoerceThrows() {
        assertThrows(IllegalStateException.class, () -> string().coerce());
    }
}

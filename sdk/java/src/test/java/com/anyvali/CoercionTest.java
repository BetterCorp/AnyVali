package com.anyvali;

import com.anyvali.parse.Coercion;
import com.anyvali.parse.CoercionConfig;
import org.junit.jupiter.api.Test;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Tests for portable coercion behavior.
 */
class CoercionTest {

    // ---- String to Int ----
    @Test
    void coerceStringToInt() {
        var config = CoercionConfig.builder().toInt(true).build();
        var ctx = new ValidationContext();
        assertEquals(42L, Coercion.applyCoercion("42", config, ctx));
        assertFalse(ctx.hasIssues());
    }

    @Test
    void coerceStringToIntWithSpaces() {
        var config = CoercionConfig.builder().toInt(true).build();
        var ctx = new ValidationContext();
        assertEquals(42L, Coercion.applyCoercion("  42  ", config, ctx));
    }

    @Test
    void coerceStringToIntFromFloat() {
        var config = CoercionConfig.builder().toInt(true).build();
        var ctx = new ValidationContext();
        assertEquals(42L, Coercion.applyCoercion("42.0", config, ctx));
    }

    @Test
    void coerceStringToIntFails() {
        var config = CoercionConfig.builder().toInt(true).build();
        var ctx = new ValidationContext();
        Coercion.applyCoercion("not a number", config, ctx);
        assertTrue(ctx.hasIssues());
        assertEquals(IssueCodes.COERCION_FAILED, ctx.getIssues().get(0).code());
    }

    @Test
    void coerceStringToIntNonWholeFloat() {
        var config = CoercionConfig.builder().toInt(true).build();
        var ctx = new ValidationContext();
        Coercion.applyCoercion("3.14", config, ctx);
        assertTrue(ctx.hasIssues());
    }

    // ---- String to Number ----
    @Test
    void coerceStringToNumber() {
        var config = CoercionConfig.builder().toNumber(true).build();
        var ctx = new ValidationContext();
        assertEquals(3.14, Coercion.applyCoercion("3.14", config, ctx));
    }

    @Test
    void coerceStringToNumberFails() {
        var config = CoercionConfig.builder().toNumber(true).build();
        var ctx = new ValidationContext();
        Coercion.applyCoercion("abc", config, ctx);
        assertTrue(ctx.hasIssues());
        assertEquals(IssueCodes.COERCION_FAILED, ctx.getIssues().get(0).code());
    }

    // ---- String to Bool ----
    @Test
    void coerceStringToBoolTrue() {
        var config = CoercionConfig.builder().toBool(true).build();
        var ctx = new ValidationContext();
        assertEquals(true, Coercion.applyCoercion("true", config, ctx));
        ctx = new ValidationContext();
        assertEquals(true, Coercion.applyCoercion("1", config, ctx));
        ctx = new ValidationContext();
        assertEquals(true, Coercion.applyCoercion("yes", config, ctx));
    }

    @Test
    void coerceStringToBoolFalse() {
        var config = CoercionConfig.builder().toBool(true).build();
        var ctx = new ValidationContext();
        assertEquals(false, Coercion.applyCoercion("false", config, ctx));
        ctx = new ValidationContext();
        assertEquals(false, Coercion.applyCoercion("0", config, ctx));
        ctx = new ValidationContext();
        assertEquals(false, Coercion.applyCoercion("no", config, ctx));
    }

    @Test
    void coerceStringToBoolFails() {
        var config = CoercionConfig.builder().toBool(true).build();
        var ctx = new ValidationContext();
        Coercion.applyCoercion("maybe", config, ctx);
        assertTrue(ctx.hasIssues());
        assertEquals(IssueCodes.COERCION_FAILED, ctx.getIssues().get(0).code());
    }

    @Test
    void coerceStringToBoolCaseInsensitive() {
        var config = CoercionConfig.builder().toBool(true).build();
        var ctx = new ValidationContext();
        assertEquals(true, Coercion.applyCoercion("TRUE", config, ctx));
        ctx = new ValidationContext();
        assertEquals(false, Coercion.applyCoercion("FALSE", config, ctx));
    }

    // ---- Trim ----
    @Test
    void trimWhitespace() {
        var config = CoercionConfig.builder().trim(true).build();
        var ctx = new ValidationContext();
        assertEquals("hello", Coercion.applyCoercion("  hello  ", config, ctx));
    }

    @Test
    void trimOnNonString() {
        var config = CoercionConfig.builder().trim(true).build();
        var ctx = new ValidationContext();
        // Trim does not apply to non-strings
        assertEquals(42, Coercion.applyCoercion(42, config, ctx));
    }

    // ---- Lower ----
    @Test
    void lowerCase() {
        var config = CoercionConfig.builder().lower(true).build();
        var ctx = new ValidationContext();
        assertEquals("hello", Coercion.applyCoercion("HELLO", config, ctx));
    }

    // ---- Upper ----
    @Test
    void upperCase() {
        var config = CoercionConfig.builder().upper(true).build();
        var ctx = new ValidationContext();
        assertEquals("HELLO", Coercion.applyCoercion("hello", config, ctx));
    }

    // ---- Combined coercion + validation ----
    @Test
    void coercionInSchemaPipeline() {
        var s = int_().coerce(CoercionConfig.builder().toInt(true).build());
        assertEquals(42L, s.parse("42"));
    }

    @Test
    void coercionFailureReturnsCoercionFailed() {
        var s = int_().coerce(CoercionConfig.builder().toInt(true).build());
        var r = s.safeParse("abc");
        assertFalse(r.success());
        assertEquals(IssueCodes.COERCION_FAILED, r.issues().get(0).code());
    }

    @Test
    void trimThenValidate() {
        var s = string().minLength(1).coerce(CoercionConfig.builder().trim(true).build());
        assertEquals("hello", s.parse("  hello  "));
    }

    @Test
    void trimMakesTooShort() {
        var s = string().minLength(1).coerce(CoercionConfig.builder().trim(true).build());
        var r = s.safeParse("   ");
        assertFalse(r.success());
        assertEquals(IssueCodes.TOO_SMALL, r.issues().get(0).code());
    }

    // ---- Coercion ordering ----
    @Test
    void coercionBeforeDefaults() {
        // Coercion only applies to present values
        var s = string().withDefault("default").coerce(
                CoercionConfig.builder().trim(true).build());
        var ctx = new ValidationContext();
        // Absent -> default is applied, not coercion
        assertEquals("default", s.runPipeline(Schema.ABSENT, ctx));
    }

    @Test
    void coercionThenValidation() {
        var s = number().min(0).coerce(CoercionConfig.builder().toNumber(true).build());
        assertEquals(42.0, s.parse("42"));
        var r = s.safeParse("-5");
        assertFalse(r.success());
    }

    // ---- Config builder ----
    @Test
    void configBuilderDefaults() {
        var config = CoercionConfig.builder().build();
        assertFalse(config.toInt());
        assertFalse(config.toNumber());
        assertFalse(config.toBool());
        assertFalse(config.trim());
        assertFalse(config.lower());
        assertFalse(config.upper());
    }

    @Test
    void configBuilderAllTrue() {
        var config = CoercionConfig.builder()
                .toInt(true).toNumber(true).toBool(true)
                .trim(true).lower(true).upper(true)
                .build();
        assertTrue(config.toInt());
        assertTrue(config.toNumber());
        assertTrue(config.toBool());
        assertTrue(config.trim());
        assertTrue(config.lower());
        assertTrue(config.upper());
    }
}

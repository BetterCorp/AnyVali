package com.anyvali;

import com.anyvali.format.FormatValidators;
import com.anyvali.interchange.Importer;
import com.anyvali.schemas.*;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.*;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Security-focused tests covering known CVE/CWE vulnerability categories.
 */
class SecurityTest {

    // ---- CVE-2016-4055 / CVE-2022-25883: ReDoS (Regular Expression Denial of Service) ----
    @Nested
    class CVE_2016_4055_CVE_2022_25883_ReDoS {

        @Test
        void catastrophicBacktracking_aPlus_Plus_Dollar() {
            // Pattern (a+)+$ is a classic ReDoS trigger with input "aaa...!"
            var s = string().pattern("(a+)+$");
            String malicious = "a".repeat(30) + "!";
            long start = System.nanoTime();
            var r = s.safeParse(malicious);
            long elapsed = System.nanoTime() - start;
            assertFalse(r.success());
            assertTrue(elapsed < 5_000_000_000L, "ReDoS: pattern (a+)+$ took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void catastrophicBacktracking_nestedQuantifiers() {
            // Pattern (a|a)+ with non-matching suffix
            var s = string().pattern("(a|a)+$");
            String malicious = "a".repeat(30) + "!";
            long start = System.nanoTime();
            var r = s.safeParse(malicious);
            long elapsed = System.nanoTime() - start;
            assertFalse(r.success());
            assertTrue(elapsed < 5_000_000_000L, "ReDoS: pattern (a|a)+$ took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void catastrophicBacktracking_aStar_aStar_b() {
            // Pattern a*a*b causes exponential backtracking on "aaa...c"
            var s = string().pattern("^a*a*b$");
            String malicious = "a".repeat(30) + "c";
            long start = System.nanoTime();
            var r = s.safeParse(malicious);
            long elapsed = System.nanoTime() - start;
            assertFalse(r.success());
            assertTrue(elapsed < 5_000_000_000L, "ReDoS: pattern a*a*b took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void catastrophicBacktracking_importedPattern() {
            // ReDoS via imported schema pattern
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "string", "pattern", "(a+)+$"));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            Schema<?> imported = Importer.importSchema(doc);
            String malicious = "a".repeat(30) + "!";
            long start = System.nanoTime();
            var r = imported.safeParse(malicious);
            long elapsed = System.nanoTime() - start;
            assertFalse(r.success());
            assertTrue(elapsed < 5_000_000_000L, "ReDoS via import took " + elapsed / 1_000_000 + "ms");
        }
    }

    // ---- CVE-2003-1564: Recursive $ref (billion laughs / self-referencing schemas) ----
    @Nested
    class CVE_2003_1564_RecursiveRef {

        @Test
        void selfReferencingRefDoesNotHang() {
            // CVE-2003-1564 parse-time: a schema where root = ref(Self) and
            // Self = ref(Self) creates an unbounded recursion at parse. JVM
            // raises StackOverflowError (an Error, not Exception) so we must
            // catch Throwable; otherwise the assertion would not see the
            // expected defense and the test could appear to "hang". The wall-
            // clock bound additionally guards against true hangs (e.g. trivial
            // ref-cycle short-circuit returning success).
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "ref", "ref", "#/definitions/Self"));

            var selfDef = new LinkedHashMap<String, Object>();
            selfDef.put("kind", "ref");
            selfDef.put("ref", "#/definitions/Self");
            doc.put("definitions", Map.of("Self", selfDef));
            doc.put("extensions", Map.of());

            long start = System.nanoTime();
            boolean threw = false;
            ParseResult<?> result = null;
            try {
                Schema<?> imported = Importer.importSchema(doc);
                result = imported.safeParse("test");
            } catch (Throwable t) {
                // StackOverflowError is acceptable: it proves the runtime
                // surfaces the cycle rather than silently looping.
                threw = true;
            }
            long elapsed = System.nanoTime() - start;
            assertTrue(elapsed < 5_000_000_000L,
                "Self-cycle parse hung " + elapsed / 1_000_000 + "ms > 5s (CVE-2003-1564)");
            // Returning success on a pure self-cycle would be a logic bug.
            if (!threw && result != null) {
                assertFalse(result.success(),
                    "Self-cycle parse must not silently succeed");
            }
        }

        @Test
        void mutuallyRecursiveRefsDoNotHang() {
            // A -> B -> A cycle
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "ref", "ref", "#/definitions/A"));

            var defA = new LinkedHashMap<String, Object>();
            defA.put("kind", "ref");
            defA.put("ref", "#/definitions/B");

            var defB = new LinkedHashMap<String, Object>();
            defB.put("kind", "ref");
            defB.put("ref", "#/definitions/A");

            var defs = new LinkedHashMap<String, Object>();
            defs.put("A", defA);
            defs.put("B", defB);
            doc.put("definitions", defs);
            doc.put("extensions", Map.of());

            long start = System.nanoTime();
            try {
                Schema<?> imported = Importer.importSchema(doc);
                imported.safeParse("test");
            } catch (Exception e) {
                // Throwing is acceptable
            }
            long elapsed = System.nanoTime() - start;
            assertTrue(elapsed < 5_000_000_000L, "Mutual $ref cycle took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void deeplyNestedSchemaImportDoesNotHang() {
            // Build a deeply nested object schema: {a: {a: {a: ... {kind: string}}}}
            Map<String, Object> inner = Map.of("kind", "string");
            for (int i = 0; i < 100; i++) {
                var props = new LinkedHashMap<String, Object>();
                props.put("a", inner);
                var obj = new LinkedHashMap<String, Object>();
                obj.put("kind", "object");
                obj.put("properties", props);
                obj.put("required", List.of("a"));
                inner = obj;
            }

            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", inner);
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            long start = System.nanoTime();
            try {
                Schema<?> imported = Importer.importSchema(doc);
                assertNotNull(imported);
            } catch (StackOverflowError | Exception e) {
                // Acceptable: either handles gracefully or fails fast
            }
            long elapsed = System.nanoTime() - start;
            assertTrue(elapsed < 5_000_000_000L, "Deeply nested import took " + elapsed / 1_000_000 + "ms");
        }
    }

    // ---- CWE-190: Integer Overflow ----
    @Nested
    class CWE_190_IntegerOverflow {

        @Test
        void int8_boundaryValues() {
            assertTrue(int8().safeParse(-128).success());
            assertTrue(int8().safeParse(127).success());
            assertFalse(int8().safeParse(-129).success());
            assertFalse(int8().safeParse(128).success());
        }

        @Test
        void int8_farOutOfRange() {
            assertFalse(int8().safeParse(1000).success());
            assertFalse(int8().safeParse(-1000).success());
            assertFalse(int8().safeParse(Long.MAX_VALUE).success());
            assertFalse(int8().safeParse(Long.MIN_VALUE).success());
        }

        @Test
        void int16_boundaryValues() {
            assertTrue(int16().safeParse(-32768).success());
            assertTrue(int16().safeParse(32767).success());
            assertFalse(int16().safeParse(-32769).success());
            assertFalse(int16().safeParse(32768).success());
        }

        @Test
        void int16_farOutOfRange() {
            assertFalse(int16().safeParse(Long.MAX_VALUE).success());
            assertFalse(int16().safeParse(Long.MIN_VALUE).success());
        }

        @Test
        void int32_boundaryValues() {
            assertTrue(int32().safeParse(Integer.MIN_VALUE).success());
            assertTrue(int32().safeParse(Integer.MAX_VALUE).success());
            assertFalse(int32().safeParse((long) Integer.MIN_VALUE - 1).success());
            assertFalse(int32().safeParse((long) Integer.MAX_VALUE + 1).success());
        }

        @Test
        void int32_farOutOfRange() {
            assertFalse(int32().safeParse(Long.MAX_VALUE).success());
            assertFalse(int32().safeParse(Long.MIN_VALUE).success());
        }

        @Test
        void int64_boundaryValues() {
            assertTrue(int64().safeParse(Long.MIN_VALUE).success());
            assertTrue(int64().safeParse(Long.MAX_VALUE).success());
        }

        @Test
        void uint8_boundaryValues() {
            assertTrue(uint8().safeParse(0).success());
            assertTrue(uint8().safeParse(255).success());
            assertFalse(uint8().safeParse(-1).success());
            assertFalse(uint8().safeParse(256).success());
        }

        @Test
        void uint16_boundaryValues() {
            assertTrue(uint16().safeParse(0).success());
            assertTrue(uint16().safeParse(65535).success());
            assertFalse(uint16().safeParse(-1).success());
            assertFalse(uint16().safeParse(65536).success());
        }

        @Test
        void uint32_boundaryValues() {
            assertTrue(uint32().safeParse(0).success());
            assertTrue(uint32().safeParse(4294967295L).success());
            assertFalse(uint32().safeParse(-1).success());
            assertFalse(uint32().safeParse(4294967296L).success());
        }

        @Test
        void uint64_boundaryValues() {
            assertTrue(uint64().safeParse(0).success());
            assertTrue(uint64().safeParse(Long.MAX_VALUE).success());
            assertFalse(uint64().safeParse(-1).success());
        }

        @Test
        void int_rejectsFloatingPoint() {
            assertFalse(int_().safeParse(3.14).success());
            assertFalse(int8().safeParse(3.14).success());
            assertFalse(int16().safeParse(3.14).success());
            assertFalse(int32().safeParse(3.14).success());
            assertFalse(uint8().safeParse(3.14).success());
            assertFalse(uint16().safeParse(3.14).success());
            assertFalse(uint32().safeParse(3.14).success());
        }

        @Test
        void int_acceptsWholeFloatWithinBounds() {
            // 42.0 should be accepted as integer 42
            assertTrue(int8().safeParse(42.0).success());
            assertTrue(int16().safeParse(42.0).success());
            assertTrue(int32().safeParse(42.0).success());
            assertTrue(int64().safeParse(42.0).success());
        }

        @Test
        void int_rejectsWholeFloatOutOfBounds() {
            // 200.0 is within float range but outside int8
            assertFalse(int8().safeParse(200.0).success());
            // 40000.0 is outside int16
            assertFalse(int16().safeParse(40000.0).success());
        }
    }

    // ---- CWE-20: NaN and Infinity rejection ----
    @Nested
    class CWE_20_NaN_Infinity {

        @Test
        void number_rejectsNaN() {
            var r = number().safeParse(Double.NaN);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_NUMBER, r.issues().get(0).code());
        }

        @Test
        void number_rejectsPositiveInfinity() {
            var r = number().safeParse(Double.POSITIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void number_rejectsNegativeInfinity() {
            var r = number().safeParse(Double.NEGATIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void float64_rejectsNaN() {
            var r = float64().safeParse(Double.NaN);
            assertFalse(r.success());
        }

        @Test
        void float64_rejectsPositiveInfinity() {
            var r = float64().safeParse(Double.POSITIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void float64_rejectsNegativeInfinity() {
            var r = float64().safeParse(Double.NEGATIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void float32_rejectsNaN() {
            var r = float32().safeParse(Double.NaN);
            assertFalse(r.success());
        }

        @Test
        void float32_rejectsPositiveInfinity() {
            var r = float32().safeParse(Double.POSITIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void float32_rejectsNegativeInfinity() {
            var r = float32().safeParse(Double.NEGATIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void int_rejectsNaN() {
            var r = int_().safeParse(Double.NaN);
            assertFalse(r.success());
        }

        @Test
        void int_rejectsInfinity() {
            assertFalse(int_().safeParse(Double.POSITIVE_INFINITY).success());
            assertFalse(int_().safeParse(Double.NEGATIVE_INFINITY).success());
        }

        @Test
        void numberWithConstraints_rejectsNaN() {
            // NaN should fail even when constraints are present
            assertFalse(number().min(0).max(100).safeParse(Double.NaN).success());
        }

        @Test
        void numberWithConstraints_rejectsInfinity() {
            // Infinity should fail even if seemingly within no max bound
            assertFalse(number().min(0).safeParse(Double.POSITIVE_INFINITY).success());
        }

        @Test
        void floatNaN_asBoxedFloat() {
            // Float.NaN promoted to double
            assertFalse(number().safeParse((double) Float.NaN).success());
            assertFalse(float32().safeParse((double) Float.NaN).success());
        }
    }

    // ---- CWE-20: Format Bypass ----
    @Nested
    class CWE_20_FormatBypass {

        @Test
        void email_rejectsNoDomain() {
            var s = string().format("email");
            assertFalse(s.safeParse("user@").success());
        }

        @Test
        void email_tamperedFormatNameNotSilentlyIgnored() {
            var s = string().format("email\u0000");
            assertFalse(s.safeParse("not-an-email").success());
        }

        @Test
        void email_importedTamperedFormatNameNotUnconstrained() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "string", "format", "email\u0000"));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());
            Schema<?> imported = Importer.importSchema(doc);
            assertFalse(imported.safeParse("not-an-email").success());
        }

        @Test
        void email_rejectsNoLocalPart() {
            var s = string().format("email");
            assertFalse(s.safeParse("@example.com").success());
        }

        @Test
        void email_rejectsPlainString() {
            var s = string().format("email");
            assertFalse(s.safeParse("notanemail").success());
        }

        @Test
        void email_rejectsEmpty() {
            var s = string().format("email");
            assertFalse(s.safeParse("").success());
        }

        @Test
        void email_acceptsValid() {
            var s = string().format("email");
            assertTrue(s.safeParse("user@example.com").success());
            assertTrue(s.safeParse("user.name+tag@example.co.uk").success());
        }

        @Test
        void url_rejectsJavascriptProtocol() {
            var s = string().format("url");
            assertFalse(s.safeParse("javascript:alert(1)").success());
        }

        @Test
        void url_rejectsDataProtocol() {
            var s = string().format("url");
            assertFalse(s.safeParse("data:text/html,<script>alert(1)</script>").success());
        }

        @Test
        void url_rejectsFtpProtocol() {
            var s = string().format("url");
            assertFalse(s.safeParse("ftp://example.com").success());
        }

        @Test
        void url_rejectsPlainString() {
            var s = string().format("url");
            assertFalse(s.safeParse("not a url").success());
        }

        @Test
        void url_rejectsEmpty() {
            var s = string().format("url");
            assertFalse(s.safeParse("").success());
        }

        @Test
        void url_acceptsHttpAndHttps() {
            var s = string().format("url");
            assertTrue(s.safeParse("http://example.com").success());
            assertTrue(s.safeParse("https://example.com/path?q=1").success());
        }

        @Test
        void ipv4_rejectsOctetAbove255() {
            var s = string().format("ipv4");
            assertFalse(s.safeParse("256.1.1.1").success());
            assertFalse(s.safeParse("1.256.1.1").success());
            assertFalse(s.safeParse("1.1.256.1").success());
            assertFalse(s.safeParse("1.1.1.256").success());
        }

        @Test
        void ipv4_rejectsLeadingZeros() {
            assertFalse(FormatValidators.validate("ipv4", "01.01.01.01"));
            assertFalse(FormatValidators.validate("ipv4", "001.1.1.1"));
        }

        @Test
        void ipv4_rejectsIncompleteAddress() {
            var s = string().format("ipv4");
            assertFalse(s.safeParse("1.2.3").success());
            assertFalse(s.safeParse("1.2").success());
            assertFalse(s.safeParse("1").success());
        }

        @Test
        void ipv4_acceptsValidAddresses() {
            var s = string().format("ipv4");
            assertTrue(s.safeParse("0.0.0.0").success());
            assertTrue(s.safeParse("255.255.255.255").success());
            assertTrue(s.safeParse("192.168.1.1").success());
        }

        @Test
        void uuid_rejectsTruncated() {
            var s = string().format("uuid");
            assertFalse(s.safeParse("550e8400-e29b-41d4-a716").success());
        }

        @Test
        void uuid_rejectsPlainString() {
            var s = string().format("uuid");
            assertFalse(s.safeParse("not-a-uuid").success());
        }

        @Test
        void date_rejectsInvalidMonth() {
            var s = string().format("date");
            assertFalse(s.safeParse("2024-13-01").success());
        }

        @Test
        void date_rejectsInvalidDay() {
            var s = string().format("date");
            assertFalse(s.safeParse("2024-02-30").success());
        }

        @Test
        void dateTime_rejectsNoTimezone() {
            var s = string().format("date-time");
            assertFalse(s.safeParse("2024-01-15T10:30:00").success());
        }
    }

    // ---- Unicode length constraints ----
    @Nested
    class UnicodeLengthConstraints {

        @Test
        void astralCodePointCountsAsOneCharacter() {
            var emoji = new String(Character.toChars(0x1F600));
            assertTrue(string().maxLength(1).safeParse(emoji).success());
            assertFalse(string().minLength(2).safeParse(emoji).success());
        }

        @Test
        void importedMaxLengthUsesCodePoints() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "string", "maxLength", 1));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());
            Schema<?> imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(new String(Character.toChars(0x1F600))).success());
        }
    }

    // ---- CWE-400: Large Inputs (Resource Exhaustion) ----
    @Nested
    class CWE_400_LargeInputs {

        @Test
        void largeStringDoesNotCrash() {
            // 1MB string
            String large = "a".repeat(1_000_000);
            long start = System.nanoTime();
            var r = string().safeParse(large);
            long elapsed = System.nanoTime() - start;
            assertTrue(r.success());
            assertTrue(elapsed < 5_000_000_000L, "1MB string took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void largeStringWithMinLength() {
            String large = "a".repeat(1_000_000);
            var s = string().minLength(1);
            var r = s.safeParse(large);
            assertTrue(r.success());
        }

        @Test
        void largeStringWithMaxLengthRejects() {
            String large = "a".repeat(1_000_000);
            var s = string().maxLength(100);
            var r = s.safeParse(large);
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_LARGE, r.issues().get(0).code());
        }

        @Test
        void largeStringWithPattern() {
            String large = "a".repeat(1_000_000);
            var s = string().pattern("^[a-z]+$");
            long start = System.nanoTime();
            var r = s.safeParse(large);
            long elapsed = System.nanoTime() - start;
            assertTrue(r.success());
            assertTrue(elapsed < 5_000_000_000L, "1MB pattern match took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void largeArrayDoesNotCrash() {
            // 10,000 element array
            List<Object> large = new ArrayList<>();
            for (int i = 0; i < 10_000; i++) {
                large.add(i);
            }
            long start = System.nanoTime();
            var r = array(int_()).safeParse(large);
            long elapsed = System.nanoTime() - start;
            assertTrue(r.success());
            assertTrue(elapsed < 5_000_000_000L, "10K array took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void largeArrayWithInvalidElements() {
            // 10,000 element array where all are wrong type
            List<Object> large = new ArrayList<>();
            for (int i = 0; i < 10_000; i++) {
                large.add("not_an_int_" + i);
            }
            long start = System.nanoTime();
            var r = array(int_()).safeParse(large);
            long elapsed = System.nanoTime() - start;
            assertFalse(r.success());
            assertTrue(elapsed < 5_000_000_000L, "10K invalid array took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void deeplyNestedObjectDoesNotCrash() {
            // Build 100-level nested object: {a: {a: {a: ...}}}
            Map<String, Object> inner = Map.of("leaf", "value");
            for (int i = 0; i < 100; i++) {
                inner = Map.of("a", inner);
            }

            // Build matching schema
            Schema<?> innerSchema = object_(Map.of("leaf", string()), null, UnknownKeyMode.ALLOW);
            for (int i = 0; i < 100; i++) {
                innerSchema = object_(Map.of("a", innerSchema), null, UnknownKeyMode.ALLOW);
            }

            long start = System.nanoTime();
            try {
                var r = innerSchema.safeParse(inner);
                assertTrue(r.success());
            } catch (StackOverflowError e) {
                // Acceptable: extremely deep nesting may exhaust stack
            }
            long elapsed = System.nanoTime() - start;
            assertTrue(elapsed < 5_000_000_000L, "Deep nesting took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void largeObjectManyKeysDoesNotCrash() {
            // Object with 1000 keys
            Map<String, Object> data = new LinkedHashMap<>();
            Map<String, Schema<?>> props = new LinkedHashMap<>();
            for (int i = 0; i < 1000; i++) {
                data.put("key" + i, "value" + i);
                props.put("key" + i, string());
            }
            var s = object_(props);
            long start = System.nanoTime();
            var r = s.safeParse(data);
            long elapsed = System.nanoTime() - start;
            assertTrue(r.success());
            assertTrue(elapsed < 5_000_000_000L, "1K-key object took " + elapsed / 1_000_000 + "ms");
        }

        @Test
        void largeRecordDoesNotCrash() {
            Map<String, Object> data = new LinkedHashMap<>();
            for (int i = 0; i < 10_000; i++) {
                data.put("key" + i, i);
            }
            long start = System.nanoTime();
            var r = record(int_()).safeParse(data);
            long elapsed = System.nanoTime() - start;
            assertTrue(r.success());
            assertTrue(elapsed < 5_000_000_000L, "10K record took " + elapsed / 1_000_000 + "ms");
        }
    }

    // ---- CVE-2019-10744: Object Key Safety (__proto__, constructor) ----
    @Nested
    class CVE_2019_10744_ObjectKeySafety {

        @Test
        void protoKeyAsObjectProperty() {
            // __proto__ should be treated as a normal key, not trigger prototype pollution
            var s = object_(Map.of("__proto__", string()), null, UnknownKeyMode.ALLOW);
            var data = new LinkedHashMap<String, Object>();
            data.put("__proto__", "safe_value");
            var r = s.safeParse(data);
            assertTrue(r.success());
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) r.data();
            assertEquals("safe_value", result.get("__proto__"));
        }

        @Test
        void constructorKeyAsObjectProperty() {
            var s = object_(Map.of("constructor", string()), null, UnknownKeyMode.ALLOW);
            var data = new LinkedHashMap<String, Object>();
            data.put("constructor", "safe_value");
            var r = s.safeParse(data);
            assertTrue(r.success());
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) r.data();
            assertEquals("safe_value", result.get("constructor"));
        }

        @Test
        void protoKeyInRecordSchema() {
            var s = record(string());
            var data = new LinkedHashMap<String, Object>();
            data.put("__proto__", "value1");
            data.put("constructor", "value2");
            data.put("toString", "value3");
            var r = s.safeParse(data);
            assertTrue(r.success());
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) r.data();
            assertEquals("value1", result.get("__proto__"));
            assertEquals("value2", result.get("constructor"));
            assertEquals("value3", result.get("toString"));
        }

        @Test
        void protoKeyRejectedAsUnknownKey() {
            // When unknownKeys is REJECT, __proto__ should be rejected like any other unknown key
            var s = object_(Map.of("name", string()), null, UnknownKeyMode.REJECT);
            var data = new LinkedHashMap<String, Object>();
            data.put("name", "Alice");
            data.put("__proto__", "malicious");
            var r = s.safeParse(data);
            assertFalse(r.success());
            assertEquals(IssueCodes.UNKNOWN_KEY, r.issues().get(0).code());
        }

        @Test
        void protoKeyStrippedWhenStripMode() {
            var s = object_(Map.of("name", string()), null, UnknownKeyMode.STRIP);
            var data = new LinkedHashMap<String, Object>();
            data.put("name", "Alice");
            data.put("__proto__", "malicious");
            data.put("constructor", "malicious");
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(data);
            assertEquals(1, result.size());
            assertEquals("Alice", result.get("name"));
            assertNull(result.get("__proto__"));
            assertNull(result.get("constructor"));
        }

        @Test
        void importedSchemaWithProtoKey() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            var props = new LinkedHashMap<String, Object>();
            props.put("__proto__", Map.of("kind", "string"));
            props.put("constructor", Map.of("kind", "int"));
            doc.put("root", Map.of("kind", "object", "properties", props, "required", List.of("__proto__", "constructor")));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            Schema<?> imported = Importer.importSchema(doc);
            var data = new LinkedHashMap<String, Object>();
            data.put("__proto__", "safe");
            data.put("constructor", 42);
            assertTrue(imported.safeParse(data).success());
        }
    }

    // ---- Schema Import Injection ----
    @Nested
    class SchemaImportInjection {

        @Test
        void unknownKindRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "evil_schema"));
            assertThrows(IllegalArgumentException.class, () -> Importer.importSchema(doc));
        }

        @Test
        void emptyKindRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", ""));
            assertThrows(Exception.class, () -> Importer.importSchema(doc));
        }

        @Test
        void missingKindRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of());
            assertThrows(Exception.class, () -> Importer.importSchema(doc));
        }

        @Test
        void nullRootRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", null);
            assertThrows(Exception.class, () -> Importer.importSchema(doc));
        }

        @Test
        void missingRootRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            assertThrows(Exception.class, () -> Importer.importSchema(doc));
        }

        @Test
        void sqlInjectionInKindRejected() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "string'; DROP TABLE users; --"));
            assertThrows(IllegalArgumentException.class, () -> Importer.importSchema(doc));
        }

        @Test
        void invalidSourceTypeRejected() {
            assertThrows(IllegalArgumentException.class, () -> Importer.importSchema(42));
            assertThrows(IllegalArgumentException.class, () -> Importer.importSchema(List.of()));
            assertThrows(IllegalArgumentException.class, () -> Importer.importSchema(true));
        }

        @Test
        void importValidDocumentSucceeds() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "string"));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            Schema<?> imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
            assertFalse(imported.safeParse(42).success());
        }

        @Test
        void importWithExtraFieldsInRoot() {
            // Extra fields in root node should not cause failures
            var root = new LinkedHashMap<String, Object>();
            root.put("kind", "string");
            root.put("unknown_field", "ignored");
            root.put("malicious", Map.of("eval", "code"));

            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", root);
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            Schema<?> imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
        }

        @Test
        void importUnresolvedRefProducesError() {
            var doc = new LinkedHashMap<String, Object>();
            doc.put("anyvaliVersion", "1.0");
            doc.put("schemaVersion", "1");
            doc.put("root", Map.of("kind", "ref", "ref", "#/definitions/DoesNotExist"));
            doc.put("definitions", Map.of());
            doc.put("extensions", Map.of());

            Schema<?> imported = Importer.importSchema(doc);
            var r = imported.safeParse("anything");
            assertFalse(r.success());
        }

        @Test
        void importInvalidJsonStringThrows() {
            assertThrows(Exception.class, () -> Importer.importSchema("{invalid json}"));
            assertThrows(Exception.class, () -> Importer.importSchema(""));
            assertThrows(Exception.class, () -> Importer.importSchema("null"));
        }
    }

    // CWE-20 / spec 3.1: regex anchor newline bypass. Java's "$" matches before
    // a final line terminator, so ^[a-z]+$ would accept "abc\n" (newline/CRLF
    // injection). Anchors are rewritten to absolute (\A/\z) to match JS.
    @Test
    void patternAnchorsRejectNewlineInjection() {
        var s = string().pattern("^[a-z]+$");
        assertTrue(s.safeParse("abc").success());
        assertFalse(s.safeParse("abc\n").success());
        assertFalse(s.safeParse("abc\nEVIL").success());

        var admin = string().pattern("^admin$");
        assertTrue(admin.safeParse("admin").success());
        assertFalse(admin.safeParse("x\nadmin").success());
        assertFalse(admin.safeParse("admin\n").success());
    }
}

package com.anyvali

import com.anyvali.interchange.Importer
import com.anyvali.schemas.*
import kotlinx.serialization.json.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue
import kotlin.test.assertFalse

class SecurityTest {

    // ========================================================================
    // CVE-2016-4055: ReDoS - Catastrophic backtracking patterns
    // ========================================================================

    @Test
    fun `CVE-2016-4055 ReDoS - exponential backtracking completes`() {
        // Classic ReDoS pattern: (a+)+ against "aaa...!" should not hang
        val s = string().pattern("^(a+)+$")
        val malicious = "a".repeat(30) + "!"
        val start = System.nanoTime()
        val result = s.safeParse(malicious)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Failure>(result)
        assertTrue(elapsed < 5_000_000_000L, "ReDoS pattern took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CVE-2016-4055 ReDoS - nested quantifier backtracking completes`() {
        // Another classic ReDoS: (a|a)+ against non-matching input
        val s = string().pattern("^(a|a)+$")
        val malicious = "a".repeat(30) + "b"
        val start = System.nanoTime()
        val result = s.safeParse(malicious)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Failure>(result)
        assertTrue(elapsed < 5_000_000_000L, "Nested quantifier took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CVE-2016-4055 ReDoS - polynomial backtracking with overlapping groups`() {
        val s = string().pattern("^([a-zA-Z0-9]+)*$")
        val malicious = "a".repeat(30) + "!"
        val start = System.nanoTime()
        val result = s.safeParse(malicious)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Failure>(result)
        assertTrue(elapsed < 5_000_000_000L, "Overlapping groups took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CVE-2016-4055 ReDoS - safe pattern still validates correctly`() {
        val s = string().pattern("^[a-z]+$")
        val result = s.safeParse("abcdef")
        assertIs<ParseResult.Success>(result)
    }

    // ========================================================================
    // CVE-2003-1564: Recursive $ref - Self-referencing schemas
    // ========================================================================

    @Test
    fun `CVE-2003-1564 recursive ref - self-referencing schema does not infinite loop`() {
        // Import a schema where "Node" references itself via $ref
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Node" },
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "value": { "kind": "string" },
                        "child": { "kind": "nullable", "schema": { "kind": "ref", "ref": "#/definitions/Node" } }
                    },
                    "required": ["value"],
                    "unknownKeys": "reject"
                }
            },
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        // Validate a simple non-recursive input
        val result = schema.safeParse(
            mapOf("value" to "root", "child" to null),
            definitions
        )
        assertIs<ParseResult.Success>(result)
    }

    @Test
    fun `CVE-2003-1564 recursive ref - nested recursive input validates`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Node" },
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "value": { "kind": "string" },
                        "child": { "kind": "nullable", "schema": { "kind": "ref", "ref": "#/definitions/Node" } }
                    },
                    "required": ["value"],
                    "unknownKeys": "reject"
                }
            },
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        // Two levels of nesting
        val nested = mapOf(
            "value" to "parent",
            "child" to mapOf(
                "value" to "child",
                "child" to null
            )
        )
        val result = schema.safeParse(nested, definitions)
        assertIs<ParseResult.Success>(result)
    }

    @Test
    fun `CVE-2003-1564 recursive ref - invalid nested data rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Node" },
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "value": { "kind": "string" },
                        "child": { "kind": "nullable", "schema": { "kind": "ref", "ref": "#/definitions/Node" } }
                    },
                    "required": ["value"],
                    "unknownKeys": "reject"
                }
            },
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        // child.value is wrong type (number instead of string)
        val nested = mapOf(
            "value" to "parent",
            "child" to mapOf(
                "value" to 42,
                "child" to null
            )
        )
        val result = schema.safeParse(nested, definitions)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CVE-2003-1564 recursive ref - unresolved ref throws`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Missing" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        assertThrows<IllegalStateException> {
            schema.parse("anything", definitions)
        }
    }

    /**
     * CVE-2003-1564 parse-time: pure self-cycle (Self -> ref Self) at parse.
     * The JVM throws StackOverflowError (an Error, not Exception) so we catch
     * Throwable. Wall-clock bound guards against silent loops or hangs.
     */
    @Test
    fun `CVE-2003-1564 recursive ref - pure self-cycle parse terminates`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Self" },
            "definitions": {
                "Self": { "kind": "ref", "ref": "#/definitions/Self" }
            },
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)

        val start = System.nanoTime()
        var threw = false
        var result: ParseResult<*>? = null
        try {
            result = schema.safeParse("anything", definitions)
        } catch (t: Throwable) {
            // StackOverflowError is acceptable defense
            threw = true
        }
        val elapsed = System.nanoTime() - start
        assertTrue(
            elapsed < 5_000_000_000L,
            "Self-cycle parse hung ${elapsed / 1_000_000}ms > 5s (CVE-2003-1564)"
        )
        // Silently succeeding on a pure self-cycle would be a logic bug.
        if (!threw && result != null) {
            assertIs<ParseResult.Failure>(result)
        }
    }

    // ========================================================================
    // CWE-190: Integer overflow - All int width boundaries
    // ========================================================================

    @Test
    fun `CWE-190 int8 overflow - max boundary 127 accepted`() {
        val i = int8()
        val result = i.safeParse(127)
        assertIs<ParseResult.Success>(result)
        assertEquals(127L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int8 overflow - 128 rejected`() {
        val i = int8()
        val result = i.safeParse(128)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int8 underflow - min boundary -128 accepted`() {
        val i = int8()
        val result = i.safeParse(-128)
        assertIs<ParseResult.Success>(result)
        assertEquals(-128L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int8 underflow - -129 rejected`() {
        val i = int8()
        val result = i.safeParse(-129)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int16 overflow - max boundary 32767 accepted`() {
        val i = int16()
        val result = i.safeParse(32767)
        assertIs<ParseResult.Success>(result)
        assertEquals(32767L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int16 overflow - 32768 rejected`() {
        val i = int16()
        val result = i.safeParse(32768)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int16 underflow - min boundary -32768 accepted`() {
        val i = int16()
        val result = i.safeParse(-32768)
        assertIs<ParseResult.Success>(result)
        assertEquals(-32768L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int16 underflow - -32769 rejected`() {
        val i = int16()
        val result = i.safeParse(-32769)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int32 overflow - max boundary 2147483647 accepted`() {
        val i = int32()
        val result = i.safeParse(2147483647)
        assertIs<ParseResult.Success>(result)
        assertEquals(2147483647L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int32 overflow - 2147483648 rejected`() {
        val i = int32()
        val result = i.safeParse(2147483648L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int32 underflow - min boundary -2147483648 accepted`() {
        val i = int32()
        val result = i.safeParse(-2147483648L)
        assertIs<ParseResult.Success>(result)
        assertEquals(-2147483648L, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int32 underflow - -2147483649 rejected`() {
        val i = int32()
        val result = i.safeParse(-2147483649L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int64 overflow - Long MAX_VALUE accepted`() {
        val i = int64()
        val result = i.safeParse(Long.MAX_VALUE)
        assertIs<ParseResult.Success>(result)
        assertEquals(Long.MAX_VALUE, result.getOrThrow())
    }

    @Test
    fun `CWE-190 int64 underflow - Long MIN_VALUE accepted`() {
        val i = int64()
        val result = i.safeParse(Long.MIN_VALUE)
        assertIs<ParseResult.Success>(result)
        assertEquals(Long.MIN_VALUE, result.getOrThrow())
    }

    @Test
    fun `CWE-190 uint8 boundaries - 0 and 255 accepted`() {
        val i = uint8()
        assertIs<ParseResult.Success>(i.safeParse(0))
        assertIs<ParseResult.Success>(i.safeParse(255))
    }

    @Test
    fun `CWE-190 uint8 overflow - 256 rejected`() {
        val i = uint8()
        val result = i.safeParse(256)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 uint8 underflow - -1 rejected`() {
        val i = uint8()
        val result = i.safeParse(-1)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `CWE-190 uint16 boundaries - 0 and 65535 accepted`() {
        val i = uint16()
        assertIs<ParseResult.Success>(i.safeParse(0))
        assertIs<ParseResult.Success>(i.safeParse(65535))
    }

    @Test
    fun `CWE-190 uint16 overflow - 65536 rejected`() {
        val i = uint16()
        val result = i.safeParse(65536)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 uint32 boundaries - 0 and 4294967295 accepted`() {
        val i = uint32()
        assertIs<ParseResult.Success>(i.safeParse(0))
        assertIs<ParseResult.Success>(i.safeParse(4294967295L))
    }

    @Test
    fun `CWE-190 uint32 overflow - 4294967296 rejected`() {
        val i = uint32()
        val result = i.safeParse(4294967296L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 uint64 underflow - -1 rejected`() {
        val i = uint64()
        val result = i.safeParse(-1)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int rejects float as integer`() {
        val i = int_()
        val result = i.safeParse(3.14)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
    }

    @Test
    fun `CWE-190 int8 rejects value way outside range`() {
        val i = int8()
        val result = i.safeParse(100000L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    // ========================================================================
    // CWE-20: NaN/Infinity - Special floating-point values rejected
    // ========================================================================

    @Test
    fun `CWE-20 NaN rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Double.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 positive infinity rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Double.POSITIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 negative infinity rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Double.NEGATIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 NaN rejected by float64 schema`() {
        val f = float64()
        val result = f.safeParse(Double.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 positive infinity rejected by float64 schema`() {
        val f = float64()
        val result = f.safeParse(Double.POSITIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 negative infinity rejected by float64 schema`() {
        val f = float64()
        val result = f.safeParse(Double.NEGATIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 NaN rejected by float32 schema`() {
        val f = float32()
        val result = f.safeParse(Double.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 Float NaN rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Float.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 Float positive infinity rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Float.POSITIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 Float negative infinity rejected by number schema`() {
        val n = number()
        val result = n.safeParse(Float.NEGATIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 NaN rejected by int schema`() {
        val i = int_()
        val result = i.safeParse(Double.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 infinity rejected by int schema`() {
        val i = int_()
        val result = i.safeParse(Double.POSITIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 number with min constraint rejects NaN`() {
        val n = number().min(0)
        val result = n.safeParse(Double.NaN)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 number with max constraint rejects infinity`() {
        val n = number().max(1000)
        val result = n.safeParse(Double.POSITIVE_INFINITY)
        assertIs<ParseResult.Failure>(result)
    }

    // ========================================================================
    // CWE-20: Format bypass - email, url, ipv4 edge cases
    // ========================================================================

    @Test
    fun `CWE-20 email rejects double at sign`() {
        val s = string().format("email")
        val result = s.safeParse("user@@example.com")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 tampered email format name is not silently ignored`() {
        val s = string().format("email\u0000")
        val result = s.safeParse("not-an-email")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 imported tampered email format name is not unconstrained`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string", "format": "email\u0000" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        val result = schema.safeParse("not-an-email", definitions)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 email rejects empty local part`() {
        val s = string().format("email")
        val result = s.safeParse("@example.com")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 email rejects missing domain`() {
        val s = string().format("email")
        val result = s.safeParse("user@")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 email rejects empty string`() {
        val s = string().format("email")
        val result = s.safeParse("")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 email rejects no TLD`() {
        val s = string().format("email")
        val result = s.safeParse("user@localhost")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 email accepts valid complex email`() {
        val s = string().format("email")
        val result = s.safeParse("user.name+tag@sub.domain.com")
        assertIs<ParseResult.Success>(result)
    }

    @Test
    fun `CWE-20 url rejects javascript protocol`() {
        val s = string().format("url")
        val result = s.safeParse("javascript:alert(1)")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 url rejects data URI`() {
        val s = string().format("url")
        val result = s.safeParse("data:text/html,<script>alert(1)</script>")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 url rejects ftp protocol`() {
        val s = string().format("url")
        val result = s.safeParse("ftp://files.example.com/file.txt")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 url rejects empty string`() {
        val s = string().format("url")
        val result = s.safeParse("")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 url accepts valid https url`() {
        val s = string().format("url")
        val result = s.safeParse("https://example.com/path?q=1&r=2#frag")
        assertIs<ParseResult.Success>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects octet above 255`() {
        val s = string().format("ipv4")
        val result = s.safeParse("256.1.1.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects leading zeros`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.01.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects too few octets`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects too many octets`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.1.1.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects negative octet`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.-1.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 rejects non-numeric octet`() {
        val s = string().format("ipv4")
        val result = s.safeParse("192.168.abc.1")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `CWE-20 ipv4 accepts boundary values`() {
        val s = string().format("ipv4")
        assertIs<ParseResult.Success>(s.safeParse("0.0.0.0"))
        assertIs<ParseResult.Success>(s.safeParse("255.255.255.255"))
    }

    // ========================================================================
    // Unicode length constraints - code points, not UTF-16 code units
    // ========================================================================

    @Test
    fun `unicode length astral code point counts as one character`() {
        val emoji = String(Character.toChars(0x1F600))
        assertIs<ParseResult.Success>(string().maxLength(1).safeParse(emoji))
        assertIs<ParseResult.Failure>(string().minLength(2).safeParse(emoji))
    }

    @Test
    fun `unicode length imported maxLength uses code points`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string", "maxLength": 1 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, definitions) = Importer.importFromJson(jsonStr)
        val result = schema.safeParse(String(Character.toChars(0x1F600)), definitions)
        assertIs<ParseResult.Success>(result)
    }

    // ========================================================================
    // CWE-400: Large inputs - Resource exhaustion
    // ========================================================================

    @Test
    fun `CWE-400 large string input validates within time limit`() {
        val s = string()
        val largeStr = "a".repeat(1_000_000)
        val start = System.nanoTime()
        val result = s.safeParse(largeStr)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Success>(result)
        assertTrue(elapsed < 5_000_000_000L, "Large string took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CWE-400 large string rejected by maxLength`() {
        val s = string().maxLength(100)
        val largeStr = "a".repeat(10_000)
        val result = s.safeParse(largeStr)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-400 large array validates within time limit`() {
        val a = array(int_())
        val largeArray = (1..10_000).toList()
        val start = System.nanoTime()
        val result = a.safeParse(largeArray)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Success>(result)
        assertTrue(elapsed < 5_000_000_000L, "Large array took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CWE-400 large array rejected by maxItems`() {
        val a = array(string()).maxItems(10)
        val largeArray = List(1000) { "item$it" }
        val result = a.safeParse(largeArray)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `CWE-400 deeply nested object validates within time limit`() {
        // Create a deeply nested map: { "a": { "a": { ... } } }
        var nested: Any? = "leaf"
        for (i in 1..100) {
            nested = mapOf("value" to nested)
        }
        val s = record(any_())
        val start = System.nanoTime()
        val result = s.safeParse(nested)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Success>(result)
        assertTrue(elapsed < 5_000_000_000L, "Deeply nested object took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CWE-400 large object with many keys validates within time limit`() {
        val properties = mutableMapOf<String, Schema<*>>()
        val input = mutableMapOf<String, Any>()
        for (i in 1..1000) {
            properties["field$i"] = string()
            input["field$i"] = "value$i"
        }
        val o = obj(properties = properties, required = properties.keys, unknownKeys = UnknownKeyMode.REJECT)
        val start = System.nanoTime()
        val result = o.safeParse(input)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Success>(result)
        assertTrue(elapsed < 5_000_000_000L, "Large object took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    @Test
    fun `CWE-400 large string with pattern validates within time limit`() {
        val s = string().pattern("^[a-z]+$")
        val largeStr = "a".repeat(100_000)
        val start = System.nanoTime()
        val result = s.safeParse(largeStr)
        val elapsed = System.nanoTime() - start
        assertIs<ParseResult.Success>(result)
        assertTrue(elapsed < 5_000_000_000L, "Pattern on large string took ${elapsed / 1_000_000}ms, expected < 5000ms")
    }

    // ========================================================================
    // Schema import injection - Unknown kinds rejected
    // ========================================================================

    @Test
    fun `schema import injection - unknown kind rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "evil_injection" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val ex = assertThrows<ValidationError> {
            Importer.importFromJson(jsonStr)
        }
        assertTrue(ex.issues.any { it.code == IssueCodes.UNSUPPORTED_SCHEMA_KIND })
    }

    @Test
    fun `schema import injection - empty kind rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val ex = assertThrows<ValidationError> {
            Importer.importFromJson(jsonStr)
        }
        assertTrue(ex.issues.any { it.code == IssueCodes.UNSUPPORTED_SCHEMA_KIND })
    }

    @Test
    fun `schema import injection - missing kind throws`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "minLength": 5 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        assertThrows<IllegalArgumentException> {
            Importer.importFromJson(jsonStr)
        }
    }

    @Test
    fun `schema import injection - SQL injection string as kind rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "'; DROP TABLE schemas; --" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val ex = assertThrows<ValidationError> {
            Importer.importFromJson(jsonStr)
        }
        assertTrue(ex.issues.any { it.code == IssueCodes.UNSUPPORTED_SCHEMA_KIND })
    }

    @Test
    fun `schema import injection - prototype pollution kind rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "__proto__" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val ex = assertThrows<ValidationError> {
            Importer.importFromJson(jsonStr)
        }
        assertTrue(ex.issues.any { it.code == IssueCodes.UNSUPPORTED_SCHEMA_KIND })
    }

    @Test
    fun `schema import injection - nested unknown kind in definition rejected`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "ref", "ref": "#/definitions/Bad" },
            "definitions": {
                "Bad": { "kind": "malicious_schema" }
            },
            "extensions": {}
        }
        """.trimIndent()
        val ex = assertThrows<ValidationError> {
            Importer.importFromJson(jsonStr)
        }
        assertTrue(ex.issues.any { it.code == IssueCodes.UNSUPPORTED_SCHEMA_KIND })
    }

    @Test
    fun `schema import injection - valid kinds still import correctly`() {
        val validKinds = listOf("string", "number", "int", "bool", "null", "any", "unknown", "never")
        for (kind in validKinds) {
            val jsonStr = """
            {
                "anyvaliVersion": "1.0",
                "schemaVersion": "1",
                "root": { "kind": "$kind" },
                "definitions": {},
                "extensions": {}
            }
            """.trimIndent()
            val (schema, _) = Importer.importFromJson(jsonStr)
            assertEquals(kind, schema.kind, "Expected kind '$kind' to import successfully")
        }
    }
}

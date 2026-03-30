package com.anyvali

import com.anyvali.schemas.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull
import kotlin.test.assertTrue

class SchemaTest {

    // ---- String Schema ----

    @Test
    fun `string accepts a simple string`() {
        val s = string()
        assertEquals("hello", s.parse("hello"))
    }

    @Test
    fun `string accepts an empty string`() {
        val s = string()
        assertEquals("", s.parse(""))
    }

    @Test
    fun `string accepts unicode`() {
        val s = string()
        assertEquals("\u00e9\u00e0\u00fc\u00f1\u00f6", s.parse("\u00e9\u00e0\u00fc\u00f1\u00f6"))
    }

    @Test
    fun `string rejects a number`() {
        val s = string()
        val result = s.safeParse(42)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("string", result.issues[0].expected)
        assertEquals("number", result.issues[0].received)
    }

    @Test
    fun `string rejects a boolean`() {
        val s = string()
        val result = s.safeParse(true)
        assertIs<ParseResult.Failure>(result)
        assertEquals("boolean", result.issues[0].received)
    }

    @Test
    fun `string rejects null`() {
        val s = string()
        val result = s.safeParse(null)
        assertIs<ParseResult.Failure>(result)
        assertEquals("null", result.issues[0].received)
    }

    @Test
    fun `string rejects an array`() {
        val s = string()
        val result = s.safeParse(listOf("a", "b"))
        assertIs<ParseResult.Failure>(result)
        assertEquals("array", result.issues[0].received)
    }

    @Test
    fun `string rejects an object`() {
        val s = string()
        val result = s.safeParse(mapOf("key" to "value"))
        assertIs<ParseResult.Failure>(result)
        assertEquals("object", result.issues[0].received)
    }

    @Test
    fun `string parse throws on invalid input`() {
        val s = string()
        assertThrows<ValidationError> { s.parse(42) }
    }

    // ---- String constraints ----

    @Test
    fun `minLength passes when long enough`() {
        val s = string().minLength(3)
        assertEquals("abc", s.parse("abc"))
    }

    @Test
    fun `minLength fails when too short`() {
        val s = string().minLength(3)
        val result = s.safeParse("ab")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
        assertEquals("3", result.issues[0].expected)
        assertEquals("2", result.issues[0].received)
    }

    @Test
    fun `maxLength passes when short enough`() {
        val s = string().maxLength(5)
        assertEquals("hello", s.parse("hello"))
    }

    @Test
    fun `maxLength fails when too long`() {
        val s = string().maxLength(5)
        val result = s.safeParse("hello!")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `pattern passes when matches`() {
        val s = string().pattern("^[a-z]+$")
        assertEquals("abc", s.parse("abc"))
    }

    @Test
    fun `pattern fails when no match`() {
        val s = string().pattern("^[a-z]+$")
        val result = s.safeParse("ABC")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_STRING, result.issues[0].code)
    }

    @Test
    fun `startsWith passes`() {
        val s = string().startsWith("hello")
        assertEquals("hello world", s.parse("hello world"))
    }

    @Test
    fun `startsWith fails`() {
        val s = string().startsWith("hello")
        val result = s.safeParse("world hello")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_STRING, result.issues[0].code)
    }

    @Test
    fun `endsWith passes`() {
        val s = string().endsWith(".json")
        assertEquals("file.json", s.parse("file.json"))
    }

    @Test
    fun `endsWith fails`() {
        val s = string().endsWith(".json")
        val result = s.safeParse("file.xml")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `includes passes`() {
        val s = string().includes("world")
        assertEquals("hello world!", s.parse("hello world!"))
    }

    @Test
    fun `includes fails`() {
        val s = string().includes("world")
        val result = s.safeParse("hello there")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Number Schema ----

    @Test
    fun `number accepts a positive integer`() {
        val n = number()
        assertEquals(42.0, n.parse(42))
    }

    @Test
    fun `number accepts zero`() {
        val n = number()
        assertEquals(0.0, n.parse(0))
    }

    @Test
    fun `number accepts negative float`() {
        val n = number()
        assertEquals(-3.14, n.parse(-3.14))
    }

    @Test
    fun `number accepts very large value`() {
        val n = number()
        assertEquals(1.7976931348623157e+308, n.parse(1.7976931348623157e+308))
    }

    @Test
    fun `number rejects a string`() {
        val n = number()
        val result = n.safeParse("42")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("number", result.issues[0].expected)
        assertEquals("string", result.issues[0].received)
    }

    @Test
    fun `number rejects a boolean`() {
        val n = number()
        val result = n.safeParse(true)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `number rejects null`() {
        val n = number()
        val result = n.safeParse(null)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `number rejects an object`() {
        val n = number()
        val result = n.safeParse(emptyMap<String, Any>())
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `number rejects an array`() {
        val n = number()
        val result = n.safeParse(listOf(1, 2, 3))
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Numeric constraints ----

    @Test
    fun `min passes when equal`() {
        val n = number().min(10)
        assertEquals(10.0, n.parse(10))
    }

    @Test
    fun `min fails when below`() {
        val n = number().min(10)
        val result = n.safeParse(9)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `max passes when equal`() {
        val n = number().max(100)
        assertEquals(100.0, n.parse(100))
    }

    @Test
    fun `max fails when above`() {
        val n = number().max(100)
        val result = n.safeParse(101)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `exclusiveMin passes when above`() {
        val n = number().exclusiveMin(0)
        assertEquals(0.001, n.parse(0.001))
    }

    @Test
    fun `exclusiveMin fails when equal`() {
        val n = number().exclusiveMin(0)
        val result = n.safeParse(0)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `exclusiveMax passes when below`() {
        val n = number().exclusiveMax(100)
        assertEquals(99.999, n.parse(99.999))
    }

    @Test
    fun `exclusiveMax fails when equal`() {
        val n = number().exclusiveMax(100)
        val result = n.safeParse(100)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `multipleOf passes when multiple`() {
        val n = number().multipleOf(3)
        assertEquals(9.0, n.parse(9))
    }

    @Test
    fun `multipleOf fails when not multiple`() {
        val n = number().multipleOf(3)
        val result = n.safeParse(10)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_NUMBER, result.issues[0].code)
    }

    @Test
    fun `multipleOf works with float divisor`() {
        val n = number().multipleOf(0.5)
        assertEquals(2.5, n.parse(2.5))
    }

    // ---- Int Schema ----

    @Test
    fun `int accepts positive integer`() {
        val i = int_()
        assertEquals(42L, i.parse(42))
    }

    @Test
    fun `int accepts zero`() {
        val i = int_()
        assertEquals(0L, i.parse(0))
    }

    @Test
    fun `int accepts negative integer`() {
        val i = int_()
        assertEquals(-100L, i.parse(-100))
    }

    @Test
    fun `int rejects float`() {
        val i = int_()
        val result = i.safeParse(3.14)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("int", result.issues[0].expected)
    }

    @Test
    fun `int rejects string`() {
        val i = int_()
        val result = i.safeParse("42")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `int with min and max`() {
        val i = int_().min(1).max(10)
        assertEquals(5L, i.parse(5))
    }

    // ---- Int Widths ----

    @Test
    fun `int8 accepts value in range`() {
        val i = int8()
        assertEquals(127L, i.parse(127))
    }

    @Test
    fun `int8 accepts minimum -128`() {
        val i = int8()
        assertEquals(-128L, i.parse(-128))
    }

    @Test
    fun `int8 rejects above range`() {
        val i = int8()
        val result = i.safeParse(128)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
        assertEquals("int8", result.issues[0].expected)
    }

    @Test
    fun `int8 rejects below range`() {
        val i = int8()
        val result = i.safeParse(-129)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `int16 accepts in range`() {
        val i = int16()
        assertEquals(32767L, i.parse(32767))
    }

    @Test
    fun `int16 rejects above range`() {
        val i = int16()
        val result = i.safeParse(32768)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `int32 accepts max value`() {
        val i = int32()
        assertEquals(2147483647L, i.parse(2147483647))
    }

    @Test
    fun `int32 rejects above range`() {
        val i = int32()
        val result = i.safeParse(2147483648L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `uint8 accepts 0`() {
        val i = uint8()
        assertEquals(0L, i.parse(0))
    }

    @Test
    fun `uint8 accepts 255`() {
        val i = uint8()
        assertEquals(255L, i.parse(255))
    }

    @Test
    fun `uint8 rejects negative`() {
        val i = uint8()
        val result = i.safeParse(-1)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `uint8 rejects 256`() {
        val i = uint8()
        val result = i.safeParse(256)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `uint16 accepts 65535`() {
        val i = uint16()
        assertEquals(65535L, i.parse(65535))
    }

    @Test
    fun `uint32 accepts 4294967295`() {
        val i = uint32()
        assertEquals(4294967295L, i.parse(4294967295L))
    }

    @Test
    fun `uint64 rejects negative`() {
        val i = uint64()
        val result = i.safeParse(-1)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    // ---- Float widths ----

    @Test
    fun `float64 accepts normal float`() {
        val f = float64()
        assertEquals(3.141592653589793, f.parse(3.141592653589793))
    }

    @Test
    fun `float64 accepts integer as float`() {
        val f = float64()
        assertEquals(42.0, f.parse(42))
    }

    @Test
    fun `float64 rejects string`() {
        val f = float64()
        val result = f.safeParse("3.14")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `float32 accepts value in range`() {
        val f = float32()
        assertEquals(1.5, f.parse(1.5))
    }

    @Test
    fun `float32 rejects boolean`() {
        val f = float32()
        val result = f.safeParse(true)
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Bool Schema ----

    @Test
    fun `bool accepts true`() {
        val b = bool()
        assertEquals(true, b.parse(true))
    }

    @Test
    fun `bool accepts false`() {
        val b = bool()
        assertEquals(false, b.parse(false))
    }

    @Test
    fun `bool rejects number`() {
        val b = bool()
        val result = b.safeParse(1)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("bool", result.issues[0].expected)
        assertEquals("number", result.issues[0].received)
    }

    @Test
    fun `bool rejects string`() {
        val b = bool()
        val result = b.safeParse("true")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `bool rejects null`() {
        val b = bool()
        val result = b.safeParse(null)
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Null Schema ----

    @Test
    fun `null accepts null`() {
        val n = null_()
        assertNull(n.parse(null))
    }

    @Test
    fun `null rejects string`() {
        val n = null_()
        val result = n.safeParse("null")
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `null rejects zero`() {
        val n = null_()
        val result = n.safeParse(0)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `null rejects false`() {
        val n = null_()
        val result = n.safeParse(false)
        assertIs<ParseResult.Failure>(result)
    }

    @Test
    fun `null rejects empty string`() {
        val n = null_()
        val result = n.safeParse("")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Any Schema ----

    @Test
    fun `any accepts string`() {
        val a = any_()
        assertEquals("hello", a.parse("hello"))
    }

    @Test
    fun `any accepts number`() {
        val a = any_()
        assertEquals(42, a.parse(42))
    }

    @Test
    fun `any accepts null`() {
        val a = any_()
        assertNull(a.parse(null))
    }

    @Test
    fun `any accepts object`() {
        val a = any_()
        val input = mapOf("key" to "value")
        assertEquals(input, a.parse(input))
    }

    @Test
    fun `any accepts array`() {
        val a = any_()
        val input = listOf(1, "two", true)
        assertEquals(input, a.parse(input))
    }

    // ---- Unknown Schema ----

    @Test
    fun `unknown accepts string`() {
        val u = unknown()
        assertEquals("hello", u.parse("hello"))
    }

    @Test
    fun `unknown accepts number`() {
        val u = unknown()
        assertEquals(99, u.parse(99))
    }

    @Test
    fun `unknown accepts null`() {
        val u = unknown()
        assertNull(u.parse(null))
    }

    @Test
    fun `unknown accepts boolean`() {
        val u = unknown()
        assertEquals(false, u.parse(false))
    }

    // ---- Never Schema ----

    @Test
    fun `never rejects string`() {
        val n = never()
        val result = n.safeParse("hello")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("never", result.issues[0].expected)
        assertEquals("string", result.issues[0].received)
    }

    @Test
    fun `never rejects number`() {
        val n = never()
        val result = n.safeParse(0)
        assertIs<ParseResult.Failure>(result)
        assertEquals("number", result.issues[0].received)
    }

    @Test
    fun `never rejects null`() {
        val n = never()
        val result = n.safeParse(null)
        assertIs<ParseResult.Failure>(result)
        assertEquals("null", result.issues[0].received)
    }

    @Test
    fun `never rejects boolean`() {
        val n = never()
        val result = n.safeParse(true)
        assertIs<ParseResult.Failure>(result)
        assertEquals("boolean", result.issues[0].received)
    }

    @Test
    fun `never rejects empty object`() {
        val n = never()
        val result = n.safeParse(emptyMap<String, Any>())
        assertIs<ParseResult.Failure>(result)
        assertEquals("object", result.issues[0].received)
    }

    // ---- Literal Schema ----

    @Test
    fun `literal accepts matching string`() {
        val l = literal("hello")
        assertEquals("hello", l.parse("hello"))
    }

    @Test
    fun `literal rejects non-matching string`() {
        val l = literal("hello")
        val result = l.safeParse("world")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_LITERAL, result.issues[0].code)
    }

    @Test
    fun `literal accepts matching number`() {
        val l = literal(42)
        assertEquals(42, l.parse(42))
    }

    @Test
    fun `literal accepts matching boolean`() {
        val l = literal(true)
        assertEquals(true, l.parse(true))
    }

    @Test
    fun `literal rejects wrong type`() {
        val l = literal(42)
        val result = l.safeParse("42")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_LITERAL, result.issues[0].code)
    }

    @Test
    fun `literal accepts null literal`() {
        val l = literal(null)
        assertNull(l.parse(null))
    }

    // ---- Enum Schema ----

    @Test
    fun `enum accepts value in list`() {
        val e = enum_("red", "green", "blue")
        assertEquals("red", e.parse("red"))
    }

    @Test
    fun `enum accepts another value in list`() {
        val e = enum_("red", "green", "blue")
        assertEquals("blue", e.parse("blue"))
    }

    @Test
    fun `enum rejects value not in list`() {
        val e = enum_("red", "green", "blue")
        val result = e.safeParse("yellow")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertTrue(result.issues[0].expected.contains("enum"))
    }

    @Test
    fun `enum accepts numeric values`() {
        val e = enum_(1L, 2L, 3L)
        assertEquals(2L, e.parse(2L))
    }

    @Test
    fun `enum rejects wrong type`() {
        val e = enum_(1L, 2L, 3L)
        val result = e.safeParse("1")
        assertIs<ParseResult.Failure>(result)
    }

    // ---- Array Schema ----

    @Test
    fun `array accepts valid elements`() {
        val a = array(string())
        assertEquals(listOf("a", "b", "c"), a.parse(listOf("a", "b", "c")))
    }

    @Test
    fun `array accepts empty array`() {
        val a = array(int_())
        assertEquals(emptyList<Any>(), a.parse(emptyList<Any>()))
    }

    @Test
    fun `array rejects non-array`() {
        val a = array(string())
        val result = a.safeParse("not an array")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("array", result.issues[0].expected)
    }

    @Test
    fun `array rejects invalid element`() {
        val a = array(int_())
        val result = a.safeParse(listOf(1, 2, "three"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals(listOf(2), result.issues[0].path)
    }

    @Test
    fun `array reports multiple invalid elements`() {
        val a = array(bool())
        val result = a.safeParse(listOf(true, "yes", false, 1))
        assertIs<ParseResult.Failure>(result)
        assertEquals(2, result.issues.size)
        assertEquals(listOf(1), result.issues[0].path)
        assertEquals(listOf(3), result.issues[1].path)
    }

    // ---- Array constraints ----

    @Test
    fun `minItems passes`() {
        val a = array(int_()).minItems(2)
        assertEquals(listOf(1L, 2L), a.parse(listOf(1, 2)))
    }

    @Test
    fun `minItems fails`() {
        val a = array(int_()).minItems(2)
        val result = a.safeParse(listOf(1))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `maxItems passes`() {
        val a = array(string()).maxItems(3)
        assertEquals(listOf("a", "b", "c"), a.parse(listOf("a", "b", "c")))
    }

    @Test
    fun `maxItems fails`() {
        val a = array(string()).maxItems(3)
        val result = a.safeParse(listOf("a", "b", "c", "d"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    // ---- Tuple Schema ----

    @Test
    fun `tuple accepts valid`() {
        val t = tuple(string(), int_())
        val result = t.parse(listOf("hello", 42)) as List<*>
        assertEquals("hello", result[0])
        assertEquals(42L, result[1])
    }

    @Test
    fun `tuple rejects too few elements`() {
        val t = tuple(string(), int_())
        val result = t.safeParse(listOf("hello"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    @Test
    fun `tuple rejects too many elements`() {
        val t = tuple(string(), int_())
        val result = t.safeParse(listOf("hello", 42, true))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `tuple rejects wrong element type`() {
        val t = tuple(string(), int_())
        val result = t.safeParse(listOf(42, "hello"))
        assertIs<ParseResult.Failure>(result)
        assertTrue(result.issues.size >= 2)
    }

    @Test
    fun `tuple rejects non-array`() {
        val t = tuple(string())
        val result = t.safeParse("not a tuple")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("tuple", result.issues[0].expected)
    }

    // ---- Object Schema ----

    @Test
    fun `object accepts with all required fields`() {
        val o = obj(
            properties = mapOf("name" to string(), "age" to int_()),
            required = setOf("name", "age")
        )
        val result = o.parse(mapOf("name" to "Alice", "age" to 30)) as Map<*, *>
        assertEquals("Alice", result["name"])
        assertEquals(30L, result["age"])
    }

    @Test
    fun `object rejects missing required field`() {
        val o = obj(
            properties = mapOf("name" to string(), "age" to int_()),
            required = setOf("name", "age")
        )
        val result = o.safeParse(mapOf("name" to "Alice"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.REQUIRED, result.issues[0].code)
        assertEquals(listOf("age"), result.issues[0].path)
    }

    @Test
    fun `object rejects all missing required fields`() {
        val o = obj(
            properties = mapOf("name" to string(), "age" to int_()),
            required = setOf("name", "age")
        )
        val result = o.safeParse(emptyMap<String, Any>())
        assertIs<ParseResult.Failure>(result)
        assertEquals(2, result.issues.size)
    }

    @Test
    fun `object accepts optional field absent`() {
        val o = obj(
            properties = mapOf("name" to string(), "nickname" to string()),
            required = setOf("name")
        )
        val result = o.parse(mapOf("name" to "Alice")) as Map<*, *>
        assertEquals("Alice", result["name"])
        assertTrue(!result.containsKey("nickname"))
    }

    @Test
    fun `object rejects non-object input`() {
        val o = obj(
            properties = mapOf("name" to string()),
            required = setOf("name")
        )
        val result = o.safeParse("not an object")
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("object", result.issues[0].expected)
    }

    // ---- Unknown keys ----

    @Test
    fun `reject mode rejects unknown keys by default`() {
        val o = obj(
            properties = mapOf("name" to string()),
            required = setOf("name")
        )
        val result = o.safeParse(mapOf("name" to "Alice", "extra" to "value"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.UNKNOWN_KEY, result.issues[0].code)
        assertEquals(listOf("extra"), result.issues[0].path)
    }

    @Test
    fun `strip mode removes unknown keys`() {
        val o = obj(
            properties = mapOf("name" to string()),
            required = setOf("name"),
            unknownKeys = UnknownKeyMode.STRIP
        )
        val result = o.parse(mapOf("name" to "Alice", "extra" to "value", "another" to 42)) as Map<*, *>
        assertEquals(mapOf("name" to "Alice"), result)
    }

    @Test
    fun `allow mode passes unknown keys`() {
        val o = obj(
            properties = mapOf("name" to string()),
            required = setOf("name"),
            unknownKeys = UnknownKeyMode.ALLOW
        )
        val result = o.parse(mapOf("name" to "Alice", "extra" to "value")) as Map<*, *>
        assertEquals("Alice", result["name"])
        assertEquals("value", result["extra"])
    }

    @Test
    fun `reject mode reports multiple unknown keys`() {
        val o = obj(
            properties = mapOf("id" to int_()),
            required = setOf("id")
        )
        val result = o.safeParse(mapOf("id" to 1, "foo" to "bar", "baz" to true))
        assertIs<ParseResult.Failure>(result)
        assertEquals(2, result.issues.size)
        assertTrue(result.issues.all { it.code == IssueCodes.UNKNOWN_KEY })
    }

    // ---- Record Schema ----

    @Test
    fun `record accepts valid values`() {
        val r = record(int_())
        val result = r.parse(mapOf("a" to 1, "b" to 2, "c" to 3)) as Map<*, *>
        assertEquals(1L, result["a"])
        assertEquals(2L, result["b"])
        assertEquals(3L, result["c"])
    }

    @Test
    fun `record accepts empty`() {
        val r = record(string())
        assertEquals(emptyMap<String, Any>(), r.parse(emptyMap<String, Any>()))
    }

    @Test
    fun `record rejects invalid value`() {
        val r = record(int_())
        val result = r.safeParse(mapOf("a" to 1, "b" to "two"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals(listOf("b"), result.issues[0].path)
    }

    @Test
    fun `record rejects non-object`() {
        val r = record(string())
        val result = r.safeParse(listOf(1, 2, 3))
        assertIs<ParseResult.Failure>(result)
        assertEquals("record", result.issues[0].expected)
    }

    // ---- Union Schema ----

    @Test
    fun `union accepts first variant`() {
        val u = union(string(), int_())
        assertEquals("hello", u.parse("hello"))
    }

    @Test
    fun `union accepts second variant`() {
        val u = union(string(), int_())
        assertEquals(42L, u.parse(42))
    }

    @Test
    fun `union rejects no match`() {
        val u = union(string(), int_())
        val result = u.safeParse(true)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_UNION, result.issues[0].code)
        assertEquals("string | int", result.issues[0].expected)
        assertEquals("boolean", result.issues[0].received)
    }

    @Test
    fun `union first matching wins`() {
        val u = union(number(), int_())
        val result = u.parse(5)
        assertEquals(5, result)
    }

    @Test
    fun `union with null variant`() {
        val u = union(string(), null_())
        assertNull(u.parse(null))
    }

    // ---- Intersection Schema ----

    @Test
    fun `intersection accepts value satisfying all`() {
        val i = intersection(
            obj(mapOf("name" to string()), setOf("name"), UnknownKeyMode.ALLOW),
            obj(mapOf("age" to int_()), setOf("age"), UnknownKeyMode.ALLOW)
        )
        val result = i.parse(mapOf("name" to "Alice", "age" to 30)) as Map<*, *>
        assertEquals("Alice", result["name"])
        assertEquals(30L, result["age"])
    }

    @Test
    fun `intersection rejects missing field`() {
        val i = intersection(
            obj(mapOf("name" to string()), setOf("name"), UnknownKeyMode.ALLOW),
            obj(mapOf("age" to int_()), setOf("age"), UnknownKeyMode.ALLOW)
        )
        val result = i.safeParse(mapOf("name" to "Alice"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.REQUIRED, result.issues[0].code)
    }

    @Test
    fun `intersection numeric ranges`() {
        val i = intersection(
            number().min(0),
            number().max(100)
        )
        assertEquals(50, i.parse(50))
    }

    @Test
    fun `intersection rejects out of range`() {
        val i = intersection(
            number().min(0),
            number().max(100)
        )
        val result = i.safeParse(-5)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_SMALL, result.issues[0].code)
    }

    // ---- Optional Schema ----

    @Test
    fun `optional accepts present valid value`() {
        val o = obj(
            properties = mapOf("name" to optional(string())),
            required = emptySet()
        )
        val result = o.parse(mapOf("name" to "Alice")) as Map<*, *>
        assertEquals("Alice", result["name"])
    }

    @Test
    fun `optional accepts absent field`() {
        val o = obj(
            properties = mapOf("name" to optional(string())),
            required = emptySet()
        )
        val result = o.parse(emptyMap<String, Any>()) as Map<*, *>
        assertTrue(!result.containsKey("name"))
    }

    @Test
    fun `optional rejects present but invalid`() {
        val o = obj(
            properties = mapOf("name" to optional(string())),
            required = emptySet()
        )
        val result = o.safeParse(mapOf("name" to 123))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals(listOf("name"), result.issues[0].path)
    }

    @Test
    fun `optional null is not treated as absent`() {
        val o = obj(
            properties = mapOf("name" to optional(string())),
            required = emptySet()
        )
        val result = o.safeParse(mapOf("name" to null))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
        assertEquals("null", result.issues[0].received)
    }

    // ---- Nullable Schema ----

    @Test
    fun `nullable accepts null`() {
        val n = nullable(string())
        assertNull(n.parse(null))
    }

    @Test
    fun `nullable accepts valid non-null`() {
        val n = nullable(string())
        assertEquals("hello", n.parse("hello"))
    }

    @Test
    fun `nullable rejects invalid non-null`() {
        val n = nullable(string())
        val result = n.safeParse(42)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
    }

    @Test
    fun `nullable int accepts null`() {
        val n = nullable(int_())
        assertNull(n.parse(null))
    }

    @Test
    fun `nullable int accepts valid int`() {
        val n = nullable(int_())
        assertEquals(99L, n.parse(99))
    }

    // ---- Numeric safety ----

    @Test
    fun `number roundtrips as float64`() {
        val n = number()
        assertEquals(1.7976931348623157e+308, n.parse(1.7976931348623157e+308))
    }

    @Test
    fun `int roundtrips as int64`() {
        val i = int_()
        assertEquals(9007199254740991L, i.parse(9007199254740991L))
    }

    @Test
    fun `int rejects non-integer number`() {
        val i = int_()
        val result = i.safeParse(3.14)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.INVALID_TYPE, result.issues[0].code)
    }

    @Test
    fun `int32 rejects value fitting int64 but not int32`() {
        val i = int32()
        val result = i.safeParse(2147483648L)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    @Test
    fun `float64 and number semantically identical`() {
        val f = float64()
        assertEquals(42.5, f.parse(42.5))
    }

    @Test
    fun `int64 and int semantically identical`() {
        val i = int64()
        assertEquals(42L, i.parse(42))
    }

    @Test
    fun `narrowing int64 to int8 rejected`() {
        val i = int8()
        val result = i.safeParse(200)
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.TOO_LARGE, result.issues[0].code)
    }

    // ---- ParseResult helpers ----

    @Test
    fun `ParseResult Success properties`() {
        val result = ParseResult.Success("hello")
        assertTrue(result.isSuccess)
        assertTrue(!result.isFailure)
        assertEquals("hello", result.getOrNull())
        assertEquals("hello", result.getOrThrow())
        assertEquals(emptyList<ValidationIssue>(), result.issuesOrEmpty())
    }

    @Test
    fun `ParseResult Failure properties`() {
        val issue = ValidationIssue(code = IssueCodes.INVALID_TYPE)
        val result = ParseResult.Failure(listOf(issue))
        assertTrue(result.isFailure)
        assertTrue(!result.isSuccess)
        assertNull(result.getOrNull())
        assertThrows<ValidationError> { result.getOrThrow() }
        assertEquals(1, result.issuesOrEmpty().size)
    }

    // ---- Export ----

    @Test
    fun `schema exports to document`() {
        val s = string().minLength(1)
        val doc = s.export()
        assertEquals("1.0", doc.anyvaliVersion)
        assertEquals("1", doc.schemaVersion)
    }

    @Test
    fun `ValidationError message includes issues`() {
        val error = ValidationError(listOf(ValidationIssue(code = "test_code")))
        assertTrue(error.message!!.contains("test_code"))
    }
}

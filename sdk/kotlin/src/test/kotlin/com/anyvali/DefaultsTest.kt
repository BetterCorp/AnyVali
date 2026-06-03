package com.anyvali

import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull

class DefaultsTest {

    @Test
    fun `missing field gets default`() {
        val schema = obj(
            mapOf(
                "name" to string(),
                "role" to string().default("user"),
            ),
            required = setOf("name"),
        )

        val result = schema.safeParse(mapOf("name" to "Alice"))

        assertIs<ParseResult.Success<Map<String, Any?>>>(result)
        assertEquals("user", result.value["role"])
    }

    @Test
    fun `present field is not overwritten`() {
        val schema = obj(mapOf("role" to string().default("user")))

        val result = schema.safeParse(mapOf("role" to "admin"))

        assertIs<ParseResult.Success<Map<String, Any?>>>(result)
        assertEquals("admin", result.value["role"])
    }

    @Test
    fun `invalid default produces default invalid`() {
        val schema = obj(mapOf("count" to int_().min(10).default(5)))

        val result = schema.safeParse(emptyMap<String, Any?>())

        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.DEFAULT_INVALID, result.issues[0].code)
        assertEquals(listOf("count"), result.issues[0].path)
    }

    @Test
    fun `null is not absent for nullable default`() {
        val schema = obj(mapOf("value" to nullable(string().default("fallback"))))

        val result = schema.safeParse(mapOf("value" to null))

        assertIs<ParseResult.Success<Map<String, Any?>>>(result)
        assertNull(result.value["value"])
    }

    @Test
    fun `falsy defaults are applied`() {
        val schema = obj(
            mapOf(
                "count" to int_().default(0),
                "name" to string().default(""),
                "active" to bool().default(false),
            )
        )

        val result = schema.safeParse(emptyMap<String, Any?>())

        assertIs<ParseResult.Success<Map<String, Any?>>>(result)
        assertEquals(0L, result.value["count"])
        assertEquals("", result.value["name"])
        assertEquals(false, result.value["active"])
    }

    @Test
    fun `nested object field gets default`() {
        val schema = obj(
            mapOf(
                "user" to obj(
                    mapOf(
                        "name" to string(),
                        "role" to string().default("guest"),
                    ),
                    required = setOf("name"),
                )
            ),
            required = setOf("user"),
        )

        val result = schema.safeParse(mapOf("user" to mapOf("name" to "Bob")))

        assertIs<ParseResult.Success<Map<String, Any?>>>(result)
        @Suppress("UNCHECKED_CAST")
        val user = result.value["user"] as Map<String, Any?>
        assertEquals("guest", user["role"])
    }
}

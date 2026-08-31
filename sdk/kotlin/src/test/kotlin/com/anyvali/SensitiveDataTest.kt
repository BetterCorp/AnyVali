package com.anyvali

import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertIs

class SensitiveDataTest {
    @Test
    fun `encrypted round trip and validation`() {
        val sensitive = DescribeOptions(sensitive = true)
        val schema = obj(
            mapOf(
                "name" to string(),
                "secret" to string().describe("secret", sensitive),
                "profile" to obj(mapOf("token" to string()), setOf("token")).describe("profile", sensitive),
                "aliases" to array(string().describe("alias", sensitive)),
            ),
            setOf("name", "secret", "profile", "aliases")
        )
        val plain = mapOf(
            "name" to "Ada", "secret" to "abc",
            "profile" to mapOf("token" to "xyz"), "aliases" to listOf("one", "two")
        )
        val encrypted = encrypt(schema, plain) { path, value -> "encrypted:${path.joinToString(".")}:$value" }

        assertIs<ParseResult.Success<*>>(safeParseEncrypted(schema, encrypted))
        assertThrows<ValidationError> { encrypt(schema, plain) { _, _ -> "broken" } }
        assertEquals(plain, decrypt(schema, encrypted) { path, _ ->
            when (path.joinToString(".")) {
                "profile" -> mapOf("token" to "xyz")
                "aliases.0" -> "one"
                "aliases.1" -> "two"
                else -> "abc"
            }
        })
    }
}

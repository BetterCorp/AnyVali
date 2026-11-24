package com.anyvali

import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertIs
import com.anyvali.interchange.Importer
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

class SensitiveDataTest {
    @Test
    fun `imported sensitive metadata encrypts instead of returning plaintext`() {
        val (schema, _) = Importer.importFromJson("""{
            "anyvaliVersion":"1.0","schemaVersion":"1.1","definitions":{},
            "root":{"kind":"string","metadata":{"sensitive":true,"vendor":{"scope":"test"}}}
        }""")
        val encrypted = encrypt(schema, "secret") { _, value -> "encrypted:$value" }
        assertEquals("encrypted:secret", encrypted)
        assertIs<ParseResult.Failure>(safeParseEncrypted(schema, "secret"))
        assertEquals(true, schema.schemaMetadata?.get("sensitive"))
        assertEquals(
            "test",
            schema.exportNode()["metadata"]?.jsonObject?.get("vendor")?.jsonObject?.get("scope")?.jsonPrimitive?.content,
        )
    }

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

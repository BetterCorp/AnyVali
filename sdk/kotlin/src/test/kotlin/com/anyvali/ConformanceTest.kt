package com.anyvali

import com.anyvali.interchange.Importer
import kotlinx.serialization.json.*
import org.junit.jupiter.api.DynamicTest
import org.junit.jupiter.api.TestFactory
import java.io.File
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue

class ConformanceTest {

    private val json = Json { ignoreUnknownKeys = true }

    private fun findCorpusDir(): File {
        // Navigate from sdk/kotlin up to repo root, then into spec/corpus
        var dir = File(System.getProperty("user.dir"))
        // If we're in sdk/kotlin, go up two levels
        val specCorpus = File(dir, "../../spec/corpus")
        if (specCorpus.exists()) return specCorpus.canonicalFile

        // Try from repo root
        val direct = File(dir, "spec/corpus")
        if (direct.exists()) return direct.canonicalFile

        throw IllegalStateException("Cannot find spec/corpus directory from $dir")
    }

    private fun loadTestSuites(): List<Pair<String, JsonObject>> {
        val corpusDir = findCorpusDir()
        val suites = mutableListOf<Pair<String, JsonObject>>()
        corpusDir.walkTopDown()
            .filter { it.isFile && it.extension == "json" }
            .forEach { file ->
                val content = file.readText()
                val obj = json.parseToJsonElement(content).jsonObject
                suites.add(file.nameWithoutExtension to obj)
            }
        return suites
    }

    @TestFactory
    fun `conformance test corpus`(): List<DynamicTest> {
        val suites = loadTestSuites()
        val tests = mutableListOf<DynamicTest>()

        for ((suiteName, suiteObj) in suites) {
            val cases = suiteObj["cases"]?.jsonArray ?: continue

            for (caseEl in cases) {
                val caseObj = caseEl.jsonObject
                val description = caseObj["description"]?.jsonPrimitive?.content ?: "unknown"
                val schemaDoc = caseObj["schema"]!!.jsonObject
                val inputEl = caseObj["input"]!!
                val valid = caseObj["valid"]!!.jsonPrimitive.boolean
                val expectedIssues = caseObj["issues"]?.jsonArray ?: JsonArray(emptyList())

                tests.add(DynamicTest.dynamicTest("$suiteName: $description") {
                    runConformanceCase(schemaDoc, inputEl, valid, expectedIssues)
                })
            }
        }

        assertTrue(tests.isNotEmpty(), "Should have found at least one test case")
        return tests
    }

    private fun runConformanceCase(
        schemaDoc: JsonObject,
        inputEl: JsonElement,
        expectedValid: Boolean,
        expectedIssues: JsonArray
    ) {
        val (schema, definitions) = Importer.importFromJsonObject(schemaDoc)
        val input = jsonElementToKotlin(inputEl)

        val result = schema.safeParse(input, definitions)

        assertEquals(
            expectedValid,
            result.isSuccess,
            "Expected valid=$expectedValid but got ${if (result.isSuccess) "success" else "failure: ${(result as ParseResult.Failure).issues}"}"
        )

        if (!expectedValid) {
            assertIs<ParseResult.Failure>(result)
            // Check that we got the expected issue codes
            for ((index, expectedIssueEl) in expectedIssues.withIndex()) {
                val expectedIssue = expectedIssueEl.jsonObject
                val expectedCode = expectedIssue["code"]!!.jsonPrimitive.content
                if (index < result.issues.size) {
                    assertEquals(
                        expectedCode,
                        result.issues[index].code,
                        "Issue[$index] code mismatch"
                    )
                    // Check path
                    val expectedPath = expectedIssue["path"]!!.jsonArray.map { pathEl ->
                        when {
                            pathEl is JsonPrimitive && pathEl.isString -> pathEl.content
                            pathEl is JsonPrimitive -> pathEl.int
                            else -> pathEl.toString()
                        }
                    }
                    assertEquals(
                        expectedPath,
                        result.issues[index].path,
                        "Issue[$index] path mismatch"
                    )
                }
            }
            assertEquals(
                expectedIssues.size,
                result.issues.size,
                "Expected ${expectedIssues.size} issues but got ${result.issues.size}: ${result.issues}"
            )
        }

        if (expectedValid) {
            assertIs<ParseResult.Success>(result)
        }
    }

    private fun jsonElementToKotlin(element: JsonElement): Any? {
        return when (element) {
            is JsonNull -> null
            is JsonPrimitive -> {
                if (element.isString) {
                    element.content
                } else {
                    val content = element.content
                    if (content == "true") return true
                    if (content == "false") return false
                    // Try long first, then double
                    content.toLongOrNull()?.let { return it }
                    content.toDoubleOrNull()?.let { return it }
                    content
                }
            }
            is JsonArray -> element.map { jsonElementToKotlin(it) }
            is JsonObject -> element.entries.associate { (k, v) -> k to jsonElementToKotlin(v) }
        }
    }
}

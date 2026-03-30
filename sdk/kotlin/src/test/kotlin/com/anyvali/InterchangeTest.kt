package com.anyvali

import com.anyvali.interchange.Exporter
import com.anyvali.interchange.Importer
import com.anyvali.schemas.*
import kotlinx.serialization.json.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull
import kotlin.test.assertTrue

class InterchangeTest {

    private val jsonParser = Json { ignoreUnknownKeys = true }

    // ---- Export tests ----

    @Test
    fun `export string schema`() {
        val s = string().minLength(1).maxLength(100)
        val doc = s.export()
        assertEquals("1.0", doc.anyvaliVersion)
        assertEquals("1", doc.schemaVersion)
        val root = doc.root
        assertEquals("string", root["kind"]?.jsonPrimitive?.content)
        assertEquals(1, root["minLength"]?.jsonPrimitive?.int)
        assertEquals(100, root["maxLength"]?.jsonPrimitive?.int)
    }

    @Test
    fun `export number schema with constraints`() {
        val n = number().min(0).max(100)
        val doc = n.export()
        val root = doc.root
        assertEquals("number", root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export int schema`() {
        val i = int_().min(1).max(10)
        val doc = i.export()
        val root = doc.root
        assertEquals("int", root["kind"]?.jsonPrimitive?.content)
        assertEquals(1L, root["min"]?.jsonPrimitive?.long)
    }

    @Test
    fun `export object schema`() {
        val o = obj(
            properties = mapOf("name" to string(), "age" to int_()),
            required = setOf("name", "age")
        )
        val doc = o.export()
        val root = doc.root
        assertEquals("object", root["kind"]?.jsonPrimitive?.content)
        assertTrue(root["properties"] is JsonObject)
        assertTrue(root["required"] is JsonArray)
    }

    @Test
    fun `export array schema`() {
        val a = array(string()).minItems(1)
        val doc = a.export()
        val root = doc.root
        assertEquals("array", root["kind"]?.jsonPrimitive?.content)
        assertEquals(1, root["minItems"]?.jsonPrimitive?.int)
    }

    @Test
    fun `export union schema`() {
        val u = union(string(), int_())
        val doc = u.export()
        val root = doc.root
        assertEquals("union", root["kind"]?.jsonPrimitive?.content)
        assertTrue(root["variants"] is JsonArray)
    }

    @Test
    fun `export literal schema`() {
        val l = literal("hello")
        val doc = l.export()
        assertEquals("literal", doc.root["kind"]?.jsonPrimitive?.content)
        assertEquals("hello", doc.root["value"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export null literal schema`() {
        val l = literal(null)
        val doc = l.export()
        assertTrue(doc.root["value"] is JsonNull)
    }

    @Test
    fun `export enum schema`() {
        val e = enum_("a", "b", "c")
        val doc = e.export()
        assertEquals("enum", doc.root["kind"]?.jsonPrimitive?.content)
        assertTrue(doc.root["values"] is JsonArray)
    }

    @Test
    fun `export tuple schema`() {
        val t = tuple(string(), int_())
        val doc = t.export()
        assertEquals("tuple", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export record schema`() {
        val r = record(string())
        val doc = r.export()
        assertEquals("record", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export intersection schema`() {
        val i = intersection(number().min(0), number().max(100))
        val doc = i.export()
        assertEquals("intersection", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export optional schema`() {
        val o = optional(string())
        val doc = o.export()
        assertEquals("optional", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export nullable schema`() {
        val n = nullable(string())
        val doc = n.export()
        assertEquals("nullable", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export ref schema`() {
        val r = ref("#/definitions/User")
        val doc = r.export()
        assertEquals("ref", doc.root["kind"]?.jsonPrimitive?.content)
        assertEquals("#/definitions/User", doc.root["ref"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export bool schema`() {
        val b = bool()
        val doc = b.export()
        assertEquals("bool", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export null schema`() {
        val n = null_()
        val doc = n.export()
        assertEquals("null", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export any schema`() {
        val a = any_()
        val doc = a.export()
        assertEquals("any", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export unknown schema`() {
        val u = unknown()
        val doc = u.export()
        assertEquals("unknown", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export never schema`() {
        val n = never()
        val doc = n.export()
        assertEquals("never", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export string with coercion`() {
        val s = string().coerce("trim")
        val doc = s.export()
        assertEquals("trim", doc.root["coerce"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export string with chained coercions`() {
        val s = string().coerce("trim", "lower")
        val doc = s.export()
        assertTrue(doc.root["coerce"] is JsonArray)
    }

    @Test
    fun `export string with format`() {
        val s = string().format("email")
        val doc = s.export()
        assertEquals("email", doc.root["format"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export string with pattern`() {
        val s = string().pattern("^[a-z]+$")
        val doc = s.export()
        assertEquals("^[a-z]+$", doc.root["pattern"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export string with startsWith endsWith includes`() {
        val s = string().startsWith("a").endsWith("z").includes("m")
        val doc = s.export()
        assertEquals("a", doc.root["startsWith"]?.jsonPrimitive?.content)
        assertEquals("z", doc.root["endsWith"]?.jsonPrimitive?.content)
        assertEquals("m", doc.root["includes"]?.jsonPrimitive?.content)
    }

    @Test
    fun `export number with all constraints`() {
        val n = number().min(0).max(100).exclusiveMin(0).exclusiveMax(100).multipleOf(5)
        val doc = n.export()
        assertTrue(doc.root.containsKey("min"))
        assertTrue(doc.root.containsKey("max"))
        assertTrue(doc.root.containsKey("exclusiveMin"))
        assertTrue(doc.root.containsKey("exclusiveMax"))
        assertTrue(doc.root.containsKey("multipleOf"))
    }

    @Test
    fun `export int with all constraints`() {
        val i = int_().min(0).max(100).exclusiveMin(0).exclusiveMax(100).multipleOf(5)
        val doc = i.export()
        assertTrue(doc.root.containsKey("min"))
        assertTrue(doc.root.containsKey("multipleOf"))
    }

    @Test
    fun `Exporter exportToJson produces valid JSON string`() {
        val s = string()
        val jsonStr = Exporter.exportToJson(s)
        val parsed = jsonParser.parseToJsonElement(jsonStr)
        assertTrue(parsed is JsonObject)
    }

    @Test
    fun `Exporter exportDocument with definitions`() {
        val userSchema = obj(
            properties = mapOf("name" to string()),
            required = setOf("name")
        )
        val doc = Exporter.exportDocument(
            schema = ref("#/definitions/User"),
            definitions = mapOf("User" to userSchema)
        )
        assertTrue(doc.definitions.containsKey("User"))
    }

    @Test
    fun `export with custom validator in portable mode fails`() {
        val schema = object : Schema<Any?>() {
            override val kind = "custom"
            override val hasCustomValidators = true
            override fun validateValue(value: Any?, ctx: ValidationContext) = emptyList<ValidationIssue>()
            override fun exportNode() = buildJsonObject { put("kind", JsonPrimitive("custom")) }
        }
        assertThrows<ValidationError> { schema.export(ExportMode.PORTABLE) }
    }

    @Test
    fun `export with custom validator in extended mode succeeds`() {
        val schema = object : Schema<Any?>() {
            override val kind = "custom"
            override val hasCustomValidators = true
            override fun validateValue(value: Any?, ctx: ValidationContext) = emptyList<ValidationIssue>()
            override fun exportNode() = buildJsonObject { put("kind", JsonPrimitive("custom")) }
        }
        val doc = schema.export(ExportMode.EXTENDED)
        assertEquals("custom", doc.root["kind"]?.jsonPrimitive?.content)
    }

    // ---- Import tests ----

    @Test
    fun `import string schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string", "minLength": 3 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<StringSchema>(schema)
        assertEquals(3, (schema as StringSchema).minLength)
    }

    @Test
    fun `import number schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "number", "min": 0 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NumberSchema>(schema)
    }

    @Test
    fun `import int schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "int" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<IntSchema>(schema)
    }

    @Test
    fun `import bool schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "bool" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<BoolSchema>(schema)
    }

    @Test
    fun `import null schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "null" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NullSchema>(schema)
    }

    @Test
    fun `import object schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "object",
                "properties": {
                    "name": { "kind": "string" }
                },
                "required": ["name"],
                "unknownKeys": "reject"
            },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<ObjectSchema>(schema)
        assertEquals(1, (schema as ObjectSchema).properties.size)
    }

    @Test
    fun `import array schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "array", "items": { "kind": "int" }, "minItems": 1 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<ArraySchema>(schema)
        assertEquals(1, (schema as ArraySchema).minItems)
    }

    @Test
    fun `import union schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "union",
                "variants": [
                    { "kind": "string" },
                    { "kind": "int" }
                ]
            },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<UnionSchema>(schema)
    }

    @Test
    fun `import with definitions and refs`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "object",
                "properties": {
                    "user": { "kind": "ref", "ref": "#/definitions/User" }
                },
                "required": ["user"],
                "unknownKeys": "reject"
            },
            "definitions": {
                "User": {
                    "kind": "object",
                    "properties": {
                        "name": { "kind": "string" }
                    },
                    "required": ["name"],
                    "unknownKeys": "reject"
                }
            },
            "extensions": {}
        }
        """.trimIndent()
        val (schema, defs) = Importer.importFromJson(jsonStr)
        assertIs<ObjectSchema>(schema)
        assertTrue(defs.containsKey("User"))

        // Validate with definitions
        val result = schema.safeParse(mapOf("user" to mapOf("name" to "Alice")), defs)
        assertIs<ParseResult.Success<*>>(result)
    }

    @Test
    fun `import literal schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "literal", "value": "hello" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<LiteralSchema>(schema)
    }

    @Test
    fun `import enum schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "enum", "values": ["a", "b", "c"] },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<EnumSchema>(schema)
    }

    @Test
    fun `import tuple schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "tuple", "elements": [{ "kind": "string" }, { "kind": "int" }] },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<TupleSchema>(schema)
    }

    @Test
    fun `import record schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "record", "values": { "kind": "int" } },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<RecordSchema>(schema)
    }

    @Test
    fun `import intersection schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "intersection", "allOf": [{ "kind": "number", "min": 0 }, { "kind": "number", "max": 100 }] },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<IntersectionSchema>(schema)
    }

    @Test
    fun `import optional schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "optional", "schema": { "kind": "string" } },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<OptionalSchema>(schema)
    }

    @Test
    fun `import nullable schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "nullable", "schema": { "kind": "string" } },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NullableSchema>(schema)
    }

    @Test
    fun `import any schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "any" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<AnySchema>(schema)
    }

    @Test
    fun `import unknown schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "unknown" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<UnknownSchema>(schema)
    }

    @Test
    fun `import never schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "never" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NeverSchema>(schema)
    }

    @Test
    fun `import float32 schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "float32" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NumberSchema>(schema)
        assertEquals("float32", (schema as NumberSchema).schemaKind)
    }

    @Test
    fun `import float64 schema`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "float64" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NumberSchema>(schema)
        assertEquals("float64", (schema as NumberSchema).schemaKind)
    }

    @Test
    fun `import int width schemas`() {
        for (kind in listOf("int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64")) {
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
            assertIs<IntSchema>(schema)
            assertEquals(kind, (schema as IntSchema).schemaKind)
        }
    }

    @Test
    fun `import unsupported schema kind fails`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "custom_unsupported" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        assertThrows<ValidationError> { Importer.importFromJson(jsonStr) }
    }

    @Test
    fun `import missing kind field fails`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {},
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        assertThrows<Exception> { Importer.importFromJson(jsonStr) }
    }

    @Test
    fun `import null empty root fails`() {
        val emptyRoot = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        assertThrows<Exception> { Importer.importFromJson(emptyRoot) }

        val nullRoot = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": null,
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        assertThrows<Exception> { Importer.importFromJson(nullRoot) }
    }

    @Test
    fun `import string with coercion`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string", "coerce": "trim" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<StringSchema>(schema)
        assertEquals(listOf("trim"), (schema as StringSchema).coerce)
    }

    @Test
    fun `import string with chained coercions`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string", "coerce": ["trim", "lower"] },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<StringSchema>(schema)
        assertEquals(listOf("trim", "lower"), (schema as StringSchema).coerce)
    }

    @Test
    fun `import int with coercion`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "int", "coerce": "string->int" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<IntSchema>(schema)
        assertEquals("string->int", (schema as IntSchema).coerce)
    }

    @Test
    fun `import number with coercion`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "number", "coerce": "string->number" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NumberSchema>(schema)
        assertEquals("string->number", (schema as NumberSchema).coerce)
    }

    @Test
    fun `import bool with default`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "bool", "default": true },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<BoolSchema>(schema)
        assertEquals(true, (schema as BoolSchema).defaultValue)
    }

    @Test
    fun `import bool with coercion`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "bool", "coerce": "string->bool" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<BoolSchema>(schema)
        assertEquals("string->bool", (schema as BoolSchema).coerce)
    }

    // ---- Round-trip tests ----

    @Test
    fun `roundtrip string schema`() {
        val original = string().minLength(1).maxLength(50).pattern("^[a-z]+$")
        val exported = Exporter.exportToJson(original)
        val (imported, _) = Importer.importFromJson(exported)
        assertIs<StringSchema>(imported)
        assertEquals("hello", imported.parse("hello"))
    }

    @Test
    fun `roundtrip object schema`() {
        val original = obj(
            properties = mapOf("name" to string(), "age" to int_()),
            required = setOf("name", "age")
        )
        val exported = Exporter.exportToJson(original)
        val (imported, _) = Importer.importFromJson(exported)
        assertIs<ObjectSchema>(imported)
        val result = imported.parse(mapOf("name" to "Bob", "age" to 25)) as Map<*, *>
        assertEquals("Bob", result["name"])
        assertEquals(25L, result["age"])
    }

    @Test
    fun `Exporter exportDocument custom validator portable fails`() {
        val schema = object : Schema<Any?>() {
            override val kind = "custom"
            override val hasCustomValidators = true
            override fun validateValue(value: Any?, ctx: ValidationContext) = emptyList<ValidationIssue>()
            override fun exportNode() = buildJsonObject { put("kind", JsonPrimitive("custom")) }
        }
        assertThrows<ValidationError> {
            Exporter.exportDocument(schema, mode = ExportMode.PORTABLE)
        }
    }

    @Test
    fun `Exporter exportDocument custom validator extended succeeds`() {
        val schema = object : Schema<Any?>() {
            override val kind = "custom"
            override val hasCustomValidators = true
            override fun validateValue(value: Any?, ctx: ValidationContext) = emptyList<ValidationIssue>()
            override fun exportNode() = buildJsonObject { put("kind", JsonPrimitive("custom")) }
        }
        val doc = Exporter.exportDocument(schema, mode = ExportMode.EXTENDED)
        assertEquals("custom", doc.root["kind"]?.jsonPrimitive?.content)
    }

    @Test
    fun `import_ top-level function works`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "string" },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = import_(jsonStr)
        assertIs<StringSchema>(schema)
    }

    @Test
    fun `import literal with null value`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "literal", "value": null },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<LiteralSchema>(schema)
        assertNull(schema.parse(null))
    }

    @Test
    fun `import literal with boolean value`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "literal", "value": true },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<LiteralSchema>(schema)
        assertEquals(true, schema.parse(true))
    }

    @Test
    fun `import literal with numeric value`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "literal", "value": 42 },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<LiteralSchema>(schema)
        assertEquals(42L, schema.parse(42L))
    }

    @Test
    fun `import enum with numeric values`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": { "kind": "enum", "values": [1, 2, 3] },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<EnumSchema>(schema)
        assertEquals(2L, schema.parse(2L))
    }

    @Test
    fun `import object with default unknown keys mode`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "object",
                "properties": { "name": { "kind": "string" } },
                "required": ["name"]
            },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<ObjectSchema>(schema)
        // Default unknown key mode should be reject
        val result = schema.safeParse(mapOf("name" to "Alice", "extra" to "val"))
        assertIs<ParseResult.Failure>(result)
        assertEquals(IssueCodes.UNKNOWN_KEY, result.issues[0].code)
    }

    @Test
    fun `import string with all constraints`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "string",
                "minLength": 1,
                "maxLength": 100,
                "pattern": "^[a-z]+$",
                "startsWith": "a",
                "endsWith": "z",
                "includes": "m",
                "format": "email"
            },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<StringSchema>(schema)
        val s = schema as StringSchema
        assertEquals(1, s.minLength)
        assertEquals(100, s.maxLength)
        assertEquals("^[a-z]+$", s.pattern)
        assertEquals("a", s.startsWith)
        assertEquals("z", s.endsWith)
        assertEquals("m", s.includes)
        assertEquals("email", s.format)
    }

    @Test
    fun `import number with all constraints`() {
        val jsonStr = """
        {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "number",
                "min": 0.0,
                "max": 100.0,
                "exclusiveMin": -1.0,
                "exclusiveMax": 101.0,
                "multipleOf": 0.5
            },
            "definitions": {},
            "extensions": {}
        }
        """.trimIndent()
        val (schema, _) = Importer.importFromJson(jsonStr)
        assertIs<NumberSchema>(schema)
        val n = schema as NumberSchema
        assertEquals(0.0, n.min)
        assertEquals(100.0, n.max)
        assertEquals(-1.0, n.exclusiveMin)
        assertEquals(101.0, n.exclusiveMax)
        assertEquals(0.5, n.multipleOf)
    }
}

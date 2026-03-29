package com.anyvali.interchange

import com.anyvali.*
import com.anyvali.schemas.*
import kotlinx.serialization.json.*

object Importer {

    private val json = Json { ignoreUnknownKeys = true }

    fun importFromJson(jsonStr: String): Pair<Schema, Map<String, Schema>> {
        val element = json.parseToJsonElement(jsonStr).jsonObject
        return importFromJsonObject(element)
    }

    fun importFromJsonObject(obj: JsonObject): Pair<Schema, Map<String, Schema>> {
        val defsObj = obj["definitions"]?.jsonObject ?: JsonObject(emptyMap())
        val definitions = mutableMapOf<String, Schema>()

        // First pass: create placeholder refs for all definitions
        for (key in defsObj.keys) {
            // Will be resolved later
        }

        // Parse definitions
        for ((key, value) in defsObj) {
            definitions[key] = importNode(value.jsonObject, definitions)
        }

        val root = importNode(obj["root"]!!.jsonObject, definitions)
        return root to definitions
    }

    fun importNode(node: JsonObject, definitions: Map<String, Schema> = emptyMap()): Schema {
        val kind = node["kind"]?.jsonPrimitive?.content
            ?: throw IllegalArgumentException("Schema node missing 'kind'")

        return when (kind) {
            "string" -> importStringSchema(node)
            "number", "float64" -> importNumberSchema(node, kind)
            "float32" -> importNumberSchema(node, kind)
            "int", "int64" -> importIntSchema(node, kind)
            "int8", "int16", "int32" -> importIntSchema(node, kind)
            "uint8", "uint16", "uint32", "uint64" -> importIntSchema(node, kind)
            "bool" -> importBoolSchema(node)
            "null" -> NullSchema()
            "any" -> AnySchema()
            "unknown" -> UnknownSchema()
            "never" -> NeverSchema()
            "literal" -> importLiteralSchema(node)
            "enum" -> importEnumSchema(node)
            "array" -> importArraySchema(node, definitions)
            "tuple" -> importTupleSchema(node, definitions)
            "object" -> importObjectSchema(node, definitions)
            "record" -> importRecordSchema(node, definitions)
            "union" -> importUnionSchema(node, definitions)
            "intersection" -> importIntersectionSchema(node, definitions)
            "optional" -> importOptionalSchema(node, definitions)
            "nullable" -> importNullableSchema(node, definitions)
            "ref" -> importRefSchema(node)
            else -> throw ValidationError(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.UNSUPPORTED_SCHEMA_KIND,
                        message = "Unsupported schema kind: $kind"
                    )
                )
            )
        }
    }

    private fun importStringSchema(node: JsonObject): StringSchema {
        var schema = StringSchema()
        node["minLength"]?.jsonPrimitive?.int?.let { schema = schema.copy(minLength = it) }
        node["maxLength"]?.jsonPrimitive?.int?.let { schema = schema.copy(maxLength = it) }
        node["pattern"]?.jsonPrimitive?.content?.let { schema = schema.copy(pattern = it) }
        node["startsWith"]?.jsonPrimitive?.content?.let { schema = schema.copy(startsWith = it) }
        node["endsWith"]?.jsonPrimitive?.content?.let { schema = schema.copy(endsWith = it) }
        node["includes"]?.jsonPrimitive?.content?.let { schema = schema.copy(includes = it) }
        node["format"]?.jsonPrimitive?.content?.let { schema = schema.copy(format = it) }
        node["default"]?.jsonPrimitive?.content?.let { schema = schema.copy(defaultValue = it) }

        val coerceEl = node["coerce"]
        if (coerceEl != null) {
            val coercions = when {
                coerceEl is JsonArray -> coerceEl.map { it.jsonPrimitive.content }
                coerceEl is JsonPrimitive -> listOf(coerceEl.content)
                else -> emptyList()
            }
            schema = schema.copy(coerce = coercions)
        }

        return schema
    }

    private fun importNumberSchema(node: JsonObject, kind: String): NumberSchema {
        var schema = NumberSchema(schemaKind = kind)
        node["min"]?.jsonPrimitive?.double?.let { schema = schema.copy(min = it) }
        node["max"]?.jsonPrimitive?.double?.let { schema = schema.copy(max = it) }
        node["exclusiveMin"]?.jsonPrimitive?.double?.let { schema = schema.copy(exclusiveMin = it) }
        node["exclusiveMax"]?.jsonPrimitive?.double?.let { schema = schema.copy(exclusiveMax = it) }
        node["multipleOf"]?.jsonPrimitive?.double?.let { schema = schema.copy(multipleOf = it) }
        node["coerce"]?.jsonPrimitive?.content?.let { schema = schema.copy(coerce = it) }

        val defaultEl = node["default"]
        if (defaultEl != null && defaultEl is JsonPrimitive && defaultEl.isString.not()) {
            schema = schema.copy(defaultValue = defaultEl.double)
        }

        return schema
    }

    private fun importIntSchema(node: JsonObject, kind: String): IntSchema {
        var schema = IntSchema(schemaKind = kind)
        node["min"]?.jsonPrimitive?.long?.let { schema = schema.copy(min = it) }
        node["max"]?.jsonPrimitive?.long?.let { schema = schema.copy(max = it) }
        node["exclusiveMin"]?.jsonPrimitive?.long?.let { schema = schema.copy(exclusiveMin = it) }
        node["exclusiveMax"]?.jsonPrimitive?.long?.let { schema = schema.copy(exclusiveMax = it) }
        node["multipleOf"]?.jsonPrimitive?.long?.let { schema = schema.copy(multipleOf = it) }
        node["coerce"]?.jsonPrimitive?.content?.let { schema = schema.copy(coerce = it) }

        val defaultEl = node["default"]
        if (defaultEl != null && defaultEl is JsonPrimitive && defaultEl.isString.not()) {
            schema = schema.copy(defaultValue = defaultEl.long)
        }

        return schema
    }

    private fun importBoolSchema(node: JsonObject): BoolSchema {
        var schema = BoolSchema()
        val defaultEl = node["default"]
        if (defaultEl != null && defaultEl is JsonPrimitive) {
            schema = schema.copy(defaultValue = defaultEl.boolean)
        }
        node["coerce"]?.jsonPrimitive?.content?.let { schema = schema.copy(coerce = it) }
        return schema
    }

    private fun importLiteralSchema(node: JsonObject): LiteralSchema {
        val valueEl = node["value"]!!
        val value: Any? = when {
            valueEl is JsonNull -> null
            valueEl is JsonPrimitive && valueEl.isString -> valueEl.content
            valueEl is JsonPrimitive && valueEl.content == "true" -> true
            valueEl is JsonPrimitive && valueEl.content == "false" -> false
            valueEl is JsonPrimitive -> {
                val longVal = valueEl.longOrNull
                if (longVal != null) longVal else valueEl.double
            }
            else -> null
        }
        return LiteralSchema(value)
    }

    private fun importEnumSchema(node: JsonObject): EnumSchema {
        val valuesArr = node["values"]!!.jsonArray
        val values = valuesArr.map { el ->
            when {
                el is JsonNull -> null
                el is JsonPrimitive && el.isString -> el.content
                el is JsonPrimitive && el.content == "true" -> true
                el is JsonPrimitive && el.content == "false" -> false
                el is JsonPrimitive -> {
                    val longVal = el.longOrNull
                    if (longVal != null) longVal else el.double
                }
                else -> null
            }
        }
        return EnumSchema(values)
    }

    private fun importArraySchema(node: JsonObject, definitions: Map<String, Schema>): ArraySchema {
        val items = importNode(node["items"]!!.jsonObject, definitions)
        var schema = ArraySchema(items)
        node["minItems"]?.jsonPrimitive?.int?.let { schema = schema.copy(minItems = it) }
        node["maxItems"]?.jsonPrimitive?.int?.let { schema = schema.copy(maxItems = it) }
        return schema
    }

    private fun importTupleSchema(node: JsonObject, definitions: Map<String, Schema>): TupleSchema {
        val elements = node["elements"]!!.jsonArray.map { importNode(it.jsonObject, definitions) }
        return TupleSchema(elements)
    }

    private fun importObjectSchema(node: JsonObject, definitions: Map<String, Schema>): ObjectSchema {
        val props = node["properties"]?.jsonObject?.mapValues { importNode(it.value.jsonObject, definitions) } ?: emptyMap()
        val required = node["required"]?.jsonArray?.map { it.jsonPrimitive.content }?.toSet() ?: emptySet()
        val unknownKeysStr = node["unknownKeys"]?.jsonPrimitive?.content ?: "reject"
        val unknownKeys = UnknownKeyMode.fromValue(unknownKeysStr)
        return ObjectSchema(props, required, unknownKeys)
    }

    private fun importRecordSchema(node: JsonObject, definitions: Map<String, Schema>): RecordSchema {
        val values = importNode(node["values"]!!.jsonObject, definitions)
        return RecordSchema(values)
    }

    private fun importUnionSchema(node: JsonObject, definitions: Map<String, Schema>): UnionSchema {
        val variants = node["variants"]!!.jsonArray.map { importNode(it.jsonObject, definitions) }
        return UnionSchema(variants)
    }

    private fun importIntersectionSchema(node: JsonObject, definitions: Map<String, Schema>): IntersectionSchema {
        val allOf = node["allOf"]!!.jsonArray.map { importNode(it.jsonObject, definitions) }
        return IntersectionSchema(allOf)
    }

    private fun importOptionalSchema(node: JsonObject, definitions: Map<String, Schema>): OptionalSchema {
        val inner = importNode(node["schema"]!!.jsonObject, definitions)
        return OptionalSchema(inner)
    }

    private fun importNullableSchema(node: JsonObject, definitions: Map<String, Schema>): NullableSchema {
        val inner = importNode(node["schema"]!!.jsonObject, definitions)
        // Check for default on the nullable wrapper
        val defaultEl = node["default"]
        if (defaultEl != null) {
            // Need to set default on the inner schema if it supports it
            // For now, wrap with the inner having default
        }
        return NullableSchema(inner)
    }

    private fun importRefSchema(node: JsonObject): RefSchema {
        val ref = node["ref"]!!.jsonPrimitive.content
        return RefSchema(ref)
    }
}

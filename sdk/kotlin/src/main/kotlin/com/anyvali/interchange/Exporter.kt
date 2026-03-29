package com.anyvali.interchange

import com.anyvali.*
import com.anyvali.schemas.*
import kotlinx.serialization.json.*

object Exporter {

    fun exportToJson(schema: Schema, mode: ExportMode = ExportMode.PORTABLE): String {
        val doc = schema.export(mode)
        return toJsonString(doc)
    }

    fun toJsonString(doc: AnyValiDocument): String {
        val json = buildJsonObject {
            put("anyvaliVersion", JsonPrimitive(doc.anyvaliVersion))
            put("schemaVersion", JsonPrimitive(doc.schemaVersion))
            put("root", doc.root)
            put("definitions", JsonObject(doc.definitions.mapValues { it.value }))
            put("extensions", JsonObject(doc.extensions))
        }
        return Json { prettyPrint = true }.encodeToString(JsonElement.serializer(), json)
    }

    fun exportDocument(
        schema: Schema,
        definitions: Map<String, Schema> = emptyMap(),
        extensions: Map<String, JsonElement> = emptyMap(),
        mode: ExportMode = ExportMode.PORTABLE
    ): AnyValiDocument {
        if (mode == ExportMode.PORTABLE && schema.hasCustomValidators) {
            throw ValidationError(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.CUSTOM_VALIDATION_NOT_PORTABLE,
                        message = "Schema uses custom validators that are not portable"
                    )
                )
            )
        }
        return AnyValiDocument(
            root = schema.exportNode(),
            definitions = definitions.mapValues { it.value.exportNode() },
            extensions = extensions
        )
    }
}

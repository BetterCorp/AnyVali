package com.anyvali

import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject

data class AnyValiDocument(
    val anyvaliVersion: String = "1.0",
    val schemaVersion: String = "1",
    val root: JsonObject,
    val definitions: Map<String, JsonObject> = emptyMap(),
    val extensions: Map<String, JsonElement> = emptyMap()
)

package com.anyvali

import kotlinx.serialization.json.*

data class DescribeOptions(
    val title: String? = null,
    val deprecated: Boolean? = null,
    val deprecatedMessage: String? = null,
    val notStable: Boolean? = null,
    val since: String? = null,
    val sensitive: Boolean? = null,
    val readonly: Boolean? = null,
    val writeonly: Boolean? = null,
    val examples: List<Any?>? = null
)

abstract class Schema<out T> {
    abstract val kind: String
    open val hasCustomValidators: Boolean get() = false

    // Metadata storage
    var schemaMetadata: MutableMap<String, Any?>? = null
        protected set

    abstract fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue>

    abstract fun exportNode(): JsonObject

    fun parse(input: Any?): T {
        val result = safeParse(input)
        return result.getOrThrow()
    }

    fun parse(input: Any?, definitions: Map<String, Schema<*>>): T {
        val result = safeParse(input, definitions)
        return result.getOrThrow()
    }

    fun safeParse(input: Any?): ParseResult<T> {
        return safeParse(input, emptyMap())
    }

    fun safeParse(input: Any?, definitions: Map<String, Schema<*>>): ParseResult<T> {
        val ctx = ValidationContext(definitions = definitions)
        return safeParseWithContext(input, ctx)
    }

    @Suppress("UNCHECKED_CAST")
    open fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult<T> {
        val issues = validateValue(input, ctx)
        return if (issues.isEmpty()) {
            ParseResult.Success(input as T)
        } else {
            ParseResult.Failure(issues)
        }
    }

    fun export(mode: ExportMode = ExportMode.PORTABLE): AnyValiDocument {
        if (mode == ExportMode.PORTABLE && hasCustomValidators) {
            throw ValidationError(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.CUSTOM_VALIDATION_NOT_PORTABLE,
                        message = "Schema uses custom validators that are not portable"
                    )
                )
            )
        }
        return AnyValiDocument(root = exportNode())
    }

    // ---- Describe & Metadata helpers ----

    protected fun applyDescribe(description: String, opts: DescribeOptions? = null) {
        if (schemaMetadata == null) schemaMetadata = mutableMapOf()
        schemaMetadata!!["description"] = description
        opts?.let {
            it.title?.let { v -> schemaMetadata!!["title"] = v }
            it.deprecated?.let { v -> schemaMetadata!!["deprecated"] = v }
            it.deprecatedMessage?.let { v ->
                require(it.deprecated == true) { "describe(): deprecatedMessage requires deprecated to be true" }
                schemaMetadata!!["deprecatedMessage"] = v
            }
            it.notStable?.let { v -> schemaMetadata!!["notStable"] = v }
            it.since?.let { v -> schemaMetadata!!["since"] = v }
            it.sensitive?.let { v -> schemaMetadata!!["sensitive"] = v }
            it.readonly?.let { v -> schemaMetadata!!["readonly"] = v }
            it.writeonly?.let { v -> schemaMetadata!!["writeonly"] = v }
            require(!(it.readonly == true && it.writeonly == true)) {
                "describe(): readonly and writeonly cannot both be true"
            }
            it.examples?.let { v -> schemaMetadata!!["examples"] = v }
        }
    }

    protected fun applyMetadata(meta: Map<String, Any?>, replace: Boolean = false) {
        for (key in meta.keys) {
            require(key !in RESERVED_METADATA_KEYS) {
                "metadata(): \"$key\" is a reserved key. Use describe() instead."
            }
        }
        if (replace) {
            val preserved = schemaMetadata?.filter { it.key in RESERVED_METADATA_KEYS }?.toMutableMap() ?: mutableMapOf()
            preserved.putAll(meta)
            schemaMetadata = preserved
        } else {
            if (schemaMetadata == null) schemaMetadata = mutableMapOf()
            schemaMetadata!!.putAll(meta)
        }
    }

    protected fun addMetadataToNode(builder: JsonObjectBuilder) {
        schemaMetadata?.takeIf { it.isNotEmpty() }?.let { meta ->
            builder.put("metadata", buildJsonObject {
                for ((k, v) in meta) {
                    when (v) {
                        is String -> put(k, v)
                        is Boolean -> put(k, v)
                        is Number -> put(k, JsonPrimitive(v))
                        is List<*> -> put(k, buildJsonArray {
                            for (item in v) {
                                when (item) {
                                    is String -> add(JsonPrimitive(item))
                                    is Boolean -> add(JsonPrimitive(item))
                                    is Number -> add(JsonPrimitive(item))
                                    null -> add(JsonNull)
                                    else -> add(JsonPrimitive(item.toString()))
                                }
                            }
                        })
                        null -> put(k, JsonNull)
                        else -> put(k, v.toString())
                    }
                }
            })
        }
    }

    companion object {
        private val RESERVED_METADATA_KEYS = setOf(
            "title", "description", "deprecated", "deprecatedMessage",
            "notStable", "since", "sensitive", "readonly", "writeonly", "examples"
        )

        fun getJsonTypeName(value: Any?): String = when (value) {
            null -> "null"
            is Boolean -> "boolean"
            is String -> "string"
            is Long, is Int, is Short, is Byte -> "number"
            is Double, is Float -> "number"
            is Number -> "number"
            is List<*> -> "array"
            is Map<*, *> -> "object"
            else -> value::class.simpleName ?: "unknown"
        }

        fun isIntegerValue(value: Any?): Boolean = when (value) {
            is Long, is Int, is Short, is Byte -> true
            is Double -> value == value.toLong().toDouble() && !value.isNaN() && !value.isInfinite()
            is Float -> value == value.toLong().toFloat() && !value.isNaN() && !value.isInfinite()
            else -> false
        }

        fun toDouble(value: Any?): Double = when (value) {
            is Double -> value
            is Float -> value.toDouble()
            is Long -> value.toDouble()
            is Int -> value.toDouble()
            is Short -> value.toDouble()
            is Byte -> value.toDouble()
            is Number -> value.toDouble()
            else -> throw IllegalArgumentException("Cannot convert to double: $value")
        }

        fun toLong(value: Any?): Long = when (value) {
            is Long -> value
            is Int -> value.toLong()
            is Short -> value.toLong()
            is Byte -> value.toLong()
            is Double -> value.toLong()
            is Float -> value.toLong()
            is Number -> value.toLong()
            else -> throw IllegalArgumentException("Cannot convert to long: $value")
        }
    }
}

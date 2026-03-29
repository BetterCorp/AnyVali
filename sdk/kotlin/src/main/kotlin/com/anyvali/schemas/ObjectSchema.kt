package com.anyvali.schemas

import com.anyvali.*
import kotlinx.serialization.json.*

data class ObjectSchema(
    val properties: Map<String, Schema> = emptyMap(),
    val required: Set<String> = emptySet(),
    val unknownKeys: UnknownKeyMode = UnknownKeyMode.REJECT
) : Schema() {
    override val kind: String = "object"

    fun unknownKeys(mode: UnknownKeyMode) = copy(unknownKeys = mode)

    override fun safeParseWithContext(input: Any?, ctx: ValidationContext): ParseResult {
        if (input !is Map<*, *>) {
            return ParseResult.Failure(
                listOf(
                    ValidationIssue(
                        code = IssueCodes.INVALID_TYPE,
                        path = ctx.path,
                        expected = "object",
                        received = getJsonTypeName(input)
                    )
                )
            )
        }

        @Suppress("UNCHECKED_CAST")
        val inputMap = input as Map<String, Any?>
        val issues = mutableListOf<ValidationIssue>()
        val result = mutableMapOf<String, Any?>()

        // Check required fields
        for (reqKey in required) {
            if (!inputMap.containsKey(reqKey)) {
                val propSchema = properties[reqKey]
                // Check for default
                val defaultVal = getDefault(propSchema)
                if (defaultVal != null) {
                    // Will be handled below
                } else {
                    issues.add(
                        ValidationIssue(
                            code = IssueCodes.REQUIRED,
                            path = ctx.path + reqKey,
                            expected = propSchema?.kind ?: "unknown",
                            received = "undefined"
                        )
                    )
                }
            }
        }

        // Validate known properties
        for ((key, schema) in properties) {
            val childCtx = ctx.child(key)

            if (inputMap.containsKey(key)) {
                val childResult = schema.safeParseWithContext(inputMap[key], childCtx)
                when (childResult) {
                    is ParseResult.Success -> result[key] = childResult.value
                    is ParseResult.Failure -> issues.addAll(childResult.issues)
                }
            } else {
                // Try default
                val defaultVal = getDefault(schema)
                if (defaultVal != null) {
                    val dv = defaultVal.second
                    // Validate the default value
                    val childResult = schema.safeParseWithContext(dv, childCtx)
                    when (childResult) {
                        is ParseResult.Success -> result[key] = childResult.value
                        is ParseResult.Failure -> {
                            issues.add(
                                ValidationIssue(
                                    code = IssueCodes.DEFAULT_INVALID,
                                    path = ctx.path + key,
                                    expected = childResult.issues.firstOrNull()?.expected ?: "",
                                    received = dv?.toString() ?: "null"
                                )
                            )
                        }
                    }
                }
                // else: optional field not present, skip
            }
        }

        // Handle unknown keys
        val knownKeys = properties.keys
        for (key in inputMap.keys) {
            if (key !in knownKeys) {
                when (unknownKeys) {
                    UnknownKeyMode.REJECT -> {
                        issues.add(
                            ValidationIssue(
                                code = IssueCodes.UNKNOWN_KEY,
                                path = ctx.path + key,
                                expected = "undefined",
                                received = key
                            )
                        )
                    }
                    UnknownKeyMode.STRIP -> { /* don't include */ }
                    UnknownKeyMode.ALLOW -> result[key] = inputMap[key]
                }
            }
        }

        return if (issues.isEmpty()) ParseResult.Success(result) else ParseResult.Failure(issues)
    }

    override fun validateValue(value: Any?, ctx: ValidationContext): List<ValidationIssue> {
        val result = safeParseWithContext(value, ctx)
        return if (result is ParseResult.Failure) result.issues else emptyList()
    }

    override fun exportNode(): JsonObject = buildJsonObject {
        put("kind", JsonPrimitive("object"))
        put("properties", JsonObject(properties.mapValues { it.value.exportNode() }))
        put("required", JsonArray(required.sorted().map { JsonPrimitive(it) }))
        put("unknownKeys", JsonPrimitive(unknownKeys.value))
    }

    companion object {
        fun getDefault(schema: Schema?): Pair<String, Any?>? {
            if (schema == null) return null
            return when (schema) {
                is StringSchema -> if (schema.defaultValue !== StringSchema.UNSET) "string" to schema.defaultValue else null
                is IntSchema -> if (schema.defaultValue !== IntSchema.UNSET) "int" to schema.defaultValue else null
                is NumberSchema -> if (schema.defaultValue !== NumberSchema.UNSET) "number" to schema.defaultValue else null
                is BoolSchema -> if (schema.defaultValue !== BoolSchema.UNSET) "bool" to schema.defaultValue else null
                is NullableSchema -> getDefault(schema.inner)
                is OptionalSchema -> getDefault(schema.inner)
                else -> null
            }
        }
    }
}

package com.anyvali

import com.anyvali.schemas.*

typealias SensitiveTransform = (path: List<Any>, value: Any?) -> Any?

/** Validate storage data while treating sensitive values as opaque envelopes. */
fun safeParseEncrypted(schema: Schema<*>, data: Any?): ParseResult<Any?> {
    @Suppress("UNCHECKED_CAST")
    return encryptedSchema(schema).safeParse(data) as ParseResult<Any?>
}

/** Parse plaintext, transform sensitive values, and validate the storage result. */
fun encrypt(schema: Schema<*>, data: Any?, transform: SensitiveTransform): Any? {
    val encrypted = transformSensitive(schema, schema.parse(data), emptyList(), "encrypt", transform, mutableMapOf())
    return safeParseEncrypted(schema, encrypted).getOrThrow()
}

/** Validate storage data, transform sensitive values, and parse plaintext. */
fun decrypt(schema: Schema<*>, data: Any?, transform: SensitiveTransform): Any? {
    val encrypted = safeParseEncrypted(schema, data).getOrThrow()
    return schema.parse(transformSensitive(schema, encrypted, emptyList(), "decrypt", transform, mutableMapOf()))
}

private fun encryptedSchema(schema: Schema<*>): Schema<*> {
    if (schema.schemaMetadata?.get("sensitive") == true && schema !is NullSchema) {
        val marker = StringSchema(minLength = "encrypted:".length + 1, startsWith = "encrypted:")
        return when (schema) {
            is OptionalSchema -> OptionalSchema(marker, schema.defaultValue)
            is NullableSchema -> NullableSchema(marker)
            else -> marker
        }
    }
    return when (schema) {
        is ArraySchema -> schema.copy(items = encryptedSchema(schema.items))
        is TupleSchema -> schema.copy(elements = schema.elements.map(::encryptedSchema))
        is ObjectSchema -> schema.copy(properties = schema.properties.mapValues { encryptedSchema(it.value) })
        is RecordSchema -> schema.copy(values = encryptedSchema(schema.values))
        is UnionSchema -> schema.copy(variants = schema.variants.map(::encryptedSchema))
        is IntersectionSchema -> schema.copy(allOf = schema.allOf.map(::encryptedSchema))
        is OptionalSchema -> schema.copy(inner = encryptedSchema(schema.inner))
        is NullableSchema -> schema.copy(inner = encryptedSchema(schema.inner))
        else -> schema
    }
}

private fun transformSensitive(
    schema: Schema<*>,
    input: Any?,
    path: List<Any>,
    mode: String,
    transform: SensitiveTransform,
    cache: MutableMap<List<Any>, Any?>,
): Any? {
    if (schema.schemaMetadata?.get("sensitive") == true && input != null) {
        if (cache.containsKey(path)) return cache[path]
        val value = if (mode == "encrypt") schema.parse(input) else input
        return transform(path.toList(), value).also { cache[path.toList()] = it }
    }

    return when (schema) {
        is ObjectSchema -> {
            @Suppress("UNCHECKED_CAST")
            val source = input as? Map<String, Any?> ?: return input
            source.toMutableMap().also { output ->
                for ((key, child) in schema.properties) {
                    if (source.containsKey(key)) {
                        output[key] = transformSensitive(child, source[key], path + key, mode, transform, cache)
                    }
                }
            }
        }
        is ArraySchema -> (input as? List<*>)?.mapIndexed { index, value ->
            transformSensitive(schema.items, value, path + index, mode, transform, cache)
        } ?: input
        is TupleSchema -> (input as? List<*>)?.mapIndexed { index, value ->
            transformSensitive(schema.elements[index], value, path + index, mode, transform, cache)
        } ?: input
        is RecordSchema -> {
            @Suppress("UNCHECKED_CAST")
            val source = input as? Map<String, Any?> ?: return input
            source.mapValues { (key, value) ->
                transformSensitive(schema.values, value, path + key, mode, transform, cache)
            }
        }
        is OptionalSchema -> transformSensitive(schema.inner, input, path, mode, transform, cache)
        is NullableSchema -> if (input == null) null else transformSensitive(schema.inner, input, path, mode, transform, cache)
        is UnionSchema -> schema.variants.firstOrNull {
            val result = if (mode == "decrypt") encryptedSchema(it).safeParse(input) else it.safeParse(input)
            result.isSuccess
        }?.let { transformSensitive(it, input, path, mode, transform, cache) } ?: input
        is IntersectionSchema -> schema.allOf.fold(input) { value, child ->
            transformSensitive(child, value, path, mode, transform, cache)
        }
        else -> input
    }
}

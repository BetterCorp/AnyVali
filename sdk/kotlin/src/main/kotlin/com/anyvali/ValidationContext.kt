package com.anyvali

data class ValidationContext(
    val path: List<Any> = emptyList(),
    val definitions: Map<String, Schema<*>> = emptyMap(),
    val inheritedUnknownKeys: UnknownKeyMode? = null
) {
    fun child(segment: Any): ValidationContext =
        copy(path = path + segment)
}

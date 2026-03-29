package com.anyvali

data class ValidationIssue(
    val code: String,
    val message: String = "",
    val path: List<Any> = emptyList(),
    val expected: String = "",
    val received: String = "",
    val meta: Map<String, Any?> = emptyMap()
)

package com.anyvali

class ValidationError(
    val issues: List<ValidationIssue>
) : Exception("Validation failed with ${issues.size} issue(s): ${issues.joinToString("; ") { "${it.code} at ${it.path}" }}")

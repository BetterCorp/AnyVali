package com.anyvali

sealed class ParseResult {
    data class Success(val value: Any?) : ParseResult()
    data class Failure(val issues: List<ValidationIssue>) : ParseResult()

    val isSuccess: Boolean get() = this is Success
    val isFailure: Boolean get() = this is Failure

    fun getOrNull(): Any? = when (this) {
        is Success -> value
        is Failure -> null
    }

    fun getOrThrow(): Any? = when (this) {
        is Success -> value
        is Failure -> throw ValidationError(issues)
    }

    fun issuesOrEmpty(): List<ValidationIssue> = when (this) {
        is Success -> emptyList()
        is Failure -> issues
    }
}

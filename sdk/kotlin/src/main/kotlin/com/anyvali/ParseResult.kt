package com.anyvali

sealed class ParseResult<out T> {
    data class Success<T>(val value: T) : ParseResult<T>()
    data class Failure(val issues: List<ValidationIssue>) : ParseResult<Nothing>()

    val isSuccess: Boolean get() = this is Success
    val isFailure: Boolean get() = this is Failure

    fun getOrNull(): T? = when (this) {
        is Success -> value
        is Failure -> null
    }

    fun getOrThrow(): T = when (this) {
        is Success -> value
        is Failure -> throw ValidationError(issues)
    }

    fun issuesOrEmpty(): List<ValidationIssue> = when (this) {
        is Success -> emptyList()
        is Failure -> issues
    }
}

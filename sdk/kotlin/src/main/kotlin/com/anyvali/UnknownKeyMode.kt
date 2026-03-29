package com.anyvali

enum class UnknownKeyMode(val value: String) {
    REJECT("reject"),
    STRIP("strip"),
    ALLOW("allow");

    companion object {
        fun fromValue(value: String): UnknownKeyMode =
            entries.first { it.value == value }
    }
}

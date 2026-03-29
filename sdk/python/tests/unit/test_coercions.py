"""Unit tests for coercion behavior."""

from __future__ import annotations

import anyvali as v


class TestCoercions:
    def test_string_to_int(self):
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("42")
        assert result.success
        assert result.data == 42

    def test_string_to_int_fails(self):
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("not-a-number")
        assert not result.success
        assert any(i.code == v.COERCION_FAILED for i in result.issues)

    def test_string_to_number(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse("3.14")
        assert result.success
        assert result.data == 3.14

    def test_string_to_bool_true(self):
        schema = v.bool_().coerce(to_bool=True)
        result = schema.safe_parse("true")
        assert result.success
        assert result.data is True

    def test_string_to_bool_false(self):
        schema = v.bool_().coerce(to_bool=True)
        result = schema.safe_parse("false")
        assert result.success
        assert result.data is False

    def test_string_trim(self):
        schema = v.string().coerce(trim=True)
        result = schema.safe_parse("  hello  ")
        assert result.success
        assert result.data == "hello"

    def test_string_lower(self):
        schema = v.string().coerce(lower=True)
        result = schema.safe_parse("HELLO")
        assert result.success
        assert result.data == "hello"

    def test_string_upper(self):
        schema = v.string().coerce(upper=True)
        result = schema.safe_parse("hello")
        assert result.success
        assert result.data == "HELLO"

    def test_coercion_before_validation(self):
        schema = v.string().min_length(3).coerce(trim=True)
        # After trimming "  ab  " -> "ab", which is < 3 chars
        result = schema.safe_parse("  ab  ")
        assert not result.success

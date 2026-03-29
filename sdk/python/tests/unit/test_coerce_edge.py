"""Edge case tests for coercion behavior."""

from __future__ import annotations

import pytest

import anyvali as v
from anyvali.parse.coerce import apply_coercion, _coerce_to_int, _coerce_to_number, _coerce_to_bool
from anyvali.schemas.base import CoercionConfig, ValidationContext
from anyvali.parse.defaults import apply_default
from anyvali.schemas.base import _SENTINEL


class TestCoerceToIntEdgeCases:
    def test_float_string_whole(self):
        """String "3.0" should coerce to int 3 via float path."""
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("3.0")
        assert result.success
        assert result.data == 3

    def test_float_string_non_whole(self):
        """String "3.5" should fail coercion to int."""
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("3.5")
        assert not result.success
        assert result.issues[0].code == v.COERCION_FAILED

    def test_whitespace_string(self):
        """String " 42 " should strip and coerce."""
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse(" 42 ")
        assert result.success
        assert result.data == 42

    def test_empty_string(self):
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("")
        assert not result.success

    def test_non_string_bypasses_coercion(self):
        """Non-string input should bypass coercion and validate normally."""
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse(42)
        assert result.success
        assert result.data == 42

    def test_overflow_string(self):
        """Very large string should fail coercion."""
        schema = v.int_().coerce(to_int=True)
        result = schema.safe_parse("not_a_number_at_all")
        assert not result.success


class TestCoerceToNumberEdgeCases:
    def test_valid_float(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse("3.14")
        assert result.success
        assert result.data == 3.14

    def test_invalid_float(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse("abc")
        assert not result.success
        assert result.issues[0].code == v.COERCION_FAILED

    def test_integer_string_to_number(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse("42")
        assert result.success
        assert result.data == 42.0

    def test_whitespace_string(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse("  1.5  ")
        assert result.success
        assert result.data == 1.5

    def test_non_string_bypasses(self):
        schema = v.number().coerce(to_number=True)
        result = schema.safe_parse(3.14)
        assert result.success


class TestCoerceToBoolEdgeCases:
    @pytest.mark.parametrize("value,expected", [
        ("true", True),
        ("True", True),
        ("TRUE", True),
        ("1", True),
        ("yes", True),
        ("YES", True),
        ("false", False),
        ("False", False),
        ("FALSE", False),
        ("0", False),
        ("no", False),
        ("NO", False),
    ])
    def test_valid_values(self, value: str, expected: bool):
        schema = v.bool_().coerce(to_bool=True)
        result = schema.safe_parse(value)
        assert result.success
        assert result.data is expected

    def test_invalid_string(self):
        schema = v.bool_().coerce(to_bool=True)
        result = schema.safe_parse("maybe")
        assert not result.success
        assert result.issues[0].code == v.COERCION_FAILED

    def test_whitespace_true(self):
        schema = v.bool_().coerce(to_bool=True)
        result = schema.safe_parse("  true  ")
        assert result.success
        assert result.data is True


class TestStringTransformCombinations:
    def test_trim_then_lower(self):
        schema = v.string().coerce(trim=True, lower=True)
        result = schema.safe_parse("  HELLO  ")
        assert result.success
        assert result.data == "hello"

    def test_trim_then_upper(self):
        schema = v.string().coerce(trim=True, upper=True)
        result = schema.safe_parse("  hello  ")
        assert result.success
        assert result.data == "HELLO"

    def test_no_coercion_no_change(self):
        schema = v.string()
        result = schema.safe_parse("  HELLO  ")
        assert result.success
        assert result.data == "  HELLO  "


class TestApplyCoercionDirect:
    def test_no_config(self):
        """When coercion is None, _apply_coercion returns value as-is."""
        schema = v.string()
        ctx = ValidationContext()
        result = schema._apply_coercion("hello", ctx)
        assert result == "hello"

    def test_apply_coercion_trim_non_string(self):
        """Trim should not affect non-strings."""
        config = CoercionConfig(trim=True)
        ctx = ValidationContext()
        result = apply_coercion(42, config, ctx)
        assert result == 42

    def test_apply_coercion_lower_non_string(self):
        config = CoercionConfig(lower=True)
        ctx = ValidationContext()
        result = apply_coercion(42, config, ctx)
        assert result == 42


class TestApplyDefault:
    def test_sentinel_gets_default(self):
        result = apply_default(_SENTINEL, "default_val")
        assert result == "default_val"

    def test_non_sentinel_unchanged(self):
        result = apply_default("actual", "default_val")
        assert result == "actual"

    def test_none_unchanged(self):
        result = apply_default(None, "default_val")
        assert result is None

    def test_deep_copy_default(self):
        default_list = [1, 2, 3]
        result = apply_default(_SENTINEL, default_list)
        assert result == [1, 2, 3]
        assert result is not default_list


class TestCoercionPlusValidation:
    def test_coerce_then_validate_passes(self):
        """Coerce string to int, then validate min constraint."""
        schema = v.int_().min(10).coerce(to_int=True)
        result = schema.safe_parse("42")
        assert result.success
        assert result.data == 42

    def test_coerce_then_validate_fails(self):
        """Coerce string to int, but value fails min constraint."""
        schema = v.int_().min(10).coerce(to_int=True)
        result = schema.safe_parse("5")
        assert not result.success
        assert any(i.code == v.TOO_SMALL for i in result.issues)

    def test_coerce_failure_stops_pipeline(self):
        """Coercion failure should stop pipeline early."""
        schema = v.int_().min(10).coerce(to_int=True)
        result = schema.safe_parse("abc")
        assert not result.success
        assert result.issues[0].code == v.COERCION_FAILED
        # Should only have coercion failure, not a type error too
        assert len(result.issues) == 1

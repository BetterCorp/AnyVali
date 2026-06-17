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


class TestDefaultCoerce:
    """Regression tests for the bare / default form of enabling coercion.

    Mirrors the JS SDK bug: calling ``.coerce()`` with no explicit source
    should enable coercion from the only portable source ("string"), so a
    string input like "3.14" is coerced instead of failing with invalid_type.
    """

    def test_number_default_coerce(self):
        schema = v.number().coerce()
        result = schema.safe_parse("3.14")
        assert result.success
        assert result.data == 3.14

    def test_int_default_coerce(self):
        schema = v.int_().coerce()
        result = schema.safe_parse("42")
        assert result.success
        assert result.data == 42

    def test_bool_default_coerce_true(self):
        schema = v.bool_().coerce()
        result = schema.safe_parse("true")
        assert result.success
        assert result.data is True

    def test_bool_default_coerce_false(self):
        schema = v.bool_().coerce()
        result = schema.safe_parse("false")
        assert result.success
        assert result.data is False

    def test_object_numeric_fields_default_coerce(self):
        schema = v.object_(
            {
                "lumpSum": v.number().coerce(),
                "monthlyContributions": v.number().coerce(),
                "investmentTerm": v.number().coerce(),
            }
        )
        result = schema.safe_parse(
            {
                "lumpSum": "1000000",
                "monthlyContributions": "1000",
                "investmentTerm": "20",
            }
        )
        assert result.success
        assert result.data == {
            "lumpSum": 1000000.0,
            "monthlyContributions": 1000.0,
            "investmentTerm": 20.0,
        }


def _assert_coercion_failed(result):
    assert result.success is False
    assert any(i.code == v.COERCION_FAILED for i in result.issues)


class TestCoercionMatrixString:
    """Canonical coercion matrix, all FROM STRING, via the bare ``.coerce()``.

    Every row mirrors the cross-SDK portable grammar (spec 5.1): ASCII only,
    no Unicode digits / underscores / leading '+' / hex / inf / yes-no.
    """

    # ── string -> int : ASCII ^-?\d+$ (trimmed) ──────────────────────
    def test_int_accept_plain(self):
        result = v.int_().coerce().safe_parse("42")
        assert result.success
        assert result.data == 42

    def test_int_accept_surrounding_whitespace(self):
        result = v.int_().coerce().safe_parse("  42  ")
        assert result.success
        assert result.data == 42

    def test_int_accept_negative(self):
        result = v.int_().coerce().safe_parse("-7")
        assert result.success
        assert result.data == -7

    def test_int_reject_decimal(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("3.14"))

    def test_int_reject_hex(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("0x10"))

    def test_int_reject_underscore_grouping(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("1_000"))

    def test_int_reject_leading_plus(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("+5"))

    def test_int_reject_infinity(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("Infinity"))

    def test_int_reject_empty(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse(""))

    def test_int_reject_non_numeric(self):
        _assert_coercion_failed(v.int_().coerce().safe_parse("abc"))

    # ── string -> number : ASCII decimal float incl exponent (trimmed) ─
    def test_number_accept_decimal(self):
        result = v.number().coerce().safe_parse("3.14")
        assert result.success
        assert result.data == 3.14

    def test_number_accept_signed_exponent(self):
        result = v.number().coerce().safe_parse("-1.5e3")
        assert result.success
        assert result.data == -1500.0

    def test_number_accept_surrounding_whitespace(self):
        result = v.number().coerce().safe_parse("  2  ")
        assert result.success
        assert result.data == 2.0

    def test_number_accept_zero(self):
        result = v.number().coerce().safe_parse("0")
        assert result.success
        assert result.data == 0.0

    def test_number_reject_hex(self):
        _assert_coercion_failed(v.number().coerce().safe_parse("0x10"))

    def test_number_reject_infinity(self):
        _assert_coercion_failed(v.number().coerce().safe_parse("Infinity"))

    def test_number_reject_nan(self):
        _assert_coercion_failed(v.number().coerce().safe_parse("NaN"))

    def test_number_reject_empty(self):
        _assert_coercion_failed(v.number().coerce().safe_parse(""))

    def test_number_reject_underscore_grouping(self):
        _assert_coercion_failed(v.number().coerce().safe_parse("1_000"))

    def test_number_reject_non_numeric(self):
        _assert_coercion_failed(v.number().coerce().safe_parse("abc"))

    # ── string -> bool : trim + case-insensitive ─────────────────────
    def test_bool_accept_true_word(self):
        result = v.bool_().coerce().safe_parse("true")
        assert result.success
        assert result.data is True

    def test_bool_accept_true_upper(self):
        result = v.bool_().coerce().safe_parse("TRUE")
        assert result.success
        assert result.data is True

    def test_bool_accept_true_one(self):
        result = v.bool_().coerce().safe_parse("1")
        assert result.success
        assert result.data is True

    def test_bool_accept_false_word(self):
        result = v.bool_().coerce().safe_parse("false")
        assert result.success
        assert result.data is False

    def test_bool_accept_false_zero(self):
        result = v.bool_().coerce().safe_parse("0")
        assert result.success
        assert result.data is False

    def test_bool_reject_yes(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("yes"))

    def test_bool_reject_no(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("no"))

    def test_bool_reject_on(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("on"))

    def test_bool_reject_off(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("off"))

    def test_bool_reject_t(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("t"))

    def test_bool_reject_f(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("f"))

    def test_bool_reject_two(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse("2"))

    def test_bool_reject_empty(self):
        _assert_coercion_failed(v.bool_().coerce().safe_parse(""))


class TestStringTransforms:
    """String-kind transforms: trim, lower, upper; chainable.

    The bare ``.coerce()`` must NOT auto-coerce a string schema; only the
    explicit transform flags apply (target kind == string).
    """

    def test_trim(self):
        result = v.string().coerce(trim=True).safe_parse("  hello  ")
        assert result.success
        assert result.data == "hello"

    def test_lower(self):
        result = v.string().coerce(lower=True).safe_parse("HELLO")
        assert result.success
        assert result.data == "hello"

    def test_upper(self):
        result = v.string().coerce(upper=True).safe_parse("hello")
        assert result.success
        assert result.data == "HELLO"

    def test_chained_trim_lower(self):
        result = v.string().coerce(trim=True, lower=True).safe_parse("  HELLO  ")
        assert result.success
        assert result.data == "hello"

    def test_chained_trim_upper(self):
        result = v.string().coerce(trim=True, upper=True).safe_parse("  hello  ")
        assert result.success
        assert result.data == "HELLO"

    def test_bare_coerce_on_string_is_noop(self):
        # No explicit transform + string target => no auto-coercion.
        result = v.string().coerce().safe_parse("  Hello  ")
        assert result.success
        assert result.data == "  Hello  "

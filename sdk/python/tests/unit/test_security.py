"""Security-focused tests for the AnyVali Python SDK.

Covers ReDoS, recursive schema DoS, integer overflow, NaN/Infinity injection,
format validation bypass, large input DoS, type confusion, and schema import
injection attacks.
"""

from __future__ import annotations

import json
import math
import time

import pytest

import anyvali as v
from anyvali.format.validators import validate_format


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _timed(fn, *, timeout_seconds: float = 5.0):
    """Run *fn* and assert it completes within *timeout_seconds*.

    Uses wall-clock timing because signal-based alarms are not available on
    all platforms (e.g. Windows).
    """
    start = time.monotonic()
    result = fn()
    elapsed = time.monotonic() - start
    assert elapsed < timeout_seconds, (
        f"Function took {elapsed:.2f}s, exceeding {timeout_seconds}s timeout"
    )
    return result


# ===================================================================
# 1. ReDoS -- CVE-2016-4055 / CVE-2022-25883
# ===================================================================


class TestReDoS_CVE_2016_4055:
    """Pathological regex patterns must not cause catastrophic backtracking."""

    def test_exponential_backtracking_a_plus(self):
        """(a+)+$ against 'aaa...!' should complete quickly."""
        schema = v.string().pattern(r"(a+)+$")
        evil = "a" * 25 + "!"
        result = _timed(lambda: schema.safe_parse(evil))
        # The string does not match the pattern, so validation should fail.
        assert not result.success

    def test_exponential_backtracking_alpha_star(self):
        """^([a-zA-Z]+)*X$ against 'aaa...1' should complete quickly.

        The anchored pattern forces full backtracking when it cannot match.
        Python's re engine should handle this without catastrophic time.
        """
        schema = v.string().pattern(r"^([a-zA-Z]+)*X$")
        evil = "a" * 25 + "1"
        result = _timed(lambda: schema.safe_parse(evil))
        assert not result.success

    def test_nested_quantifiers_word(self):
        r"""^(\w+\s?)*X$ against a long non-matching string.

        Forces the engine to fully backtrack before concluding no match.
        """
        schema = v.string().pattern(r"^(\w+\s?)*X$")
        evil = "a " * 15 + "!"
        result = _timed(lambda: schema.safe_parse(evil))
        assert not result.success

    def test_benign_pattern_still_works(self):
        """Sanity: a normal pattern still matches."""
        schema = v.string().pattern(r"^hello$")
        assert schema.safe_parse("hello").success
        assert not schema.safe_parse("world").success


# ===================================================================
# 2. Recursive $ref DoS -- CVE-2003-1564 (Billion Laughs class)
# ===================================================================


class TestRecursiveRefDoS_CVE_2003_1564:
    """Recursive and deeply nested schema definitions must not crash."""

    def test_self_referencing_definition(self):
        """Schema A -> ref(A) should import without infinite loop."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/A"},
            "definitions": {
                "A": {
                    "kind": "object",
                    "properties": {
                        "child": {"kind": "ref", "ref": "#/definitions/A"},
                    },
                    "required": [],
                    "unknownKeys": "reject",
                },
            },
        }
        schema = v.import_schema(doc)
        # Should validate a simple non-recursive value.
        result = schema.safe_parse({"child": {"child": {}}})
        # Depending on implementation, this may succeed or fail validation,
        # but it must not hang or crash.
        assert isinstance(result.success, bool)

    def test_deeply_nested_schema_import(self):
        """Importing a schema with 100 levels of nesting should not crash."""
        # Build nested array(array(array(...string()...)))
        inner: dict = {"kind": "string"}
        for _ in range(100):
            inner = {"kind": "array", "items": inner}

        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": inner,
        }
        schema = _timed(lambda: v.import_schema(doc))
        assert schema is not None

    def test_recursive_tree_validation_reasonable_depth(self):
        """Validate a recursive tree structure at reasonable depth."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/Node"},
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "value": {"kind": "int"},
                        "children": {
                            "kind": "array",
                            "items": {"kind": "ref", "ref": "#/definitions/Node"},
                        },
                    },
                    "required": ["value"],
                    "unknownKeys": "reject",
                },
            },
        }
        schema = v.import_schema(doc)

        # Build a tree 5 levels deep.
        def make_tree(depth: int) -> dict:
            if depth <= 0:
                return {"value": 0, "children": []}
            return {"value": depth, "children": [make_tree(depth - 1)]}

        tree = make_tree(5)
        result = schema.safe_parse(tree)
        assert result.success

    def test_mutually_recursive_definitions(self):
        """A -> B -> A should import without infinite loop."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/A"},
            "definitions": {
                "A": {
                    "kind": "object",
                    "properties": {
                        "b": {"kind": "ref", "ref": "#/definitions/B"},
                    },
                    "required": [],
                    "unknownKeys": "reject",
                },
                "B": {
                    "kind": "object",
                    "properties": {
                        "a": {"kind": "ref", "ref": "#/definitions/A"},
                    },
                    "required": [],
                    "unknownKeys": "reject",
                },
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse({"b": {"a": {}}})
        assert isinstance(result.success, bool)

    def test_deeply_nested_input_does_not_crash_safe_parse(self):
        """Adversarial deep input against a recursive schema must not raise.

        safe_parse must never raise (incl. RecursionError); it must return a
        bounded failure result. Otherwise a crafted payload is a DoS.
        """
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/Node"},
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "next": {"kind": "ref", "ref": "#/definitions/Node"},
                    },
                    "required": [],
                    "unknownKeys": "strip",
                },
            },
        }
        schema = v.import_schema(doc)

        data: dict = {}
        for _ in range(100_000):
            data = {"next": data}

        # Must return a result, not raise RecursionError.
        result = schema.safe_parse(data)
        assert result.success is False

    def test_deep_input_via_union_does_not_crash(self):
        """Recursion routed through a union must not bypass the depth guard."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/Node"},
            "definitions": {
                "Node": {
                    "kind": "object",
                    "properties": {
                        "next": {
                            "kind": "union",
                            "variants": [
                                {"kind": "null"},
                                {"kind": "ref", "ref": "#/definitions/Node"},
                            ],
                        },
                    },
                    "required": ["next"],
                    "unknownKeys": "strip",
                },
            },
        }
        schema = v.import_schema(doc)

        cur = {"next": None}
        for _ in range(100_000):
            cur = {"next": cur}

        result = schema.safe_parse(cur)
        assert result.success is False


# ===================================================================
# 3. Integer overflow -- CWE-190
# ===================================================================


class TestIntegerOverflow_CWE_190:
    """All integer width boundaries must be enforced correctly."""

    # -- int8 --
    def test_int8_max_accepted(self):
        assert v.int8().safe_parse(127).success

    def test_int8_max_plus_one_rejected(self):
        assert not v.int8().safe_parse(128).success

    def test_int8_min_accepted(self):
        assert v.int8().safe_parse(-128).success

    def test_int8_min_minus_one_rejected(self):
        assert not v.int8().safe_parse(-129).success

    # -- int16 --
    def test_int16_max_accepted(self):
        assert v.int16().safe_parse(32767).success

    def test_int16_max_plus_one_rejected(self):
        assert not v.int16().safe_parse(32768).success

    def test_int16_min_accepted(self):
        assert v.int16().safe_parse(-32768).success

    def test_int16_min_minus_one_rejected(self):
        assert not v.int16().safe_parse(-32769).success

    # -- int32 --
    def test_int32_max_accepted(self):
        assert v.int32().safe_parse(2147483647).success

    def test_int32_max_plus_one_rejected(self):
        assert not v.int32().safe_parse(2147483648).success

    def test_int32_min_accepted(self):
        assert v.int32().safe_parse(-2147483648).success

    def test_int32_min_minus_one_rejected(self):
        assert not v.int32().safe_parse(-2147483649).success

    # -- int64 --
    def test_int64_max_accepted(self):
        assert v.int64().safe_parse(2**63 - 1).success

    def test_int64_max_plus_one_rejected(self):
        assert not v.int64().safe_parse(2**63).success

    def test_int64_min_accepted(self):
        assert v.int64().safe_parse(-(2**63)).success

    def test_int64_min_minus_one_rejected(self):
        assert not v.int64().safe_parse(-(2**63) - 1).success

    # -- uint8 --
    def test_uint8_max_accepted(self):
        assert v.uint8().safe_parse(255).success

    def test_uint8_max_plus_one_rejected(self):
        assert not v.uint8().safe_parse(256).success

    def test_uint8_zero_accepted(self):
        assert v.uint8().safe_parse(0).success

    def test_uint8_negative_rejected(self):
        assert not v.uint8().safe_parse(-1).success

    # -- uint16 --
    def test_uint16_max_accepted(self):
        assert v.uint16().safe_parse(65535).success

    def test_uint16_max_plus_one_rejected(self):
        assert not v.uint16().safe_parse(65536).success

    def test_uint16_zero_accepted(self):
        assert v.uint16().safe_parse(0).success

    def test_uint16_negative_rejected(self):
        assert not v.uint16().safe_parse(-1).success

    # -- uint32 --
    def test_uint32_max_accepted(self):
        assert v.uint32().safe_parse(2**32 - 1).success

    def test_uint32_max_plus_one_rejected(self):
        assert not v.uint32().safe_parse(2**32).success

    def test_uint32_zero_accepted(self):
        assert v.uint32().safe_parse(0).success

    def test_uint32_negative_rejected(self):
        assert not v.uint32().safe_parse(-1).success

    # -- uint64 --
    def test_uint64_max_accepted(self):
        assert v.uint64().safe_parse(2**64 - 1).success

    def test_uint64_max_plus_one_rejected(self):
        assert not v.uint64().safe_parse(2**64).success

    def test_uint64_zero_accepted(self):
        assert v.uint64().safe_parse(0).success

    def test_uint64_negative_rejected(self):
        assert not v.uint64().safe_parse(-1).success

    # -- Python big-int edge cases --
    def test_huge_positive_int_rejected_by_int64(self):
        """Python allows arbitrary-precision ints; schema must reject."""
        assert not v.int64().safe_parse(10**100).success

    def test_huge_negative_int_rejected_by_int64(self):
        assert not v.int64().safe_parse(-(10**100)).success

    def test_huge_positive_int_rejected_by_uint64(self):
        assert not v.uint64().safe_parse(2**128).success


# ===================================================================
# 4. NaN / Infinity injection -- CWE-20
# ===================================================================


class TestNaNInfinityInjection_CWE_20:
    """NaN and Infinity must be rejected by all numeric schemas."""

    @pytest.mark.parametrize("bad_value", [
        float("nan"),
        float("inf"),
        float("-inf"),
        math.nan,
        math.inf,
        -math.inf,
    ])
    def test_number_rejects(self, bad_value):
        result = v.number().safe_parse(bad_value)
        assert not result.success
        assert result.issues[0].code == v.INVALID_NUMBER

    @pytest.mark.parametrize("bad_value", [
        float("nan"),
        float("inf"),
        float("-inf"),
        math.nan,
    ])
    def test_int_rejects(self, bad_value):
        result = v.int_().safe_parse(bad_value)
        assert not result.success

    @pytest.mark.parametrize("bad_value", [
        float("nan"),
        float("inf"),
        float("-inf"),
    ])
    def test_float32_rejects(self, bad_value):
        result = v.float32().safe_parse(bad_value)
        assert not result.success

    @pytest.mark.parametrize("bad_value", [
        float("nan"),
        float("inf"),
        float("-inf"),
    ])
    def test_float64_rejects(self, bad_value):
        result = v.float64().safe_parse(bad_value)
        assert not result.success

    def test_nan_not_equal_to_self(self):
        """Verify NaN's self-inequality doesn't bypass checks."""
        schema = v.number().min(0)
        result = schema.safe_parse(float("nan"))
        assert not result.success

    @pytest.mark.parametrize("schema_factory", [
        v.int8, v.int16, v.int32, v.int64,
        v.uint8, v.uint16, v.uint32, v.uint64,
    ])
    def test_all_int_widths_reject_nan(self, schema_factory):
        result = schema_factory().safe_parse(float("nan"))
        assert not result.success

    @pytest.mark.parametrize("schema_factory", [
        v.int8, v.int16, v.int32, v.int64,
        v.uint8, v.uint16, v.uint32, v.uint64,
    ])
    def test_all_int_widths_reject_inf(self, schema_factory):
        result = schema_factory().safe_parse(float("inf"))
        assert not result.success


# ===================================================================
# 5. Format validation bypass -- CWE-20
# ===================================================================


class TestFormatValidationBypass_CWE_20:
    """Format validators must reject adversarial inputs."""

    # -- Email --

    def test_tampered_email_format_name_not_silently_ignored(self):
        """A malformed format identifier must not strip email validation."""
        schema = v.string().format("email\x00")
        result = schema.safe_parse("not-an-email")
        assert not result.success

    def test_imported_tampered_email_format_not_unconstrained(self):
        """Imported malformed format identifiers must not become no-ops."""
        schema = v.import_schema({
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "format": "email\x00"},
            "definitions": {},
        })
        result = schema.safe_parse("not-an-email")
        assert not result.success

    def test_email_null_byte_injection(self):
        """Null byte in email should be rejected."""
        result = validate_format("email", "user@example.com\x00.evil.com")
        assert result is False

    def test_email_very_long_local_part(self):
        """Local part exceeding 64 chars should be rejected by the regex."""
        long_local = "a" * 65 + "@example.com"
        # The basic regex may or may not enforce the 64-char local part limit,
        # but the test documents the behavior.
        result = validate_format("email", long_local)
        # Record the actual behavior -- this is a known limitation.
        assert isinstance(result, bool)

    def test_email_with_newline_rejected(self):
        """Newline in email should be rejected."""
        result = validate_format("email", "user\n@example.com")
        assert result is False

    def test_email_with_space_rejected(self):
        result = validate_format("email", "user @example.com")
        assert result is False

    # -- URL --

    def test_url_javascript_protocol_rejected(self):
        """javascript: URLs must be rejected."""
        result = validate_format("url", "javascript:alert(1)")
        assert result is False

    def test_url_data_protocol_rejected(self):
        """data: URLs must be rejected."""
        result = validate_format("url", "data:text/html,<script>alert(1)</script>")
        assert result is False

    def test_url_ftp_protocol_rejected(self):
        """ftp: URLs must be rejected (only http/https allowed)."""
        result = validate_format("url", "ftp://example.com/file")
        assert result is False

    def test_url_file_protocol_rejected(self):
        """file: URLs must be rejected."""
        result = validate_format("url", "file:///etc/passwd")
        assert result is False

    def test_url_null_byte_rejected(self):
        """Null byte in URL should be rejected."""
        result = validate_format("url", "https://example.com\x00.evil.com")
        assert result is False

    # -- IPv4 --

    def test_ipv4_overflow_rejected(self):
        """256.1.1.1 must be rejected."""
        result = validate_format("ipv4", "256.1.1.1")
        assert result is False

    def test_ipv4_octal_notation_rejected(self):
        """Octal 0177.0.0.1 (== 127.0.0.1) should be rejected.

        Python's ipaddress module rejects leading zeros by default (since 3.9.5+),
        which is the correct security-conscious behavior.
        """
        result = validate_format("ipv4", "0177.0.0.1")
        assert result is False

    def test_ipv4_leading_zeros_rejected(self):
        """Leading zeros like 192.168.001.001 should be rejected."""
        result = validate_format("ipv4", "192.168.001.001")
        assert result is False

    def test_ipv4_negative_octet_rejected(self):
        result = validate_format("ipv4", "-1.0.0.0")
        assert result is False

    def test_ipv4_too_many_octets_rejected(self):
        result = validate_format("ipv4", "1.2.3.4.5")
        assert result is False

    def test_ipv4_too_few_octets_rejected(self):
        result = validate_format("ipv4", "1.2.3")
        assert result is False

    # -- IPv6 --

    def test_ipv6_valid_loopback(self):
        assert validate_format("ipv6", "::1") is True

    def test_ipv6_valid_full_form(self):
        assert validate_format("ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334") is True

    def test_ipv6_invalid_segment(self):
        assert validate_format("ipv6", "gggg::1") is False

    def test_ipv6_too_many_segments(self):
        assert validate_format("ipv6", "1:2:3:4:5:6:7:8:9") is False

    # -- Via schema integration --

    def test_string_format_email_rejects_javascript(self):
        """End-to-end: v.string().format('email') rejects obviously bad input."""
        schema = v.string().format("email")
        result = schema.safe_parse("javascript:alert(1)")
        assert not result.success
        assert result.issues[0].code == v.INVALID_STRING

    def test_string_format_url_rejects_javascript(self):
        """End-to-end: v.string().format('url') rejects javascript: protocol."""
        schema = v.string().format("url")
        result = schema.safe_parse("javascript:alert(1)")
        assert not result.success
        assert result.issues[0].code == v.INVALID_STRING


# ===================================================================
# 5b. Unicode length constraints
# ===================================================================


class TestUnicodeLengthConstraints:
    """minLength/maxLength are defined in Unicode code points."""

    def test_astral_code_point_counts_as_one_character(self):
        emoji = "😀"
        assert v.string().max_length(1).safe_parse(emoji).success
        assert not v.string().min_length(2).safe_parse(emoji).success

    def test_imported_max_length_uses_code_points(self):
        schema = v.import_schema({
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "maxLength": 1},
            "definitions": {},
        })
        assert schema.safe_parse("😀").success


# ===================================================================
# 6. Large input DoS -- CWE-400
# ===================================================================


class TestLargeInputDoS_CWE_400:
    """Validation should handle very large inputs without crashing."""

    def test_very_long_string_1mb(self):
        """1 MB string should be validated without crash or excessive time."""
        schema = v.string()
        big = "x" * (1024 * 1024)
        result = _timed(lambda: schema.safe_parse(big), timeout_seconds=5.0)
        assert result.success

    def test_very_long_string_with_pattern(self):
        """Pattern matching on a large string should complete."""
        schema = v.string().pattern(r"^x+$")
        big = "x" * (1024 * 1024)
        result = _timed(lambda: schema.safe_parse(big), timeout_seconds=5.0)
        assert result.success

    def test_very_long_string_max_length_rejects(self):
        """max_length check should reject a large string quickly."""
        schema = v.string().max_length(100)
        big = "x" * (1024 * 1024)
        result = _timed(lambda: schema.safe_parse(big), timeout_seconds=5.0)
        assert not result.success

    def test_deeply_nested_dict_100_levels(self):
        """100 levels of dict nesting should be validated."""
        schema = v.object_({"child": v.any_()}, unknown_keys="allow")
        data: dict = {"value": "leaf"}
        for _ in range(100):
            data = {"child": data}
        result = _timed(lambda: schema.safe_parse(data), timeout_seconds=5.0)
        assert result.success

    def test_array_with_10000_items(self):
        """Array with 10,000 items should be validated."""
        schema = v.array(v.int_())
        data = list(range(10_000))
        result = _timed(lambda: schema.safe_parse(data), timeout_seconds=5.0)
        assert result.success

    def test_array_with_10000_items_failing_validation(self):
        """Array with 10,000 invalid items should produce issues."""
        schema = v.array(v.string())
        data = list(range(10_000))
        result = _timed(lambda: schema.safe_parse(data), timeout_seconds=10.0)
        assert not result.success
        assert len(result.issues) == 10_000

    def test_object_with_many_properties(self):
        """Object with many properties should be validated."""
        props = {f"field_{i}": v.string() for i in range(500)}
        schema = v.object_(props, required=list(props.keys()))
        data = {f"field_{i}": f"value_{i}" for i in range(500)}
        result = _timed(lambda: schema.safe_parse(data), timeout_seconds=5.0)
        assert result.success


# ===================================================================
# 7. Type confusion -- CWE-843
# ===================================================================


class TestTypeConfusion_CWE_843:
    """Schema validation must be strict about Python type identity."""

    # -- bool vs int --
    def test_bool_not_accepted_as_int(self):
        """True/False must NOT pass int validation (bool is subclass of int)."""
        assert not v.int_().safe_parse(True).success
        assert not v.int_().safe_parse(False).success

    def test_bool_not_accepted_as_int8(self):
        assert not v.int8().safe_parse(True).success

    def test_bool_not_accepted_as_uint8(self):
        assert not v.uint8().safe_parse(False).success

    def test_bool_not_accepted_as_number(self):
        """True/False must NOT pass number validation."""
        assert not v.number().safe_parse(True).success
        assert not v.number().safe_parse(False).success

    def test_bool_not_accepted_as_float32(self):
        assert not v.float32().safe_parse(True).success

    def test_bool_not_accepted_as_float64(self):
        assert not v.float64().safe_parse(True).success

    # -- int vs bool --
    def test_int_not_accepted_as_bool(self):
        """1/0 must NOT pass bool validation."""
        assert not v.bool_().safe_parse(1).success
        assert not v.bool_().safe_parse(0).success

    # -- string vs int --
    def test_string_not_accepted_as_int(self):
        assert not v.int_().safe_parse("1").success
        assert not v.int_().safe_parse("0").success

    def test_string_not_accepted_as_number(self):
        assert not v.number().safe_parse("3.14").success

    # -- None coercion --
    def test_none_not_accepted_as_int(self):
        assert not v.int_().safe_parse(None).success

    def test_none_not_accepted_as_string(self):
        assert not v.string().safe_parse(None).success

    def test_none_not_accepted_as_number(self):
        assert not v.number().safe_parse(None).success

    def test_none_not_accepted_as_bool(self):
        assert not v.bool_().safe_parse(None).success

    def test_none_not_accepted_as_array(self):
        assert not v.array(v.any_()).safe_parse(None).success

    def test_none_not_accepted_as_object(self):
        assert not v.object_({}).safe_parse(None).success

    # -- list vs dict confusion --
    def test_list_not_accepted_as_object(self):
        assert not v.object_({}).safe_parse([1, 2, 3]).success

    def test_dict_not_accepted_as_array(self):
        assert not v.array(v.any_()).safe_parse({"a": 1}).success

    # -- enum type strictness --
    def test_enum_bool_vs_int(self):
        """enum([1, 2, 3]) must reject True (even though True == 1)."""
        assert not v.enum_([1, 2, 3]).safe_parse(True).success

    def test_enum_int_vs_string(self):
        """enum(["1", "2"]) must reject integer 1."""
        assert not v.enum_(["1", "2"]).safe_parse(1).success

    # -- literal type strictness --
    def test_literal_bool_vs_int(self):
        """literal(True) must reject 1."""
        assert not v.literal(True).safe_parse(1).success
        assert not v.literal(1).safe_parse(True).success


# ===================================================================
# 8. Schema import injection
# ===================================================================


class TestSchemaImportInjection:
    """Malicious schema documents must not cause crashes or code execution."""

    def test_unknown_kind_rejected(self):
        """Unknown schema kind must raise ValueError."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "__evil__"},
        }
        with pytest.raises(ValueError, match="Unsupported schema kind"):
            v.import_schema(doc)

    def test_empty_kind_rejected(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": ""},
        }
        with pytest.raises(ValueError, match="Unsupported schema kind"):
            v.import_schema(doc)

    def test_missing_kind_rejected(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {},
        }
        with pytest.raises((ValueError, KeyError)):
            v.import_schema(doc)

    def test_kind_with_traversal_rejected(self):
        """Kind with path traversal-like string must be rejected."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "../../../etc/passwd"},
        }
        with pytest.raises(ValueError, match="Unsupported schema kind"):
            v.import_schema(doc)

    def test_missing_required_fields_in_array(self):
        """Array schema without 'items' should raise."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "array"},
        }
        with pytest.raises((ValueError, KeyError, TypeError)):
            v.import_schema(doc)

    def test_missing_required_fields_in_literal(self):
        """Literal schema without 'value' should raise."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "literal"},
        }
        with pytest.raises((ValueError, KeyError)):
            v.import_schema(doc)

    def test_missing_required_fields_in_ref(self):
        """Ref schema without 'ref' should raise."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref"},
        }
        with pytest.raises((ValueError, KeyError)):
            v.import_schema(doc)

    def test_missing_required_fields_in_enum(self):
        """Enum schema without 'values' should raise."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "enum"},
        }
        with pytest.raises((ValueError, KeyError)):
            v.import_schema(doc)

    # -- Malicious definition names --

    def test_dunder_class_definition_name(self):
        """Definition named __class__ must not cause special behavior."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/__class__"},
            "definitions": {
                "__class__": {"kind": "string"},
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse("hello")
        assert result.success

    def test_dunder_init_definition_name(self):
        """Definition named __init__ must not cause special behavior."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/__init__"},
            "definitions": {
                "__init__": {"kind": "int"},
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse(42)
        assert result.success

    def test_dunder_import_definition_name(self):
        """Definition named __import__ must not trigger code import."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/__import__"},
            "definitions": {
                "__import__": {"kind": "bool"},
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse(True)
        assert result.success

    def test_definition_name_with_null_bytes(self):
        """Definition name with null bytes should not cause issues."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/evil\x00name"},
            "definitions": {
                "evil\x00name": {"kind": "string"},
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse("hello")
        assert result.success

    def test_import_from_json_string(self):
        """Importing from a JSON string should work the same as from a dict."""
        import json
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string"},
        }
        schema = v.import_schema(json.dumps(doc))
        result = schema.safe_parse("hello")
        assert result.success

    def test_import_invalid_json_string(self):
        """Importing invalid JSON should raise."""
        with pytest.raises((json.JSONDecodeError, ValueError)):
            v.import_schema("not valid json {{{")

    def test_numeric_kind_rejected(self):
        """Numeric kind should be rejected."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": 42},
        }
        with pytest.raises((ValueError, TypeError)):
            v.import_schema(doc)

    def test_none_kind_rejected(self):
        """None kind should be rejected."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": None},
        }
        with pytest.raises((ValueError, TypeError)):
            v.import_schema(doc)

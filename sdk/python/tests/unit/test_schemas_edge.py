"""Edge case tests for schema types, wrappers, and validation paths."""

from __future__ import annotations

import math

import pytest

import anyvali as v
from anyvali.schemas.base import (
    ValidationContext,
    _SENTINEL,
    _anyvali_type_name,
)
from anyvali.types import ValidationError


# ── _anyvali_type_name helper ─────────────────────────────────────


class TestAnyvaliTypeName:
    def test_none(self):
        assert _anyvali_type_name(None) == "null"

    def test_bool(self):
        assert _anyvali_type_name(True) == "boolean"

    def test_int(self):
        assert _anyvali_type_name(42) == "integer"

    def test_float(self):
        assert _anyvali_type_name(3.14) == "number"

    def test_string(self):
        assert _anyvali_type_name("hello") == "string"

    def test_list(self):
        assert _anyvali_type_name([1, 2]) == "array"

    def test_dict(self):
        assert _anyvali_type_name({"a": 1}) == "object"

    def test_custom_object(self):
        class Foo:
            pass
        assert _anyvali_type_name(Foo()) == "Foo"


# ── ValidationContext ─────────────────────────────────────────────


class TestValidationContext:
    def test_add_issue(self):
        ctx = ValidationContext()
        ctx.add_issue("invalid_type", "bad", expected="string", received="int")
        assert len(ctx.issues) == 1
        assert ctx.issues[0].code == "invalid_type"
        assert ctx.issues[0].path == []

    def test_add_issue_with_meta(self):
        ctx = ValidationContext()
        ctx.add_issue("custom", "msg", meta={"key": "val"})
        assert ctx.issues[0].meta == {"key": "val"}

    def test_child_inherits_issues(self):
        ctx = ValidationContext()
        child = ctx.child("name")
        child.add_issue("required", "missing")
        # Issues are shared
        assert len(ctx.issues) == 1
        assert ctx.issues[0].path == ["name"]

    def test_nested_child(self):
        ctx = ValidationContext()
        child = ctx.child("items")
        grandchild = child.child(0)
        grandchild.add_issue("invalid_type", "bad")
        assert ctx.issues[0].path == ["items", 0]

    def test_definitions_shared(self):
        defs = {"Foo": v.string()}
        ctx = ValidationContext(definitions=defs)
        child = ctx.child("x")
        assert child.definitions is defs


# ── Optional + Nullable combinations ─────────────────────────────


class TestOptionalNullable:
    def test_optional_nullable(self):
        """optional(nullable(string)) should accept None and absent."""
        schema = v.optional(v.nullable(v.string()))
        # None accepted
        assert schema.safe_parse(None).success
        # String accepted
        assert schema.safe_parse("hello").success
        # Wrong type rejected
        assert not schema.safe_parse(42).success

    def test_nullable_optional(self):
        """nullable(optional(string)) should accept None."""
        schema = v.nullable(v.optional(v.string()))
        assert schema.safe_parse(None).success
        assert schema.safe_parse("hello").success

    def test_chained_optional(self):
        """string().optional() returns OptionalSchema."""
        schema = v.string().optional()
        ctx = ValidationContext()
        result = schema._run_pipeline(_SENTINEL, ctx)
        assert not ctx.issues
        assert result is None

    def test_chained_nullable(self):
        """string().nullable() returns NullableSchema."""
        schema = v.string().nullable()
        result = schema.safe_parse(None)
        assert result.success
        assert result.data is None

    def test_optional_with_default(self):
        """Optional with default should use default when absent."""
        inner = v.string().default("fallback")
        schema = v.optional(inner)
        ctx = ValidationContext()
        result = schema._run_pipeline(_SENTINEL, ctx)
        assert not ctx.issues
        assert result == "fallback"


# ── parse() throwing ──────────────────────────────────────────────


class TestParseThrows:
    def test_string_parse_throws(self):
        with pytest.raises(ValidationError):
            v.string().parse(42)

    def test_int_parse_throws(self):
        with pytest.raises(ValidationError):
            v.int_().parse("not int")

    def test_bool_parse_throws(self):
        with pytest.raises(ValidationError):
            v.bool_().parse(1)

    def test_null_parse_throws(self):
        with pytest.raises(ValidationError):
            v.null().parse("not null")

    def test_never_parse_throws(self):
        with pytest.raises(ValidationError):
            v.never().parse("anything")

    def test_array_parse_throws(self):
        with pytest.raises(ValidationError):
            v.array(v.string()).parse("not array")

    def test_object_parse_throws(self):
        with pytest.raises(ValidationError):
            v.object_({"x": v.string()}).parse("not obj")

    def test_union_parse_throws(self):
        with pytest.raises(ValidationError):
            v.union([v.string(), v.int_()]).parse(True)

    def test_enum_parse_throws(self):
        with pytest.raises(ValidationError):
            v.enum_(["a", "b"]).parse("c")

    def test_literal_parse_throws(self):
        with pytest.raises(ValidationError):
            v.literal("x").parse("y")

    def test_record_parse_throws(self):
        with pytest.raises(ValidationError):
            v.record(v.string()).parse("not dict")

    def test_tuple_parse_throws(self):
        with pytest.raises(ValidationError):
            v.tuple_([v.string()]).parse("not tuple")


# ── Number edge cases ─────────────────────────────────────────────


class TestNumberEdgeCases:
    def test_nan_rejected(self):
        result = v.number().safe_parse(float("nan"))
        assert not result.success
        assert result.issues[0].code == v.INVALID_NUMBER

    def test_inf_rejected(self):
        result = v.number().safe_parse(float("inf"))
        assert not result.success

    def test_neg_inf_rejected(self):
        result = v.number().safe_parse(float("-inf"))
        assert not result.success

    def test_float32_out_of_range(self):
        schema = v.float32()
        result = schema.safe_parse(1e39)
        assert not result.success

    def test_float32_in_range(self):
        schema = v.float32()
        result = schema.safe_parse(1.0)
        assert result.success

    def test_float64_accepts_large(self):
        schema = v.float64()
        result = schema.safe_parse(1e39)
        assert result.success

    def test_multiple_of_zero_ignored(self):
        """multiple_of(0) should not cause division error."""
        schema = v.number().multiple_of(0)
        result = schema.safe_parse(42)
        assert result.success

    def test_number_rejects_string(self):
        result = v.number().safe_parse("hello")
        assert not result.success
        assert result.issues[0].code == v.INVALID_TYPE

    def test_number_rejects_none(self):
        result = v.number().safe_parse(None)
        assert not result.success

    def test_number_rejects_list(self):
        result = v.number().safe_parse([1, 2])
        assert not result.success


# ── Integer edge cases ────────────────────────────────────────────


class TestIntegerEdgeCases:
    def test_rejects_nan(self):
        result = v.int_().safe_parse(float("nan"))
        assert not result.success

    def test_rejects_inf(self):
        result = v.int_().safe_parse(float("inf"))
        assert not result.success

    def test_rejects_fractional_float(self):
        result = v.int_().safe_parse(3.7)
        assert not result.success

    def test_accepts_whole_float(self):
        result = v.int_().safe_parse(5.0)
        assert result.success
        assert result.data == 5

    def test_rejects_string(self):
        result = v.int_().safe_parse("hello")
        assert not result.success

    def test_int16_range(self):
        schema = v.int16()
        assert schema.safe_parse(32767).success
        assert schema.safe_parse(-32768).success
        assert not schema.safe_parse(32768).success
        assert not schema.safe_parse(-32769).success

    def test_uint16_range(self):
        schema = v.uint16()
        assert schema.safe_parse(0).success
        assert schema.safe_parse(65535).success
        assert not schema.safe_parse(-1).success
        assert not schema.safe_parse(65536).success

    def test_uint32_range(self):
        schema = v.uint32()
        assert schema.safe_parse(0).success
        assert schema.safe_parse(2**32 - 1).success
        assert not schema.safe_parse(2**32).success
        assert not schema.safe_parse(-1).success

    def test_int64_range(self):
        schema = v.int64()
        assert schema.safe_parse(2**63 - 1).success
        assert not schema.safe_parse(2**63).success
        assert schema.safe_parse(-(2**63)).success
        assert not schema.safe_parse(-(2**63) - 1).success

    def test_uint64_range(self):
        schema = v.uint64()
        assert schema.safe_parse(2**64 - 1).success
        assert not schema.safe_parse(2**64).success

    def test_int_constraints(self):
        schema = v.int_().min(0).max(10).exclusive_min(-1).exclusive_max(11).multiple_of(2)
        assert schema.safe_parse(4).success
        assert not schema.safe_parse(5).success  # not multiple of 2

    def test_int_exclusive_min_fail(self):
        schema = v.int_().exclusive_min(5)
        assert not schema.safe_parse(5).success
        assert schema.safe_parse(6).success

    def test_int_exclusive_max_fail(self):
        schema = v.int_().exclusive_max(10)
        assert not schema.safe_parse(10).success
        assert schema.safe_parse(9).success

    def test_int_multiple_of_zero(self):
        schema = v.int_().multiple_of(0)
        result = schema.safe_parse(42)
        assert result.success


# ── Enum edge cases ───────────────────────────────────────────────


class TestEnumEdgeCases:
    def test_type_strictness(self):
        """Enum should reject type mismatches: 1 (int) != True (bool)."""
        schema = v.enum_([1, 2, 3])
        assert not schema.safe_parse(True).success

    def test_string_int_no_match(self):
        """Enum ["1", "2"] should not match integer 1."""
        schema = v.enum_(["1", "2"])
        assert not schema.safe_parse(1).success

    def test_none_in_enum(self):
        schema = v.enum_([None, "a"])
        assert schema.safe_parse(None).success


# ── Literal edge cases ────────────────────────────────────────────


class TestLiteralEdgeCases:
    def test_none_literal(self):
        schema = v.literal(None)
        assert schema.safe_parse(None).success
        assert not schema.safe_parse(0).success

    def test_bool_int_strict(self):
        """literal(True) should not match 1."""
        assert not v.literal(True).safe_parse(1).success

    def test_accepts_none_for_none_literal(self):
        schema = v.literal(None)
        assert schema._accepts_none() is True

    def test_not_accepts_none_for_non_none_literal(self):
        schema = v.literal("hello")
        assert schema._accepts_none() is False


# ── Object edge cases ─────────────────────────────────────────────


class TestObjectEdgeCases:
    def test_optional_field_with_default(self):
        """Non-required field with default should get default value."""
        schema = v.object_(
            {"name": v.string(), "role": v.string().default("user")},
            required=["name"],
        )
        result = schema.safe_parse({"name": "Alice"})
        assert result.success
        assert result.data["role"] == "user"

    def test_required_field_with_default(self):
        """Required field with default should use default when missing."""
        schema = v.object_({"name": v.string().default("anon")})
        result = schema.safe_parse({})
        assert result.success
        assert result.data["name"] == "anon"

    def test_multiple_unknown_keys_sorted(self):
        schema = v.object_({"a": v.string()})
        result = schema.safe_parse({"a": "x", "c": 1, "b": 2})
        assert not result.success
        unknown_keys = [i.received for i in result.issues if i.code == v.UNKNOWN_KEY]
        assert unknown_keys == ["b", "c"]

    def test_strip_multiple_unknown(self):
        schema = v.object_({"a": v.string()}, unknown_keys="strip")
        result = schema.safe_parse({"a": "x", "b": 1, "c": 2})
        assert result.success
        assert "b" not in result.data
        assert "c" not in result.data

    def test_allow_preserves_all(self):
        schema = v.object_({"a": v.string()}, unknown_keys="allow")
        result = schema.safe_parse({"a": "x", "b": 1, "c": [2]})
        assert result.success
        assert result.data["b"] == 1
        assert result.data["c"] == [2]


# ── Record edge cases ────────────────────────────────────────────


class TestRecordEdgeCases:
    def test_non_string_key(self):
        """Record should reject non-string keys (if dict has int keys)."""
        schema = v.record(v.int_())
        # In Python, dicts can have int keys
        result = schema.safe_parse({1: 42})
        assert not result.success
        assert result.issues[0].code == v.INVALID_TYPE

    def test_non_dict_rejected(self):
        result = v.record(v.string()).safe_parse([1, 2])
        assert not result.success

    def test_empty_dict(self):
        result = v.record(v.string()).safe_parse({})
        assert result.success
        assert result.data == {}


# ── Union edge cases ─────────────────────────────────────────────


class TestUnionEdgeCases:
    def test_first_match_wins(self):
        """Union should return result from first matching schema."""
        schema = v.union([v.literal("a"), v.string()])
        result = schema.safe_parse("a")
        assert result.success
        assert result.data == "a"

    def test_all_fail_error(self):
        schema = v.union([v.string(), v.int_()])
        result = schema.safe_parse([1, 2])
        assert not result.success
        assert result.issues[0].code == v.INVALID_UNION

    def test_accepts_none_when_variant_does(self):
        schema = v.union([v.null(), v.string()])
        assert schema._accepts_none() is True
        result = schema.safe_parse(None)
        assert result.success

    def test_not_accepts_none(self):
        schema = v.union([v.string(), v.int_()])
        assert schema._accepts_none() is False


# ── Intersection edge cases ──────────────────────────────────────


class TestIntersectionEdgeCases:
    def test_merges_dicts(self):
        s1 = v.object_({"a": v.string()}, unknown_keys="allow")
        s2 = v.object_({"b": v.int_()}, unknown_keys="allow")
        schema = v.intersection([s1, s2])
        result = schema.safe_parse({"a": "hello", "b": 42})
        assert result.success
        assert result.data == {"a": "hello", "b": 42}

    def test_failure_in_one(self):
        s1 = v.object_({"a": v.string()}, unknown_keys="allow")
        s2 = v.object_({"b": v.int_()}, unknown_keys="allow")
        schema = v.intersection([s1, s2])
        result = schema.safe_parse({"a": "hello", "b": "not int"})
        assert not result.success

    def test_non_dict_returns_last(self):
        """Intersection of non-object schemas returns last result."""
        s1 = v.int_().min(0)
        s2 = v.int_().max(100)
        schema = v.intersection([s1, s2])
        result = schema.safe_parse(50)
        assert result.success
        assert result.data == 50

    def test_accepts_none_only_if_all_do(self):
        schema = v.intersection([v.null(), v.any_()])
        assert schema._accepts_none() is True

    def test_not_accepts_none(self):
        schema = v.intersection([v.string(), v.any_()])
        assert schema._accepts_none() is False

    def test_empty_results_returns_merged_empty(self):
        """Edge: intersection with empty schema list returns empty merged dict."""
        schema = v.intersection([])
        result = schema.safe_parse("hello")
        # With empty list, all(isinstance(r, dict) for r in []) is vacuously True
        # so it returns an empty merged dict
        assert result.success
        assert result.data == {}


# ── Ref schema ────────────────────────────────────────────────────


class TestRefSchema:
    def test_unresolved_ref_fails(self):
        schema = v.ref("Nonexistent")
        result = schema.safe_parse("hello")
        assert not result.success

    def test_resolved_ref_validates(self):
        ref = v.ref("MyString")
        ref.resolve(v.string())
        result = ref.safe_parse("hello")
        assert result.success
        assert result.data == "hello"

    def test_resolved_ref_rejects(self):
        ref = v.ref("MyString")
        ref.resolve(v.string())
        result = ref.safe_parse(42)
        assert not result.success

    def test_lazy_resolution_via_definitions(self):
        ref = v.ref("MyInt")
        ref.set_definitions({"MyInt": v.int_()})
        result = ref.safe_parse(42)
        assert result.success
        assert result.data == 42

    def test_lazy_resolution_with_hash_prefix(self):
        ref = v.ref("#/definitions/MyBool")
        ref.set_definitions({"MyBool": v.bool_()})
        result = ref.safe_parse(True)
        assert result.success

    def test_context_definitions(self):
        """Ref should try context definitions when not otherwise resolved."""
        ref = v.ref("FromCtx")
        ctx = ValidationContext(definitions={"FromCtx": v.string()})
        result = ref._run_pipeline("hello", ctx)
        assert not ctx.issues
        assert result == "hello"

    def test_context_definitions_with_hash(self):
        ref = v.ref("#/definitions/FromCtx")
        ctx = ValidationContext(definitions={"FromCtx": v.int_()})
        result = ref._run_pipeline(42, ctx)
        assert not ctx.issues

    def test_accepts_none_resolved(self):
        ref = v.ref("NullRef")
        ref.resolve(v.null())
        assert ref._accepts_none() is True

    def test_accepts_none_unresolved(self):
        ref = v.ref("Unresolved")
        assert ref._accepts_none() is False

    def test_validate_unresolved(self):
        ref = v.ref("Missing")
        ctx = ValidationContext()
        ref._validate("whatever", ctx)
        assert len(ctx.issues) == 1

    def test_validate_resolved(self):
        ref = v.ref("Str")
        ref.resolve(v.string())
        ctx = ValidationContext()
        result = ref._validate("hello", ctx)
        assert not ctx.issues
        assert result == "hello"


# ── Nullable/Optional direct _validate/_accepts_none coverage ────


class TestNullableDirectMethods:
    def test_nullable_accepts_none_returns_true(self):
        schema = v.nullable(v.string())
        assert schema._accepts_none() is True

    def test_nullable_validate_null_returns_none(self):
        schema = v.nullable(v.string())
        ctx = ValidationContext()
        result = schema._validate(None, ctx)
        assert result is None
        assert not ctx.issues

    def test_nullable_validate_non_null_delegates(self):
        schema = v.nullable(v.string())
        ctx = ValidationContext()
        result = schema._validate("hello", ctx)
        assert result == "hello"
        assert not ctx.issues

    def test_nullable_validate_non_null_invalid_delegates(self):
        schema = v.nullable(v.int_())
        ctx = ValidationContext()
        schema._validate("not int", ctx)
        assert len(ctx.issues) >= 1


class TestOptionalDirectMethods:
    def test_optional_accepts_none_delegates_to_inner(self):
        # inner is string, which doesn't accept None
        schema = v.optional(v.string())
        assert schema._accepts_none() is False

    def test_optional_accepts_none_with_nullable_inner(self):
        schema = v.optional(v.nullable(v.string()))
        assert schema._accepts_none() is True

    def test_optional_validate_delegates(self):
        schema = v.optional(v.string())
        ctx = ValidationContext()
        result = schema._validate("hello", ctx)
        assert result == "hello"
        assert not ctx.issues


# ── Integer exclusive_min coverage ──────────────────────────────


class TestIntegerExclusiveConstraints:
    def test_exclusive_min_rejects_equal(self):
        schema = v.int_().exclusive_min(5)
        result = schema.safe_parse(5)
        assert not result.success
        assert result.issues[0].code == v.TOO_SMALL

    def test_exclusive_min_accepts_above(self):
        schema = v.int_().exclusive_min(5)
        result = schema.safe_parse(6)
        assert result.success

    def test_exclusive_max_rejects_equal(self):
        schema = v.int_().exclusive_max(10)
        result = schema.safe_parse(10)
        assert not result.success
        assert result.issues[0].code == v.TOO_LARGE

    def test_exclusive_max_accepts_below(self):
        schema = v.int_().exclusive_max(10)
        result = schema.safe_parse(9)
        assert result.success

    def test_max_rejects_above(self):
        schema = v.int_().max(10)
        result = schema.safe_parse(11)
        assert not result.success
        assert result.issues[0].code == v.TOO_LARGE

    def test_max_accepts_equal(self):
        schema = v.int_().max(10)
        result = schema.safe_parse(10)
        assert result.success


# ── Default value edge cases ─────────────────────────────────────


class TestDefaultEdgeCases:
    def test_default_invalid_produces_default_invalid_issue(self):
        """When default fails validation, a DEFAULT_INVALID issue is produced."""
        schema = v.int_().min(0).default(-5)
        ctx = ValidationContext()
        schema._run_pipeline(_SENTINEL, ctx)
        assert len(ctx.issues) == 1
        assert ctx.issues[0].code == v.DEFAULT_INVALID

    def test_default_valid_passes(self):
        schema = v.int_().min(0).default(10)
        ctx = ValidationContext()
        result = schema._run_pipeline(_SENTINEL, ctx)
        assert not ctx.issues
        assert result == 10

    def test_default_deep_copied(self):
        """Default values should be deep copied to avoid mutation."""
        default_list = [1, 2, 3]
        schema = v.array(v.int_()).default(default_list)
        ctx1 = ValidationContext()
        result1 = schema._run_pipeline(_SENTINEL, ctx1)
        ctx2 = ValidationContext()
        result2 = schema._run_pipeline(_SENTINEL, ctx2)
        assert result1 is not result2
        assert result1 == result2 == [1, 2, 3]


# ── Tuple edge cases ─────────────────────────────────────────────


class TestTupleEdgeCases:
    def test_accepts_python_tuple(self):
        schema = v.tuple_([v.string(), v.int_()])
        result = schema.safe_parse(("hello", 42))
        assert result.success

    def test_too_few_items(self):
        schema = v.tuple_([v.string(), v.int_(), v.bool_()])
        result = schema.safe_parse(["a"])
        assert not result.success
        assert result.issues[0].code == v.TOO_SMALL

    def test_too_many_items(self):
        schema = v.tuple_([v.string()])
        result = schema.safe_parse(["a", "b"])
        assert not result.success
        assert result.issues[0].code == v.TOO_LARGE

    def test_element_validation_failure(self):
        schema = v.tuple_([v.string(), v.int_()])
        result = schema.safe_parse(["hello", "not int"])
        assert not result.success


# ── Any/Unknown edge cases ────────────────────────────────────────


class TestAnyUnknownEdgeCases:
    def test_any_accepts_none(self):
        assert v.any_()._accepts_none() is True

    def test_unknown_accepts_none(self):
        assert v.unknown()._accepts_none() is True

    def test_any_returns_value(self):
        result = v.any_().safe_parse({"complex": [1, 2]})
        assert result.success
        assert result.data == {"complex": [1, 2]}

    def test_unknown_returns_value(self):
        result = v.unknown().safe_parse(42)
        assert result.success
        assert result.data == 42


# ── String format via schema ──────────────────────────────────────


class TestStringFormat:
    def test_format_url_valid(self):
        schema = v.string().format("url")
        assert schema.safe_parse("https://example.com").success

    def test_format_url_invalid(self):
        schema = v.string().format("url")
        result = schema.safe_parse("not a url")
        assert not result.success
        assert result.issues[0].code == v.INVALID_STRING

    def test_format_ipv6_valid(self):
        schema = v.string().format("ipv6")
        assert schema.safe_parse("::1").success

    def test_format_ipv6_invalid(self):
        schema = v.string().format("ipv6")
        assert not schema.safe_parse("not ipv6").success

    def test_format_datetime_valid(self):
        schema = v.string().format("date-time")
        assert schema.safe_parse("2024-01-15T10:30:00Z").success

    def test_format_datetime_invalid(self):
        schema = v.string().format("date-time")
        assert not schema.safe_parse("2024-01-15").success

    def test_unknown_format_passes(self):
        schema = v.string().format("x-custom")
        assert schema.safe_parse("anything").success


# ── Copy behavior ─────────────────────────────────────────────────


class TestCopyBehavior:
    def test_copy_preserves_constraints(self):
        original = v.string().min_length(1)
        copy = original._copy()
        assert copy.safe_parse("").success is False
        assert copy.safe_parse("a").success is True

    def test_copy_does_not_affect_original(self):
        original = v.string()
        modified = original.min_length(5)
        # Original should still accept short strings
        assert original.safe_parse("hi").success
        assert not modified.safe_parse("hi").success

    def test_coerce_returns_new_schema(self):
        original = v.string()
        coerced = original.coerce(trim=True)
        assert original._coercion is None
        assert coerced._coercion is not None

    def test_default_returns_new_schema(self):
        original = v.string()
        with_default = original.default("hi")
        assert original._has_default is False
        assert with_default._has_default is True


# ── Import round-trip for every schema kind ───────────────────────


class TestImportRoundTrip:
    def test_float32_roundtrip(self):
        doc = v.float32().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(1.0).success

    def test_float64_roundtrip(self):
        doc = v.float64().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(3.14).success

    def test_int8_roundtrip(self):
        doc = v.int8().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(42).success
        assert not imported.safe_parse(200).success

    def test_int16_roundtrip(self):
        doc = v.int16().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(1000).success

    def test_int32_roundtrip(self):
        doc = v.int32().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(100000).success

    def test_int64_roundtrip(self):
        doc = v.int64().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(42).success

    def test_uint8_roundtrip(self):
        doc = v.uint8().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(200).success
        assert not imported.safe_parse(-1).success

    def test_uint16_roundtrip(self):
        doc = v.uint16().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(50000).success

    def test_uint32_roundtrip(self):
        doc = v.uint32().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(100000).success

    def test_uint64_roundtrip(self):
        doc = v.uint64().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(100000).success

    def test_null_roundtrip(self):
        doc = v.null().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(None).success

    def test_any_roundtrip(self):
        doc = v.any_().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("anything").success

    def test_unknown_roundtrip(self):
        doc = v.unknown().export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(42).success

    def test_never_roundtrip(self):
        doc = v.never().export()
        imported = v.import_schema(doc)
        assert not imported.safe_parse("x").success

    def test_literal_roundtrip(self):
        doc = v.literal("hello").export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("hello").success
        assert not imported.safe_parse("world").success

    def test_enum_roundtrip(self):
        doc = v.enum_(["a", "b"]).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("a").success
        assert not imported.safe_parse("c").success

    def test_tuple_roundtrip(self):
        doc = v.tuple_([v.string(), v.int_()]).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(["hello", 42]).success

    def test_record_roundtrip(self):
        doc = v.record(v.int_()).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse({"a": 1}).success

    def test_union_roundtrip(self):
        doc = v.union([v.string(), v.int_()]).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("hello").success
        assert imported.safe_parse(42).success

    def test_intersection_roundtrip(self):
        s1 = v.object_({"a": v.string()}, unknown_keys="allow")
        s2 = v.object_({"b": v.int_()}, unknown_keys="allow")
        doc = v.intersection([s1, s2]).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse({"a": "x", "b": 1}).success

    def test_optional_roundtrip(self):
        doc = v.optional(v.string()).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("hello").success

    def test_nullable_roundtrip(self):
        doc = v.nullable(v.string()).export()
        imported = v.import_schema(doc)
        assert imported.safe_parse(None).success
        assert imported.safe_parse("hello").success

    def test_ref_roundtrip(self):
        doc = v.ref("#/definitions/Foo").export()
        assert doc["root"]["kind"] == "ref"

    def test_with_default_roundtrip(self):
        doc = v.string().default("hi").export()
        imported = v.import_schema(doc)
        assert imported._has_default is True
        assert imported._default_value == "hi"

    def test_with_coercion_roundtrip(self):
        doc = v.string().coerce(trim=True, lower=True).export()
        imported = v.import_schema(doc)
        assert imported._coercion is not None
        assert imported._coercion.trim is True
        assert imported._coercion.lower is True

    def test_import_unsupported_kind_raises(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "nonexistent_kind"},
        }
        with pytest.raises(ValueError, match="Unsupported schema kind"):
            v.import_schema(doc)

    def test_import_with_definitions(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "ref", "ref": "#/definitions/MyStr"},
            "definitions": {"MyStr": {"kind": "string"}},
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse("hello")
        assert result.success

    def test_import_coercion_as_string(self):
        """Import coercion from corpus string format."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "int", "coerce": "string->int"},
        }
        schema = v.import_schema(doc)
        assert schema._coercion is not None
        assert schema._coercion.to_int is True

    def test_import_coercion_as_list(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "coerce": ["trim", "lower"]},
        }
        schema = v.import_schema(doc)
        assert schema._coercion is not None
        assert schema._coercion.trim is True
        assert schema._coercion.lower is True

    def test_import_coercion_string_to_number(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "number", "coerce": "string->number"},
        }
        schema = v.import_schema(doc)
        assert schema._coercion.to_number is True

    def test_import_coercion_string_to_bool(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "bool", "coerce": "string->bool"},
        }
        schema = v.import_schema(doc)
        assert schema._coercion.to_bool is True

    def test_import_coercion_upper(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "coerce": "upper"},
        }
        schema = v.import_schema(doc)
        assert schema._coercion.upper is True

    def test_import_coercion_unknown_token_returns_none(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "coerce": "unknown_coercion"},
        }
        schema = v.import_schema(doc)
        assert schema._coercion is None

    def test_import_coercion_non_string_non_list_non_dict(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "coerce": 42},
        }
        schema = v.import_schema(doc)
        assert schema._coercion is None

    def test_import_number_with_constraints(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "number",
                "min": 0,
                "max": 100,
                "exclusiveMin": -1,
                "exclusiveMax": 101,
                "multipleOf": 5,
            },
        }
        schema = v.import_schema(doc)
        assert schema.safe_parse(50).success
        assert not schema.safe_parse(-2).success

    def test_import_coercion_dict_format(self):
        """Import coercion from dict format (our own export format)."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "string",
                "coerce": {"trim": True, "lower": True, "upper": False,
                           "toInt": False, "toNumber": False, "toBool": False},
            },
        }
        schema = v.import_schema(doc)
        assert schema._coercion is not None
        assert schema._coercion.trim is True
        assert schema._coercion.lower is True

    def test_import_coercion_null_value(self):
        """When coerce key exists but has null value."""
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "coerce": None},
        }
        schema = v.import_schema(doc)
        assert schema._coercion is None

    def test_import_string_all_constraints(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "string",
                "minLength": 1,
                "maxLength": 10,
                "pattern": "^a",
                "startsWith": "a",
                "endsWith": "z",
                "includes": "mid",
                "format": "email",
            },
        }
        schema = v.import_schema(doc)
        # Will fail because it can't match all constraints at once, but import works
        assert schema is not None

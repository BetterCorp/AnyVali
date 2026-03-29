"""Unit tests for composite schema types."""

from __future__ import annotations

import pytest

import anyvali as v


class TestArray:
    def test_valid(self):
        schema = v.array(v.string())
        result = schema.safe_parse(["a", "b", "c"])
        assert result.success
        assert result.data == ["a", "b", "c"]

    def test_invalid_type(self):
        schema = v.array(v.string())
        result = schema.safe_parse("not-an-array")
        assert not result.success

    def test_invalid_items(self):
        schema = v.array(v.int_())
        result = schema.safe_parse([1, "two", 3])
        assert not result.success

    def test_min_items(self):
        schema = v.array(v.string()).min_items(2)
        assert schema.safe_parse(["a", "b"]).success
        assert not schema.safe_parse(["a"]).success

    def test_max_items(self):
        schema = v.array(v.string()).max_items(2)
        assert schema.safe_parse(["a", "b"]).success
        assert not schema.safe_parse(["a", "b", "c"]).success


class TestTuple:
    def test_valid(self):
        schema = v.tuple_([v.string(), v.int_()])
        result = schema.safe_parse(["hello", 42])
        assert result.success
        assert result.data == ["hello", 42]

    def test_wrong_length(self):
        schema = v.tuple_([v.string(), v.int_()])
        assert not schema.safe_parse(["hello"]).success
        assert not schema.safe_parse(["hello", 42, "extra"]).success

    def test_invalid_type(self):
        schema = v.tuple_([v.string(), v.int_()])
        result = schema.safe_parse("not-a-tuple")
        assert not result.success


class TestObject:
    def test_valid(self):
        schema = v.object_({"name": v.string(), "age": v.int_()})
        result = schema.safe_parse({"name": "Alice", "age": 30})
        assert result.success
        assert result.data == {"name": "Alice", "age": 30}

    def test_missing_required(self):
        schema = v.object_({"name": v.string(), "age": v.int_()})
        result = schema.safe_parse({"name": "Alice"})
        assert not result.success
        assert any(i.code == v.REQUIRED for i in result.issues)

    def test_unknown_keys_reject(self):
        schema = v.object_({"name": v.string()})
        result = schema.safe_parse({"name": "Alice", "extra": 1})
        assert not result.success
        assert any(i.code == v.UNKNOWN_KEY for i in result.issues)

    def test_unknown_keys_strip(self):
        schema = v.object_({"name": v.string()}, unknown_keys="strip")
        result = schema.safe_parse({"name": "Alice", "extra": 1})
        assert result.success
        assert "extra" not in result.data

    def test_unknown_keys_allow(self):
        schema = v.object_({"name": v.string()}, unknown_keys="allow")
        result = schema.safe_parse({"name": "Alice", "extra": 1})
        assert result.success
        assert result.data["extra"] == 1

    def test_optional_field(self):
        schema = v.object_(
            {"name": v.string(), "age": v.int_()},
            required=["name"],
        )
        result = schema.safe_parse({"name": "Alice"})
        assert result.success

    def test_invalid_type(self):
        schema = v.object_({"name": v.string()})
        result = schema.safe_parse("not-an-object")
        assert not result.success


class TestRecord:
    def test_valid(self):
        schema = v.record(v.int_())
        result = schema.safe_parse({"a": 1, "b": 2})
        assert result.success
        assert result.data == {"a": 1, "b": 2}

    def test_invalid_values(self):
        schema = v.record(v.int_())
        result = schema.safe_parse({"a": 1, "b": "two"})
        assert not result.success


class TestUnion:
    def test_first_match(self):
        schema = v.union([v.string(), v.int_()])
        assert schema.safe_parse("hello").success
        assert schema.safe_parse(42).success

    def test_no_match(self):
        schema = v.union([v.string(), v.int_()])
        result = schema.safe_parse(True)
        assert not result.success
        assert result.issues[0].code == v.INVALID_UNION


class TestIntersection:
    def test_both_pass(self):
        s1 = v.object_({"name": v.string()}, unknown_keys="allow")
        s2 = v.object_({"age": v.int_()}, unknown_keys="allow")
        schema = v.intersection([s1, s2])
        result = schema.safe_parse({"name": "Alice", "age": 30})
        assert result.success


class TestOptional:
    def test_present_value(self):
        schema = v.optional(v.string())
        result = schema.safe_parse("hello")
        assert result.success
        assert result.data == "hello"

    def test_rejects_wrong_type(self):
        schema = v.optional(v.string())
        result = schema.safe_parse(42)
        assert not result.success


class TestNullable:
    def test_null_value(self):
        schema = v.nullable(v.string())
        result = schema.safe_parse(None)
        assert result.success
        assert result.data is None

    def test_valid_value(self):
        schema = v.nullable(v.string())
        result = schema.safe_parse("hello")
        assert result.success

    def test_invalid_value(self):
        schema = v.nullable(v.string())
        result = schema.safe_parse(42)
        assert not result.success

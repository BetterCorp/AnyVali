"""Unit tests for primitive schema types."""

from __future__ import annotations

import pytest

import anyvali as v


class TestString:
    def test_valid(self):
        result = v.string().safe_parse("hello")
        assert result.success
        assert result.data == "hello"

    def test_invalid_type(self):
        result = v.string().safe_parse(42)
        assert not result.success
        assert result.issues[0].code == v.INVALID_TYPE

    def test_min_length(self):
        schema = v.string().min_length(3)
        assert schema.safe_parse("abc").success
        assert not schema.safe_parse("ab").success

    def test_max_length(self):
        schema = v.string().max_length(3)
        assert schema.safe_parse("abc").success
        assert not schema.safe_parse("abcd").success

    def test_pattern(self):
        schema = v.string().pattern(r"^\d+$")
        assert schema.safe_parse("123").success
        assert not schema.safe_parse("abc").success

    def test_starts_with(self):
        schema = v.string().starts_with("hello")
        assert schema.safe_parse("hello world").success
        assert not schema.safe_parse("world hello").success

    def test_ends_with(self):
        schema = v.string().ends_with("world")
        assert schema.safe_parse("hello world").success
        assert not schema.safe_parse("world hello").success

    def test_includes(self):
        schema = v.string().includes("mid")
        assert schema.safe_parse("in the middle").success
        assert not schema.safe_parse("nothing here").success

    def test_format_email(self):
        schema = v.string().format("email")
        assert schema.safe_parse("user@example.com").success
        assert not schema.safe_parse("not-an-email").success

    def test_format_uuid(self):
        schema = v.string().format("uuid")
        assert schema.safe_parse("550e8400-e29b-41d4-a716-446655440000").success
        assert not schema.safe_parse("not-a-uuid").success

    def test_format_ipv4(self):
        schema = v.string().format("ipv4")
        assert schema.safe_parse("192.168.1.1").success
        assert not schema.safe_parse("999.999.999.999").success

    def test_format_date(self):
        schema = v.string().format("date")
        assert schema.safe_parse("2024-01-15").success
        assert not schema.safe_parse("not-a-date").success


class TestNumber:
    def test_valid_int(self):
        result = v.number().safe_parse(42)
        assert result.success
        assert result.data == 42

    def test_valid_float(self):
        result = v.number().safe_parse(3.14)
        assert result.success
        assert result.data == 3.14

    def test_invalid_type(self):
        result = v.number().safe_parse("hello")
        assert not result.success

    def test_rejects_bool(self):
        result = v.number().safe_parse(True)
        assert not result.success

    def test_min(self):
        schema = v.number().min(5)
        assert schema.safe_parse(5).success
        assert not schema.safe_parse(4).success

    def test_max(self):
        schema = v.number().max(10)
        assert schema.safe_parse(10).success
        assert not schema.safe_parse(11).success

    def test_exclusive_min(self):
        schema = v.number().exclusive_min(5)
        assert schema.safe_parse(6).success
        assert not schema.safe_parse(5).success

    def test_exclusive_max(self):
        schema = v.number().exclusive_max(10)
        assert schema.safe_parse(9).success
        assert not schema.safe_parse(10).success

    def test_multiple_of(self):
        schema = v.number().multiple_of(3)
        assert schema.safe_parse(9).success
        assert not schema.safe_parse(10).success


class TestInteger:
    def test_valid(self):
        result = v.int_().safe_parse(42)
        assert result.success
        assert result.data == 42

    def test_rejects_float(self):
        result = v.int_().safe_parse(3.14)
        assert not result.success

    def test_accepts_whole_float(self):
        result = v.int_().safe_parse(42.0)
        assert result.success
        assert result.data == 42

    def test_rejects_bool(self):
        result = v.int_().safe_parse(True)
        assert not result.success

    def test_int8_range(self):
        schema = v.int8()
        assert schema.safe_parse(127).success
        assert schema.safe_parse(-128).success
        assert not schema.safe_parse(128).success
        assert not schema.safe_parse(-129).success

    def test_uint8_range(self):
        schema = v.uint8()
        assert schema.safe_parse(0).success
        assert schema.safe_parse(255).success
        assert not schema.safe_parse(-1).success
        assert not schema.safe_parse(256).success

    def test_int32_range(self):
        schema = v.int32()
        assert schema.safe_parse(2147483647).success
        assert not schema.safe_parse(2147483648).success

    def test_uint64_range(self):
        schema = v.uint64()
        assert schema.safe_parse(0).success
        assert not schema.safe_parse(-1).success


class TestBool:
    def test_valid_true(self):
        assert v.bool_().safe_parse(True).success

    def test_valid_false(self):
        assert v.bool_().safe_parse(False).success

    def test_rejects_int(self):
        assert not v.bool_().safe_parse(1).success

    def test_rejects_string(self):
        assert not v.bool_().safe_parse("true").success


class TestNull:
    def test_valid(self):
        assert v.null().safe_parse(None).success

    def test_rejects_non_null(self):
        assert not v.null().safe_parse(0).success
        assert not v.null().safe_parse("").success


class TestAny:
    def test_accepts_anything(self):
        schema = v.any_()
        assert schema.safe_parse(42).success
        assert schema.safe_parse("hello").success
        assert schema.safe_parse(None).success
        assert schema.safe_parse([1, 2, 3]).success


class TestNever:
    def test_rejects_everything(self):
        schema = v.never()
        assert not schema.safe_parse(42).success
        assert not schema.safe_parse("hello").success
        assert not schema.safe_parse(None).success


class TestLiteral:
    def test_valid(self):
        assert v.literal("hello").safe_parse("hello").success

    def test_invalid_value(self):
        assert not v.literal("hello").safe_parse("world").success

    def test_type_strict(self):
        # 1 != True for literal check
        assert not v.literal(1).safe_parse(True).success


class TestEnum:
    def test_valid(self):
        schema = v.enum_(["a", "b", "c"])
        assert schema.safe_parse("a").success
        assert schema.safe_parse("b").success

    def test_invalid(self):
        schema = v.enum_(["a", "b", "c"])
        assert not schema.safe_parse("d").success

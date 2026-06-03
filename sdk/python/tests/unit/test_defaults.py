"""Unit tests for default value application."""

from __future__ import annotations

import anyvali as v


class TestDefaults:
    def test_missing_field_gets_default(self):
        schema = v.object_({
            "name": v.string(),
            "role": v.string().default("user"),
        })
        result = schema.safe_parse({"name": "Alice"})
        assert result.success
        assert result.data["role"] == "user"

    def test_present_field_not_overwritten(self):
        schema = v.object_({
            "name": v.string(),
            "role": v.string().default("user"),
        })
        result = schema.safe_parse({"name": "Alice", "role": "admin"})
        assert result.success
        assert result.data["role"] == "admin"

    def test_default_on_primitive(self):
        schema = v.int_().default(42)
        # When used standalone, we pass the sentinel to trigger default
        from anyvali.schemas.base import _SENTINEL, ValidationContext
        ctx = ValidationContext()
        result = schema._run_pipeline(_SENTINEL, ctx)
        assert not ctx.issues
        assert result == 42

    def test_default_value_is_validated(self):
        schema = v.object_({
            "count": v.int_().min(0).default(-1),
        })
        result = schema.safe_parse({})
        # Default of -1 violates min(0), so validation should fail
        assert not result.success

    def test_null_is_not_absent_for_nullable_default(self):
        schema = v.object_({
            "value": v.nullable(v.string()).default("fallback"),
        })
        result = schema.safe_parse({"value": None})
        assert result.success
        assert result.data == {"value": None}

    def test_falsy_defaults_are_applied(self):
        schema = v.object_({
            "count": v.int_().default(0),
            "name": v.string().default(""),
            "active": v.bool_().default(False),
        })
        result = schema.safe_parse({})
        assert result.success
        assert result.data == {"count": 0, "name": "", "active": False}

    def test_optional_wrapper_field_gets_default(self):
        schema = v.object_({
            "host": v.optional(v.string()).default("localhost"),
        })
        result = schema.safe_parse({})
        assert result.success
        assert result.data == {"host": "localhost"}

    def test_optional_wrapper_default_does_not_override_present_field(self):
        schema = v.object_({
            "host": v.optional(v.string()).default("localhost"),
        })
        result = schema.safe_parse({"host": "example.com"})
        assert result.success
        assert result.data == {"host": "example.com"}

    def test_optional_wrapper_default_is_validated(self):
        schema = v.object_({
            "host": v.optional(v.string().min_length(5)).default("hi"),
        })
        result = schema.safe_parse({})
        assert not result.success
        assert result.issues[0].code == v.DEFAULT_INVALID
        assert result.issues[0].path == ["host"]

    def test_optional_wrapper_default_is_exported(self):
        doc = v.optional(v.string()).default("localhost").export()
        assert doc["root"]["kind"] == "optional"
        assert doc["root"]["default"] == "localhost"

    def test_nested_object_field_gets_default(self):
        schema = v.object_({
            "user": v.object_({
                "name": v.string(),
                "role": v.string().default("guest"),
            }, required=["name"]),
        }, required=["user"])
        result = schema.safe_parse({"user": {"name": "Bob"}})
        assert result.success
        assert result.data == {"user": {"name": "Bob", "role": "guest"}}

    def test_mutable_default_is_not_shared_between_parses(self):
        schema = v.object_({
            "tags": v.array(v.string()).default([]),
        })
        first = schema.parse({})
        first["tags"].append("mutated")
        second = schema.parse({})
        assert second == {"tags": []}

    def test_mutable_optional_wrapper_default_is_not_shared_between_parses(self):
        schema = v.object_({
            "meta": v.optional(v.any_()).default({"items": []}),
        })
        first = schema.parse({})
        first["meta"]["items"].append("mutated")
        second = schema.parse({})
        assert second == {"meta": {"items": []}}

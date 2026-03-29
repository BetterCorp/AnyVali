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

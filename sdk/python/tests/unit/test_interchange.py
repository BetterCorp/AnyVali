"""Unit tests for import/export round-trip."""

from __future__ import annotations

import json

import pytest

import anyvali as v


class TestExport:
    def test_string_export(self):
        schema = v.string().min_length(1)
        doc = schema.export()
        assert doc["root"]["kind"] == "string"
        assert doc["root"]["minLength"] == 1

    def test_object_export(self):
        schema = v.object_({
            "name": v.string(),
            "age": v.int_(),
        })
        doc = schema.export()
        assert doc["root"]["kind"] == "object"
        assert "name" in doc["root"]["properties"]
        assert "age" in doc["root"]["properties"]

    def test_number_alias(self):
        schema = v.number()
        doc = schema.export()
        assert doc["root"]["kind"] == "number"

    def test_int_alias(self):
        schema = v.int_()
        doc = schema.export()
        assert doc["root"]["kind"] == "int"

    def test_with_default(self):
        schema = v.string().default("hello")
        doc = schema.export()
        assert doc["root"]["default"] == "hello"

    def test_with_coercion(self):
        schema = v.string().coerce(trim=True, lower=True)
        doc = schema.export()
        assert doc["root"]["coerce"]["trim"] is True
        assert doc["root"]["coerce"]["lower"] is True


class TestImport:
    def test_string_import(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "minLength": 1},
        }
        schema = v.import_schema(doc)
        assert schema.safe_parse("hello").success
        assert not schema.safe_parse("").success

    def test_object_import(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "object",
                "properties": {
                    "name": {"kind": "string"},
                    "age": {"kind": "int"},
                },
                "required": ["name", "age"],
                "unknownKeys": "reject",
            },
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse({"name": "Alice", "age": 30})
        assert result.success

    def test_json_string_import(self):
        doc = json.dumps({
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "bool"},
        })
        schema = v.import_schema(doc)
        assert schema.safe_parse(True).success

    def test_import_unsupported_kind(self):
        doc = {"root": {"kind": "bogus_xyz"}}
        with pytest.raises(Exception):
            v.import_schema(doc)

    def test_import_missing_kind(self):
        doc = {"root": {}}
        with pytest.raises(Exception):
            v.import_schema(doc)

    def test_import_null_empty_root(self):
        with pytest.raises(Exception):
            v.import_schema({})
        with pytest.raises(Exception):
            v.import_schema({"root": None})


class TestRoundTrip:
    def test_string_roundtrip(self):
        original = v.string().min_length(1).max_length(100)
        doc = original.export()
        imported = v.import_schema(doc)
        assert imported.safe_parse("hello").success
        assert not imported.safe_parse("").success

    def test_object_roundtrip(self):
        original = v.object_({
            "name": v.string(),
            "tags": v.array(v.string()),
        })
        doc = original.export()
        imported = v.import_schema(doc)
        result = imported.safe_parse({"name": "test", "tags": ["a", "b"]})
        assert result.success

    def test_number_roundtrip_as_float64(self):
        original = v.number()
        doc = original.export()
        assert doc["root"]["kind"] == "number"
        imported = v.import_schema(doc)
        result = imported.safe_parse(3.14)
        assert result.success

    def test_int_roundtrip_as_int64(self):
        original = v.int_()
        doc = original.export()
        assert doc["root"]["kind"] == "int"
        imported = v.import_schema(doc)
        result = imported.safe_parse(42)
        assert result.success

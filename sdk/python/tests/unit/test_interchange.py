"""Unit tests for import/export round-trip."""

from __future__ import annotations

import json

import pytest

import anyvali as v


class TestExport:
    @pytest.mark.parametrize("left,right,equal", [
        ({"min": 1}, {"min": 1.0}, True),
        ({"min": 1}, {"min": 1.5}, False),
        ({"metadata": {"custom": {"a": [1, None], "b": 2}}},
         {"metadata": {"custom": {"b": 2.0, "a": [1.0, None]}}}, True),
        ({"metadata": {"custom": [True]}}, {"metadata": {"custom": [1]}}, False),
        ({"metadata": {"custom": [1]}}, {"metadata": {"custom": ["1"]}}, False),
        ({"metadata": {"custom": [1, 2]}}, {"metadata": {"custom": [2, 1]}}, False),
        ({"metadata": {"custom": [1]}}, {"metadata": {"custom": [1, 2]}}, False),
    ])
    def test_definition_json_equality(self, left, right, equal):
        def child(fields):
            return v.import_schema({
                "anyvaliVersion": "1.0", "schemaVersion": "1.1",
                "root": {"kind": "ref", "ref": "#/definitions/Value"},
                "definitions": {"Value": {"kind": "number", **fields}}, "extensions": {},
            })

        parent = v.object_({"first": child(left), "second": child(right)})
        if equal:
            assert v.import_schema(parent.export()).parse({"first": 2, "second": 3}) == {"first": 2, "second": 3}
        else:
            with pytest.raises(ValueError, match="Conflicting definition: Value"):
                parent.export()

    def test_composite_definitions_and_conflicts(self):
        doc = {"anyvaliVersion": "1.0", "schemaVersion": "1.1",
               "root": {"kind": "ref", "ref": "#/definitions/Value"},
               "definitions": {"Value": {"kind": "string", "minLength": 1}}, "extensions": {}}
        child = v.import_schema(doc)
        for parent, value in (
            (v.object_({"value": child}), {"value": "ok"}),
            (v.array(child), ["ok"]), (v.record(child), {"key": "ok"}),
            (v.tuple_([child]), ["ok"]), (v.optional(child), "ok"), (v.nullable(child), "ok"),
            (v.union([child, v.bool_()]), "ok"), (v.intersection([child, v.string()]), "ok"),
        ):
            assert parent.parse(value) == value
            exported = parent.export()
            assert exported["definitions"] == doc["definitions"]
            assert v.import_schema(exported).parse(value) == value
        reordered = v.import_schema({**doc, "definitions": {"Value": {"minLength": 1, "kind": "string"}}})
        v.object_({"first": child, "second": child, "third": reordered}).export()
        conflicting = v.import_schema({**doc, "definitions": {"Value": {"kind": "bool"}}})
        parent = v.object_({"first": child, "second": conflicting})
        assert parent.parse({"first": "ok", "second": True}) == {"first": "ok", "second": True}
        with pytest.raises(ValueError, match="Conflicting definition: Value"):
            v.export_schema(parent)
        with pytest.raises(ValueError, match="Conflicting definition: Value"):
            v.export_schema(child, definitions={"Value": v.bool_()})

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
    @pytest.mark.parametrize("kind,value,expected", [
        ("int", "7", 7), ("number", "1.5", 1.5), ("bool", "true", True),
    ])
    @pytest.mark.parametrize("coerce", [{}, {"from": "string"}])
    def test_sdk_coercion_object_survives_export(self, kind, value, expected, coerce):
        schema = v.import_schema({
            "anyvaliVersion": "1.0", "schemaVersion": "1.1",
            "root": {"kind": kind, "coerce": coerce},
            "definitions": {}, "extensions": {},
        })
        assert schema.parse(value) == expected
        exported = schema.export()
        assert "coerce" in exported["root"]
        assert v.import_schema(exported).parse(value) == expected

    def test_reserved_and_custom_metadata_roundtrip(self):
        metadata = {"description": "Private", "sensitive": True, "custom": {"owner": "ports"}}
        for root in (
            {"kind": "string", "metadata": metadata},
            {"kind": "optional", "inner": {"kind": "string"}, "metadata": metadata},
            {"kind": "nullable", "inner": {"kind": "string"}, "metadata": metadata},
            {"kind": "ref", "ref": "#/definitions/Secret", "metadata": metadata},
        ):
            doc = {"anyvaliVersion": "1.0", "schemaVersion": "1.1", "root": root,
                   "definitions": {"Secret": {"kind": "string"}}, "extensions": {}}
            schema = v.import_schema(doc)
            assert schema.export()["root"]["metadata"] == metadata
            assert v.export_schema(v.import_schema(schema.export()))["root"]["metadata"] == metadata
        with pytest.raises(ValueError, match="Schema metadata must be an object"):
            v.import_schema({**doc, "root": {"kind": "string", "metadata": None}})

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

    def test_string_import_with_invalid_pattern_fails_without_throwing(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {"kind": "string", "pattern": "("},
        }
        schema = v.import_schema(doc)
        result = schema.safe_parse("abc")
        assert not result.success

    def test_array_import_with_canonical_items_key(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "array",
                "items": {"kind": "int"},
            },
        }
        schema = v.import_schema(doc)
        assert schema.safe_parse([1, 2]).success
        assert not schema.safe_parse(["a"]).success

    def test_union_import_with_canonical_variants_key(self):
        doc = {
            "anyvaliVersion": "1.0",
            "schemaVersion": "1",
            "root": {
                "kind": "union",
                "variants": [{"kind": "string"}, {"kind": "int"}],
            },
        }
        schema = v.import_schema(doc)
        assert schema.safe_parse("hello").success
        assert schema.safe_parse(42).success
        assert not schema.safe_parse(True).success


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

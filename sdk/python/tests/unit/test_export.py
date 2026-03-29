"""Unit tests for _to_node() and export() for every schema type."""

from __future__ import annotations

import json

import anyvali as v
from anyvali.interchange.exporter import export_schema, export_schema_json


class TestStringExport:
    def test_plain(self):
        node = v.string()._to_node()
        assert node == {"kind": "string"}

    def test_all_constraints(self):
        schema = (
            v.string()
            .min_length(1)
            .max_length(100)
            .pattern(r"^\w+$")
            .starts_with("a")
            .ends_with("z")
            .includes("mid")
            .format("email")
        )
        node = schema._to_node()
        assert node["kind"] == "string"
        assert node["minLength"] == 1
        assert node["maxLength"] == 100
        assert node["pattern"] == r"^\w+$"
        assert node["startsWith"] == "a"
        assert node["endsWith"] == "z"
        assert node["includes"] == "mid"
        assert node["format"] == "email"


class TestNumberExport:
    def test_number_plain(self):
        assert v.number()._to_node() == {"kind": "number"}

    def test_float64_plain(self):
        assert v.float64()._to_node() == {"kind": "float64"}

    def test_float32_plain(self):
        assert v.float32()._to_node() == {"kind": "float32"}

    def test_with_constraints(self):
        schema = v.number().min(0).max(100).exclusive_min(-1).exclusive_max(101).multiple_of(5)
        node = schema._to_node()
        assert node["min"] == 0
        assert node["max"] == 100
        assert node["exclusiveMin"] == -1
        assert node["exclusiveMax"] == 101
        assert node["multipleOf"] == 5


class TestIntegerExport:
    def test_int_plain(self):
        assert v.int_()._to_node()["kind"] == "int"

    def test_int8(self):
        assert v.int8()._to_node()["kind"] == "int8"

    def test_int16(self):
        assert v.int16()._to_node()["kind"] == "int16"

    def test_int32(self):
        assert v.int32()._to_node()["kind"] == "int32"

    def test_int64(self):
        assert v.int64()._to_node()["kind"] == "int64"

    def test_uint8(self):
        assert v.uint8()._to_node()["kind"] == "uint8"

    def test_uint16(self):
        assert v.uint16()._to_node()["kind"] == "uint16"

    def test_uint32(self):
        assert v.uint32()._to_node()["kind"] == "uint32"

    def test_uint64(self):
        assert v.uint64()._to_node()["kind"] == "uint64"

    def test_int_with_constraints(self):
        schema = v.int_().min(0).max(100).exclusive_min(-1).exclusive_max(101).multiple_of(2)
        node = schema._to_node()
        assert node["min"] == 0
        assert node["max"] == 100
        assert node["exclusiveMin"] == -1
        assert node["exclusiveMax"] == 101
        assert node["multipleOf"] == 2


class TestBoolExport:
    def test_plain(self):
        assert v.bool_()._to_node() == {"kind": "bool"}


class TestNullExport:
    def test_plain(self):
        assert v.null()._to_node() == {"kind": "null"}


class TestAnyExport:
    def test_plain(self):
        assert v.any_()._to_node() == {"kind": "any"}


class TestUnknownExport:
    def test_plain(self):
        assert v.unknown()._to_node() == {"kind": "unknown"}


class TestNeverExport:
    def test_plain(self):
        assert v.never()._to_node() == {"kind": "never"}


class TestLiteralExport:
    def test_string_literal(self):
        node = v.literal("hello")._to_node()
        assert node == {"kind": "literal", "value": "hello"}

    def test_int_literal(self):
        node = v.literal(42)._to_node()
        assert node == {"kind": "literal", "value": 42}

    def test_bool_literal(self):
        node = v.literal(True)._to_node()
        assert node == {"kind": "literal", "value": True}

    def test_null_literal(self):
        node = v.literal(None)._to_node()
        assert node == {"kind": "literal", "value": None}


class TestEnumExport:
    def test_string_enum(self):
        node = v.enum_(["a", "b", "c"])._to_node()
        assert node == {"kind": "enum", "values": ["a", "b", "c"]}


class TestArrayExport:
    def test_plain(self):
        node = v.array(v.string())._to_node()
        assert node["kind"] == "array"
        assert node["items"] == {"kind": "string"}

    def test_with_constraints(self):
        schema = v.array(v.int_()).min_items(1).max_items(10)
        node = schema._to_node()
        assert node["minItems"] == 1
        assert node["maxItems"] == 10


class TestTupleExport:
    def test_basic(self):
        node = v.tuple_([v.string(), v.int_()])._to_node()
        assert node["kind"] == "tuple"
        assert len(node["elements"]) == 2
        assert node["elements"][0]["kind"] == "string"
        assert node["elements"][1]["kind"] == "int"


class TestObjectExport:
    def test_basic(self):
        node = v.object_({"name": v.string()})._to_node()
        assert node["kind"] == "object"
        assert "name" in node["properties"]
        assert node["unknownKeys"] == "reject"

    def test_with_required(self):
        schema = v.object_({"a": v.string(), "b": v.int_()}, required=["a"])
        node = schema._to_node()
        assert node["required"] == ["a"]

    def test_strip_mode(self):
        schema = v.object_({"x": v.string()}, unknown_keys="strip")
        node = schema._to_node()
        assert node["unknownKeys"] == "strip"

    def test_allow_mode(self):
        schema = v.object_({"x": v.string()}, unknown_keys="allow")
        node = schema._to_node()
        assert node["unknownKeys"] == "allow"


class TestRecordExport:
    def test_basic(self):
        node = v.record(v.int_())._to_node()
        assert node["kind"] == "record"
        assert node["values"]["kind"] == "int"


class TestUnionExport:
    def test_basic(self):
        node = v.union([v.string(), v.int_()])._to_node()
        assert node["kind"] == "union"
        assert len(node["variants"]) == 2


class TestIntersectionExport:
    def test_basic(self):
        s1 = v.object_({"a": v.string()}, unknown_keys="allow")
        s2 = v.object_({"b": v.int_()}, unknown_keys="allow")
        node = v.intersection([s1, s2])._to_node()
        assert node["kind"] == "intersection"
        assert len(node["allOf"]) == 2


class TestOptionalExport:
    def test_basic(self):
        node = v.optional(v.string())._to_node()
        assert node["kind"] == "optional"
        assert node["schema"]["kind"] == "string"


class TestNullableExport:
    def test_basic(self):
        node = v.nullable(v.string())._to_node()
        assert node["kind"] == "nullable"
        assert node["schema"]["kind"] == "string"


class TestRefExport:
    def test_basic(self):
        node = v.ref("#/definitions/Foo")._to_node()
        assert node["kind"] == "ref"
        assert node["ref"] == "#/definitions/Foo"


class TestCommonNodeFields:
    def test_with_default(self):
        schema = v.string().default("hi")
        node = schema._to_node()
        assert node["default"] == "hi"

    def test_with_coercion(self):
        schema = v.string().coerce(trim=True, lower=True)
        node = schema._to_node()
        assert node["coerce"]["trim"] is True
        assert node["coerce"]["lower"] is True

    def test_with_all_coercions(self):
        schema = v.int_().coerce(to_int=True, to_number=True, to_bool=True, upper=True)
        node = schema._to_node()
        assert node["coerce"]["toInt"] is True
        assert node["coerce"]["toNumber"] is True
        assert node["coerce"]["toBool"] is True
        assert node["coerce"]["upper"] is True

    def test_no_coercion_when_none(self):
        node = v.string()._to_node()
        assert "coerce" not in node


class TestExportHelpers:
    def test_export_schema_with_definitions(self):
        doc = export_schema(
            v.ref("#/definitions/User"),
            definitions={"User": v.object_({"name": v.string()})},
        )
        assert "definitions" in doc
        assert "User" in doc["definitions"]

    def test_export_schema_extended_with_extensions(self):
        doc = export_schema(
            v.string(),
            mode="extended",
            extensions={"x-custom": True},
        )
        assert doc["extensions"] == {"x-custom": True}

    def test_export_schema_portable_ignores_extensions(self):
        doc = export_schema(
            v.string(),
            mode="portable",
            extensions={"x-custom": True},
        )
        assert "extensions" not in doc or doc.get("extensions") == {}

    def test_export_schema_json(self):
        json_str = export_schema_json(v.string(), indent=None)
        parsed = json.loads(json_str)
        assert parsed["root"]["kind"] == "string"

    def test_export_schema_json_pretty(self):
        json_str = export_schema_json(v.string(), indent=2)
        assert "\n" in json_str
        parsed = json.loads(json_str)
        assert parsed["root"]["kind"] == "string"

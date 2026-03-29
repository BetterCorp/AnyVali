"""Unit tests for standalone parse/safe_parse functions and document helpers."""

from __future__ import annotations

import pytest

import anyvali as v
from anyvali.parse.parser import parse, safe_parse
from anyvali.interchange.document import (
    create_document,
    document_from_json,
    document_to_json,
)
from anyvali.types import ValidationError


class TestStandaloneParse:
    def test_parse_success(self):
        result = parse(v.string(), "hello")
        assert result == "hello"

    def test_parse_failure_raises(self):
        with pytest.raises(ValidationError) as exc_info:
            parse(v.string(), 42)
        assert len(exc_info.value.issues) == 1
        assert exc_info.value.issues[0].code == v.INVALID_TYPE

    def test_parse_int(self):
        result = parse(v.int_(), 42)
        assert result == 42

    def test_parse_int_failure(self):
        with pytest.raises(ValidationError):
            parse(v.int_(), "not an int")

    def test_parse_bool(self):
        result = parse(v.bool_(), True)
        assert result is True

    def test_parse_null(self):
        result = parse(v.null(), None)
        assert result is None

    def test_parse_array(self):
        result = parse(v.array(v.int_()), [1, 2, 3])
        assert result == [1, 2, 3]


class TestStandaloneSafeParse:
    def test_safe_parse_success(self):
        result = safe_parse(v.string(), "hello")
        assert result.success is True
        assert result.data == "hello"

    def test_safe_parse_failure(self):
        result = safe_parse(v.string(), 42)
        assert result.success is False
        assert len(result.issues) > 0

    def test_safe_parse_number(self):
        result = safe_parse(v.number(), 3.14)
        assert result.success
        assert result.data == 3.14

    def test_safe_parse_object(self):
        schema = v.object_({"x": v.int_()})
        result = safe_parse(schema, {"x": 1})
        assert result.success


class TestDocumentHelpers:
    def test_create_document(self):
        doc = create_document({"kind": "string"})
        assert doc.root == {"kind": "string"}
        assert doc.definitions == {}
        assert doc.extensions == {}

    def test_create_document_with_defs_and_extensions(self):
        doc = create_document(
            {"kind": "ref", "ref": "#/definitions/A"},
            definitions={"A": {"kind": "string"}},
            extensions={"x": 1},
        )
        assert doc.definitions == {"A": {"kind": "string"}}
        assert doc.extensions == {"x": 1}

    def test_document_to_json(self):
        doc = create_document({"kind": "bool"})
        json_str = document_to_json(doc)
        assert '"kind": "bool"' in json_str

    def test_document_to_json_no_indent(self):
        doc = create_document({"kind": "bool"})
        json_str = document_to_json(doc, indent=None)
        assert "\n" not in json_str

    def test_document_from_json(self):
        json_str = '{"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"int"}}'
        doc = document_from_json(json_str)
        assert doc.root == {"kind": "int"}
        assert doc.anyvali_version == "1.0"

    def test_roundtrip(self):
        doc = create_document(
            {"kind": "string"},
            definitions={"Foo": {"kind": "int"}},
        )
        json_str = document_to_json(doc)
        doc2 = document_from_json(json_str)
        assert doc2.root == doc.root
        assert doc2.definitions == doc.definitions

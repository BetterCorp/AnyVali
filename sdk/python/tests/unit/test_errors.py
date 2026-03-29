"""Unit tests for ValidationError, ParseResult, ValidationIssue, and AnyValiDocument."""

from __future__ import annotations

import pytest

import anyvali as v
from anyvali.types import AnyValiDocument, ParseResult, ValidationError, ValidationIssue


class TestValidationIssue:
    def test_basic_fields(self):
        issue = ValidationIssue(
            code="invalid_type",
            message="Expected string, received integer",
            path=["name"],
            expected="string",
            received="integer",
        )
        assert issue.code == "invalid_type"
        assert issue.message == "Expected string, received integer"
        assert issue.path == ["name"]
        assert issue.expected == "string"
        assert issue.received == "integer"
        assert issue.meta is None

    def test_defaults(self):
        issue = ValidationIssue(code="test", message="test msg")
        assert issue.path == []
        assert issue.expected is None
        assert issue.received is None
        assert issue.meta is None

    def test_with_meta(self):
        issue = ValidationIssue(
            code="custom",
            message="msg",
            meta={"key": "value"},
        )
        assert issue.meta == {"key": "value"}

    def test_with_int_path(self):
        issue = ValidationIssue(code="test", message="msg", path=[0, "name"])
        assert issue.path == [0, "name"]

    def test_frozen(self):
        issue = ValidationIssue(code="test", message="msg")
        with pytest.raises(AttributeError):
            issue.code = "other"  # type: ignore


class TestParseResult:
    def test_success(self):
        result = ParseResult(success=True, data="hello")
        assert result.success is True
        assert result.data == "hello"
        assert result.issues == []

    def test_failure(self):
        issues = [ValidationIssue(code="invalid_type", message="bad")]
        result = ParseResult(success=False, issues=issues)
        assert result.success is False
        assert result.data is None
        assert len(result.issues) == 1

    def test_frozen(self):
        result = ParseResult(success=True)
        with pytest.raises(AttributeError):
            result.success = False  # type: ignore


class TestValidationError:
    def test_construction(self):
        issues = [
            ValidationIssue(code="invalid_type", message="Expected string"),
            ValidationIssue(code="too_small", message="Too short"),
        ]
        err = ValidationError(issues)
        assert err.issues == issues
        assert "Validation failed:" in str(err)
        assert "Expected string" in str(err)
        assert "Too short" in str(err)

    def test_single_issue(self):
        issues = [ValidationIssue(code="required", message="Field missing")]
        err = ValidationError(issues)
        assert "Field missing" in str(err)

    def test_is_exception(self):
        err = ValidationError([ValidationIssue(code="x", message="y")])
        assert isinstance(err, Exception)

    def test_parse_raises_validation_error(self):
        schema = v.string()
        with pytest.raises(ValidationError) as exc_info:
            schema.parse(42)
        assert len(exc_info.value.issues) == 1
        assert exc_info.value.issues[0].code == v.INVALID_TYPE

    def test_parse_returns_value_on_success(self):
        schema = v.string()
        result = schema.parse("hello")
        assert result == "hello"


class TestAnyValiDocument:
    def test_to_dict_basic(self):
        doc = AnyValiDocument(root={"kind": "string"})
        d = doc.to_dict()
        assert d["anyvaliVersion"] == "1.0"
        assert d["schemaVersion"] == "1"
        assert d["root"] == {"kind": "string"}
        assert "definitions" not in d
        assert "extensions" not in d

    def test_to_dict_with_definitions(self):
        doc = AnyValiDocument(
            root={"kind": "ref", "ref": "#/definitions/Foo"},
            definitions={"Foo": {"kind": "string"}},
        )
        d = doc.to_dict()
        assert d["definitions"] == {"Foo": {"kind": "string"}}

    def test_to_dict_with_extensions(self):
        doc = AnyValiDocument(
            root={"kind": "string"},
            extensions={"x-custom": True},
        )
        d = doc.to_dict()
        assert d["extensions"] == {"x-custom": True}

    def test_from_dict(self):
        raw = {
            "anyvaliVersion": "2.0",
            "schemaVersion": "2",
            "root": {"kind": "int"},
            "definitions": {"X": {"kind": "bool"}},
            "extensions": {"ext": 1},
        }
        doc = AnyValiDocument.from_dict(raw)
        assert doc.anyvali_version == "2.0"
        assert doc.schema_version == "2"
        assert doc.root == {"kind": "int"}
        assert doc.definitions == {"X": {"kind": "bool"}}
        assert doc.extensions == {"ext": 1}

    def test_from_dict_defaults(self):
        doc = AnyValiDocument.from_dict({})
        assert doc.anyvali_version == "1.0"
        assert doc.schema_version == "1"
        assert doc.root == {}
        assert doc.definitions == {}
        assert doc.extensions == {}

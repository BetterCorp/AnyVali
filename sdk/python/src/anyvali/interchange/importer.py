"""Import schemas from AnyVali interchange format."""

from __future__ import annotations

import json
from typing import Any

from ..schemas.any import AnySchema
from ..schemas.array import ArraySchema
from ..schemas.base import BaseSchema, CoercionConfig
from ..schemas.bool import BoolSchema
from ..schemas.enum import EnumSchema
from ..schemas.integer import (
    Int8Schema,
    Int16Schema,
    Int32Schema,
    Int64Schema,
    IntSchema,
    Uint8Schema,
    Uint16Schema,
    Uint32Schema,
    Uint64Schema,
)
from ..schemas.intersection import IntersectionSchema
from ..schemas.literal import LiteralSchema
from ..schemas.never import NeverSchema
from ..schemas.null import NullSchema
from ..schemas.nullable import NullableSchema
from ..schemas.number import Float32Schema, Float64Schema, NumberSchema
from ..schemas.object import ObjectSchema
from ..schemas.optional import OptionalSchema
from ..schemas.record import RecordSchema
from ..schemas.ref import RefSchema
from ..schemas.string import StringSchema
from ..schemas.tuple import TupleSchema
from ..schemas.union import UnionSchema
from ..schemas.unknown import UnknownSchema
from ..types import AnyValiDocument


def import_schema(source: dict[str, Any] | str) -> BaseSchema[Any]:
    """Import a schema from an AnyVali document dict or JSON string.

    Returns the root schema.
    """
    if isinstance(source, str):
        source = json.loads(source)
    assert isinstance(source, dict)

    doc = AnyValiDocument.from_dict(source)

    # Two-pass definition building to handle recursive refs:
    # Pass 1: Create placeholder RefSchema entries so recursive refs can find them
    definitions: dict[str, BaseSchema] = {}

    # Pass 2: Build actual definitions (refs inside will get lazy resolution)
    for name, node in doc.definitions.items():
        definitions[name] = _import_node(node, definitions)

    root = _import_node(doc.root, definitions)

    # Resolve all ref schemas (including those nested in definitions)
    for defn_schema in definitions.values():
        _resolve_refs(defn_schema, definitions)
    _resolve_refs(root, definitions)

    return root


def _parse_coercion(raw: Any) -> CoercionConfig | None:
    """Parse coercion from interchange format.

    Corpus format: either a single string like "string->int" or "trim",
    or a list like ["trim", "lower"].
    """
    if raw is None:
        return None

    flags: dict[str, bool] = {}

    if isinstance(raw, str):
        tokens = [raw]
    elif isinstance(raw, list):
        tokens = raw
    elif isinstance(raw, dict):
        # Dict format (our own export format)
        return CoercionConfig(
            to_int=raw.get("toInt", False),
            to_number=raw.get("toNumber", False),
            to_bool=raw.get("toBool", False),
            trim=raw.get("trim", False),
            lower=raw.get("lower", False),
            upper=raw.get("upper", False),
        )
    else:
        return None

    for token in tokens:
        if token == "string->int":
            flags["to_int"] = True
        elif token == "string->number":
            flags["to_number"] = True
        elif token == "string->bool":
            flags["to_bool"] = True
        elif token == "trim":
            flags["trim"] = True
        elif token == "lower":
            flags["lower"] = True
        elif token == "upper":
            flags["upper"] = True

    if not flags:
        return None

    return CoercionConfig(
        to_int=flags.get("to_int", False),
        to_number=flags.get("to_number", False),
        to_bool=flags.get("to_bool", False),
        trim=flags.get("trim", False),
        lower=flags.get("lower", False),
        upper=flags.get("upper", False),
    )


def _import_node(node: dict[str, Any], definitions: dict[str, BaseSchema]) -> BaseSchema:
    """Import a single schema node from its dict representation."""
    kind = node.get("kind", "")

    schema = _build_schema(kind, node, definitions)

    # Apply common fields
    if "default" in node:
        schema._has_default = True
        schema._default_value = node["default"]

    if "coerce" in node:
        schema._coercion = _parse_coercion(node["coerce"])

    return schema


def _build_int_kwargs(node: dict[str, Any]) -> dict[str, Any]:
    """Extract numeric constraint kwargs from a node."""
    kw: dict[str, Any] = {}
    if "min" in node:
        kw["min"] = node["min"]
    if "max" in node:
        kw["max"] = node["max"]
    if "exclusiveMin" in node:
        kw["exclusive_min"] = node["exclusiveMin"]
    if "exclusiveMax" in node:
        kw["exclusive_max"] = node["exclusiveMax"]
    if "multipleOf" in node:
        kw["multiple_of"] = node["multipleOf"]
    return kw


def _build_schema(
    kind: str, node: dict[str, Any], definitions: dict[str, BaseSchema]
) -> BaseSchema:
    """Build a schema from a kind string and node dict."""
    if kind == "string":
        return StringSchema(
            min_length=node.get("minLength"),
            max_length=node.get("maxLength"),
            pattern=node.get("pattern"),
            starts_with=node.get("startsWith"),
            ends_with=node.get("endsWith"),
            includes=node.get("includes"),
            format_=node.get("format"),
        )
    elif kind in ("number", "float64"):
        cls = NumberSchema if kind == "number" else Float64Schema
        return cls(**_build_int_kwargs(node))
    elif kind == "float32":
        return Float32Schema(**_build_int_kwargs(node))
    elif kind in ("int", "int64"):
        cls = IntSchema if kind == "int" else Int64Schema
        return cls(**_build_int_kwargs(node))
    elif kind == "int8":
        return Int8Schema(**_build_int_kwargs(node))
    elif kind == "int16":
        return Int16Schema(**_build_int_kwargs(node))
    elif kind == "int32":
        return Int32Schema(**_build_int_kwargs(node))
    elif kind == "uint8":
        return Uint8Schema(**_build_int_kwargs(node))
    elif kind == "uint16":
        return Uint16Schema(**_build_int_kwargs(node))
    elif kind == "uint32":
        return Uint32Schema(**_build_int_kwargs(node))
    elif kind == "uint64":
        return Uint64Schema(**_build_int_kwargs(node))
    elif kind == "bool":
        return BoolSchema()
    elif kind == "null":
        return NullSchema()
    elif kind == "any":
        return AnySchema()
    elif kind == "unknown":
        return UnknownSchema()
    elif kind == "never":
        return NeverSchema()
    elif kind == "literal":
        return LiteralSchema(node["value"])
    elif kind == "enum":
        return EnumSchema(node["values"])
    elif kind == "array":
        items_schema = _import_node(node["items"], definitions)
        return ArraySchema(
            items_schema,
            min_items=node.get("minItems"),
            max_items=node.get("maxItems"),
        )
    elif kind == "tuple":
        # Support both "items" and "elements" keys
        items_raw = node.get("elements") or node.get("items", [])
        items = [_import_node(item, definitions) for item in items_raw]
        return TupleSchema(items)
    elif kind == "object":
        props: dict[str, BaseSchema] = {}
        for pname, pnode in node.get("properties", {}).items():
            props[pname] = _import_node(pnode, definitions)
        required = node.get("required")
        unknown_keys = node.get("unknownKeys", "reject")
        return ObjectSchema(props, required=required, unknown_keys=unknown_keys)
    elif kind == "record":
        value_schema = _import_node(node["values"], definitions)
        return RecordSchema(value_schema)
    elif kind == "union":
        # Support both "schemas" and "variants" keys
        schemas_raw = node.get("variants") or node.get("schemas", [])
        schemas = [_import_node(s, definitions) for s in schemas_raw]
        return UnionSchema(schemas)
    elif kind == "intersection":
        # Support both "schemas" and "allOf" keys
        schemas_raw = node.get("allOf") or node.get("schemas", [])
        schemas = [_import_node(s, definitions) for s in schemas_raw]
        return IntersectionSchema(schemas)
    elif kind == "optional":
        inner = _import_node(node["schema"], definitions)
        return OptionalSchema(inner)
    elif kind == "nullable":
        inner = _import_node(node["schema"], definitions)
        return NullableSchema(inner)
    elif kind == "ref":
        return RefSchema(node["ref"])
    else:
        raise ValueError(f"Unsupported schema kind: {kind!r}")


def _resolve_refs(schema: BaseSchema, definitions: dict[str, BaseSchema]) -> None:
    """Resolve RefSchema instances to their target schemas."""
    if isinstance(schema, RefSchema):
        # Give ref access to the full definitions dict for lazy resolution
        schema.set_definitions(definitions)
        ref_name = schema._ref
        if ref_name.startswith("#/definitions/"):
            ref_name = ref_name[len("#/definitions/"):]
        if ref_name in definitions:
            # Don't eagerly resolve recursive refs -- lazy resolution handles it
            schema.resolve(definitions[ref_name])
    elif isinstance(schema, (OptionalSchema, NullableSchema)):
        _resolve_refs(schema._inner, definitions)
    elif isinstance(schema, ArraySchema):
        _resolve_refs(schema._items, definitions)
    elif isinstance(schema, TupleSchema):
        for item in schema._items:
            _resolve_refs(item, definitions)
    elif isinstance(schema, ObjectSchema):
        for prop_schema in schema._properties.values():
            _resolve_refs(prop_schema, definitions)
    elif isinstance(schema, RecordSchema):
        _resolve_refs(schema._value_schema, definitions)
    elif isinstance(schema, (UnionSchema, IntersectionSchema)):
        for s in schema._schemas:
            _resolve_refs(s, definitions)

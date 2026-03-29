"""AnyVali Python SDK – portable schema validation.

Usage::

    import anyvali as v

    schema = v.object_({
        "name": v.string().min_length(1),
        "age": v.int_().min(0),
    })

    result = schema.safe_parse({"name": "Alice", "age": 30})
    assert result.success
"""

from __future__ import annotations

from .interchange.exporter import export_schema, export_schema_json
from .interchange.importer import import_schema
from .issue_codes import (
    COERCION_FAILED,
    CUSTOM_VALIDATION_NOT_PORTABLE,
    DEFAULT_INVALID,
    INVALID_LITERAL,
    INVALID_NUMBER,
    INVALID_STRING,
    INVALID_TYPE,
    INVALID_UNION,
    REQUIRED,
    TOO_LARGE,
    TOO_SMALL,
    UNKNOWN_KEY,
    UNSUPPORTED_EXTENSION,
    UNSUPPORTED_SCHEMA_KIND,
)
from .parse.parser import parse, safe_parse
from .schemas.any import AnySchema
from .schemas.array import ArraySchema
from .schemas.base import BaseSchema
from .schemas.bool import BoolSchema
from .schemas.enum import EnumSchema
from .schemas.integer import (
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
from .schemas.intersection import IntersectionSchema
from .schemas.literal import LiteralSchema
from .schemas.never import NeverSchema
from .schemas.null import NullSchema
from .schemas.nullable import NullableSchema
from .schemas.number import Float32Schema, Float64Schema, NumberSchema
from .schemas.object import ObjectSchema
from .schemas.optional import OptionalSchema
from .schemas.record import RecordSchema
from .schemas.ref import RefSchema
from .schemas.string import StringSchema
from .schemas.tuple import TupleSchema
from .schemas.union import UnionSchema
from .schemas.unknown import UnknownSchema
from .types import (
    AnyValiDocument,
    ExportMode,
    ParseResult,
    UnknownKeyMode,
    ValidationError,
    ValidationIssue,
)

__version__ = "0.0.1"


# ── Builder functions (public API) ────────────────────────────────


def string(**kw: object) -> StringSchema:
    """Create a string schema."""
    return StringSchema()


def number(**kw: object) -> NumberSchema:
    """Create a number schema (float64)."""
    return NumberSchema()


def float32(**kw: object) -> Float32Schema:
    """Create a float32 schema."""
    return Float32Schema()


def float64(**kw: object) -> Float64Schema:
    """Create a float64 schema."""
    return Float64Schema()


def int_(**kw: object) -> IntSchema:
    """Create an int schema (int64)."""
    return IntSchema()


def int8(**kw: object) -> Int8Schema:
    """Create an int8 schema."""
    return Int8Schema()


def int16(**kw: object) -> Int16Schema:
    """Create an int16 schema."""
    return Int16Schema()


def int32(**kw: object) -> Int32Schema:
    """Create an int32 schema."""
    return Int32Schema()


def int64(**kw: object) -> Int64Schema:
    """Create an int64 schema."""
    return Int64Schema()


def uint8(**kw: object) -> Uint8Schema:
    """Create a uint8 schema."""
    return Uint8Schema()


def uint16(**kw: object) -> Uint16Schema:
    """Create a uint16 schema."""
    return Uint16Schema()


def uint32(**kw: object) -> Uint32Schema:
    """Create a uint32 schema."""
    return Uint32Schema()


def uint64(**kw: object) -> Uint64Schema:
    """Create a uint64 schema."""
    return Uint64Schema()


def bool_(**kw: object) -> BoolSchema:
    """Create a boolean schema."""
    return BoolSchema()


def null(**kw: object) -> NullSchema:
    """Create a null schema."""
    return NullSchema()


def any_(**kw: object) -> AnySchema:
    """Create an any schema."""
    return AnySchema()


def unknown(**kw: object) -> UnknownSchema:
    """Create an unknown schema."""
    return UnknownSchema()


def never(**kw: object) -> NeverSchema:
    """Create a never schema."""
    return NeverSchema()


def literal(value: object) -> LiteralSchema:
    """Create a literal schema for a specific value."""
    return LiteralSchema(value)


def enum_(values: list[object]) -> EnumSchema:
    """Create an enum schema."""
    return EnumSchema(values)


def array(items: BaseSchema, **kw: object) -> ArraySchema:
    """Create an array schema."""
    return ArraySchema(items)


def tuple_(items: list[BaseSchema]) -> TupleSchema:
    """Create a tuple schema."""
    return TupleSchema(items)


def object_(
    properties: dict[str, BaseSchema],
    *,
    required: list[str] | None = None,
    unknown_keys: UnknownKeyMode = "reject",
) -> ObjectSchema:
    """Create an object schema."""
    return ObjectSchema(properties, required=required, unknown_keys=unknown_keys)


def record(value_schema: BaseSchema) -> RecordSchema:
    """Create a record schema (dict with uniform value type)."""
    return RecordSchema(value_schema)


def union(schemas: list[BaseSchema]) -> UnionSchema:
    """Create a union schema."""
    return UnionSchema(schemas)


def intersection(schemas: list[BaseSchema]) -> IntersectionSchema:
    """Create an intersection schema."""
    return IntersectionSchema(schemas)


def optional(schema: BaseSchema) -> OptionalSchema:
    """Create an optional schema wrapper."""
    return OptionalSchema(schema)


def nullable(schema: BaseSchema) -> NullableSchema:
    """Create a nullable schema wrapper."""
    return NullableSchema(schema)


def ref(reference: str) -> RefSchema:
    """Create a ref schema."""
    return RefSchema(reference)


__all__ = [
    # Builder functions
    "string",
    "number",
    "float32",
    "float64",
    "int_",
    "int8",
    "int16",
    "int32",
    "int64",
    "uint8",
    "uint16",
    "uint32",
    "uint64",
    "bool_",
    "null",
    "any_",
    "unknown",
    "never",
    "literal",
    "enum_",
    "array",
    "tuple_",
    "object_",
    "record",
    "union",
    "intersection",
    "optional",
    "nullable",
    "ref",
    # Parse
    "parse",
    "safe_parse",
    # Interchange
    "export_schema",
    "export_schema_json",
    "import_schema",
    # Types
    "BaseSchema",
    "ParseResult",
    "ValidationIssue",
    "ValidationError",
    "AnyValiDocument",
    "ExportMode",
    "UnknownKeyMode",
    # Issue codes
    "INVALID_TYPE",
    "REQUIRED",
    "UNKNOWN_KEY",
    "TOO_SMALL",
    "TOO_LARGE",
    "INVALID_STRING",
    "INVALID_NUMBER",
    "INVALID_LITERAL",
    "INVALID_UNION",
    "CUSTOM_VALIDATION_NOT_PORTABLE",
    "UNSUPPORTED_EXTENSION",
    "UNSUPPORTED_SCHEMA_KIND",
    "COERCION_FAILED",
    "DEFAULT_INVALID",
]

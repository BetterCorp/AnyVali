pub mod format;
pub mod interchange;
pub mod issue_codes;
pub mod parse;
pub mod schema;
pub mod schemas;
pub mod typed;
pub mod types;

// Re-export key types
pub use schema::{ParseContext, Schema};
pub use schemas::*;
pub use typed::{parse_as, TypedSchema};

/// Any AnyVali schema. Use as a generic constraint. Equivalent to Zod's ZodTypeAny.
pub type SchemaAny = Box<dyn Schema>;
pub use types::{
    AnyValiDocument, ExportMode, ParseResult, PathSegment, UnknownKeyMode, ValidationError,
    ValidationIssue,
};

// Convenience builder functions
use serde_json::Value;

/// Create a new string schema.
pub fn string() -> StringSchema {
    StringSchema::new()
}

/// Create a new number (float64) schema.
pub fn number() -> NumberSchema {
    NumberSchema::new()
}

/// Create a new float64 schema.
pub fn float64() -> NumberSchema {
    NumberSchema::float64()
}

/// Create a new float32 schema.
pub fn float32() -> NumberSchema {
    NumberSchema::float32()
}

/// Create a new int (int64) schema.
pub fn int() -> IntSchema {
    IntSchema::new()
}

/// Create an int8 schema.
pub fn int8() -> IntSchema {
    IntSchema::with_width(IntWidth::Int8)
}

/// Create an int16 schema.
pub fn int16() -> IntSchema {
    IntSchema::with_width(IntWidth::Int16)
}

/// Create an int32 schema.
pub fn int32() -> IntSchema {
    IntSchema::with_width(IntWidth::Int32)
}

/// Create an int64 schema.
pub fn int64() -> IntSchema {
    IntSchema::with_width(IntWidth::Int64)
}

/// Create a uint8 schema.
pub fn uint8() -> IntSchema {
    IntSchema::with_width(IntWidth::UInt8)
}

/// Create a uint16 schema.
pub fn uint16() -> IntSchema {
    IntSchema::with_width(IntWidth::UInt16)
}

/// Create a uint32 schema.
pub fn uint32() -> IntSchema {
    IntSchema::with_width(IntWidth::UInt32)
}

/// Create a uint64 schema.
pub fn uint64() -> IntSchema {
    IntSchema::with_width(IntWidth::UInt64)
}

/// Create a bool schema.
pub fn bool_() -> BoolSchema {
    BoolSchema::new()
}

/// Create a null schema.
pub fn null() -> NullSchema {
    NullSchema::new()
}

/// Create an any schema.
pub fn any() -> AnySchema {
    AnySchema::new()
}

/// Create an unknown schema.
pub fn unknown() -> UnknownSchema {
    UnknownSchema::new()
}

/// Create a never schema.
pub fn never() -> NeverSchema {
    NeverSchema::new()
}

/// Create a literal schema.
pub fn literal(value: Value) -> LiteralSchema {
    LiteralSchema::new(value)
}

/// Create an enum schema.
pub fn enum_(values: Vec<Value>) -> EnumSchema {
    EnumSchema::new(values)
}

/// Create an array schema.
pub fn array(items: Box<dyn Schema>) -> ArraySchema {
    ArraySchema::new(items)
}

/// Create a tuple schema.
pub fn tuple(elements: Vec<Box<dyn Schema>>) -> TupleSchema {
    TupleSchema::new(elements)
}

/// Create an object schema.
pub fn object() -> ObjectSchema {
    ObjectSchema::new()
}

/// Create a record schema.
pub fn record(values: Box<dyn Schema>) -> RecordSchema {
    RecordSchema::new(values)
}

/// Create a union schema.
pub fn union(variants: Vec<Box<dyn Schema>>) -> UnionSchema {
    UnionSchema::new(variants)
}

/// Create an intersection schema.
pub fn intersection(all_of: Vec<Box<dyn Schema>>) -> IntersectionSchema {
    IntersectionSchema::new(all_of)
}

/// Create an optional schema.
pub fn optional(schema: Box<dyn Schema>) -> OptionalSchema {
    OptionalSchema::new(schema)
}

/// Create a nullable schema.
pub fn nullable(schema: Box<dyn Schema>) -> NullableSchema {
    NullableSchema::new(schema)
}

/// Create a ref schema.
pub fn ref_(ref_path: &str) -> RefSchema {
    RefSchema::new(ref_path)
}

/// Export a schema to an AnyValiDocument.
pub fn export(
    schema: &dyn Schema,
    mode: ExportMode,
    definitions: &std::collections::HashMap<String, Box<dyn Schema>>,
) -> Result<AnyValiDocument, String> {
    interchange::exporter::export_schema(schema, mode, definitions)
}

/// Import a schema from a JSON string.
pub fn import(json_str: &str) -> Result<(Box<dyn Schema>, ParseContext), String> {
    interchange::importer::import_from_json(json_str)
}

/// Import a schema from a serde_json::Value.
pub fn import_value(value: &Value) -> Result<(Box<dyn Schema>, ParseContext), String> {
    interchange::importer::import_from_value(value)
}

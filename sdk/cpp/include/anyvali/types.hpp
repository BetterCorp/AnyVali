#pragma once

#include <stdexcept>
#include <string>
#include <variant>
#include <vector>
#include <cstdint>

namespace anyvali {

enum class ExportMode {
    Portable,
    Extended
};

enum class UnknownKeyMode {
    Reject,
    Strip,
    Allow
};

enum class SchemaKind {
    Any,
    Unknown,
    Never,
    Null,
    Bool,
    String,
    Number,    // alias for Float64
    Int,       // alias for Int64
    Float32,
    Float64,
    Int8,
    Int16,
    Int32,
    Int64,
    Uint8,
    Uint16,
    Uint32,
    Uint64,
    Literal,
    Enum,
    Array,
    Tuple,
    Object,
    Record,
    Union,
    Intersection,
    Optional,
    Nullable,
    Ref
};

inline std::string kind_to_string(SchemaKind k) {
    switch (k) {
        case SchemaKind::Any: return "any";
        case SchemaKind::Unknown: return "unknown";
        case SchemaKind::Never: return "never";
        case SchemaKind::Null: return "null";
        case SchemaKind::Bool: return "bool";
        case SchemaKind::String: return "string";
        case SchemaKind::Number: return "number";
        case SchemaKind::Int: return "int";
        case SchemaKind::Float32: return "float32";
        case SchemaKind::Float64: return "float64";
        case SchemaKind::Int8: return "int8";
        case SchemaKind::Int16: return "int16";
        case SchemaKind::Int32: return "int32";
        case SchemaKind::Int64: return "int64";
        case SchemaKind::Uint8: return "uint8";
        case SchemaKind::Uint16: return "uint16";
        case SchemaKind::Uint32: return "uint32";
        case SchemaKind::Uint64: return "uint64";
        case SchemaKind::Literal: return "literal";
        case SchemaKind::Enum: return "enum";
        case SchemaKind::Array: return "array";
        case SchemaKind::Tuple: return "tuple";
        case SchemaKind::Object: return "object";
        case SchemaKind::Record: return "record";
        case SchemaKind::Union: return "union";
        case SchemaKind::Intersection: return "intersection";
        case SchemaKind::Optional: return "optional";
        case SchemaKind::Nullable: return "nullable";
        case SchemaKind::Ref: return "ref";
    }
    return "unknown";
}

inline SchemaKind string_to_kind(const std::string& s) {
    if (s == "any") return SchemaKind::Any;
    if (s == "unknown") return SchemaKind::Unknown;
    if (s == "never") return SchemaKind::Never;
    if (s == "null") return SchemaKind::Null;
    if (s == "bool") return SchemaKind::Bool;
    if (s == "string") return SchemaKind::String;
    if (s == "number") return SchemaKind::Number;
    if (s == "int") return SchemaKind::Int;
    if (s == "float32") return SchemaKind::Float32;
    if (s == "float64") return SchemaKind::Float64;
    if (s == "int8") return SchemaKind::Int8;
    if (s == "int16") return SchemaKind::Int16;
    if (s == "int32") return SchemaKind::Int32;
    if (s == "int64") return SchemaKind::Int64;
    if (s == "uint8") return SchemaKind::Uint8;
    if (s == "uint16") return SchemaKind::Uint16;
    if (s == "uint32") return SchemaKind::Uint32;
    if (s == "uint64") return SchemaKind::Uint64;
    if (s == "literal") return SchemaKind::Literal;
    if (s == "enum") return SchemaKind::Enum;
    if (s == "array") return SchemaKind::Array;
    if (s == "tuple") return SchemaKind::Tuple;
    if (s == "object") return SchemaKind::Object;
    if (s == "record") return SchemaKind::Record;
    if (s == "union") return SchemaKind::Union;
    if (s == "intersection") return SchemaKind::Intersection;
    if (s == "optional") return SchemaKind::Optional;
    if (s == "nullable") return SchemaKind::Nullable;
    if (s == "ref") return SchemaKind::Ref;
    throw std::runtime_error("Unsupported schema kind: " + s);
}

using PathSegment = std::variant<std::string, int>;
using Path = std::vector<PathSegment>;

inline std::string json_type_name(int type) {
    // nlohmann::json type enum: null=0, object=1, array=2, string=3,
    // boolean=4, number_integer=5, number_unsigned=6, number_float=7, binary=8, discarded=9
    switch (type) {
        case 0: return "null";
        case 1: return "object";
        case 2: return "array";
        case 3: return "string";
        case 4: return "boolean";
        case 5: return "number";
        case 6: return "number";
        case 7: return "number";
        default: return "unknown";
    }
}

} // namespace anyvali

#pragma once

// Core types
#include "types.hpp"
#include "issue_codes.hpp"
#include "validation_issue.hpp"
#include "validation_error.hpp"
#include "parse_result.hpp"
#include "validation_context.hpp"
#include "anyvali_document.hpp"
#include "schema.hpp"

// Schema types
#include "schemas/any_schema.hpp"
#include "schemas/unknown_schema.hpp"
#include "schemas/never_schema.hpp"
#include "schemas/null_schema.hpp"
#include "schemas/bool_schema.hpp"
#include "schemas/string_schema.hpp"
#include "schemas/number_schema.hpp"
#include "schemas/int_schema.hpp"
#include "schemas/literal_schema.hpp"
#include "schemas/enum_schema.hpp"
#include "schemas/array_schema.hpp"
#include "schemas/tuple_schema.hpp"
#include "schemas/object_schema.hpp"
#include "schemas/record_schema.hpp"
#include "schemas/union_schema.hpp"
#include "schemas/intersection_schema.hpp"
#include "schemas/optional_schema.hpp"
#include "schemas/nullable_schema.hpp"
#include "schemas/ref_schema.hpp"

// Parsing support
#include "parse/coercion.hpp"
#include "parse/coercion_config.hpp"
#include "parse/defaults.hpp"

// Format validators
#include "format/validators.hpp"

// Typed parse helpers
#include "typed_parse.hpp"

// Interchange
#include "interchange/importer.hpp"
#include "interchange/exporter.hpp"

namespace anyvali {

// ---- Builder functions (factory helpers) ----

inline std::shared_ptr<StringSchema> string_() {
    return std::make_shared<StringSchema>();
}

inline std::shared_ptr<NumberSchema> number() {
    return std::make_shared<NumberSchema>(SchemaKind::Number);
}

inline std::shared_ptr<NumberSchema> int_() {
    return std::make_shared<NumberSchema>(SchemaKind::Int);
}

inline std::shared_ptr<NumberSchema> int8() {
    return std::make_shared<NumberSchema>(SchemaKind::Int8);
}

inline std::shared_ptr<NumberSchema> int16() {
    return std::make_shared<NumberSchema>(SchemaKind::Int16);
}

inline std::shared_ptr<NumberSchema> int32() {
    return std::make_shared<NumberSchema>(SchemaKind::Int32);
}

inline std::shared_ptr<NumberSchema> int64() {
    return std::make_shared<NumberSchema>(SchemaKind::Int64);
}

inline std::shared_ptr<NumberSchema> uint8() {
    return std::make_shared<NumberSchema>(SchemaKind::Uint8);
}

inline std::shared_ptr<NumberSchema> uint16() {
    return std::make_shared<NumberSchema>(SchemaKind::Uint16);
}

inline std::shared_ptr<NumberSchema> uint32() {
    return std::make_shared<NumberSchema>(SchemaKind::Uint32);
}

inline std::shared_ptr<NumberSchema> uint64() {
    return std::make_shared<NumberSchema>(SchemaKind::Uint64);
}

inline std::shared_ptr<NumberSchema> float32() {
    return std::make_shared<NumberSchema>(SchemaKind::Float32);
}

inline std::shared_ptr<NumberSchema> float64() {
    return std::make_shared<NumberSchema>(SchemaKind::Float64);
}

inline std::shared_ptr<BoolSchema> bool_() {
    return std::make_shared<BoolSchema>();
}

inline std::shared_ptr<NullSchema> null() {
    return std::make_shared<NullSchema>();
}

inline std::shared_ptr<AnySchema> any() {
    return std::make_shared<AnySchema>();
}

inline std::shared_ptr<UnknownSchema> unknown() {
    return std::make_shared<UnknownSchema>();
}

inline std::shared_ptr<NeverSchema> never() {
    return std::make_shared<NeverSchema>();
}

inline std::shared_ptr<LiteralSchema> literal(const nlohmann::json& value) {
    return std::make_shared<LiteralSchema>(value);
}

inline std::shared_ptr<EnumSchema> enum_(std::vector<nlohmann::json> values) {
    return std::make_shared<EnumSchema>(std::move(values));
}

inline std::shared_ptr<ArraySchema> array(std::shared_ptr<Schema> items) {
    return std::make_shared<ArraySchema>(std::move(items));
}

inline std::shared_ptr<TupleSchema> tuple(std::vector<std::shared_ptr<Schema>> elements) {
    return std::make_shared<TupleSchema>(std::move(elements));
}

inline std::shared_ptr<ObjectSchema> object() {
    return std::make_shared<ObjectSchema>();
}

inline std::shared_ptr<RecordSchema> record(std::shared_ptr<Schema> values) {
    return std::make_shared<RecordSchema>(std::move(values));
}

inline std::shared_ptr<UnionSchema> union_(std::vector<std::shared_ptr<Schema>> variants) {
    return std::make_shared<UnionSchema>(std::move(variants));
}

inline std::shared_ptr<IntersectionSchema> intersection(std::vector<std::shared_ptr<Schema>> schemas) {
    return std::make_shared<IntersectionSchema>(std::move(schemas));
}

inline std::shared_ptr<OptionalSchema> optional_(std::shared_ptr<Schema> inner) {
    return std::make_shared<OptionalSchema>(std::move(inner));
}

inline std::shared_ptr<NullableSchema> nullable(std::shared_ptr<Schema> inner) {
    return std::make_shared<NullableSchema>(std::move(inner));
}

inline std::shared_ptr<RefSchema> ref(const std::string& ref_path) {
    return std::make_shared<RefSchema>(ref_path);
}

// ---- Document-level parse helpers ----

inline ParseResult parse_document(const AnyValiDocument& doc, const nlohmann::json& input) {
    return doc.root->safe_parse_with_defs(input, doc.definitions);
}

inline nlohmann::json parse_document_or_throw(const AnyValiDocument& doc, const nlohmann::json& input) {
    return doc.root->parse_with_defs(input, doc.definitions);
}

// ---- Import/Export convenience ----

inline AnyValiDocument import_schema(const nlohmann::json& doc) {
    return interchange::import_document(doc);
}

inline nlohmann::json export_schema(const AnyValiDocument& doc,
                                     ExportMode mode = ExportMode::Portable) {
    return interchange::export_document(doc, mode);
}

} // namespace anyvali

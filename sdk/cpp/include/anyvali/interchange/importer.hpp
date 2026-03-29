#pragma once

#include <memory>
#include <stdexcept>
#include <nlohmann/json.hpp>
#include "../schema.hpp"
#include "../anyvali_document.hpp"
#include "../schemas/string_schema.hpp"
#include "../schemas/number_schema.hpp"
#include "../schemas/bool_schema.hpp"
#include "../schemas/null_schema.hpp"
#include "../schemas/any_schema.hpp"
#include "../schemas/unknown_schema.hpp"
#include "../schemas/never_schema.hpp"
#include "../schemas/literal_schema.hpp"
#include "../schemas/enum_schema.hpp"
#include "../schemas/array_schema.hpp"
#include "../schemas/tuple_schema.hpp"
#include "../schemas/object_schema.hpp"
#include "../schemas/record_schema.hpp"
#include "../schemas/union_schema.hpp"
#include "../schemas/intersection_schema.hpp"
#include "../schemas/optional_schema.hpp"
#include "../schemas/nullable_schema.hpp"
#include "../schemas/ref_schema.hpp"

namespace anyvali {
namespace interchange {

inline std::shared_ptr<Schema> import_node(const nlohmann::json& node);

inline std::shared_ptr<Schema> import_node(const nlohmann::json& node) {
    if (!node.is_object() || !node.contains("kind")) {
        throw std::runtime_error("Invalid schema node: missing 'kind'");
    }

    std::string kind_str = node["kind"].get<std::string>();
    SchemaKind k = string_to_kind(kind_str);

    std::shared_ptr<Schema> schema;

    switch (k) {
        case SchemaKind::Any:
            schema = std::make_shared<AnySchema>();
            break;

        case SchemaKind::Unknown:
            schema = std::make_shared<UnknownSchema>();
            break;

        case SchemaKind::Never:
            schema = std::make_shared<NeverSchema>();
            break;

        case SchemaKind::Null:
            schema = std::make_shared<NullSchema>();
            break;

        case SchemaKind::Bool: {
            auto bs = std::make_shared<BoolSchema>();
            if (node.contains("coerce")) {
                if (node["coerce"].is_string()) {
                    bs->coerce(node["coerce"].get<std::string>());
                }
            }
            schema = bs;
            break;
        }

        case SchemaKind::String: {
            auto ss = std::make_shared<StringSchema>();
            if (node.contains("minLength")) ss->minLength(node["minLength"].get<int64_t>());
            if (node.contains("maxLength")) ss->maxLength(node["maxLength"].get<int64_t>());
            if (node.contains("pattern")) ss->pattern(node["pattern"].get<std::string>());
            if (node.contains("startsWith")) ss->startsWith(node["startsWith"].get<std::string>());
            if (node.contains("endsWith")) ss->endsWith(node["endsWith"].get<std::string>());
            if (node.contains("includes")) ss->includes(node["includes"].get<std::string>());
            if (node.contains("format")) ss->format(node["format"].get<std::string>());
            if (node.contains("coerce")) {
                if (node["coerce"].is_string()) {
                    ss->coerce(node["coerce"].get<std::string>());
                } else if (node["coerce"].is_array()) {
                    std::vector<std::string> coercions;
                    for (const auto& c : node["coerce"]) {
                        coercions.push_back(c.get<std::string>());
                    }
                    ss->coerce(coercions);
                }
            }
            schema = ss;
            break;
        }

        case SchemaKind::Number:
        case SchemaKind::Float32:
        case SchemaKind::Float64:
        case SchemaKind::Int:
        case SchemaKind::Int8:
        case SchemaKind::Int16:
        case SchemaKind::Int32:
        case SchemaKind::Int64:
        case SchemaKind::Uint8:
        case SchemaKind::Uint16:
        case SchemaKind::Uint32:
        case SchemaKind::Uint64: {
            auto ns = std::make_shared<NumberSchema>(k);
            if (node.contains("min")) ns->min(node["min"].get<double>());
            if (node.contains("max")) ns->max(node["max"].get<double>());
            if (node.contains("exclusiveMin")) ns->exclusiveMin(node["exclusiveMin"].get<double>());
            if (node.contains("exclusiveMax")) ns->exclusiveMax(node["exclusiveMax"].get<double>());
            if (node.contains("multipleOf")) ns->multipleOf(node["multipleOf"].get<double>());
            if (node.contains("coerce")) {
                if (node["coerce"].is_string()) {
                    ns->coerce(node["coerce"].get<std::string>());
                }
            }
            schema = ns;
            break;
        }

        case SchemaKind::Literal:
            schema = std::make_shared<LiteralSchema>(node["value"]);
            break;

        case SchemaKind::Enum: {
            std::vector<nlohmann::json> values;
            for (const auto& v : node["values"]) {
                values.push_back(v);
            }
            schema = std::make_shared<EnumSchema>(std::move(values));
            break;
        }

        case SchemaKind::Array: {
            auto items = import_node(node["items"]);
            auto as = std::make_shared<ArraySchema>(items);
            if (node.contains("minItems")) as->minItems(node["minItems"].get<int64_t>());
            if (node.contains("maxItems")) as->maxItems(node["maxItems"].get<int64_t>());
            schema = as;
            break;
        }

        case SchemaKind::Tuple: {
            std::vector<std::shared_ptr<Schema>> elements;
            for (const auto& e : node["elements"]) {
                elements.push_back(import_node(e));
            }
            schema = std::make_shared<TupleSchema>(std::move(elements));
            break;
        }

        case SchemaKind::Object: {
            auto os = std::make_shared<ObjectSchema>();
            if (node.contains("properties")) {
                for (auto it = node["properties"].begin(); it != node["properties"].end(); ++it) {
                    os->prop(it.key(), import_node(it.value()));
                }
            }
            if (node.contains("required")) {
                std::vector<std::string> req;
                for (const auto& r : node["required"]) {
                    req.push_back(r.get<std::string>());
                }
                os->required(req);
            }
            if (node.contains("unknownKeys")) {
                std::string mode = node["unknownKeys"].get<std::string>();
                if (mode == "strip") os->unknownKeys(UnknownKeyMode::Strip);
                else if (mode == "allow") os->unknownKeys(UnknownKeyMode::Allow);
                else os->unknownKeys(UnknownKeyMode::Reject);
            }
            // default unknownKeys is reject (already set)
            schema = os;
            break;
        }

        case SchemaKind::Record: {
            auto values = import_node(node["values"]);
            schema = std::make_shared<RecordSchema>(values);
            break;
        }

        case SchemaKind::Union: {
            std::vector<std::shared_ptr<Schema>> variants;
            for (const auto& v : node["variants"]) {
                variants.push_back(import_node(v));
            }
            schema = std::make_shared<UnionSchema>(std::move(variants));
            break;
        }

        case SchemaKind::Intersection: {
            std::vector<std::shared_ptr<Schema>> schemas;
            for (const auto& s : node["allOf"]) {
                schemas.push_back(import_node(s));
            }
            schema = std::make_shared<IntersectionSchema>(std::move(schemas));
            break;
        }

        case SchemaKind::Optional: {
            auto inner = import_node(node["schema"]);
            schema = std::make_shared<OptionalSchema>(inner);
            break;
        }

        case SchemaKind::Nullable: {
            auto inner = import_node(node["schema"]);
            schema = std::make_shared<NullableSchema>(inner);
            break;
        }

        case SchemaKind::Ref:
            schema = std::make_shared<RefSchema>(node["ref"].get<std::string>());
            break;

        default:
            throw std::runtime_error("Unsupported schema kind: " + kind_str);
    }

    // Apply default if present
    if (node.contains("default")) {
        schema->set_default(node["default"]);
    }

    return schema;
}

inline AnyValiDocument import_document(const nlohmann::json& doc) {
    if (!doc.contains("anyvaliVersion") || !doc.contains("schemaVersion") || !doc.contains("root")) {
        throw std::runtime_error("Invalid AnyVali document: missing required fields");
    }

    AnyValiDocument result;
    result.anyvali_version = doc["anyvaliVersion"].get<std::string>();
    result.schema_version = doc["schemaVersion"].get<std::string>();

    // Import definitions first (for refs)
    if (doc.contains("definitions") && doc["definitions"].is_object()) {
        for (auto it = doc["definitions"].begin(); it != doc["definitions"].end(); ++it) {
            result.definitions[it.key()] = import_node(it.value());
        }
    }

    result.root = import_node(doc["root"]);

    if (doc.contains("extensions")) {
        result.extensions = doc["extensions"];
    }

    return result;
}

} // namespace interchange
} // namespace anyvali

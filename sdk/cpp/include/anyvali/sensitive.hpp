#pragma once

#include <functional>
#include <map>

#include <nlohmann/json.hpp>

#include "interchange/importer.hpp"
#include "schema.hpp"

namespace anyvali {

using SensitiveTransform =
    std::function<nlohmann::json(const Path&, const nlohmann::json&)>;

namespace sensitive_detail {

inline constexpr const char* encrypted_prefix = "encrypted:";

inline bool is_sensitive(const nlohmann::json& node) {
    return node.contains("metadata") && node["metadata"].is_object() &&
           node["metadata"].value("sensitive", false);
}

inline nlohmann::json encrypted_node(const nlohmann::json& node) {
    if (!node.is_object()) return node;

    if (is_sensitive(node)) {
        auto marker = nlohmann::json{{"kind", "string"},
                                     {"minLength", 11},
                                     {"startsWith", encrypted_prefix}};
        const auto kind = node.value("kind", "");
        if (kind == "optional" || kind == "nullable") {
            return nlohmann::json{{"kind", kind}, {"schema", marker}};
        }
        return marker;
    }

    auto result = node;
    const auto kind = node.value("kind", "");
    if (kind == "object" && node.contains("properties")) {
        for (auto& [name, child] : result["properties"].items()) {
            child = encrypted_node(child);
        }
    } else if ((kind == "array" || kind == "record") &&
               node.contains(kind == "array" ? "items" : "values")) {
        const auto key = kind == "array" ? "items" : "values";
        result[key] = encrypted_node(node[key]);
    } else if (kind == "tuple" || kind == "union" || kind == "intersection") {
        const auto key = kind == "tuple" ? "elements" : kind == "union" ? "variants" : "allOf";
        for (auto& child : result[key]) child = encrypted_node(child);
    } else if ((kind == "optional" || kind == "nullable") && node.contains("schema")) {
        result["schema"] = encrypted_node(node["schema"]);
    }
    return result;
}

inline nlohmann::json path_key(const Path& path) {
    auto result = nlohmann::json::array();
    for (const auto& segment : path) {
        if (std::holds_alternative<std::string>(segment)) {
            result.push_back(std::get<std::string>(segment));
        } else {
            result.push_back(std::get<int>(segment));
        }
    }
    return result;
}

inline nlohmann::json transform_node(const nlohmann::json& node,
                                     const nlohmann::json& value,
                                     const SensitiveTransform& transform,
                                     bool encrypting,
                                     Path& path,
                                     std::map<std::string, nlohmann::json>& cache) {
    if (!node.is_object()) return value;

    if (is_sensitive(node) && !value.is_null()) {
        if (encrypting) interchange::import_node(node)->parse(value);
        const auto key = path_key(path).dump();
        if (const auto found = cache.find(key); found != cache.end()) return found->second;
        auto result = transform(path, value);
        cache.emplace(key, result);
        return result;
    }

    const auto kind = node.value("kind", "");
    if (kind == "object" && value.is_object()) {
        auto result = value;
        if (node.contains("properties")) {
            for (const auto& [name, child] : node["properties"].items()) {
                if (!value.contains(name)) continue;
                path.emplace_back(name);
                result[name] = transform_node(child, value[name], transform, encrypting, path, cache);
                path.pop_back();
            }
        }
        return result;
    }
    if ((kind == "array" || kind == "tuple") && value.is_array()) {
        auto result = value;
        for (std::size_t index = 0; index < value.size(); ++index) {
            const nlohmann::json* child = nullptr;
            if (kind == "array" && node.contains("items")) child = &node["items"];
            if (kind == "tuple" && node.contains("elements") && index < node["elements"].size()) {
                child = &node["elements"][index];
            }
            if (!child) continue;
            path.emplace_back(static_cast<int>(index));
            result[index] = transform_node(*child, value[index], transform, encrypting, path, cache);
            path.pop_back();
        }
        return result;
    }
    if (kind == "record" && value.is_object() && node.contains("values")) {
        auto result = value;
        for (const auto& [name, child_value] : value.items()) {
            path.emplace_back(name);
            result[name] = transform_node(node["values"], child_value, transform, encrypting, path, cache);
            path.pop_back();
        }
        return result;
    }
    if ((kind == "optional" || kind == "nullable") && node.contains("schema")) {
        return transform_node(node["schema"], value, transform, encrypting, path, cache);
    }
    if (kind == "union" && node.contains("variants")) {
        for (const auto& child : node["variants"]) {
            const auto candidate = encrypting ? child : encrypted_node(child);
            if (interchange::import_node(candidate)->safe_parse(value).success) {
                return transform_node(child, value, transform, encrypting, path, cache);
            }
        }
    }
    if (kind == "intersection" && node.contains("allOf")) {
        auto result = value;
        for (const auto& child : node["allOf"]) {
            result = transform_node(child, result, transform, encrypting, path, cache);
        }
        return result;
    }
    return value;
}

} // namespace sensitive_detail

inline ParseResult safe_parse_encrypted(const Schema& schema, const nlohmann::json& data) {
    return interchange::import_node(sensitive_detail::encrypted_node(schema.export_node()))
        ->safe_parse(data);
}

inline nlohmann::json encrypt(const Schema& schema,
                              const nlohmann::json& data,
                              const SensitiveTransform& transform) {
    const auto plain = schema.parse(data);
    Path path;
    std::map<std::string, nlohmann::json> cache;
    const auto encrypted = sensitive_detail::transform_node(
        schema.export_node(), plain, transform, true, path, cache);
    return interchange::import_node(sensitive_detail::encrypted_node(schema.export_node()))
        ->parse(encrypted);
}

inline nlohmann::json decrypt(const Schema& schema,
                              const nlohmann::json& data,
                              const SensitiveTransform& transform) {
    const auto encrypted = interchange::import_node(
        sensitive_detail::encrypted_node(schema.export_node()))->parse(data);
    Path path;
    std::map<std::string, nlohmann::json> cache;
    const auto plain = sensitive_detail::transform_node(
        schema.export_node(), encrypted, transform, false, path, cache);
    return schema.parse(plain);
}

} // namespace anyvali

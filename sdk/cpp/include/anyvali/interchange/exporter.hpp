#pragma once

#include <nlohmann/json.hpp>
#include "../schema.hpp"
#include "../anyvali_document.hpp"
#include "../types.hpp"
#include <stdexcept>

namespace anyvali {
namespace interchange {

inline nlohmann::json export_document(const AnyValiDocument& doc,
                                       ExportMode mode = ExportMode::Portable) {
    if (mode == ExportMode::Portable && doc.root->has_custom_validators()) {
        throw std::runtime_error("Cannot export in portable mode: schema contains custom validators");
    }

    nlohmann::json result;
    result["anyvaliVersion"] = doc.anyvali_version;
    result["schemaVersion"] = doc.schema_version;
    result["root"] = doc.root->export_node();

    nlohmann::json defs = nlohmann::json::object();
    for (const auto& [name, schema] : doc.definitions) {
        if (mode == ExportMode::Portable && schema->has_custom_validators()) {
            throw std::runtime_error(
                "Cannot export in portable mode: definition '" + name + "' contains custom validators");
        }
        defs[name] = schema->export_node();
    }
    result["definitions"] = defs;

    if (mode == ExportMode::Extended) {
        result["extensions"] = doc.extensions;
    } else {
        result["extensions"] = nlohmann::json::object();
    }

    return result;
}

} // namespace interchange
} // namespace anyvali

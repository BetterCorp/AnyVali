#pragma once

#include <string>
#include <map>
#include <memory>
#include <nlohmann/json.hpp>

namespace anyvali {

class Schema;

struct AnyValiDocument {
    std::string anyvali_version = "1.0";
    std::string schema_version = "1";
    std::shared_ptr<Schema> root;
    std::map<std::string, std::shared_ptr<Schema>> definitions;
    nlohmann::json extensions;

    AnyValiDocument() : extensions(nlohmann::json::object()) {}
};

} // namespace anyvali

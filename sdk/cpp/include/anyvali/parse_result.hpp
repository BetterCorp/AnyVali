#pragma once

#include <vector>
#include <optional>
#include <nlohmann/json.hpp>
#include "validation_issue.hpp"

namespace anyvali {

struct ParseResult {
    bool success;
    nlohmann::json value;
    std::vector<ValidationIssue> issues;

    static ParseResult ok(nlohmann::json val) {
        return ParseResult{true, std::move(val), {}};
    }

    static ParseResult fail(std::vector<ValidationIssue> issues) {
        return ParseResult{false, nullptr, std::move(issues)};
    }

    nlohmann::json to_json() const {
        nlohmann::json j;
        j["success"] = success;
        j["value"] = value;
        nlohmann::json iss = nlohmann::json::array();
        for (const auto& i : issues) {
            iss.push_back(i.to_json());
        }
        j["issues"] = iss;
        return j;
    }
};

} // namespace anyvali

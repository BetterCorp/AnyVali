#pragma once

#include <string>
#include <vector>
#include "types.hpp"
#include <nlohmann/json.hpp>

namespace anyvali {

struct ValidationIssue {
    std::string code;
    std::string message;
    Path path;
    std::string expected;
    std::string received;
    nlohmann::json meta;

    ValidationIssue() = default;
    ValidationIssue(std::string code_, Path path_, std::string expected_, std::string received_,
                    std::string message_ = "")
        : code(std::move(code_)), message(std::move(message_)),
          path(std::move(path_)), expected(std::move(expected_)),
          received(std::move(received_)) {}

    nlohmann::json to_json() const {
        nlohmann::json j;
        j["code"] = code;
        j["message"] = message;

        nlohmann::json p = nlohmann::json::array();
        for (const auto& seg : path) {
            if (std::holds_alternative<std::string>(seg)) {
                p.push_back(std::get<std::string>(seg));
            } else {
                p.push_back(std::get<int>(seg));
            }
        }
        j["path"] = p;
        j["expected"] = expected;
        j["received"] = received;
        if (!meta.is_null()) {
            j["meta"] = meta;
        }
        return j;
    }
};

} // namespace anyvali

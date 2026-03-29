#pragma once

#include <string>
#include <optional>
#include <algorithm>
#include <cctype>
#include <nlohmann/json.hpp>

namespace anyvali {
namespace coercion {

// Attempt coercion of a JSON value according to the given coercion name.
// Returns the coerced value on success, or std::nullopt on failure.
inline std::optional<nlohmann::json> apply(const std::string& coercion_name,
                                            const nlohmann::json& value) {
    if (coercion_name == "string->int") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        // Trim whitespace
        size_t start = s.find_first_not_of(" \t\n\r");
        size_t end = s.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return std::nullopt;
        s = s.substr(start, end - start + 1);
        try {
            size_t pos = 0;
            int64_t result = std::stoll(s, &pos);
            if (pos != s.size()) return std::nullopt;
            return nlohmann::json(result);
        } catch (...) {
            return std::nullopt;
        }
    }

    if (coercion_name == "string->number") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        size_t start = s.find_first_not_of(" \t\n\r");
        size_t end = s.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return std::nullopt;
        s = s.substr(start, end - start + 1);
        try {
            size_t pos = 0;
            double result = std::stod(s, &pos);
            if (pos != s.size()) return std::nullopt;
            return nlohmann::json(result);
        } catch (...) {
            return std::nullopt;
        }
    }

    if (coercion_name == "string->bool") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        // Trim
        size_t start = s.find_first_not_of(" \t\n\r");
        size_t end = s.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return std::nullopt;
        s = s.substr(start, end - start + 1);
        // Lowercase
        std::string lower;
        lower.reserve(s.size());
        for (char c : s) lower += static_cast<char>(std::tolower(static_cast<unsigned char>(c)));
        if (lower == "true" || lower == "1") return nlohmann::json(true);
        if (lower == "false" || lower == "0") return nlohmann::json(false);
        return std::nullopt;
    }

    if (coercion_name == "trim") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        size_t start = s.find_first_not_of(" \t\n\r");
        size_t end = s.find_last_not_of(" \t\n\r");
        if (start == std::string::npos) return nlohmann::json("");
        return nlohmann::json(s.substr(start, end - start + 1));
    }

    if (coercion_name == "lower") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        std::string lower;
        lower.reserve(s.size());
        for (char c : s) lower += static_cast<char>(std::tolower(static_cast<unsigned char>(c)));
        return nlohmann::json(lower);
    }

    if (coercion_name == "upper") {
        if (!value.is_string()) return std::nullopt;
        std::string s = value.get<std::string>();
        std::string upper;
        upper.reserve(s.size());
        for (char c : s) upper += static_cast<char>(std::toupper(static_cast<unsigned char>(c)));
        return nlohmann::json(upper);
    }

    return std::nullopt;
}

// Apply a chain of coercions left to right
inline std::optional<nlohmann::json> apply_chain(const std::vector<std::string>& coercions,
                                                  const nlohmann::json& value) {
    nlohmann::json current = value;
    for (const auto& c : coercions) {
        auto result = apply(c, current);
        if (!result.has_value()) return std::nullopt;
        current = result.value();
    }
    return current;
}

} // namespace coercion
} // namespace anyvali

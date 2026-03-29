#pragma once

#include "../schema.hpp"
#include "../format/validators.hpp"
#include "../parse/coercion.hpp"
#include <optional>
#include <regex>

namespace anyvali {

class StringSchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::String; }

    StringSchema& minLength(int64_t n) { min_length_ = n; return *this; }
    StringSchema& maxLength(int64_t n) { max_length_ = n; return *this; }
    StringSchema& pattern(const std::string& p) { pattern_ = p; return *this; }
    StringSchema& startsWith(const std::string& p) { starts_with_ = p; return *this; }
    StringSchema& endsWith(const std::string& p) { ends_with_ = p; return *this; }
    StringSchema& includes(const std::string& p) { includes_ = p; return *this; }
    StringSchema& format(const std::string& f) { format_ = f; return *this; }
    StringSchema& coerce(const std::string& c) { coercions_ = {c}; return *this; }
    StringSchema& coerce(const std::vector<std::string>& c) { coercions_ = c; return *this; }
    StringSchema& defaultValue(const nlohmann::json& v) { default_value_ = v; return *this; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        nlohmann::json value = input;

        // Apply coercions if present
        if (!coercions_.empty() && !value.is_null()) {
            auto coerced = coercion::apply_chain(coercions_, value);
            if (coerced.has_value()) {
                value = coerced.value();
            } else {
                ctx.add_issue(issue_codes::COERCION_FAILED, kind_to_string(kind()),
                              json_to_string(input), "Coercion failed");
                return nullptr;
            }
        }

        // Type check
        if (!value.is_string()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "string", get_json_type(value));
            return nullptr;
        }

        std::string s = value.get<std::string>();

        // Constraints
        if (min_length_.has_value() && static_cast<int64_t>(s.size()) < min_length_.value()) {
            ctx.add_issue(issue_codes::TOO_SMALL, std::to_string(min_length_.value()),
                          std::to_string(s.size()));
            return nullptr;
        }
        if (max_length_.has_value() && static_cast<int64_t>(s.size()) > max_length_.value()) {
            ctx.add_issue(issue_codes::TOO_LARGE, std::to_string(max_length_.value()),
                          std::to_string(s.size()));
            return nullptr;
        }
        if (pattern_.has_value()) {
            std::regex re(pattern_.value());
            if (!std::regex_match(s, re)) {
                ctx.add_issue(issue_codes::INVALID_STRING, pattern_.value(), s);
                return nullptr;
            }
        }
        if (starts_with_.has_value()) {
            if (s.find(starts_with_.value()) != 0) {
                ctx.add_issue(issue_codes::INVALID_STRING, starts_with_.value(), s);
                return nullptr;
            }
        }
        if (ends_with_.has_value()) {
            const auto& suffix = ends_with_.value();
            if (s.size() < suffix.size() || s.compare(s.size() - suffix.size(), suffix.size(), suffix) != 0) {
                ctx.add_issue(issue_codes::INVALID_STRING, suffix, s);
                return nullptr;
            }
        }
        if (includes_.has_value()) {
            if (s.find(includes_.value()) == std::string::npos) {
                ctx.add_issue(issue_codes::INVALID_STRING, includes_.value(), s);
                return nullptr;
            }
        }
        if (format_.has_value()) {
            if (!format::validate_format(format_.value(), s)) {
                ctx.add_issue(issue_codes::INVALID_STRING, format_.value(), s);
                return nullptr;
            }
        }

        return value;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "string";
        if (min_length_.has_value()) node["minLength"] = min_length_.value();
        if (max_length_.has_value()) node["maxLength"] = max_length_.value();
        if (pattern_.has_value()) node["pattern"] = pattern_.value();
        if (starts_with_.has_value()) node["startsWith"] = starts_with_.value();
        if (ends_with_.has_value()) node["endsWith"] = ends_with_.value();
        if (includes_.has_value()) node["includes"] = includes_.value();
        if (format_.has_value()) node["format"] = format_.value();
        export_common_fields(node);
        return node;
    }

private:
    std::optional<int64_t> min_length_;
    std::optional<int64_t> max_length_;
    std::optional<std::string> pattern_;
    std::optional<std::string> starts_with_;
    std::optional<std::string> ends_with_;
    std::optional<std::string> includes_;
    std::optional<std::string> format_;
};

} // namespace anyvali

#pragma once

#include <memory>
#include <optional>
#include <string>
#include <vector>
#include <nlohmann/json.hpp>
#include "types.hpp"
#include "parse_result.hpp"
#include "validation_error.hpp"
#include "validation_context.hpp"
#include "issue_codes.hpp"

namespace anyvali {

class Schema : public std::enable_shared_from_this<Schema> {
public:
    virtual ~Schema() = default;

    virtual SchemaKind kind() const = 0;

    // Whether this schema has non-portable custom validators
    virtual bool has_custom_validators() const { return false; }

    // Default value (if set)
    virtual std::optional<nlohmann::json> get_default() const { return default_value_; }
    void set_default(const nlohmann::json& val) { default_value_ = val; }

    // Coercion config (if set)
    virtual std::vector<std::string> get_coercions() const { return coercions_; }
    void set_coercions(const std::vector<std::string>& c) { coercions_ = c; }

    // Core validation: validate input, apply coercion & defaults, return modified value
    virtual nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const = 0;

    // Export to JSON node
    virtual nlohmann::json export_node() const = 0;

    // Throwing parse
    nlohmann::json parse(const nlohmann::json& input) const {
        auto result = safe_parse(input);
        if (!result.success) {
            throw ValidationError(std::move(result.issues));
        }
        return result.value;
    }

    // Non-throwing parse
    ParseResult safe_parse(const nlohmann::json& input) const {
        ValidationContext ctx;
        auto output = validate(input, ctx);
        if (ctx.has_issues()) {
            return ParseResult::fail(std::move(ctx.issues));
        }
        return ParseResult::ok(std::move(output));
    }

    // Parse with definitions context (for refs)
    ParseResult safe_parse_with_defs(const nlohmann::json& input,
                                     const std::map<std::string, std::shared_ptr<Schema>>& defs) const {
        ValidationContext ctx;
        ctx.definitions = &defs;
        auto output = validate(input, ctx);
        if (ctx.has_issues()) {
            return ParseResult::fail(std::move(ctx.issues));
        }
        return ParseResult::ok(std::move(output));
    }

    nlohmann::json parse_with_defs(const nlohmann::json& input,
                                   const std::map<std::string, std::shared_ptr<Schema>>& defs) const {
        auto result = safe_parse_with_defs(input, defs);
        if (!result.success) {
            throw ValidationError(std::move(result.issues));
        }
        return result.value;
    }

protected:
    std::optional<nlohmann::json> default_value_;
    std::vector<std::string> coercions_;

    // Helper to get JSON type name for error messages
    static std::string get_json_type(const nlohmann::json& val) {
        if (val.is_null()) return "null";
        if (val.is_boolean()) return "boolean";
        if (val.is_number_integer() || val.is_number_unsigned()) return "number";
        if (val.is_number_float()) return "number";
        if (val.is_string()) return "string";
        if (val.is_array()) return "array";
        if (val.is_object()) return "object";
        return "unknown";
    }

    // Helper to stringify a JSON value for error messages
    static std::string json_to_string(const nlohmann::json& val) {
        if (val.is_string()) return val.get<std::string>();
        return val.dump();
    }

    // Export coercion and default fields
    void export_common_fields(nlohmann::json& node) const {
        if (!coercions_.empty()) {
            if (coercions_.size() == 1) {
                node["coerce"] = coercions_[0];
            } else {
                node["coerce"] = coercions_;
            }
        }
        if (default_value_.has_value()) {
            node["default"] = default_value_.value();
        }
    }
};

} // namespace anyvali

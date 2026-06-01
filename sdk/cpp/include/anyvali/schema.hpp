#pragma once

#include <memory>
#include <optional>
#include <set>
#include <string>
#include <vector>
#include <nlohmann/json.hpp>
#include "types.hpp"
#include "parse_result.hpp"
#include "validation_error.hpp"
#include "validation_context.hpp"
#include "issue_codes.hpp"

namespace anyvali {

struct DescribeOpts {
    std::optional<std::string> title;
    std::optional<bool> deprecated;
    std::optional<std::string> deprecated_message;
    std::optional<bool> not_stable;
    std::optional<std::string> since;
    std::optional<bool> sensitive;
    std::optional<bool> readonly_flag;  // 'readonly' may be reserved in some contexts
    std::optional<bool> writeonly_flag;
    std::optional<nlohmann::json> examples;  // should be json array
};

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

    // Metadata
    virtual std::optional<nlohmann::json> get_metadata() const { return metadata_; }
    void set_metadata_value(const nlohmann::json& val) { metadata_ = val; }

    // Reserved metadata key check
    static bool is_reserved_metadata_key(const std::string& key) {
        static const std::set<std::string> reserved = {
            "title", "description", "deprecated", "deprecatedMessage",
            "notStable", "since", "sensitive", "readonly", "writeonly", "examples"
        };
        return reserved.count(key) > 0;
    }

    // Describe - sets reserved metadata keys
    void describe(const std::string& description, const DescribeOpts& opts = {}) {
        if (!metadata_.has_value()) {
            metadata_ = nlohmann::json::object();
        }
        auto& meta = metadata_.value();
        meta["description"] = description;

        if (opts.title.has_value()) meta["title"] = opts.title.value();
        if (opts.deprecated.has_value()) meta["deprecated"] = opts.deprecated.value();
        if (opts.deprecated_message.has_value()) {
            if (!opts.deprecated.has_value() || !opts.deprecated.value()) {
                throw std::invalid_argument("describe(): deprecatedMessage requires deprecated to be true");
            }
            meta["deprecatedMessage"] = opts.deprecated_message.value();
        }
        if (opts.not_stable.has_value()) meta["notStable"] = opts.not_stable.value();
        if (opts.since.has_value()) meta["since"] = opts.since.value();
        if (opts.sensitive.has_value()) meta["sensitive"] = opts.sensitive.value();
        if (opts.readonly_flag.has_value()) meta["readonly"] = opts.readonly_flag.value();
        if (opts.writeonly_flag.has_value()) meta["writeonly"] = opts.writeonly_flag.value();
        if (opts.readonly_flag.value_or(false) && opts.writeonly_flag.value_or(false)) {
            throw std::invalid_argument("describe(): readonly and writeonly cannot both be true");
        }
        if (opts.examples.has_value()) meta["examples"] = opts.examples.value();
    }

    // Metadata - sets non-reserved keys
    void add_metadata(const nlohmann::json& meta, bool replace = false) {
        if (!meta.is_object()) {
            throw std::invalid_argument("metadata must be a JSON object");
        }
        for (auto& [key, val] : meta.items()) {
            if (is_reserved_metadata_key(key)) {
                throw std::invalid_argument(
                    "metadata(): \"" + key + "\" is a reserved key. Use describe() instead.");
            }
        }
        if (replace) {
            nlohmann::json preserved = nlohmann::json::object();
            if (metadata_.has_value()) {
                for (auto& [k, v] : metadata_.value().items()) {
                    if (is_reserved_metadata_key(k)) {
                        preserved[k] = v;
                    }
                }
            }
            for (auto& [k, v] : meta.items()) {
                preserved[k] = v;
            }
            metadata_ = preserved;
        } else {
            if (!metadata_.has_value()) {
                metadata_ = nlohmann::json::object();
            }
            for (auto& [k, v] : meta.items()) {
                metadata_.value()[k] = v;
            }
        }
    }

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
    std::optional<nlohmann::json> metadata_;

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
        if (metadata_.has_value() && !metadata_.value().empty()) {
            node["metadata"] = metadata_.value();
        }
    }
};

} // namespace anyvali

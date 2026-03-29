#pragma once

#include "../schema.hpp"
#include <map>
#include <set>

namespace anyvali {

class ObjectSchema : public Schema {
public:
    ObjectSchema() = default;

    SchemaKind kind() const override { return SchemaKind::Object; }

    ObjectSchema& prop(const std::string& name, std::shared_ptr<Schema> schema) {
        properties_[name] = std::move(schema);
        return *this;
    }

    ObjectSchema& required(const std::vector<std::string>& fields) {
        required_fields_.insert(fields.begin(), fields.end());
        return *this;
    }

    ObjectSchema& unknownKeys(UnknownKeyMode mode) {
        unknown_key_mode_ = mode;
        return *this;
    }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!input.is_object()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "object", get_json_type(input));
            return nullptr;
        }

        nlohmann::json result = nlohmann::json::object();
        bool has_errors = false;

        // Check required fields
        for (const auto& req : required_fields_) {
            if (!input.contains(req)) {
                auto it = properties_.find(req);
                // Check if property has a default
                if (it != properties_.end() && it->second->get_default().has_value()) {
                    // Will be handled below in property processing
                } else {
                    ctx.push_path(req);
                    std::string expected_type = "unknown";
                    auto pit = properties_.find(req);
                    if (pit != properties_.end()) {
                        expected_type = kind_to_string(pit->second->kind());
                    }
                    ctx.add_issue(issue_codes::REQUIRED, expected_type, "undefined");
                    ctx.pop_path();
                    has_errors = true;
                }
            }
        }

        // Process known properties
        for (const auto& [name, schema] : properties_) {
            if (input.contains(name)) {
                ctx.push_path(name);
                auto val = schema->validate(input[name], ctx);
                ctx.pop_path();
                if (val.is_null() && !input[name].is_null()) {
                    has_errors = true;
                } else {
                    result[name] = val;
                }
            } else if (schema->get_default().has_value()) {
                // Apply default
                auto default_val = schema->get_default().value();
                // Validate default
                ValidationContext def_ctx;
                def_ctx.definitions = ctx.definitions;
                auto validated = schema->validate(default_val, def_ctx);
                if (def_ctx.has_issues()) {
                    // Default is invalid
                    ctx.push_path(name);
                    // Use the first issue from default validation to build default_invalid
                    auto& first = def_ctx.issues[0];
                    ctx.add_issue(issue_codes::DEFAULT_INVALID, first.expected,
                                  first.received);
                    ctx.pop_path();
                    has_errors = true;
                } else {
                    result[name] = validated;
                }
            }
        }

        // Check unknown keys
        for (auto it = input.begin(); it != input.end(); ++it) {
            if (properties_.find(it.key()) == properties_.end()) {
                if (unknown_key_mode_ == UnknownKeyMode::Reject) {
                    ctx.push_path(it.key());
                    ctx.add_issue(issue_codes::UNKNOWN_KEY, "undefined", it.key());
                    ctx.pop_path();
                    has_errors = true;
                } else if (unknown_key_mode_ == UnknownKeyMode::Allow) {
                    result[it.key()] = it.value();
                }
                // Strip mode: just don't add it
            }
        }

        if (has_errors) return nullptr;
        return result;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "object";
        nlohmann::json props = nlohmann::json::object();
        for (const auto& [name, schema] : properties_) {
            props[name] = schema->export_node();
        }
        node["properties"] = props;
        nlohmann::json req = nlohmann::json::array();
        for (const auto& r : required_fields_) {
            req.push_back(r);
        }
        node["required"] = req;
        switch (unknown_key_mode_) {
            case UnknownKeyMode::Reject: node["unknownKeys"] = "reject"; break;
            case UnknownKeyMode::Strip: node["unknownKeys"] = "strip"; break;
            case UnknownKeyMode::Allow: node["unknownKeys"] = "allow"; break;
        }
        export_common_fields(node);
        return node;
    }

    const std::map<std::string, std::shared_ptr<Schema>>& properties() const { return properties_; }
    const std::set<std::string>& required_fields() const { return required_fields_; }
    UnknownKeyMode unknown_key_mode() const { return unknown_key_mode_; }

private:
    std::map<std::string, std::shared_ptr<Schema>> properties_;
    std::set<std::string> required_fields_;
    UnknownKeyMode unknown_key_mode_ = UnknownKeyMode::Reject;
};

} // namespace anyvali

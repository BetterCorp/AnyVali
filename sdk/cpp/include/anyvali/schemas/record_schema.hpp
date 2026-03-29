#pragma once

#include "../schema.hpp"

namespace anyvali {

class RecordSchema : public Schema {
public:
    explicit RecordSchema(std::shared_ptr<Schema> values) : values_(std::move(values)) {}

    SchemaKind kind() const override { return SchemaKind::Record; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!input.is_object()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "record", get_json_type(input));
            return nullptr;
        }

        nlohmann::json result = nlohmann::json::object();
        bool has_errors = false;

        for (auto it = input.begin(); it != input.end(); ++it) {
            ctx.push_path(it.key());
            auto val = values_->validate(it.value(), ctx);
            ctx.pop_path();
            if (val.is_null() && !it.value().is_null()) {
                has_errors = true;
            } else {
                result[it.key()] = val;
            }
        }

        if (has_errors) return nullptr;
        return result;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "record";
        node["values"] = values_->export_node();
        export_common_fields(node);
        return node;
    }

private:
    std::shared_ptr<Schema> values_;
};

} // namespace anyvali

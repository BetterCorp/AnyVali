#pragma once

#include "../schema.hpp"

namespace anyvali {

class IntersectionSchema : public Schema {
public:
    explicit IntersectionSchema(std::vector<std::shared_ptr<Schema>> schemas)
        : schemas_(std::move(schemas)) {}

    SchemaKind kind() const override { return SchemaKind::Intersection; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        nlohmann::json result = input;

        // Validate against all schemas; collect all issues
        for (const auto& schema : schemas_) {
            ValidationContext sub_ctx;
            sub_ctx.path = ctx.path;
            sub_ctx.definitions = ctx.definitions;
            auto val = schema->validate(input, sub_ctx);
            if (sub_ctx.has_issues()) {
                for (auto& issue : sub_ctx.issues) {
                    ctx.issues.push_back(std::move(issue));
                }
            } else {
                // Merge result (for objects, later validations add fields)
                if (val.is_object() && result.is_object()) {
                    for (auto it = val.begin(); it != val.end(); ++it) {
                        result[it.key()] = it.value();
                    }
                } else {
                    result = val;
                }
            }
        }

        if (ctx.has_issues()) return nullptr;
        return result;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "intersection";
        nlohmann::json all = nlohmann::json::array();
        for (const auto& s : schemas_) {
            all.push_back(s->export_node());
        }
        node["allOf"] = all;
        export_common_fields(node);
        return node;
    }

private:
    std::vector<std::shared_ptr<Schema>> schemas_;
};

} // namespace anyvali

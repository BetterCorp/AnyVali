#pragma once

#include "../schema.hpp"
#include <sstream>

namespace anyvali {

class UnionSchema : public Schema {
public:
    explicit UnionSchema(std::vector<std::shared_ptr<Schema>> variants)
        : variants_(std::move(variants)) {}

    SchemaKind kind() const override { return SchemaKind::Union; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        // Try each variant - first match wins
        for (const auto& variant : variants_) {
            ValidationContext trial_ctx;
            trial_ctx.definitions = ctx.definitions;
            auto val = variant->validate(input, trial_ctx);
            if (!trial_ctx.has_issues()) {
                return val;
            }
        }

        // None matched
        std::ostringstream expected;
        for (size_t i = 0; i < variants_.size(); i++) {
            if (i > 0) expected << " | ";
            expected << kind_to_string(variants_[i]->kind());
        }
        ctx.add_issue(issue_codes::INVALID_UNION, expected.str(), get_json_type(input));
        return nullptr;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "union";
        nlohmann::json vars = nlohmann::json::array();
        for (const auto& v : variants_) {
            vars.push_back(v->export_node());
        }
        node["variants"] = vars;
        export_common_fields(node);
        return node;
    }

private:
    std::vector<std::shared_ptr<Schema>> variants_;
};

} // namespace anyvali

#pragma once

#include "../schema.hpp"

namespace anyvali {

class NeverSchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::Never; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        ctx.add_issue(issue_codes::INVALID_TYPE, "never", get_json_type(input));
        return nullptr;
    }

    nlohmann::json export_node() const override {
        return {{"kind", "never"}};
    }
};

} // namespace anyvali

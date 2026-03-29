#pragma once

#include "../schema.hpp"

namespace anyvali {

class NullSchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::Null; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!input.is_null()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "null", get_json_type(input));
            return nullptr;
        }
        return input;
    }

    nlohmann::json export_node() const override {
        return {{"kind", "null"}};
    }
};

} // namespace anyvali

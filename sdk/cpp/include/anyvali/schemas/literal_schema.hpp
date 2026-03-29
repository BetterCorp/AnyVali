#pragma once

#include "../schema.hpp"

namespace anyvali {

class LiteralSchema : public Schema {
public:
    explicit LiteralSchema(nlohmann::json value) : expected_value_(std::move(value)) {}

    SchemaKind kind() const override { return SchemaKind::Literal; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (input != expected_value_) {
            ctx.add_issue(issue_codes::INVALID_LITERAL,
                          json_to_string(expected_value_),
                          json_to_string(input));
            return nullptr;
        }
        return input;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "literal";
        node["value"] = expected_value_;
        export_common_fields(node);
        return node;
    }

private:
    nlohmann::json expected_value_;
};

} // namespace anyvali

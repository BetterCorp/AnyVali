#pragma once

#include "../schema.hpp"

namespace anyvali {

class NullableSchema : public Schema {
public:
    explicit NullableSchema(std::shared_ptr<Schema> inner) : inner_(std::move(inner)) {}

    SchemaKind kind() const override { return SchemaKind::Nullable; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (input.is_null()) {
            return input;
        }
        return inner_->validate(input, ctx);
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "nullable";
        node["schema"] = inner_->export_node();
        export_common_fields(node);
        return node;
    }

    std::shared_ptr<Schema> inner_schema() const { return inner_; }

private:
    std::shared_ptr<Schema> inner_;
};

} // namespace anyvali

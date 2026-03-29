#pragma once

#include "../schema.hpp"

namespace anyvali {

class AnySchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::Any; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& /*ctx*/) const override {
        return input;
    }

    nlohmann::json export_node() const override {
        return {{"kind", "any"}};
    }
};

} // namespace anyvali

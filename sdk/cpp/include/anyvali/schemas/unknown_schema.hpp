#pragma once

#include "../schema.hpp"

namespace anyvali {

class UnknownSchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::Unknown; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& /*ctx*/) const override {
        return input;
    }

    nlohmann::json export_node() const override {
        return {{"kind", "unknown"}};
    }
};

} // namespace anyvali

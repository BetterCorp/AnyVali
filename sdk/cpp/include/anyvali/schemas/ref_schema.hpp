#pragma once

#include "../schema.hpp"
#include <stdexcept>

namespace anyvali {

class RefSchema : public Schema {
public:
    explicit RefSchema(std::string ref) : ref_(std::move(ref)) {}

    SchemaKind kind() const override { return SchemaKind::Ref; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!ctx.definitions) {
            throw std::runtime_error("RefSchema: no definitions context available");
        }

        // Parse ref: "#/definitions/Name" -> "Name"
        std::string name = resolve_name();
        auto it = ctx.definitions->find(name);
        if (it == ctx.definitions->end()) {
            throw std::runtime_error("RefSchema: definition '" + name + "' not found");
        }

        return it->second->validate(input, ctx);
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "ref";
        node["ref"] = ref_;
        export_common_fields(node);
        return node;
    }

    const std::string& ref() const { return ref_; }

    std::string resolve_name() const {
        const std::string prefix = "#/definitions/";
        if (ref_.find(prefix) == 0) {
            return ref_.substr(prefix.size());
        }
        return ref_;
    }

private:
    std::string ref_;
};

} // namespace anyvali

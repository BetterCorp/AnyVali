#pragma once

#include "../schema.hpp"
#include <sstream>

namespace anyvali {

class EnumSchema : public Schema {
public:
    explicit EnumSchema(std::vector<nlohmann::json> values) : values_(std::move(values)) {}

    SchemaKind kind() const override { return SchemaKind::Enum; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        for (const auto& v : values_) {
            if (input == v) return input;
        }

        // Build expected string: "enum(val1,val2,val3)"
        std::ostringstream oss;
        oss << "enum(";
        for (size_t i = 0; i < values_.size(); i++) {
            if (i > 0) oss << ",";
            oss << json_to_string(values_[i]);
        }
        oss << ")";

        ctx.add_issue(issue_codes::INVALID_TYPE, oss.str(), json_to_string(input));
        return nullptr;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "enum";
        node["values"] = values_;
        export_common_fields(node);
        return node;
    }

private:
    std::vector<nlohmann::json> values_;
};

} // namespace anyvali

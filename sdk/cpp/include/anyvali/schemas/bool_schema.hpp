#pragma once

#include "../schema.hpp"
#include "../parse/coercion.hpp"

namespace anyvali {

class BoolSchema : public Schema {
public:
    SchemaKind kind() const override { return SchemaKind::Bool; }

    BoolSchema& coerce(const std::string& c) { coercions_ = {c}; return *this; }
    BoolSchema& defaultValue(const nlohmann::json& v) { default_value_ = v; return *this; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        nlohmann::json value = input;

        if (!coercions_.empty() && !value.is_null()) {
            auto coerced = coercion::apply_chain(coercions_, value);
            if (coerced.has_value()) {
                value = coerced.value();
            } else {
                ctx.add_issue(issue_codes::COERCION_FAILED, "bool",
                              json_to_string(input), "Coercion failed");
                return nullptr;
            }
        }

        if (!value.is_boolean()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "bool", get_json_type(value));
            return nullptr;
        }
        return value;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "bool";
        export_common_fields(node);
        return node;
    }
};

} // namespace anyvali

#pragma once

#include "../schema.hpp"

namespace anyvali {

class TupleSchema : public Schema {
public:
    explicit TupleSchema(std::vector<std::shared_ptr<Schema>> elements)
        : elements_(std::move(elements)) {}

    SchemaKind kind() const override { return SchemaKind::Tuple; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!input.is_array()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "tuple", get_json_type(input));
            return nullptr;
        }

        auto expected = static_cast<int64_t>(elements_.size());
        auto actual = static_cast<int64_t>(input.size());

        if (actual < expected) {
            ctx.add_issue(issue_codes::TOO_SMALL, std::to_string(expected),
                          std::to_string(actual));
            return nullptr;
        }
        if (actual > expected) {
            ctx.add_issue(issue_codes::TOO_LARGE, std::to_string(expected),
                          std::to_string(actual));
            return nullptr;
        }

        nlohmann::json result = nlohmann::json::array();
        bool has_errors = false;

        for (int i = 0; i < static_cast<int>(elements_.size()); i++) {
            ctx.push_path(i);
            auto val = elements_[i]->validate(input[i], ctx);
            ctx.pop_path();
            if (val.is_null() && !input[i].is_null()) {
                has_errors = true;
            } else {
                result.push_back(val);
            }
        }

        if (has_errors) return nullptr;
        return result;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = "tuple";
        nlohmann::json elems = nlohmann::json::array();
        for (const auto& e : elements_) {
            elems.push_back(e->export_node());
        }
        node["elements"] = elems;
        export_common_fields(node);
        return node;
    }

private:
    std::vector<std::shared_ptr<Schema>> elements_;
};

} // namespace anyvali

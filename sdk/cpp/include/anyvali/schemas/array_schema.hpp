#pragma once

#include "../schema.hpp"
#include <optional>

namespace anyvali {

class ArraySchema : public Schema {
public:
    explicit ArraySchema(std::shared_ptr<Schema> items) : items_(std::move(items)) {}

    SchemaKind kind() const override { return SchemaKind::Array; }

    ArraySchema& minItems(int64_t n) { min_items_ = n; return *this; }
    ArraySchema& maxItems(int64_t n) { max_items_ = n; return *this; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        if (!input.is_array()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, "array", get_json_type(input));
            return nullptr;
        }

        auto size = static_cast<int64_t>(input.size());

        if (min_items_.has_value() && size < min_items_.value()) {
            ctx.add_issue(issue_codes::TOO_SMALL, std::to_string(min_items_.value()),
                          std::to_string(size));
            return nullptr;
        }
        if (max_items_.has_value() && size > max_items_.value()) {
            ctx.add_issue(issue_codes::TOO_LARGE, std::to_string(max_items_.value()),
                          std::to_string(size));
            return nullptr;
        }

        nlohmann::json result = nlohmann::json::array();
        bool has_errors = false;

        for (int i = 0; i < static_cast<int>(input.size()); i++) {
            ctx.push_path(i);
            auto val = items_->validate(input[i], ctx);
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
        node["kind"] = "array";
        node["items"] = items_->export_node();
        if (min_items_.has_value()) node["minItems"] = min_items_.value();
        if (max_items_.has_value()) node["maxItems"] = max_items_.value();
        export_common_fields(node);
        return node;
    }

    std::shared_ptr<Schema> items_schema() const { return items_; }

private:
    std::shared_ptr<Schema> items_;
    std::optional<int64_t> min_items_;
    std::optional<int64_t> max_items_;
};

} // namespace anyvali

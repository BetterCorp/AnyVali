#pragma once

#include "../schema.hpp"
#include "../parse/coercion.hpp"
#include <optional>
#include <cmath>
#include <cstdint>
#include <limits>
#include <sstream>

namespace anyvali {

class NumberSchema : public Schema {
public:
    explicit NumberSchema(SchemaKind k = SchemaKind::Number) : kind_(k) {}

    SchemaKind kind() const override { return kind_; }

    NumberSchema& min(double v) { min_ = v; return *this; }
    NumberSchema& max(double v) { max_ = v; return *this; }
    NumberSchema& exclusiveMin(double v) { exclusive_min_ = v; return *this; }
    NumberSchema& exclusiveMax(double v) { exclusive_max_ = v; return *this; }
    NumberSchema& multipleOf(double v) { multiple_of_ = v; return *this; }
    NumberSchema& coerce(const std::string& c) { coercions_ = {c}; return *this; }
    NumberSchema& defaultValue(const nlohmann::json& v) { default_value_ = v; return *this; }

    nlohmann::json validate(const nlohmann::json& input, ValidationContext& ctx) const override {
        nlohmann::json value = input;

        // Apply coercions
        if (!coercions_.empty() && !value.is_null()) {
            auto coerced = coercion::apply_chain(coercions_, value);
            if (coerced.has_value()) {
                value = coerced.value();
            } else {
                ctx.add_issue(issue_codes::COERCION_FAILED, kind_to_string(kind_),
                              json_to_string(input), "Coercion failed");
                return nullptr;
            }
        }

        // Type check - must be a number
        if (!value.is_number()) {
            ctx.add_issue(issue_codes::INVALID_TYPE, kind_to_string(kind_), get_json_type(value));
            return nullptr;
        }

        double dval = value.get<double>();

        // For float types, accept any number
        // For int types, check it's a whole number and within range
        if (is_int_kind()) {
            // Check it's actually an integer
            if (value.is_number_float() && std::floor(dval) != dval) {
                ctx.add_issue(issue_codes::INVALID_TYPE, kind_to_string(kind_), "number");
                return nullptr;
            }

            // Check width bounds
            int64_t ival = 0;
            if (value.is_number_unsigned()) {
                uint64_t uval = value.get<uint64_t>();
                if (is_unsigned_kind()) {
                    if (!check_unsigned_range(uval, ctx)) return nullptr;
                } else {
                    // Signed kind - check it fits
                    auto [smin, smax] = signed_range();
                    if (uval > static_cast<uint64_t>(smax)) {
                        ctx.add_issue(issue_codes::TOO_LARGE, kind_to_string(kind_),
                                      std::to_string(uval));
                        return nullptr;
                    }
                    ival = static_cast<int64_t>(uval);
                }
            } else if (value.is_number_integer()) {
                ival = value.get<int64_t>();
                if (is_unsigned_kind()) {
                    if (ival < 0) {
                        ctx.add_issue(issue_codes::TOO_SMALL, kind_to_string(kind_),
                                      std::to_string(ival));
                        return nullptr;
                    }
                    uint64_t uval = static_cast<uint64_t>(ival);
                    if (!check_unsigned_range(uval, ctx)) return nullptr;
                } else {
                    auto [smin, smax] = signed_range();
                    if (ival < smin) {
                        ctx.add_issue(issue_codes::TOO_SMALL, kind_to_string(kind_),
                                      std::to_string(ival));
                        return nullptr;
                    }
                    if (ival > smax) {
                        ctx.add_issue(issue_codes::TOO_LARGE, kind_to_string(kind_),
                                      std::to_string(ival));
                        return nullptr;
                    }
                }
            } else {
                // float value that happens to be whole
                if (is_unsigned_kind()) {
                    if (dval < 0) {
                        ctx.add_issue(issue_codes::TOO_SMALL, kind_to_string(kind_),
                                      format_number(dval));
                        return nullptr;
                    }
                    uint64_t uval = static_cast<uint64_t>(dval);
                    if (!check_unsigned_range(uval, ctx)) return nullptr;
                } else {
                    auto [smin, smax] = signed_range();
                    int64_t iv = static_cast<int64_t>(dval);
                    if (iv < smin) {
                        ctx.add_issue(issue_codes::TOO_SMALL, kind_to_string(kind_),
                                      std::to_string(iv));
                        return nullptr;
                    }
                    if (iv > smax) {
                        ctx.add_issue(issue_codes::TOO_LARGE, kind_to_string(kind_),
                                      std::to_string(iv));
                        return nullptr;
                    }
                }
            }
        }

        // Numeric constraints
        if (min_.has_value() && dval < min_.value()) {
            ctx.add_issue(issue_codes::TOO_SMALL, format_number(min_.value()),
                          format_number(dval));
            return nullptr;
        }
        if (max_.has_value() && dval > max_.value()) {
            ctx.add_issue(issue_codes::TOO_LARGE, format_number(max_.value()),
                          format_number(dval));
            return nullptr;
        }
        if (exclusive_min_.has_value() && dval <= exclusive_min_.value()) {
            ctx.add_issue(issue_codes::TOO_SMALL, format_number(exclusive_min_.value()),
                          format_number(dval));
            return nullptr;
        }
        if (exclusive_max_.has_value() && dval >= exclusive_max_.value()) {
            ctx.add_issue(issue_codes::TOO_LARGE, format_number(exclusive_max_.value()),
                          format_number(dval));
            return nullptr;
        }
        if (multiple_of_.has_value()) {
            double remainder = std::fmod(dval, multiple_of_.value());
            if (std::abs(remainder) > 1e-10 && std::abs(remainder - multiple_of_.value()) > 1e-10) {
                ctx.add_issue(issue_codes::INVALID_NUMBER, format_number(multiple_of_.value()),
                              format_number(dval));
                return nullptr;
            }
        }

        return value;
    }

    nlohmann::json export_node() const override {
        nlohmann::json node;
        node["kind"] = kind_to_string(kind_);
        if (min_.has_value()) node["min"] = min_.value();
        if (max_.has_value()) node["max"] = max_.value();
        if (exclusive_min_.has_value()) node["exclusiveMin"] = exclusive_min_.value();
        if (exclusive_max_.has_value()) node["exclusiveMax"] = exclusive_max_.value();
        if (multiple_of_.has_value()) node["multipleOf"] = multiple_of_.value();
        export_common_fields(node);
        return node;
    }

private:
    SchemaKind kind_;
    std::optional<double> min_;
    std::optional<double> max_;
    std::optional<double> exclusive_min_;
    std::optional<double> exclusive_max_;
    std::optional<double> multiple_of_;

    bool is_int_kind() const {
        return kind_ == SchemaKind::Int || kind_ == SchemaKind::Int8 ||
               kind_ == SchemaKind::Int16 || kind_ == SchemaKind::Int32 ||
               kind_ == SchemaKind::Int64 || kind_ == SchemaKind::Uint8 ||
               kind_ == SchemaKind::Uint16 || kind_ == SchemaKind::Uint32 ||
               kind_ == SchemaKind::Uint64;
    }

    bool is_unsigned_kind() const {
        return kind_ == SchemaKind::Uint8 || kind_ == SchemaKind::Uint16 ||
               kind_ == SchemaKind::Uint32 || kind_ == SchemaKind::Uint64;
    }

    std::pair<int64_t, int64_t> signed_range() const {
        switch (kind_) {
            case SchemaKind::Int8: return {-128, 127};
            case SchemaKind::Int16: return {-32768, 32767};
            case SchemaKind::Int32: return {INT32_MIN, INT32_MAX};
            case SchemaKind::Int: case SchemaKind::Int64:
                return {INT64_MIN, INT64_MAX};
            default: return {INT64_MIN, INT64_MAX};
        }
    }

    std::pair<uint64_t, uint64_t> unsigned_range() const {
        switch (kind_) {
            case SchemaKind::Uint8: return {0, 255};
            case SchemaKind::Uint16: return {0, 65535};
            case SchemaKind::Uint32: return {0, UINT32_MAX};
            case SchemaKind::Uint64: return {0, UINT64_MAX};
            default: return {0, UINT64_MAX};
        }
    }

    bool check_unsigned_range(uint64_t val, ValidationContext& ctx) const {
        auto [umin, umax] = unsigned_range();
        if (val > umax) {
            ctx.add_issue(issue_codes::TOO_LARGE, kind_to_string(kind_),
                          std::to_string(val));
            return false;
        }
        return true;
    }

    static std::string format_number(double v) {
        if (v == std::floor(v) && std::abs(v) < 1e15) {
            int64_t iv = static_cast<int64_t>(v);
            return std::to_string(iv);
        }
        std::ostringstream oss;
        oss << v;
        return oss.str();
    }
};

} // namespace anyvali

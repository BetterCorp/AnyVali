#pragma once

#include "schema.hpp"
#include "parse_result.hpp"
#include <nlohmann/json.hpp>

namespace anyvali {

template<typename T>
struct TypedParseResult {
    bool success;
    T value;
    std::vector<ValidationIssue> issues;
};

/// Parse input using a schema and convert the result to a specific C++ type.
/// Throws ValidationError on validation failure, or nlohmann::json conversion errors
/// if the validated JSON value cannot be converted to T.
template<typename T>
T parse_as(const Schema& schema, const nlohmann::json& input) {
    auto result = schema.parse(input);
    return result.get<T>();
}

/// Safe parse with typed result.
/// Returns a TypedParseResult<T> containing the converted value on success,
/// or a default-constructed T with validation issues on failure.
template<typename T>
TypedParseResult<T> safe_parse_as(const Schema& schema, const nlohmann::json& input) {
    auto result = schema.safe_parse(input);
    if (result.success) {
        return TypedParseResult<T>{true, result.value.get<T>(), {}};
    }
    return TypedParseResult<T>{false, T{}, std::move(result.issues)};
}

} // namespace anyvali

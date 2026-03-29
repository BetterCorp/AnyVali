#pragma once

#include <string>

namespace anyvali {
namespace issue_codes {

inline const std::string INVALID_TYPE = "invalid_type";
inline const std::string REQUIRED = "required";
inline const std::string UNKNOWN_KEY = "unknown_key";
inline const std::string TOO_SMALL = "too_small";
inline const std::string TOO_LARGE = "too_large";
inline const std::string INVALID_STRING = "invalid_string";
inline const std::string INVALID_NUMBER = "invalid_number";
inline const std::string INVALID_LITERAL = "invalid_literal";
inline const std::string INVALID_UNION = "invalid_union";
inline const std::string CUSTOM_VALIDATION_NOT_PORTABLE = "custom_validation_not_portable";
inline const std::string UNSUPPORTED_EXTENSION = "unsupported_extension";
inline const std::string UNSUPPORTED_SCHEMA_KIND = "unsupported_schema_kind";
inline const std::string COERCION_FAILED = "coercion_failed";
inline const std::string DEFAULT_INVALID = "default_invalid";

} // namespace issue_codes
} // namespace anyvali

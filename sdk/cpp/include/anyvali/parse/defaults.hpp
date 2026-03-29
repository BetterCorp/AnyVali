#pragma once

#include <nlohmann/json.hpp>
#include <optional>

namespace anyvali {
namespace defaults {

// Check if a value is "absent" (used for default application).
// Absent means the key was not present in the object at all.
// null is NOT treated as absent.
inline bool is_absent(const nlohmann::json& /*value*/, bool key_present) {
    return !key_present;
}

} // namespace defaults
} // namespace anyvali

#pragma once

#include <string>
#include <vector>

namespace anyvali {

struct CoercionConfig {
    std::vector<std::string> coercions;

    bool has_coercion(const std::string& name) const {
        for (const auto& c : coercions) {
            if (c == name) return true;
        }
        return false;
    }
};

} // namespace anyvali

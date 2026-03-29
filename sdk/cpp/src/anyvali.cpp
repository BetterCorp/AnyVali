// Non-inline implementations for AnyVali C++ SDK.
// Currently the SDK is fully header-only; this file exists as a placeholder
// for any future non-inline implementations and to satisfy the CMake static
// library target.

#include <anyvali/anyvali.hpp>

namespace anyvali {

// Library version
const char* version() {
    return "1.0.0";
}

} // namespace anyvali

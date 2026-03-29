// Simple test framework - no external dependencies required
#include <iostream>
#include <string>
#include <vector>
#include <functional>
#include <sstream>

struct TestCase {
    std::string name;
    std::function<void()> func;
};

static std::vector<TestCase>& test_registry() {
    static std::vector<TestCase> reg;
    return reg;
}

static int& fail_count() {
    static int c = 0;
    return c;
}

static int& pass_count() {
    static int c = 0;
    return c;
}

static std::string& current_test() {
    static std::string s;
    return s;
}

struct TestRegistrar {
    TestRegistrar(const char* name, std::function<void()> func) {
        test_registry().push_back({name, std::move(func)});
    }
};

#define TEST_CASE(name) \
    static void test_##__LINE__(); \
    static TestRegistrar registrar_##__LINE__(name, test_##__LINE__); \
    static void test_##__LINE__()

// Unique name macro with concatenation
#define CONCAT_IMPL(a, b) a##b
#define CONCAT(a, b) CONCAT_IMPL(a, b)
#define TEST(name) \
    static void CONCAT(test_func_, __LINE__)(); \
    static TestRegistrar CONCAT(registrar_, __LINE__)(name, CONCAT(test_func_, __LINE__)); \
    static void CONCAT(test_func_, __LINE__)()

inline void assert_true(bool condition, const char* expr, const char* file, int line) {
    if (!condition) {
        std::cerr << "  FAIL: " << file << ":" << line << ": " << expr << std::endl;
        fail_count()++;
        throw std::runtime_error("Assertion failed");
    }
}

inline void assert_eq_json(const nlohmann::json& actual, const nlohmann::json& expected,
                            const char* file, int line) {
    if (actual != expected) {
        std::cerr << "  FAIL: " << file << ":" << line << std::endl;
        std::cerr << "    expected: " << expected.dump() << std::endl;
        std::cerr << "    actual:   " << actual.dump() << std::endl;
        fail_count()++;
        throw std::runtime_error("JSON assertion failed");
    }
}

#define ASSERT(cond) assert_true(cond, #cond, __FILE__, __LINE__)
#define ASSERT_EQ(a, b) do { \
    auto _a = (a); auto _b = (b); \
    if (_a != _b) { \
        std::cerr << "  FAIL: " << __FILE__ << ":" << __LINE__ << ": " \
                  << #a << " == " << #b << std::endl; \
        std::cerr << "    lhs: " << _a << std::endl; \
        std::cerr << "    rhs: " << _b << std::endl; \
        fail_count()++; \
        throw std::runtime_error("Assertion failed"); \
    } \
} while(0)
#define ASSERT_JSON_EQ(a, b) assert_eq_json(a, b, __FILE__, __LINE__)
#define ASSERT_THROWS(expr) do { \
    bool threw = false; \
    try { expr; } catch (...) { threw = true; } \
    if (!threw) { \
        std::cerr << "  FAIL: " << __FILE__ << ":" << __LINE__ << ": expected exception from " << #expr << std::endl; \
        fail_count()++; \
        throw std::runtime_error("Expected exception"); \
    } \
} while(0)

// Include all test files - they register tests via TEST() macro
// (linked as separate translation units)

int main(int argc, char** argv) {
    std::string filter = "";
    if (argc > 1) filter = argv[1];

    int total = 0;
    int passed = 0;
    int failed = 0;
    int skipped = 0;

    for (auto& tc : test_registry()) {
        if (!filter.empty() && tc.name.find(filter) == std::string::npos) {
            skipped++;
            continue;
        }
        total++;
        current_test() = tc.name;
        int prev_fails = fail_count();
        try {
            tc.func();
            if (fail_count() == prev_fails) {
                passed++;
                std::cout << "  PASS: " << tc.name << std::endl;
            } else {
                failed++;
                std::cerr << "  FAIL: " << tc.name << std::endl;
            }
        } catch (const std::exception& e) {
            failed++;
            std::cerr << "  FAIL: " << tc.name << " (" << e.what() << ")" << std::endl;
        }
    }

    std::cout << "\n========================================" << std::endl;
    std::cout << "Results: " << passed << " passed, " << failed << " failed, "
              << skipped << " skipped, " << total << " total" << std::endl;

    return failed > 0 ? 1 : 0;
}

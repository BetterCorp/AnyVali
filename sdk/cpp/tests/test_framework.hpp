#pragma once

#include <iostream>
#include <string>
#include <vector>
#include <functional>
#include <sstream>
#include <stdexcept>
#include <nlohmann/json.hpp>

struct TestCase {
    std::string name;
    std::function<void()> func;
};

inline std::vector<TestCase>& test_registry() {
    static std::vector<TestCase> reg;
    return reg;
}

inline int& fail_count() {
    static int c = 0;
    return c;
}

inline int& pass_count() {
    static int c = 0;
    return c;
}

inline std::string& current_test() {
    static std::string s;
    return s;
}

struct TestRegistrar {
    TestRegistrar(const char* name, std::function<void()> func) {
        test_registry().push_back({name, std::move(func)});
    }
};

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

#define CONCAT_IMPL(a, b) a##b
#define CONCAT(a, b) CONCAT_IMPL(a, b)
#define TEST(name) \
    static void CONCAT(test_func_, __LINE__)(); \
    static TestRegistrar CONCAT(registrar_, __LINE__)(name, CONCAT(test_func_, __LINE__)); \
    static void CONCAT(test_func_, __LINE__)()

#define ASSERT(cond) assert_true(cond, #cond, __FILE__, __LINE__)
#define ASSERT_EQ(a, b) do { \
    auto _a = (a); auto _b = (b); \
    if (_a != _b) { \
        std::cerr << "  FAIL: " << __FILE__ << ":" << __LINE__ << ": " \
                  << #a << " == " << #b << std::endl; \
        fail_count()++; \
        throw std::runtime_error("Assertion failed"); \
    } \
} while(0)
#define ASSERT_JSON_EQ(a, b) assert_eq_json(a, b, __FILE__, __LINE__)
#define ASSERT_THROWS(expr) do { \
    bool threw = false; \
    try { expr; } catch (...) { threw = true; } \
    assert_true(threw, "expected exception from " #expr, __FILE__, __LINE__); \
} while(0)

#include <anyvali/anyvali.hpp>
#include <iostream>
#include <stdexcept>

using namespace anyvali;
using json = nlohmann::json;

struct TestRegistrar;
extern void assert_true(bool condition, const char* expr, const char* file, int line);
extern void assert_eq_json(const json& actual, const json& expected, const char* file, int line);

#define CONCAT_IMPL(a, b) a##b
#define CONCAT(a, b) CONCAT_IMPL(a, b)
#define TEST(name) \
    static void CONCAT(test_func_, __LINE__)(); \
    static TestRegistrar CONCAT(registrar_, __LINE__)(name, CONCAT(test_func_, __LINE__)); \
    static void CONCAT(test_func_, __LINE__)()

#define ASSERT(cond) assert_true(cond, #cond, __FILE__, __LINE__)
#define ASSERT_JSON_EQ(a, b) assert_eq_json(a, b, __FILE__, __LINE__)

TEST("coercion: string->int valid") {
    auto s = int_();
    s->coerce("string->int");
    auto result = s->safe_parse(json("42"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(42));
}

TEST("coercion: string->int trims whitespace") {
    auto s = int_();
    s->coerce("string->int");
    auto result = s->safe_parse(json("  42  "));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(42));
}

TEST("coercion: string->number valid") {
    auto s = number();
    s->coerce("string->number");
    auto result = s->safe_parse(json("3.14"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(3.14));
}

TEST("coercion: string->bool true") {
    auto s = bool_();
    s->coerce("string->bool");
    ASSERT(s->safe_parse(json("true")).success);
    ASSERT_JSON_EQ(s->safe_parse(json("true")).value, json(true));
}

TEST("coercion: string->bool false") {
    auto s = bool_();
    s->coerce("string->bool");
    ASSERT_JSON_EQ(s->safe_parse(json("false")).value, json(false));
}

TEST("coercion: string->bool 1") {
    auto s = bool_();
    s->coerce("string->bool");
    ASSERT_JSON_EQ(s->safe_parse(json("1")).value, json(true));
}

TEST("coercion: string->bool 0") {
    auto s = bool_();
    s->coerce("string->bool");
    ASSERT_JSON_EQ(s->safe_parse(json("0")).value, json(false));
}

TEST("coercion: string->bool case insensitive") {
    auto s = bool_();
    s->coerce("string->bool");
    ASSERT_JSON_EQ(s->safe_parse(json("TRUE")).value, json(true));
}

TEST("coercion: trim removes whitespace") {
    auto s = string_();
    s->coerce("trim");
    auto result = s->safe_parse(json("  hello  "));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello"));
}

TEST("coercion: lower converts to lowercase") {
    auto s = string_();
    s->coerce("lower");
    auto result = s->safe_parse(json("HELLO World"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello world"));
}

TEST("coercion: upper converts to uppercase") {
    auto s = string_();
    s->coerce("upper");
    auto result = s->safe_parse(json("hello world"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("HELLO WORLD"));
}

TEST("coercion: failure produces coercion_failed") {
    auto s = int_();
    s->coerce("string->int");
    auto result = s->safe_parse(json("not-a-number"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "coercion_failed");
}

TEST("coercion: happens before validation") {
    auto s = int_();
    s->coerce("string->int");
    s->min(10);
    auto result = s->safe_parse(json("5"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_small");
}

TEST("coercion: then validation success") {
    auto s = int_();
    s->coerce("string->int");
    s->min(1);
    s->max(100);
    auto result = s->safe_parse(json("50"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(50));
}

TEST("coercion: chained coercions left to right") {
    auto s = string_();
    s->coerce({"trim", "lower"});
    auto result = s->safe_parse(json("  HELLO  "));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello"));
}

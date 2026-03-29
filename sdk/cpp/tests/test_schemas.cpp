#include <anyvali/anyvali.hpp>
#include <iostream>
#include <stdexcept>
#include <nlohmann/json.hpp>

using namespace anyvali;
using json = nlohmann::json;

// Forward-declare test framework macros from main.cpp
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
#define ASSERT_THROWS(expr) do { \
    bool threw = false; \
    try { expr; } catch (...) { threw = true; } \
    assert_true(threw, "expected exception from " #expr, __FILE__, __LINE__); \
} while(0)

// ---- String Schema Tests ----

TEST("string: accepts valid string") {
    auto s = string_();
    auto result = s->safe_parse(json("hello"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello"));
}

TEST("string: rejects number") {
    auto s = string_();
    auto result = s->safe_parse(json(42));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_type");
}

TEST("string: rejects null") {
    auto s = string_();
    auto result = s->safe_parse(json(nullptr));
    ASSERT(!result.success);
}

TEST("string: rejects boolean") {
    auto s = string_();
    auto result = s->safe_parse(json(true));
    ASSERT(!result.success);
}

TEST("string: rejects array") {
    auto s = string_();
    auto result = s->safe_parse(json::array({"a", "b"}));
    ASSERT(!result.success);
}

TEST("string: rejects object") {
    auto s = string_();
    auto result = s->safe_parse(json::object({{"key", "value"}}));
    ASSERT(!result.success);
}

TEST("string: minLength pass") {
    auto s = string_();
    s->minLength(3);
    auto result = s->safe_parse(json("abc"));
    ASSERT(result.success);
}

TEST("string: minLength fail") {
    auto s = string_();
    s->minLength(3);
    auto result = s->safe_parse(json("ab"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_small");
}

TEST("string: maxLength pass") {
    auto s = string_();
    s->maxLength(5);
    auto result = s->safe_parse(json("hello"));
    ASSERT(result.success);
}

TEST("string: maxLength fail") {
    auto s = string_();
    s->maxLength(5);
    auto result = s->safe_parse(json("hello!"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_large");
}

TEST("string: pattern pass") {
    auto s = string_();
    s->pattern("^[a-z]+$");
    auto result = s->safe_parse(json("abc"));
    ASSERT(result.success);
}

TEST("string: pattern fail") {
    auto s = string_();
    s->pattern("^[a-z]+$");
    auto result = s->safe_parse(json("ABC"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_string");
}

TEST("string: startsWith pass") {
    auto s = string_();
    s->startsWith("hello");
    auto result = s->safe_parse(json("hello world"));
    ASSERT(result.success);
}

TEST("string: startsWith fail") {
    auto s = string_();
    s->startsWith("hello");
    auto result = s->safe_parse(json("world hello"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_string");
}

TEST("string: endsWith pass") {
    auto s = string_();
    s->endsWith(".json");
    auto result = s->safe_parse(json("file.json"));
    ASSERT(result.success);
}

TEST("string: endsWith fail") {
    auto s = string_();
    s->endsWith(".json");
    auto result = s->safe_parse(json("file.xml"));
    ASSERT(!result.success);
}

TEST("string: includes pass") {
    auto s = string_();
    s->includes("world");
    auto result = s->safe_parse(json("hello world!"));
    ASSERT(result.success);
}

TEST("string: includes fail") {
    auto s = string_();
    s->includes("world");
    auto result = s->safe_parse(json("hello there"));
    ASSERT(!result.success);
}

TEST("string: empty string accepted") {
    auto s = string_();
    auto result = s->safe_parse(json(""));
    ASSERT(result.success);
}

TEST("string: unicode accepted") {
    auto s = string_();
    auto result = s->safe_parse(json("\xc3\xa9\xc3\xa0\xc3\xbc"));
    ASSERT(result.success);
}

// ---- Number Schema Tests ----

TEST("number: accepts positive integer") {
    auto s = number();
    auto result = s->safe_parse(json(42));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(42));
}

TEST("number: accepts zero") {
    auto s = number();
    auto result = s->safe_parse(json(0));
    ASSERT(result.success);
}

TEST("number: accepts negative float") {
    auto s = number();
    auto result = s->safe_parse(json(-3.14));
    ASSERT(result.success);
}

TEST("number: rejects string") {
    auto s = number();
    auto result = s->safe_parse(json("42"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_type");
}

TEST("number: rejects boolean") {
    auto s = number();
    auto result = s->safe_parse(json(true));
    ASSERT(!result.success);
}

TEST("number: rejects null") {
    auto s = number();
    auto result = s->safe_parse(json(nullptr));
    ASSERT(!result.success);
}

TEST("number: rejects object") {
    auto s = number();
    auto result = s->safe_parse(json::object());
    ASSERT(!result.success);
}

TEST("number: rejects array") {
    auto s = number();
    auto result = s->safe_parse(json::array({1, 2, 3}));
    ASSERT(!result.success);
}

TEST("number: min pass") {
    auto s = number();
    s->min(10);
    auto result = s->safe_parse(json(10));
    ASSERT(result.success);
}

TEST("number: min fail") {
    auto s = number();
    s->min(10);
    auto result = s->safe_parse(json(9));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_small");
}

TEST("number: max pass") {
    auto s = number();
    s->max(100);
    auto result = s->safe_parse(json(100));
    ASSERT(result.success);
}

TEST("number: max fail") {
    auto s = number();
    s->max(100);
    auto result = s->safe_parse(json(101));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_large");
}

TEST("number: exclusiveMin pass") {
    auto s = number();
    s->exclusiveMin(0);
    auto result = s->safe_parse(json(0.001));
    ASSERT(result.success);
}

TEST("number: exclusiveMin fail at boundary") {
    auto s = number();
    s->exclusiveMin(0);
    auto result = s->safe_parse(json(0));
    ASSERT(!result.success);
}

TEST("number: exclusiveMax pass") {
    auto s = number();
    s->exclusiveMax(100);
    auto result = s->safe_parse(json(99.999));
    ASSERT(result.success);
}

TEST("number: exclusiveMax fail at boundary") {
    auto s = number();
    s->exclusiveMax(100);
    auto result = s->safe_parse(json(100));
    ASSERT(!result.success);
}

TEST("number: multipleOf pass") {
    auto s = number();
    s->multipleOf(3);
    auto result = s->safe_parse(json(9));
    ASSERT(result.success);
}

TEST("number: multipleOf fail") {
    auto s = number();
    s->multipleOf(3);
    auto result = s->safe_parse(json(10));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_number");
}

TEST("number: multipleOf float") {
    auto s = number();
    s->multipleOf(0.5);
    auto result = s->safe_parse(json(2.5));
    ASSERT(result.success);
}

// ---- Int Schema Tests ----

TEST("int: accepts positive integer") {
    auto s = int_();
    auto result = s->safe_parse(json(42));
    ASSERT(result.success);
}

TEST("int: accepts zero") {
    auto s = int_();
    auto result = s->safe_parse(json(0));
    ASSERT(result.success);
}

TEST("int: accepts negative") {
    auto s = int_();
    auto result = s->safe_parse(json(-100));
    ASSERT(result.success);
}

TEST("int: rejects float") {
    auto s = int_();
    auto result = s->safe_parse(json(3.14));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_type");
}

TEST("int: rejects string") {
    auto s = int_();
    auto result = s->safe_parse(json("42"));
    ASSERT(!result.success);
}

TEST("int: min and max constraints") {
    auto s = int_();
    s->min(1);
    s->max(10);
    ASSERT(s->safe_parse(json(5)).success);
    ASSERT(!s->safe_parse(json(0)).success);
    ASSERT(!s->safe_parse(json(11)).success);
}

// ---- Int Width Tests ----

TEST("int8: accepts in range") {
    auto s = int8();
    ASSERT(s->safe_parse(json(127)).success);
    ASSERT(s->safe_parse(json(-128)).success);
}

TEST("int8: rejects out of range") {
    auto s = int8();
    ASSERT(!s->safe_parse(json(128)).success);
    ASSERT(!s->safe_parse(json(-129)).success);
}

TEST("int16: accepts max") {
    auto s = int16();
    ASSERT(s->safe_parse(json(32767)).success);
}

TEST("int16: rejects above range") {
    auto s = int16();
    ASSERT(!s->safe_parse(json(32768)).success);
}

TEST("int32: accepts max") {
    auto s = int32();
    ASSERT(s->safe_parse(json(2147483647)).success);
}

TEST("int32: rejects above range") {
    auto s = int32();
    ASSERT(!s->safe_parse(json(2147483648LL)).success);
}

TEST("uint8: accepts 0 and 255") {
    auto s = uint8();
    ASSERT(s->safe_parse(json(0)).success);
    ASSERT(s->safe_parse(json(255)).success);
}

TEST("uint8: rejects negative") {
    auto s = uint8();
    ASSERT(!s->safe_parse(json(-1)).success);
}

TEST("uint8: rejects 256") {
    auto s = uint8();
    ASSERT(!s->safe_parse(json(256)).success);
}

TEST("uint16: accepts 65535") {
    auto s = uint16();
    ASSERT(s->safe_parse(json(65535)).success);
}

TEST("uint32: accepts max") {
    auto s = uint32();
    ASSERT(s->safe_parse(json(4294967295ULL)).success);
}

TEST("uint64: rejects negative") {
    auto s = uint64();
    ASSERT(!s->safe_parse(json(-1)).success);
}

// ---- Float Width Tests ----

TEST("float64: accepts normal float") {
    auto s = float64();
    ASSERT(s->safe_parse(json(3.141592653589793)).success);
}

TEST("float64: accepts integer") {
    auto s = float64();
    ASSERT(s->safe_parse(json(42)).success);
}

TEST("float64: rejects string") {
    auto s = float64();
    ASSERT(!s->safe_parse(json("3.14")).success);
}

TEST("float32: accepts valid float") {
    auto s = float32();
    ASSERT(s->safe_parse(json(1.5)).success);
}

TEST("float32: rejects boolean") {
    auto s = float32();
    ASSERT(!s->safe_parse(json(true)).success);
}

// ---- Bool Schema Tests ----

TEST("bool: accepts true") {
    auto s = bool_();
    ASSERT(s->safe_parse(json(true)).success);
}

TEST("bool: accepts false") {
    auto s = bool_();
    ASSERT(s->safe_parse(json(false)).success);
}

TEST("bool: rejects number") {
    auto s = bool_();
    auto result = s->safe_parse(json(1));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_type");
}

TEST("bool: rejects string") {
    auto s = bool_();
    ASSERT(!s->safe_parse(json("true")).success);
}

TEST("bool: rejects null") {
    auto s = bool_();
    ASSERT(!s->safe_parse(json(nullptr)).success);
}

// ---- Null Schema Tests ----

TEST("null: accepts null") {
    auto s = null();
    ASSERT(s->safe_parse(json(nullptr)).success);
}

TEST("null: rejects string") {
    auto s = null();
    ASSERT(!s->safe_parse(json("null")).success);
}

TEST("null: rejects zero") {
    auto s = null();
    ASSERT(!s->safe_parse(json(0)).success);
}

TEST("null: rejects false") {
    auto s = null();
    ASSERT(!s->safe_parse(json(false)).success);
}

// ---- Any Schema Tests ----

TEST("any: accepts string") {
    auto s = any();
    ASSERT(s->safe_parse(json("hello")).success);
}

TEST("any: accepts number") {
    auto s = any();
    ASSERT(s->safe_parse(json(42)).success);
}

TEST("any: accepts null") {
    auto s = any();
    ASSERT(s->safe_parse(json(nullptr)).success);
}

TEST("any: accepts object") {
    auto s = any();
    ASSERT(s->safe_parse(json::object({{"key", "value"}})).success);
}

TEST("any: accepts array") {
    auto s = any();
    ASSERT(s->safe_parse(json::array({1, "two", true})).success);
}

// ---- Unknown Schema Tests ----

TEST("unknown: accepts anything") {
    auto s = unknown();
    ASSERT(s->safe_parse(json("hello")).success);
    ASSERT(s->safe_parse(json(99)).success);
    ASSERT(s->safe_parse(json(nullptr)).success);
    ASSERT(s->safe_parse(json(false)).success);
}

// ---- Never Schema Tests ----

TEST("never: rejects everything") {
    auto s = never();
    ASSERT(!s->safe_parse(json("hello")).success);
    ASSERT(!s->safe_parse(json(0)).success);
    ASSERT(!s->safe_parse(json(nullptr)).success);
    ASSERT(!s->safe_parse(json(true)).success);
    ASSERT(!s->safe_parse(json::object()).success);
}

// ---- Literal Schema Tests ----

TEST("literal: accepts matching string") {
    auto s = literal("hello");
    ASSERT(s->safe_parse(json("hello")).success);
}

TEST("literal: rejects non-matching string") {
    auto s = literal("hello");
    auto result = s->safe_parse(json("world"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_literal");
}

TEST("literal: accepts matching number") {
    auto s = literal(42);
    ASSERT(s->safe_parse(json(42)).success);
}

TEST("literal: accepts matching boolean") {
    auto s = literal(true);
    ASSERT(s->safe_parse(json(true)).success);
}

TEST("literal: rejects wrong type") {
    auto s = literal(42);
    ASSERT(!s->safe_parse(json("42")).success);
}

TEST("literal: accepts null literal") {
    auto s = literal(nullptr);
    ASSERT(s->safe_parse(json(nullptr)).success);
}

// ---- Enum Schema Tests ----

TEST("enum: accepts value in set") {
    auto s = enum_({json("red"), json("green"), json("blue")});
    ASSERT(s->safe_parse(json("red")).success);
    ASSERT(s->safe_parse(json("blue")).success);
}

TEST("enum: rejects value not in set") {
    auto s = enum_({json("red"), json("green"), json("blue")});
    auto result = s->safe_parse(json("yellow"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_type");
}

TEST("enum: accepts numeric enum") {
    auto s = enum_({json(1), json(2), json(3)});
    ASSERT(s->safe_parse(json(2)).success);
}

TEST("enum: rejects wrong type") {
    auto s = enum_({json(1), json(2), json(3)});
    ASSERT(!s->safe_parse(json("1")).success);
}

// ---- Array Schema Tests ----

TEST("array: accepts valid array") {
    auto s = array(string_());
    auto result = s->safe_parse(json::array({"a", "b", "c"}));
    ASSERT(result.success);
}

TEST("array: accepts empty array") {
    auto s = array(int_());
    ASSERT(s->safe_parse(json::array()).success);
}

TEST("array: rejects non-array") {
    auto s = array(string_());
    ASSERT(!s->safe_parse(json("not an array")).success);
}

TEST("array: rejects invalid element") {
    auto s = array(int_());
    auto result = s->safe_parse(json::array({1, 2, "three"}));
    ASSERT(!result.success);
    ASSERT(result.issues[0].path.size() == 1);
}

TEST("array: reports multiple errors") {
    auto s = array(bool_());
    auto result = s->safe_parse(json::array({true, "yes", false, 1}));
    ASSERT(!result.success);
    ASSERT(result.issues.size() == 2);
}

TEST("array: minItems pass") {
    auto s = array(int_());
    s->minItems(2);
    ASSERT(s->safe_parse(json::array({1, 2})).success);
}

TEST("array: minItems fail") {
    auto s = array(int_());
    s->minItems(2);
    auto result = s->safe_parse(json::array({1}));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_small");
}

TEST("array: maxItems pass") {
    auto s = array(string_());
    s->maxItems(3);
    ASSERT(s->safe_parse(json::array({"a", "b", "c"})).success);
}

TEST("array: maxItems fail") {
    auto s = array(string_());
    s->maxItems(3);
    ASSERT(!s->safe_parse(json::array({"a", "b", "c", "d"})).success);
}

// ---- Tuple Schema Tests ----

TEST("tuple: accepts valid tuple") {
    auto s = tuple({string_(), int_()});
    auto result = s->safe_parse(json::array({"hello", 42}));
    ASSERT(result.success);
}

TEST("tuple: rejects too few") {
    auto s = tuple({string_(), int_()});
    ASSERT(!s->safe_parse(json::array({"hello"})).success);
}

TEST("tuple: rejects too many") {
    auto s = tuple({string_(), int_()});
    ASSERT(!s->safe_parse(json::array({"hello", 42, true})).success);
}

TEST("tuple: rejects wrong type") {
    auto s = tuple({string_(), int_()});
    auto result = s->safe_parse(json::array({42, "hello"}));
    ASSERT(!result.success);
}

TEST("tuple: rejects non-array") {
    auto s = tuple({string_()});
    ASSERT(!s->safe_parse(json("not a tuple")).success);
}

// ---- Object Schema Tests ----

TEST("object: accepts valid object") {
    auto s = object();
    s->prop("name", string_());
    s->prop("age", int_());
    s->required({"name", "age"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}, {"age", 30}}));
    ASSERT(result.success);
}

TEST("object: rejects missing required") {
    auto s = object();
    s->prop("name", string_());
    s->prop("age", int_());
    s->required({"name", "age"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}}));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "required");
}

TEST("object: accepts optional absent") {
    auto s = object();
    s->prop("name", string_());
    s->prop("nickname", string_());
    s->required({"name"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}}));
    ASSERT(result.success);
}

TEST("object: rejects non-object") {
    auto s = object();
    s->prop("name", string_());
    s->required({"name"});
    ASSERT(!s->safe_parse(json("not an object")).success);
}

TEST("object: rejects unknown keys by default") {
    auto s = object();
    s->prop("name", string_());
    s->required({"name"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}, {"extra", "value"}}));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "unknown_key");
}

TEST("object: strip mode removes unknown keys") {
    auto s = object();
    s->prop("name", string_());
    s->required({"name"});
    s->unknownKeys(UnknownKeyMode::Strip);
    auto result = s->safe_parse(json::object({{"name", "Alice"}, {"extra", "value"}}));
    ASSERT(result.success);
    ASSERT(!result.value.contains("extra"));
}

TEST("object: allow mode passes unknown keys") {
    auto s = object();
    s->prop("name", string_());
    s->required({"name"});
    s->unknownKeys(UnknownKeyMode::Allow);
    auto result = s->safe_parse(json::object({{"name", "Alice"}, {"extra", "value"}}));
    ASSERT(result.success);
    ASSERT(result.value.contains("extra"));
}

TEST("object: reports multiple unknown keys") {
    auto s = object();
    s->prop("id", int_());
    s->required({"id"});
    auto result = s->safe_parse(json::object({{"id", 1}, {"foo", "bar"}, {"baz", true}}));
    ASSERT(!result.success);
    ASSERT(result.issues.size() == 2);
}

// ---- Record Schema Tests ----

TEST("record: accepts valid record") {
    auto s = record(int_());
    auto result = s->safe_parse(json::object({{"a", 1}, {"b", 2}}));
    ASSERT(result.success);
}

TEST("record: accepts empty") {
    auto s = record(string_());
    ASSERT(s->safe_parse(json::object()).success);
}

TEST("record: rejects invalid value") {
    auto s = record(int_());
    auto result = s->safe_parse(json::object({{"a", 1}, {"b", "two"}}));
    ASSERT(!result.success);
}

TEST("record: rejects non-object") {
    auto s = record(string_());
    ASSERT(!s->safe_parse(json::array({1, 2})).success);
}

// ---- Union Schema Tests ----

TEST("union: accepts first variant") {
    auto s = union_({string_(), int_()});
    ASSERT(s->safe_parse(json("hello")).success);
}

TEST("union: accepts second variant") {
    auto s = union_({string_(), int_()});
    ASSERT(s->safe_parse(json(42)).success);
}

TEST("union: rejects no match") {
    auto s = union_({string_(), int_()});
    auto result = s->safe_parse(json(true));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_union");
}

TEST("union: first matching wins") {
    auto s = union_({number(), int_()});
    auto result = s->safe_parse(json(5));
    ASSERT(result.success);
}

TEST("union: null variant") {
    auto s = union_({string_(), null()});
    ASSERT(s->safe_parse(json(nullptr)).success);
}

// ---- Intersection Schema Tests ----

TEST("intersection: accepts value satisfying all") {
    auto s1 = object();
    s1->prop("name", string_());
    s1->required({"name"});
    s1->unknownKeys(UnknownKeyMode::Allow);

    auto s2 = object();
    s2->prop("age", int_());
    s2->required({"age"});
    s2->unknownKeys(UnknownKeyMode::Allow);

    auto is = intersection({s1, s2});
    auto result = is->safe_parse(json::object({{"name", "Alice"}, {"age", 30}}));
    ASSERT(result.success);
}

TEST("intersection: rejects missing field") {
    auto s1 = object();
    s1->prop("name", string_());
    s1->required({"name"});
    s1->unknownKeys(UnknownKeyMode::Allow);

    auto s2 = object();
    s2->prop("age", int_());
    s2->required({"age"});
    s2->unknownKeys(UnknownKeyMode::Allow);

    auto is = intersection({s1, s2});
    auto result = is->safe_parse(json::object({{"name", "Alice"}}));
    ASSERT(!result.success);
}

TEST("intersection: numeric range") {
    auto s1 = number();
    s1->min(0);
    auto s2 = number();
    s2->max(100);
    auto is = intersection({s1, s2});
    ASSERT(is->safe_parse(json(50)).success);
    ASSERT(!is->safe_parse(json(-5)).success);
}

// ---- Optional Schema Tests ----

TEST("optional: accepts present valid value") {
    auto s = object();
    s->prop("name", optional_(string_()));
    auto result = s->safe_parse(json::object({{"name", "Alice"}}));
    ASSERT(result.success);
}

TEST("optional: accepts absent field") {
    auto s = object();
    s->prop("name", optional_(string_()));
    auto result = s->safe_parse(json::object());
    ASSERT(result.success);
}

TEST("optional: rejects invalid present value") {
    auto s = object();
    s->prop("name", optional_(string_()));
    auto result = s->safe_parse(json::object({{"name", 123}}));
    ASSERT(!result.success);
}

TEST("optional: null is not absent") {
    auto s = object();
    s->prop("name", optional_(string_()));
    auto result = s->safe_parse(json::object({{"name", nullptr}}));
    ASSERT(!result.success);
}

// ---- Nullable Schema Tests ----

TEST("nullable: accepts null") {
    auto s = nullable(string_());
    ASSERT(s->safe_parse(json(nullptr)).success);
}

TEST("nullable: accepts valid non-null") {
    auto s = nullable(string_());
    ASSERT(s->safe_parse(json("hello")).success);
}

TEST("nullable: rejects invalid non-null") {
    auto s = nullable(string_());
    auto result = s->safe_parse(json(42));
    ASSERT(!result.success);
}

TEST("nullable int: accepts null") {
    auto s = nullable(int_());
    ASSERT(s->safe_parse(json(nullptr)).success);
}

TEST("nullable int: accepts valid integer") {
    auto s = nullable(int_());
    ASSERT(s->safe_parse(json(99)).success);
}

// ---- Numeric Safety Tests ----

TEST("numeric safety: number roundtrips as float64") {
    auto s = number();
    auto result = s->safe_parse(json(1.7976931348623157e+308));
    ASSERT(result.success);
}

TEST("numeric safety: int roundtrips as int64") {
    auto s = int_();
    auto result = s->safe_parse(json(9007199254740991LL));
    ASSERT(result.success);
}

TEST("numeric safety: float64 and number identical") {
    auto s = float64();
    ASSERT(s->safe_parse(json(42.5)).success);
}

TEST("numeric safety: int64 and int identical") {
    auto s = int64();
    ASSERT(s->safe_parse(json(42)).success);
}

TEST("numeric safety: narrowing int8 rejected") {
    auto s = int8();
    auto result = s->safe_parse(json(200));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_large");
}

// ---- Throwing parse tests ----

TEST("parse: throws on invalid input") {
    auto s = string_();
    ASSERT_THROWS(s->parse(json(42)));
}

TEST("parse: returns value on valid input") {
    auto s = string_();
    auto result = s->parse(json("hello"));
    ASSERT(result == json("hello"));
}

// ---- Default Tests ----

TEST("default: missing field gets default") {
    auto s = object();
    auto role_schema = string_();
    role_schema->defaultValue("user");
    s->prop("name", string_());
    s->prop("role", role_schema);
    s->required({"name"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}}));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["role"], json("user"));
}

TEST("default: present field not overwritten") {
    auto s = object();
    auto role_schema = string_();
    role_schema->defaultValue("user");
    s->prop("name", string_());
    s->prop("role", role_schema);
    s->required({"name"});
    auto result = s->safe_parse(json::object({{"name", "Alice"}, {"role", "admin"}}));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["role"], json("admin"));
}

TEST("default: defaulted value is validated") {
    auto s = object();
    auto count_schema = int_();
    count_schema->min(1);
    count_schema->defaultValue(5);
    s->prop("count", count_schema);
    auto result = s->safe_parse(json::object());
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["count"], json(5));
}

TEST("default: invalid default produces error") {
    auto s = object();
    auto count_schema = int_();
    count_schema->min(10);
    count_schema->defaultValue(5);
    s->prop("count", count_schema);
    auto result = s->safe_parse(json::object());
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "default_invalid");
}

TEST("default: boolean default") {
    auto s = object();
    auto active = bool_();
    active->defaultValue(true);
    s->prop("active", active);
    auto result = s->safe_parse(json::object());
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["active"], json(true));
}

TEST("default: null not treated as absent for defaults") {
    auto s = object();
    auto inner = nullable(string_());
    inner->set_default(json("fallback"));
    s->prop("value", inner);
    auto result = s->safe_parse(json::object({{"value", nullptr}}));
    ASSERT(result.success);
    ASSERT(result.value["value"].is_null());
}

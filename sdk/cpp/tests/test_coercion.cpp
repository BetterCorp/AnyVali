#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>

using namespace anyvali;
using json = nlohmann::json;

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
    s->coerce(std::vector<std::string>{"trim", "lower"});
    auto result = s->safe_parse(json("  HELLO  "));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello"));
}

// ---------------------------------------------------------------------------
// CWE-20 / spec 5.1: non-portable coercion bypass
//
// std::stoll/std::stod are more permissive than the ECMA-262 reference (JS):
// they accept a leading "+", and std::stod accepts hex floats ("0x1p4") and
// "inf"/"nan". Each let a string the JS reference rejects coerce into a number.
// Coercion must accept ASCII decimals only so behaviour matches across SDKs.
// ---------------------------------------------------------------------------

TEST("security: string->int rejects non-decimal forms") {
    auto s = int_();
    s->coerce("string->int");
    for (const auto& bad : {"+5", "1.0", "1e3", "0x10"}) {
        ASSERT(!s->safe_parse(json(bad)).success);
    }
    ASSERT(s->safe_parse(json("42")).success);
    ASSERT(s->safe_parse(json("  42  ")).success);
}

TEST("security: string->number rejects hex-float and infinities") {
    auto s = number();
    s->coerce("string->number");
    for (const auto& bad : {"0x1p4", "0x10", "inf", "nan", "infinity"}) {
        ASSERT(!s->safe_parse(json(bad)).success);
    }
    ASSERT(s->safe_parse(json("3.14")).success);
    ASSERT(s->safe_parse(json("+5")).success);
    ASSERT(s->safe_parse(json(".5")).success);
}

// ---------------------------------------------------------------------------
// Default / no-explicit-source coercion (cross-SDK regression).
//
// Background: the JS SDK lets you enable coercion via an object config whose
// source ("from") is optional; when omitted on a numeric/bool schema the
// coercion silently no-ops, so a string like "3.14" fails with invalid_type
// instead of being coerced. The only portable coercion source is "string".
//
// The C++ builder has NO separate no-source form: the *simplest* way to enable
// coercion is to name the portable token (e.g. coerce("string->int")), which
// intrinsically binds the source to "string". These tests exercise that
// default/idiomatic form end-to-end through safe_parse, mirroring the regression
// cases added in the other SDKs. They MUST coerce (not invalid_type).
// ---------------------------------------------------------------------------

TEST("default-coerce: number coerces \"3.14\" -> 3.14") {
    auto s = number();
    s->coerce("string->number");
    auto result = s->safe_parse(json("3.14"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(3.14));
}

TEST("default-coerce: int coerces \"42\" -> 42") {
    auto s = int_();
    s->coerce("string->int");
    auto result = s->safe_parse(json("42"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(42));
}

TEST("default-coerce: bool coerces \"true\"/\"false\"") {
    auto s = bool_();
    s->coerce("string->bool");
    auto t = s->safe_parse(json("true"));
    ASSERT(t.success);
    ASSERT_JSON_EQ(t.value, json(true));
    auto f = s->safe_parse(json("false"));
    ASSERT(f.success);
    ASSERT_JSON_EQ(f.value, json(false));
}

TEST("default-coerce: object with numeric fields coerces all string inputs") {
    auto s = object();
    s->prop("age", [] { auto a = int_(); a->coerce("string->int"); return a; }());
    s->prop("score", [] { auto sc = number(); sc->coerce("string->number"); return sc; }());
    s->prop("active", [] { auto b = bool_(); b->coerce("string->bool"); return b; }());
    s->required({"age", "score", "active"});

    json input = {{"age", "42"}, {"score", "3.14"}, {"active", "true"}};
    auto result = s->safe_parse(input);
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["age"], json(42));
    ASSERT_JSON_EQ(result.value["score"], json(3.14));
    ASSERT_JSON_EQ(result.value["active"], json(true));
}

// ===========================================================================
// No-arg coerce() ergonomic + full FROM-STRING coercion matrix.
//
// coerce() with NO arguments enables string coercion with the target inferred
// from the schema kind:
//   - int-family kinds  -> "string->int"   (ASCII ^-?\d+$, trimmed)
//   - float/number kinds -> "string->number" (ASCII decimal float incl exponent)
//   - bool              -> "string->bool"   (trim + case-insensitive)
// String transforms (trim/lower/upper) stay explicit (string kind only) since
// they cannot be inferred from kind.
//
// Each ACCEPT/REJECT row of the canonical matrix is exercised below through the
// no-arg form end-to-end via safe_parse. Failures surface as coercion_failed.
// ===========================================================================

// ---- string -> int : ACCEPT "42","  42  ","-7" ----
TEST("matrix int noarg: accepts 42") {
    auto s = int_(); s->coerce();
    auto r = s->safe_parse(json("42"));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(42));
}

TEST("matrix int noarg: accepts padded 42") {
    auto s = int_(); s->coerce();
    auto r = s->safe_parse(json("  42  "));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(42));
}

TEST("matrix int noarg: accepts -7") {
    auto s = int_(); s->coerce();
    auto r = s->safe_parse(json("-7"));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(-7));
}

// ---- string -> int : REJECT "3.14","0x10","1_000","+5","Infinity","","abc" ----
TEST("matrix int noarg: rejects non-decimal forms") {
    auto s = int_(); s->coerce();
    for (const auto& bad : {"3.14", "0x10", "1_000", "+5", "Infinity", "", "abc"}) {
        auto r = s->safe_parse(json(bad));
        ASSERT(!r.success);
        ASSERT(r.issues[0].code == "coercion_failed");
    }
}

// ---- string -> number : ACCEPT "3.14","-1.5e3","  2  ","0" ----
TEST("matrix number noarg: accepts 3.14") {
    auto s = number(); s->coerce();
    auto r = s->safe_parse(json("3.14"));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(3.14));
}

TEST("matrix number noarg: accepts -1.5e3") {
    auto s = number(); s->coerce();
    auto r = s->safe_parse(json("-1.5e3"));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(-1500.0));
}

TEST("matrix number noarg: accepts padded 2") {
    auto s = number(); s->coerce();
    auto r = s->safe_parse(json("  2  "));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(2.0));
}

TEST("matrix number noarg: accepts 0") {
    auto s = number(); s->coerce();
    auto r = s->safe_parse(json("0"));
    ASSERT(r.success);
    ASSERT_JSON_EQ(r.value, json(0.0));
}

// ---- string -> number : REJECT "0x10","Infinity","NaN","","1_000","abc" ----
TEST("matrix number noarg: rejects non-decimal forms") {
    auto s = number(); s->coerce();
    for (const auto& bad : {"0x10", "Infinity", "NaN", "", "1_000", "abc"}) {
        auto r = s->safe_parse(json(bad));
        ASSERT(!r.success);
        ASSERT(r.issues[0].code == "coercion_failed");
    }
}

// ---- string -> bool : ACCEPT true<-"true","TRUE","1"; false<-"false","0" ----
TEST("matrix bool noarg: accepts true forms") {
    auto s = bool_(); s->coerce();
    for (const auto& good : {"true", "TRUE", "1"}) {
        auto r = s->safe_parse(json(good));
        ASSERT(r.success);
        ASSERT_JSON_EQ(r.value, json(true));
    }
}

TEST("matrix bool noarg: accepts false forms") {
    auto s = bool_(); s->coerce();
    for (const auto& good : {"false", "0"}) {
        auto r = s->safe_parse(json(good));
        ASSERT(r.success);
        ASSERT_JSON_EQ(r.value, json(false));
    }
}

TEST("matrix bool noarg: trims and is case-insensitive") {
    auto s = bool_(); s->coerce();
    ASSERT_JSON_EQ(s->safe_parse(json("  TrUe ")).value, json(true));
    ASSERT_JSON_EQ(s->safe_parse(json(" FALSE ")).value, json(false));
}

// ---- string -> bool : REJECT "yes","no","on","off","t","f","2","" ----
TEST("matrix bool noarg: rejects ambiguous forms") {
    auto s = bool_(); s->coerce();
    for (const auto& bad : {"yes", "no", "on", "off", "t", "f", "2", ""}) {
        auto r = s->safe_parse(json(bad));
        ASSERT(!r.success);
        ASSERT(r.issues[0].code == "coercion_failed");
    }
}

// ---- no-arg target inference per kind ----
TEST("matrix noarg: int-family infers string->int (rejects float text)") {
    // int8/int32/uint16 etc. must reject "3.14" like plain int does.
    {
        auto s = int32(); s->coerce();
        ASSERT(!s->safe_parse(json("3.14")).success);
        ASSERT_JSON_EQ(s->safe_parse(json("5")).value, json(5));
    }
    {
        auto s = uint16(); s->coerce();
        ASSERT(!s->safe_parse(json("3.14")).success);
        ASSERT_JSON_EQ(s->safe_parse(json("5")).value, json(5));
    }
}

TEST("matrix noarg: float32 infers string->number (accepts float text)") {
    auto s = float32(); s->coerce();
    auto r = s->safe_parse(json("3.14"));
    ASSERT(r.success);  // float kind -> string->number accepts decimals
}

// ---- no-arg form composes with downstream numeric validation ----
TEST("matrix noarg: coercion happens before range validation") {
    auto s = int_(); s->coerce(); s->min(10);
    auto r = s->safe_parse(json("5"));
    ASSERT(!r.success);
    ASSERT(r.issues[0].code == "too_small");  // coerced to 5, then range fails
}

TEST("matrix noarg: typed-token overload still works alongside") {
    auto s = int_(); s->coerce("string->int");
    ASSERT_JSON_EQ(s->safe_parse(json("42")).value, json(42));
}

// ---- string transforms stay explicit (string kind only), chainable ----
TEST("matrix string transforms: trim/lower/upper chainable") {
    {
        auto s = string_(); s->coerce("trim");
        ASSERT_JSON_EQ(s->safe_parse(json("  hi  ")).value, json("hi"));
    }
    {
        auto s = string_(); s->coerce("lower");
        ASSERT_JSON_EQ(s->safe_parse(json("HeLLo")).value, json("hello"));
    }
    {
        auto s = string_(); s->coerce("upper");
        ASSERT_JSON_EQ(s->safe_parse(json("HeLLo")).value, json("HELLO"));
    }
    {
        auto s = string_(); s->coerce(std::vector<std::string>{"trim", "upper"});
        ASSERT_JSON_EQ(s->safe_parse(json("  hi  ")).value, json("HI"));
    }
}

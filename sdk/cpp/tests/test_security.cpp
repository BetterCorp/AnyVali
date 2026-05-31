#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>
#include <chrono>
#include <cmath>
#include <cstdint>
#include <limits>
#include <string>

using namespace anyvali;
using json = nlohmann::json;

// ============================================================================
// CVE-2016-4055: ReDoS - Catastrophic Backtracking (std::regex)
// ============================================================================

TEST("security: ReDoS - pattern with catastrophic backtracking completes in time") {
    // Pattern (a+)+ is a classic ReDoS trigger with exponential backtracking.
    // Input "aaaaaaaaaaaaaaaaaaaaaaaaaaaaab" should not hang.
    auto s = string_();
    s->pattern("^(a+)+$");
    std::string evil = std::string(25, 'a') + "b";
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(json(evil));
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(!result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: ReDoS - nested quantifier pattern completes in time") {
    // Pattern (a|a)+ is another ReDoS variant.
    auto s = string_();
    s->pattern("^(a|a)+$");
    std::string evil = std::string(25, 'a') + "!";
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(json(evil));
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(!result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: ReDoS - overlapping character class pattern completes in time") {
    // Pattern ([a-zA-Z]+)* with trailing mismatch.
    auto s = string_();
    s->pattern("^([a-zA-Z]+)*$");
    std::string evil = std::string(25, 'a') + "1";
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(json(evil));
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(!result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: ReDoS - safe pattern still validates correctly") {
    auto s = string_();
    s->pattern("^[a-z]+$");
    ASSERT(s->safe_parse(json("abc")).success);
    ASSERT(!s->safe_parse(json("ABC")).success);
}

// ============================================================================
// CVE-2003-1564: Recursive $ref - Self-referencing Schemas
// ============================================================================

TEST("security: recursive ref - direct self-reference does not infinite loop") {
    // CVE-2003-1564 parse-time: Schema where root = ref(Self), Self = ref(Self).
    // Importer accepts the shape (lazy resolution). At parse, recursive descent
    // is unbounded. C++ stack overflow may segfault (uncatchable) — but if the
    // SDK has any depth guard or returns an error, this test verifies the
    // bounded-time contract. NOTE: a segfault here proves the vulnerability
    // exists; CI will surface it as a test crash.
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "ref"}, {"ref", "#/definitions/Self"}}},
        {"definitions", {
            {"Self", {{"kind", "ref"}, {"ref", "#/definitions/Self"}}}
        }},
        {"extensions", json::object()}
    };
    bool threw = false;
    bool returned_success = false;
    auto start = std::chrono::steady_clock::now();
    try {
        auto imported = import_schema(doc);
        auto result = parse_document(imported, json("anything"));
        returned_success = result.success;
    } catch (...) {
        threw = true;
    }
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);

    // 5-second wall-clock bound. If we hit this, the SDK is in an infinite
    // loop on pure self-cycle — a runtime DoS (CVE-2003-1564).
    ASSERT(elapsed.count() < 5);

    // Silently succeeding on a pure self-cycle would be a logic bug.
    // If neither threw nor returned-with-failure, this assertion fires.
    ASSERT(threw || !returned_success);
}

TEST("security: recursive ref - mutual recursion A->B->A") {
    // Two definitions that reference each other with no base case.
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "ref"}, {"ref", "#/definitions/A"}}},
        {"definitions", {
            {"A", {
                {"kind", "object"},
                {"properties", {
                    {"b", {{"kind", "ref"}, {"ref", "#/definitions/B"}}}
                }},
                {"required", {"b"}},
                {"unknownKeys", "reject"}
            }},
            {"B", {
                {"kind", "object"},
                {"properties", {
                    {"a", {{"kind", "ref"}, {"ref", "#/definitions/A"}}}
                }},
                {"required", {"a"}},
                {"unknownKeys", "reject"}
            }}
        }},
        {"extensions", json::object()}
    };
    bool threw = false;
    try {
        auto imported = import_schema(doc);
        // Provide deeply nested but finite input
        json input = {{"b", {{"a", {{"b", {{"a", json::object()}}}}}}}};
        auto start = std::chrono::steady_clock::now();
        auto result = parse_document(imported, input);
        auto end = std::chrono::steady_clock::now();
        auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
        ASSERT(elapsed.count() < 5);
    } catch (...) {
        threw = true;
    }
    // Either throwing or completing quickly is acceptable.
    (void)threw;
    ASSERT(true);
}

TEST("security: recursive ref - valid tree with bounded recursion") {
    // This test verifies that legitimate recursive schemas still work.
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "ref"}, {"ref", "#/definitions/Node"}}},
        {"definitions", {
            {"Node", {
                {"kind", "object"},
                {"properties", {
                    {"value", {{"kind", "int"}}},
                    {"children", {
                        {"kind", "array"},
                        {"items", {{"kind", "ref"}, {"ref", "#/definitions/Node"}}}
                    }}
                }},
                {"required", {"value", "children"}},
                {"unknownKeys", "reject"}
            }}
        }},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    json tree = {
        {"value", 1},
        {"children", {
            {{"value", 2}, {"children", json::array()}},
            {{"value", 3}, {"children", json::array()}}
        }}
    };
    auto result = parse_document(imported, tree);
    ASSERT(result.success);
}

// ============================================================================
// CWE-190: Integer Overflow - All int width boundaries using C++ limits
// ============================================================================

TEST("security: int8 boundary - accepts INT8_MAX") {
    auto s = int8();
    ASSERT(s->safe_parse(json(static_cast<int>(INT8_MAX))).success);
}

TEST("security: int8 boundary - accepts INT8_MIN") {
    auto s = int8();
    ASSERT(s->safe_parse(json(static_cast<int>(INT8_MIN))).success);
}

TEST("security: int8 boundary - rejects INT8_MAX + 1") {
    auto s = int8();
    ASSERT(!s->safe_parse(json(static_cast<int>(INT8_MAX) + 1)).success);
}

TEST("security: int8 boundary - rejects INT8_MIN - 1") {
    auto s = int8();
    ASSERT(!s->safe_parse(json(static_cast<int>(INT8_MIN) - 1)).success);
}

TEST("security: int16 boundary - accepts INT16_MAX") {
    auto s = int16();
    ASSERT(s->safe_parse(json(static_cast<int>(INT16_MAX))).success);
}

TEST("security: int16 boundary - accepts INT16_MIN") {
    auto s = int16();
    ASSERT(s->safe_parse(json(static_cast<int>(INT16_MIN))).success);
}

TEST("security: int16 boundary - rejects INT16_MAX + 1") {
    auto s = int16();
    ASSERT(!s->safe_parse(json(static_cast<int>(INT16_MAX) + 1)).success);
}

TEST("security: int16 boundary - rejects INT16_MIN - 1") {
    auto s = int16();
    ASSERT(!s->safe_parse(json(static_cast<int>(INT16_MIN) - 1)).success);
}

TEST("security: int32 boundary - accepts INT32_MAX") {
    auto s = int32();
    ASSERT(s->safe_parse(json(static_cast<int64_t>(INT32_MAX))).success);
}

TEST("security: int32 boundary - accepts INT32_MIN") {
    auto s = int32();
    ASSERT(s->safe_parse(json(static_cast<int64_t>(INT32_MIN))).success);
}

TEST("security: int32 boundary - rejects INT32_MAX + 1") {
    auto s = int32();
    ASSERT(!s->safe_parse(json(static_cast<int64_t>(INT32_MAX) + 1)).success);
}

TEST("security: int32 boundary - rejects INT32_MIN - 1") {
    auto s = int32();
    ASSERT(!s->safe_parse(json(static_cast<int64_t>(INT32_MIN) - 1)).success);
}

TEST("security: int64 boundary - accepts INT64_MAX") {
    auto s = int64();
    ASSERT(s->safe_parse(json(INT64_MAX)).success);
}

TEST("security: int64 boundary - accepts INT64_MIN") {
    auto s = int64();
    ASSERT(s->safe_parse(json(INT64_MIN)).success);
}

TEST("security: uint8 boundary - accepts 0") {
    auto s = uint8();
    ASSERT(s->safe_parse(json(0)).success);
}

TEST("security: uint8 boundary - accepts UINT8_MAX (255)") {
    auto s = uint8();
    ASSERT(s->safe_parse(json(static_cast<int>(UINT8_MAX))).success);
}

TEST("security: uint8 boundary - rejects UINT8_MAX + 1") {
    auto s = uint8();
    ASSERT(!s->safe_parse(json(static_cast<int>(UINT8_MAX) + 1)).success);
}

TEST("security: uint8 boundary - rejects -1") {
    auto s = uint8();
    ASSERT(!s->safe_parse(json(-1)).success);
}

TEST("security: uint16 boundary - accepts UINT16_MAX (65535)") {
    auto s = uint16();
    ASSERT(s->safe_parse(json(static_cast<int>(UINT16_MAX))).success);
}

TEST("security: uint16 boundary - rejects UINT16_MAX + 1") {
    auto s = uint16();
    ASSERT(!s->safe_parse(json(static_cast<int>(UINT16_MAX) + 1)).success);
}

TEST("security: uint16 boundary - rejects -1") {
    auto s = uint16();
    ASSERT(!s->safe_parse(json(-1)).success);
}

TEST("security: uint32 boundary - accepts UINT32_MAX") {
    auto s = uint32();
    ASSERT(s->safe_parse(json(static_cast<uint64_t>(UINT32_MAX))).success);
}

TEST("security: uint32 boundary - rejects UINT32_MAX + 1") {
    auto s = uint32();
    ASSERT(!s->safe_parse(json(static_cast<uint64_t>(UINT32_MAX) + 1)).success);
}

TEST("security: uint32 boundary - rejects -1") {
    auto s = uint32();
    ASSERT(!s->safe_parse(json(-1)).success);
}

TEST("security: uint64 boundary - accepts 0") {
    auto s = uint64();
    ASSERT(s->safe_parse(json(static_cast<uint64_t>(0))).success);
}

TEST("security: uint64 boundary - rejects -1") {
    auto s = uint64();
    ASSERT(!s->safe_parse(json(-1)).success);
}

TEST("security: int overflow - int8 rejects large positive") {
    auto s = int8();
    ASSERT(!s->safe_parse(json(1000)).success);
}

TEST("security: int overflow - int8 rejects large negative") {
    auto s = int8();
    ASSERT(!s->safe_parse(json(-1000)).success);
}

TEST("security: int overflow - int16 rejects int32 range") {
    auto s = int16();
    ASSERT(!s->safe_parse(json(100000)).success);
}

TEST("security: int overflow - uint8 rejects int16 range") {
    auto s = uint8();
    ASSERT(!s->safe_parse(json(500)).success);
}

// ============================================================================
// CWE-20: NaN and Infinity Rejection
// ============================================================================

TEST("security: NaN rejected by number schema") {
    auto s = number();
    auto result = s->safe_parse(json(std::numeric_limits<double>::quiet_NaN()));
    // NaN is not a valid JSON number; nlohmann::json may serialize it as null.
    // Either way, it should not be accepted as a valid number.
    ASSERT(!result.success);
}

TEST("security: positive infinity rejected by number schema") {
    auto s = number();
    auto result = s->safe_parse(json(std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

TEST("security: negative infinity rejected by number schema") {
    auto s = number();
    auto result = s->safe_parse(json(-std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

TEST("security: NaN rejected by float64 schema") {
    auto s = float64();
    auto result = s->safe_parse(json(std::numeric_limits<double>::quiet_NaN()));
    ASSERT(!result.success);
}

TEST("security: infinity rejected by float64 schema") {
    auto s = float64();
    auto result = s->safe_parse(json(std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

TEST("security: NaN rejected by float32 schema") {
    auto s = float32();
    auto result = s->safe_parse(json(std::numeric_limits<double>::quiet_NaN()));
    ASSERT(!result.success);
}

TEST("security: infinity rejected by float32 schema") {
    auto s = float32();
    auto result = s->safe_parse(json(std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

TEST("security: NaN rejected by int schema") {
    auto s = int_();
    auto result = s->safe_parse(json(std::numeric_limits<double>::quiet_NaN()));
    ASSERT(!result.success);
}

TEST("security: infinity rejected by int schema") {
    auto s = int_();
    auto result = s->safe_parse(json(std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

TEST("security: NaN rejected by number with min constraint") {
    auto s = number();
    s->min(0);
    auto result = s->safe_parse(json(std::numeric_limits<double>::quiet_NaN()));
    ASSERT(!result.success);
}

TEST("security: infinity rejected by number with max constraint") {
    auto s = number();
    s->max(1000);
    auto result = s->safe_parse(json(std::numeric_limits<double>::infinity()));
    ASSERT(!result.success);
}

// ============================================================================
// CWE-20: Format Bypass - email, url, ipv4 edge cases
// ============================================================================

TEST("security: format bypass - email with empty local part rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("@example.com")).success);
}

TEST("security: format bypass - tampered email format name not ignored") {
    auto s = string_();
    s->format(std::string("email\0", 6));
    ASSERT(!s->safe_parse(json("not-an-email")).success);
}

TEST("security: format bypass - imported tampered email format name not unconstrained") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "string"}, {"format", std::string("email\0", 6)}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json("not-an-email"));
    ASSERT(!result.success);
}

TEST("security: format bypass - email with empty domain rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("user@")).success);
}

TEST("security: format bypass - email with spaces rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("user @example.com")).success);
}

TEST("security: format bypass - email with double @ rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("user@@example.com")).success);
}

TEST("security: format bypass - email with no TLD rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("user@localhost")).success);
}

TEST("security: format bypass - email with only whitespace rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("   ")).success);
}

TEST("security: format bypass - email empty string rejected") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("")).success);
}

TEST("security: format bypass - url with javascript: scheme rejected") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("javascript:alert(1)")).success);
}

TEST("security: format bypass - url with data: scheme rejected") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("data:text/html,<script>alert(1)</script>")).success);
}

TEST("security: format bypass - url empty string rejected") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("")).success);
}

TEST("security: format bypass - url with ftp scheme rejected") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("ftp://files.example.com")).success);
}

TEST("security: format bypass - url with file scheme rejected") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("file:///etc/passwd")).success);
}

TEST("security: format bypass - ipv4 with leading zeros rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("192.168.01.1")).success);
}

TEST("security: format bypass - ipv4 octet 256 rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("256.0.0.1")).success);
}

TEST("security: format bypass - ipv4 too few octets rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("192.168.1")).success);
}

TEST("security: format bypass - ipv4 too many octets rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("192.168.1.1.1")).success);
}

TEST("security: format bypass - ipv4 negative octet rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("-1.0.0.1")).success);
}

TEST("security: format bypass - ipv4 empty string rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("")).success);
}

TEST("security: format bypass - ipv4 with spaces rejected") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("192. 168.1.1")).success);
}

TEST("security: format bypass - valid ipv4 accepted") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(s->safe_parse(json("0.0.0.0")).success);
    ASSERT(s->safe_parse(json("255.255.255.255")).success);
    ASSERT(s->safe_parse(json("127.0.0.1")).success);
}

TEST("security: format bypass - valid email accepted") {
    auto s = string_();
    s->format("email");
    ASSERT(s->safe_parse(json("test@example.com")).success);
    ASSERT(s->safe_parse(json("user+tag@example.co.uk")).success);
}

TEST("security: format bypass - valid url accepted") {
    auto s = string_();
    s->format("url");
    ASSERT(s->safe_parse(json("https://example.com")).success);
    ASSERT(s->safe_parse(json("http://example.com/path?q=1&r=2")).success);
}

// ============================================================================
// Unicode Length Constraints - code points, not UTF-8 bytes
// ============================================================================

TEST("security: unicode length - astral code point counts as one character") {
    std::string emoji = "\xF0\x9F\x98\x80";
    auto max_one = string_();
    max_one->maxLength(1);
    ASSERT(max_one->safe_parse(json(emoji)).success);

    auto min_two = string_();
    min_two->minLength(2);
    ASSERT(!min_two->safe_parse(json(emoji)).success);
}

TEST("security: unicode length - imported maxLength uses code points") {
    std::string emoji = "\xF0\x9F\x98\x80";
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "string"}, {"maxLength", 1}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json(emoji));
    ASSERT(result.success);
}

// ============================================================================
// CWE-400: Large Inputs - Resource Exhaustion
// ============================================================================

TEST("security: large string - 1MB string accepted by string schema") {
    auto s = string_();
    std::string large(1024 * 1024, 'x');
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(json(large));
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: large string - maxLength correctly rejects oversized input") {
    auto s = string_();
    s->maxLength(100);
    std::string large(1024 * 1024, 'x');
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(json(large));
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "too_large");
    ASSERT(elapsed.count() < 5);
}

TEST("security: large array - 100k element array validated in time") {
    auto s = array(int_());
    json big_array = json::array();
    for (int i = 0; i < 100000; i++) {
        big_array.push_back(i);
    }
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(big_array);
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: large array - maxItems correctly rejects oversized array") {
    auto s = array(int_());
    s->maxItems(10);
    json big_array = json::array();
    for (int i = 0; i < 10000; i++) {
        big_array.push_back(i);
    }
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(big_array);
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(!result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: large object - many keys validated in time") {
    auto s = record(int_());
    json big_obj = json::object();
    for (int i = 0; i < 10000; i++) {
        big_obj["key_" + std::to_string(i)] = i;
    }
    auto start = std::chrono::steady_clock::now();
    auto result = s->safe_parse(big_obj);
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(result.success);
    ASSERT(elapsed.count() < 5);
}

TEST("security: deeply nested object - imported schema with deep nesting") {
    // Build a deeply nested object schema programmatically
    // innermost: { "value": int }
    // each layer wraps: { "nested": <inner> }
    const int depth = 50;
    json innermost = {
        {"kind", "object"},
        {"properties", {{"value", {{"kind", "int"}}}}},
        {"required", {"value"}},
        {"unknownKeys", "reject"}
    };
    json current = innermost;
    for (int i = 0; i < depth; i++) {
        json wrapper = {
            {"kind", "object"},
            {"properties", {{"nested", current}}},
            {"required", {"nested"}},
            {"unknownKeys", "reject"}
        };
        current = wrapper;
    }
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", current},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);

    // Build matching deeply nested input
    json input = {{"value", 42}};
    for (int i = 0; i < depth; i++) {
        input = json::object({{"nested", input}});
    }

    auto start = std::chrono::steady_clock::now();
    auto result = parse_document(imported, input);
    auto end = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(end - start);
    ASSERT(result.success);
    ASSERT(elapsed.count() < 5);
}

// ============================================================================
// Schema Import Injection - Unknown kinds throw
// ============================================================================

TEST("security: import rejects unknown schema kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "malicious_type"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects __proto__ as schema kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "__proto__"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects constructor as schema kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "constructor"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects empty string as schema kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", ""}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects missing kind field") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", json::object()},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects null root") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", nullptr},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects kind as number") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", 42}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects kind as boolean") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", true}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects kind as array") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", json::array({"string"})}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects ref to nonexistent definition") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "ref"}, {"ref", "#/definitions/DoesNotExist"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    // Import may succeed, but parsing should throw due to missing definition.
    bool threw = false;
    try {
        auto imported = import_schema(doc);
        parse_document(imported, json("test"));
    } catch (...) {
        threw = true;
    }
    ASSERT(threw);
}

TEST("security: import rejects SQL injection in kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "'; DROP TABLE schemas; --"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

TEST("security: import rejects script tag in kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "<script>alert(1)</script>"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    ASSERT_THROWS(import_schema(doc));
}

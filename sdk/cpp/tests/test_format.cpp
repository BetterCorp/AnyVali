#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>

using namespace anyvali;
using json = nlohmann::json;

TEST("format: email valid") {
    auto s = string_();
    s->format("email");
    ASSERT(s->safe_parse(json("user@example.com")).success);
}

TEST("format: email missing @") {
    auto s = string_();
    s->format("email");
    auto result = s->safe_parse(json("not-an-email"));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "invalid_string");
}

TEST("format: email no dot after @") {
    auto s = string_();
    s->format("email");
    ASSERT(!s->safe_parse(json("user@localhost")).success);
}

TEST("format: url https") {
    auto s = string_();
    s->format("url");
    ASSERT(s->safe_parse(json("https://example.com")).success);
}

TEST("format: url http with path") {
    auto s = string_();
    s->format("url");
    ASSERT(s->safe_parse(json("http://example.com/path?q=1")).success);
}

TEST("format: url rejects ftp") {
    auto s = string_();
    s->format("url");
    ASSERT(!s->safe_parse(json("ftp://files.example.com")).success);
}

TEST("format: uuid valid") {
    auto s = string_();
    s->format("uuid");
    ASSERT(s->safe_parse(json("550e8400-e29b-41d4-a716-446655440000")).success);
}

TEST("format: uuid invalid") {
    auto s = string_();
    s->format("uuid");
    ASSERT(!s->safe_parse(json("not-a-uuid")).success);
}

TEST("format: ipv4 valid") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(s->safe_parse(json("192.168.1.1")).success);
}

TEST("format: ipv4 rejects leading zeros") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("192.168.01.1")).success);
}

TEST("format: ipv4 rejects out of range") {
    auto s = string_();
    s->format("ipv4");
    ASSERT(!s->safe_parse(json("256.1.1.1")).success);
}

TEST("format: ipv6 valid full") {
    auto s = string_();
    s->format("ipv6");
    ASSERT(s->safe_parse(json("2001:0db8:85a3:0000:0000:8a2e:0370:7334")).success);
}

TEST("format: ipv6 compressed loopback") {
    auto s = string_();
    s->format("ipv6");
    ASSERT(s->safe_parse(json("::1")).success);
}

TEST("format: ipv6 invalid") {
    auto s = string_();
    s->format("ipv6");
    ASSERT(!s->safe_parse(json("not:an:ipv6")).success);
}

TEST("format: date valid") {
    auto s = string_();
    s->format("date");
    ASSERT(s->safe_parse(json("2024-02-29")).success);
}

TEST("format: date invalid leap day") {
    auto s = string_();
    s->format("date");
    ASSERT(!s->safe_parse(json("2023-02-29")).success);
}

TEST("format: date-time with Z") {
    auto s = string_();
    s->format("date-time");
    ASSERT(s->safe_parse(json("2024-01-15T10:30:00Z")).success);
}

TEST("format: date-time with offset") {
    auto s = string_();
    s->format("date-time");
    ASSERT(s->safe_parse(json("2024-01-15T10:30:00+05:30")).success);
}

TEST("format: date-time without timezone rejected") {
    auto s = string_();
    s->format("date-time");
    ASSERT(!s->safe_parse(json("2024-01-15T10:30:00")).success);
}

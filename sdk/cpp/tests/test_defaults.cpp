#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>

using namespace anyvali;
using json = nlohmann::json;

TEST("defaults: missing field gets default") {
    auto s = object();
    auto role = string_();
    role->defaultValue("user");
    s->prop("name", string_());
    s->prop("role", role);
    s->required({"name"});

    auto result = s->safe_parse(json::object({{"name", "Alice"}}));

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["role"], json("user"));
}

TEST("defaults: present field is not overwritten") {
    auto s = object();
    auto role = string_();
    role->defaultValue("user");
    s->prop("role", role);

    auto result = s->safe_parse(json::object({{"role", "admin"}}));

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["role"], json("admin"));
}

TEST("defaults: invalid default produces default_invalid") {
    auto s = object();
    auto count = int_();
    count->min(10);
    count->set_default(json(5));
    s->prop("count", count);

    auto result = s->safe_parse(json::object());

    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "default_invalid");
    ASSERT(result.issues[0].path.size() == 1);
}

TEST("defaults: null is not absent for nullable default") {
    auto s = object();
    auto value = nullable(string_());
    value->set_default(json("fallback"));
    s->prop("value", value);

    auto result = s->safe_parse(json::object({{"value", nullptr}}));

    ASSERT(result.success);
    ASSERT(result.value["value"].is_null());
}

TEST("defaults: falsy defaults are applied") {
    auto s = object();
    auto count = int_();
    count->set_default(json(0));
    auto name = string_();
    name->defaultValue("");
    auto active = bool_();
    active->defaultValue(false);
    s->prop("count", count);
    s->prop("name", name);
    s->prop("active", active);

    auto result = s->safe_parse(json::object());

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json::object({{"count", 0}, {"name", ""}, {"active", false}}));
}

TEST("defaults: optional wrapper field gets default") {
    auto s = object();
    auto host = optional_(string_());
    host->set_default(json("localhost"));
    s->prop("host", host);

    auto result = s->safe_parse(json::object());

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json::object({{"host", "localhost"}}));
}

TEST("defaults: optional wrapper default does not override present field") {
    auto s = object();
    auto host = optional_(string_());
    host->set_default(json("localhost"));
    s->prop("host", host);

    auto result = s->safe_parse(json::object({{"host", "example.com"}}));

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json::object({{"host", "example.com"}}));
}

TEST("defaults: optional wrapper default is validated") {
    auto s = object();
    auto inner = string_();
    inner->minLength(5);
    auto host = optional_(inner);
    host->set_default(json("hi"));
    s->prop("host", host);

    auto result = s->safe_parse(json::object());

    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "default_invalid");
    ASSERT(result.issues[0].path.size() == 1);
}

TEST("defaults: optional wrapper default is exported") {
    auto host = optional_(string_());
    host->set_default(json("localhost"));

    auto node = host->export_node();

    ASSERT_JSON_EQ(node["kind"], json("optional"));
    ASSERT_JSON_EQ(node["default"], json("localhost"));
}

TEST("defaults: mutable optional wrapper default is not shared between parses") {
    auto s = object();
    auto meta = optional_(any());
    meta->set_default(json::object({{"items", json::array()}}));
    s->prop("meta", meta);

    auto first = s->parse(json::object());
    first["meta"]["items"].push_back("mutated");

    auto second = s->parse(json::object());
    ASSERT_JSON_EQ(second, json::object({{"meta", json::object({{"items", json::array()}})}}));
}

TEST("defaults: nested object field gets default") {
    auto s = object();
    auto user = object();
    auto role = string_();
    role->defaultValue("guest");
    user->prop("name", string_());
    user->prop("role", role);
    user->required({"name"});
    s->prop("user", user);
    s->required({"user"});

    auto result = s->safe_parse(json::object({{"user", json::object({{"name", "Bob"}})}}));

    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["user"]["role"], json("guest"));
}

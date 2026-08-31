#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>

using namespace anyvali;
using json = nlohmann::json;

TEST("sensitive: encrypt validates storage and decrypt restores plaintext") {
    auto secret = string_();
    secret->minLength(3);
    DescribeOpts opts;
    opts.sensitive = true;
    secret->describe("secret", opts);

    auto schema = object();
    schema->prop("id", int_());
    schema->prop("secret", secret);
    schema->required({"id", "secret"});
    const auto input = json{{"id", 1}, {"secret", "clear"}};

    const auto encrypted = encrypt(*schema, input, [](const Path& path, const json& value) {
        ASSERT(path.size() == 1);
        ASSERT(std::get<std::string>(path[0]) == "secret");
        return json("encrypted:" + value.get<std::string>());
    });
    ASSERT_JSON_EQ(encrypted, json({{"id", 1}, {"secret", "encrypted:clear"}}));
    ASSERT(safe_parse_encrypted(*schema, encrypted).success);
    ASSERT(!safe_parse_encrypted(*schema, input).success);

    const auto decrypted = decrypt(*schema, encrypted, [](const Path&, const json& value) {
        return json(value.get<std::string>().substr(10));
    });
    ASSERT_JSON_EQ(decrypted, input);
}

TEST("sensitive: composite nodes are opaque and bad callbacks fail") {
    auto credentials = object();
    credentials->prop("user", string_());
    credentials->required({"user"});
    DescribeOpts opts;
    opts.sensitive = true;
    credentials->describe("credentials", opts);
    auto schema = object();
    schema->prop("credentials", credentials);
    schema->required({"credentials"});

    ASSERT_THROWS(encrypt(*schema, json{{"credentials", {{"user", "alice"}}}},
                          [](const Path&, const json&) { return json("broken"); }));
}

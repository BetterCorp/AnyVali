#include "test_framework.hpp"
#include <anyvali/anyvali.hpp>

using namespace anyvali;
using json = nlohmann::json;

TEST("import: basic string schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "string"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(imported.root->kind() == SchemaKind::String);
    auto result = parse_document(imported, json("hello"));
    ASSERT(result.success);
}

TEST("import: number with constraints") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "number"}, {"min", 0}, {"max", 100}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json(50)).success);
    ASSERT(!parse_document(imported, json(-1)).success);
}

TEST("import: object schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {
                {"name", {{"kind", "string"}}},
                {"age", {{"kind", "int"}}}
            }},
            {"required", {"name", "age"}},
            {"unknownKeys", "reject"}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json::object({{"name", "Alice"}, {"age", 30}}));
    ASSERT(result.success);
}

TEST("import: union schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "union"},
            {"variants", {{{"kind", "string"}}, {{"kind", "int"}}}}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json("hello")).success);
    ASSERT(parse_document(imported, json(42)).success);
    ASSERT(!parse_document(imported, json(true)).success);
}

TEST("import: intersection schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "intersection"},
            {"allOf", {
                {{"kind", "number"}, {"min", 0}},
                {{"kind", "number"}, {"max", 100}}
            }}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json(50)).success);
}

TEST("import: nullable schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "nullable"}, {"schema", {{"kind", "string"}}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json(nullptr)).success);
    ASSERT(parse_document(imported, json("hello")).success);
}

TEST("import: array schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "array"}, {"items", {{"kind", "int"}}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json::array({1, 2, 3})).success);
}

TEST("import: tuple schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "tuple"}, {"elements", {{{"kind", "string"}}, {{"kind", "int"}}}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json::array({"hello", 42})).success);
}

TEST("import: record schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "record"}, {"values", {{"kind", "int"}}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json::object({{"a", 1}, {"b", 2}})).success);
}

TEST("import: literal schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "literal"}, {"value", "hello"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json("hello")).success);
    ASSERT(!parse_document(imported, json("world")).success);
}

TEST("import: enum schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "enum"}, {"values", {"red", "green", "blue"}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json("red")).success);
    ASSERT(!parse_document(imported, json("yellow")).success);
}

TEST("import: ref schema with definitions") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {
                {"user", {{"kind", "ref"}, {"ref", "#/definitions/User"}}}
            }},
            {"required", {"user"}},
            {"unknownKeys", "reject"}
        }},
        {"definitions", {
            {"User", {
                {"kind", "object"},
                {"properties", {
                    {"name", {{"kind", "string"}}},
                    {"age", {{"kind", "int"}}}
                }},
                {"required", {"name", "age"}},
                {"unknownKeys", "reject"}
            }}
        }},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported,
        json::object({{"user", json::object({{"name", "Alice"}, {"age", 30}})}}));
    ASSERT(result.success);
}

TEST("import: recursive ref") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "ref"}, {"ref", "#/definitions/TreeNode"}}},
        {"definitions", {
            {"TreeNode", {
                {"kind", "object"},
                {"properties", {
                    {"value", {{"kind", "int"}}},
                    {"children", {
                        {"kind", "array"},
                        {"items", {{"kind", "ref"}, {"ref", "#/definitions/TreeNode"}}}
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
            {{"value", 3}, {"children", {
                {{"value", 4}, {"children", json::array()}}
            }}}
        }}
    };
    ASSERT(parse_document(imported, tree).success);
}

TEST("import: coercion config") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "int"}, {"coerce", "string->int"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json("42"));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json(42));
}

TEST("import: chained coercion config") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "string"}, {"coerce", {"trim", "lower"}}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json("  HELLO  "));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json("hello"));
}

TEST("import: default value") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {
                {"name", {{"kind", "string"}}},
                {"role", {{"kind", "string"}, {"default", "user"}}}
            }},
            {"required", {"name"}},
            {"unknownKeys", "reject"}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported, json::object({{"name", "Alice"}}));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value["role"], json("user"));
}

TEST("export: round-trip simple schema") {
    AnyValiDocument doc;
    doc.root = string_();
    auto exported = export_schema(doc);
    ASSERT(exported["root"]["kind"] == "string");
    ASSERT(exported["anyvaliVersion"] == "1.0");
}

TEST("export: round-trip object schema") {
    AnyValiDocument doc;
    auto obj = object();
    obj->prop("name", string_());
    obj->prop("age", int_());
    obj->required({"name", "age"});
    doc.root = obj;
    auto exported = export_schema(doc);
    ASSERT(exported["root"]["kind"] == "object");
    ASSERT(exported["root"]["properties"]["name"]["kind"] == "string");
}

TEST("export: round-trip with definitions") {
    AnyValiDocument doc;
    auto user_schema = object();
    user_schema->prop("name", string_());
    user_schema->required({"name"});
    doc.definitions["User"] = user_schema;
    doc.root = ref("#/definitions/User");
    auto exported = export_schema(doc);
    ASSERT(exported["definitions"]["User"]["kind"] == "object");
    ASSERT(exported["root"]["kind"] == "ref");
}

TEST("export: portable mode") {
    AnyValiDocument doc;
    doc.root = string_();
    auto exported = export_schema(doc, ExportMode::Portable);
    ASSERT(exported["extensions"].empty());
}

TEST("export: extended mode with extensions") {
    AnyValiDocument doc;
    doc.root = string_();
    doc.extensions = {{"cpp", {{"custom", true}}}};
    auto exported = export_schema(doc, ExportMode::Extended);
    ASSERT(exported["extensions"]["cpp"]["custom"] == true);
}

TEST("import: all primitive kinds") {
    std::vector<std::string> kinds = {
        "any", "unknown", "never", "null", "bool", "string",
        "number", "int", "float32", "float64",
        "int8", "int16", "int32", "int64",
        "uint8", "uint16", "uint32", "uint64"
    };
    for (const auto& k : kinds) {
        json doc = {
            {"anyvaliVersion", "1.0"},
            {"schemaVersion", "1"},
            {"root", {{"kind", k}}},
            {"definitions", json::object()},
            {"extensions", json::object()}
        };
        auto imported = import_schema(doc);
        ASSERT(kind_to_string(imported.root->kind()) == k);
    }
}

TEST("import: optional schema") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {
                {"name", {{"kind", "optional"}, {"schema", {{"kind", "string"}}}}}
            }},
            {"required", json::array()},
            {"unknownKeys", "reject"}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    ASSERT(parse_document(imported, json::object({{"name", "Alice"}})).success);
    ASSERT(parse_document(imported, json::object()).success);
}

TEST("import: strip unknown keys") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {{"name", {{"kind", "string"}}}}},
            {"required", {"name"}},
            {"unknownKeys", "strip"}
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported,
        json::object({{"name", "Alice"}, {"extra", "value"}, {"another", 42}}));
    ASSERT(result.success);
    ASSERT_JSON_EQ(result.value, json::object({{"name", "Alice"}}));
}

TEST("import: unsupported schema kind") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {{"kind", "bogus_xyz"}}},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    bool threw = false;
    try {
        import_schema(doc);
    } catch (...) {
        threw = true;
    }
    ASSERT(threw);
}

TEST("import: missing kind field") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", json::object()},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    bool threw = false;
    try {
        import_schema(doc);
    } catch (...) {
        threw = true;
    }
    ASSERT(threw);
}

TEST("import: null empty root") {
    json doc_empty = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    bool threw1 = false;
    try {
        import_schema(doc_empty);
    } catch (...) {
        threw1 = true;
    }
    ASSERT(threw1);

    json doc_null = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", nullptr},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    bool threw2 = false;
    try {
        import_schema(doc_null);
    } catch (...) {
        threw2 = true;
    }
    ASSERT(threw2);
}

TEST("import: default unknown keys is reject") {
    json doc = {
        {"anyvaliVersion", "1.0"},
        {"schemaVersion", "1"},
        {"root", {
            {"kind", "object"},
            {"properties", {{"name", {{"kind", "string"}}}}},
            {"required", {"name"}}
            // no unknownKeys field -> defaults to reject
        }},
        {"definitions", json::object()},
        {"extensions", json::object()}
    };
    auto imported = import_schema(doc);
    auto result = parse_document(imported,
        json::object({{"name", "Alice"}, {"extra", "value"}}));
    ASSERT(!result.success);
    ASSERT(result.issues[0].code == "unknown_key");
}

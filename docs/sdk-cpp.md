# AnyVali C++ SDK Reference

## Installation

### CMake with FetchContent

Add AnyVali to your `CMakeLists.txt`:

```cmake
include(FetchContent)

FetchContent_Declare(
  anyvali
  GIT_REPOSITORY https://github.com/BetterCorp/AnyVali-cpp.git
  GIT_TAG        v1.0.0
)
FetchContent_MakeAvailable(anyvali)

target_link_libraries(your_target PRIVATE anyvali::anyvali)
```

AnyVali C++ is header-only. It depends on [nlohmann/json](https://github.com/nlohmann/json) for JSON value representation, which is fetched automatically.

Requires C++17 or later.

## Quick Start

```cpp
#include <anyvali/anyvali.hpp>
#include <iostream>

using namespace anyvali;
using json = nlohmann::json;

int main() {
    // Define a schema
    auto user_schema = object()
        ->prop("id", int64())
        ->prop("name", string_()->min_length(1)->max_length(100))
        ->prop("email", string_()->format("email"))
        ->prop("age", optional_(int_()->min(0)->max(150)))
        ->prop("role", enum_({"admin", "user", "guest"}))
        ->required({"id", "name", "email"});

    // Parse input (throws ValidationError on failure)
    json input = {
        {"id", 42},
        {"name", "Alice"},
        {"email", "alice@example.com"},
        {"role", "admin"},
    };
    json user = user_schema->parse(input);

    // Safe parse (never throws)
    json bad_input = {{"id", "not-a-number"}, {"name", ""}};
    auto result = user_schema->safe_parse(bad_input);
    if (!result.success) {
        for (const auto& issue : result.issues) {
            std::cout << issue.path_string() << ": ["
                      << issue.code << "] " << issue.message << "\n";
        }
    }

    return 0;
}
```

## Schema Builders

All schema builders are free functions in the `anyvali` namespace. They return `std::shared_ptr` to schema objects that can be further refined with chained constraints.

Several builder names use trailing underscores to avoid conflicts with C++ keywords and standard library names: `string_()`, `bool_()`, `int_()`, `union_()`, `optional_()`.

### Special Types

```cpp
#include <anyvali/anyvali.hpp>
using namespace anyvali;

auto a = any();       // accepts any value
auto u = unknown();   // accepts any value, requires explicit handling
auto n = never();     // rejects all values
```

### Primitives

```cpp
auto s = string_();   // trailing underscore avoids std::string conflict
auto n = number();    // IEEE 754 float64
auto b = bool_();     // trailing underscore avoids bool keyword
auto z = null();
```

### Numeric Types

AnyVali provides sized integer and float types for cross-language safety:

```cpp
// Default integer (int64)
auto i = int_();      // trailing underscore avoids int keyword

// Signed integers
auto i8  = int8();    // -128 to 127
auto i16 = int16();   // -32,768 to 32,767
auto i32 = int32();   // -2^31 to 2^31-1
auto i64 = int64();   // -2^63 to 2^63-1

// Unsigned integers
auto u8  = uint8();   // 0 to 255
auto u16 = uint16();  // 0 to 65,535
auto u32 = uint32();  // 0 to 2^32-1
auto u64 = uint64();  // 0 to 2^64-1

// Floats
auto f32 = float32();
auto f64 = float64(); // same as number()
```

### Literal and Enum

```cpp
// Literal matches a single exact value
auto active = literal("active");
auto zero   = literal(0);
auto yes    = literal(true);

// Enum matches one of several allowed values
auto status = enum_({"pending", "active", "disabled"});
```

Note that `enum_` uses a trailing underscore because `enum` is a C++ keyword.

### Arrays and Tuples

```cpp
// Homogeneous array of strings
auto tags = array(string_());

// Array with length constraints
auto top_three = array(string_())
    ->min_items(1)
    ->max_items(3);

// Tuple with fixed positional types
auto coordinate = tuple({
    float64(),
    float64(),
});
```

### Objects

Objects use a builder pattern with `->prop()` and `->required()`:

```cpp
// Basic object with required and optional fields
auto user = object()
    ->prop("name", string_())
    ->prop("email", string_()->format("email"))
    ->prop("age", int_())
    ->required({"name", "email"});

// Unknown key handling
// Reject: error on unexpected keys (default)
auto strict = object()
    ->prop("name", string_())
    ->required({"name"})
    ->unknown_keys(UnknownKeyMode::Reject);

// Strip: silently remove unexpected keys
auto stripped = object()
    ->prop("name", string_())
    ->required({"name"})
    ->unknown_keys(UnknownKeyMode::Strip);

// Allow: keep unexpected keys as-is
auto passthrough = object()
    ->prop("name", string_())
    ->required({"name"})
    ->unknown_keys(UnknownKeyMode::Allow);
```

### Records

A record validates a JSON object where all values conform to a single schema:

```cpp
// Map of string keys to integer values
auto scores = record(int_());

json input = {{"alice", 95}, {"bob", 87}};
json result = scores->parse(input);
// => {"alice": 95, "bob": 87}
```

### Composition

```cpp
// Union: value must match at least one variant
auto string_or_int = union_({
    string_(),
    int_(),
});

// Intersection: value must match all schemas
auto named_and_aged = intersection({
    object()->prop("name", string_())->required({"name"}),
    object()->prop("age", int_())->required({"age"}),
});
```

Note that `union_` uses a trailing underscore because `union` is a C++ keyword.

### Modifiers

```cpp
// Optional: value may be absent (field can be omitted)
auto maybe_age = optional_(int_());

// Nullable: value may be null
auto nullable_name = nullable(string_());

// Combine both: field can be omitted, or present as null or string
auto opt_nullable_name = optional_(
    nullable(string_())
);
```

Note that `optional_` uses a trailing underscore to avoid conflict with `std::optional`.

### References

References allow recursive and reusable schema definitions:

```cpp
auto tree_node = object()
    ->prop("value", string_())
    ->prop("children", array(ref("#/definitions/TreeNode")))
    ->required({"value"});
```

## String Constraints

All string constraints return the schema pointer for chaining:

```cpp
auto schema = string_()
    ->min_length(1)          // minimum character count
    ->max_length(255)        // maximum character count
    ->pattern("^[A-Z]")     // regex pattern the value must match
    ->starts_with("Hello")  // value must start with this prefix
    ->ends_with("!")         // value must end with this suffix
    ->includes("world")     // value must contain this substring
    ->format("email");      // named format (email, uri, uuid, date, etc.)
```

Each constraint can be used independently:

```cpp
auto slug = string_()
    ->pattern("^[a-z0-9]+(?:-[a-z0-9]+)*$")
    ->min_length(1)
    ->max_length(100);

auto email = string_()->format("email");
auto url   = string_()->format("uri");
auto uuid  = string_()->format("uuid");
```

## Number Constraints

Number constraints apply to all numeric types (`number`, `int_`, `int8`--`int64`, `uint8`--`uint64`, `float32`, `float64`):

```cpp
auto price = float64()
    ->min(0)               // value >= 0
    ->max(999999.99)       // value <= 999999.99
    ->multiple_of(0.01);   // must be a multiple of 0.01

auto rating = int_()
    ->min(1)
    ->max(5);

auto temperature = float64()
    ->exclusive_min(-273.15)    // value > -273.15
    ->exclusive_max(1000.0);    // value < 1000.0
```

## Array Constraints

```cpp
auto tags = array(string_())
    ->min_items(1)      // at least 1 element
    ->max_items(10);    // at most 10 elements
```

## Parsing

### Throwing Parse

`parse()` returns the validated and coerced value as `nlohmann::json`. If validation fails, it throws `anyvali::ValidationError`:

```cpp
#include <anyvali/anyvali.hpp>
using namespace anyvali;

auto schema = string_()->min_length(1);

try {
    json value = schema->parse("hello");
    // value == "hello"
} catch (const ValidationError& e) {
    std::cerr << e.what() << "\n";
    for (const auto& issue : e.issues()) {
        std::cerr << "[" << issue.code << "] " << issue.message << "\n";
    }
}
```

### Safe Parse

`safe_parse()` never throws. It returns a `ParseResult` with structured success/failure information:

```cpp
auto schema = object()
    ->prop("name", string_()->min_length(1))
    ->prop("email", string_()->format("email"))
    ->required({"name", "email"});

json input = {{"name", ""}, {"email", "not-an-email"}};
auto result = schema->safe_parse(input);

if (result.success) {
    // result.value contains the parsed data
    json user = result.value;
} else {
    // result.issues contains all validation errors
    for (const auto& issue : result.issues) {
        std::cout << issue.path_string() << ": ["
                  << issue.code << "] " << issue.message << "\n";
        // "name: [too_small] String must have at least 1 character"
        // "email: [invalid_format] Invalid email format"
    }
}
```

### Typed Parse Helpers

The C++ SDK provides template helpers that parse and cast the result to a native C++ type:

```cpp
#include <anyvali/anyvali.hpp>
using namespace anyvali;

auto name_schema = string_()->min_length(1);

// Throwing version: returns T directly
std::string name = parse_as<std::string>(name_schema, "Alice");

auto age_schema = int_()->min(0)->max(150);
int64_t age = parse_as<int64_t>(age_schema, 25);

// Safe version: returns TypedParseResult<T>
auto result = safe_parse_as<std::string>(name_schema, "");
if (result.success) {
    std::string value = result.value;   // typed as std::string
} else {
    for (const auto& issue : result.issues) {
        std::cerr << issue.message << "\n";
    }
}
```

`TypedParseResult<T>` mirrors `ParseResult` but carries a typed `value`:

| Field | Type | Description |
|---|---|---|
| `success` | `bool` | Whether parsing succeeded |
| `value` | `T` | The parsed value (only meaningful when `success` is `true`) |
| `issues` | `std::vector<ValidationIssue>` | Validation issues (empty when `success` is `true`) |

### Issue Structure

Each `ValidationIssue` contains:

| Field | Type | Description |
|---|---|---|
| `code` | `std::string` | Machine-readable error code (e.g., `invalid_type`, `too_small`) |
| `message` | `std::string` | Human-readable error description |
| `path` | `std::vector<PathSegment>` | Path to the invalid value |
| `expected` | `std::string` | What was expected (e.g., `string`, `int64`) |
| `received` | `std::string` | What was received (e.g., `null`, `bool`) |

The `path_string()` method on `ValidationIssue` returns a dot-separated string representation of the path (e.g., `"users.0.email"`).

## Export and Import

### Exporting Schemas

Any schema can be exported to AnyVali's portable JSON format:

```cpp
auto schema = object()
    ->prop("id", int64())
    ->prop("name", string_()->min_length(1)->max_length(100))
    ->required({"id", "name"});

// Export to portable JSON
json doc = export_schema(schema, ExportMode::Portable);

std::cout << doc.dump(2) << "\n";
// {
//   "anyvaliVersion": "1.0",
//   "schemaVersion": "1",
//   "root": {
//     "kind": "object",
//     "properties": {
//       "id": { "kind": "int64" },
//       "name": { "kind": "string", "minLength": 1, "maxLength": 100 }
//     },
//     "required": ["id", "name"],
//     "unknownKeys": "reject"
//   },
//   "definitions": {},
//   "extensions": {}
// }
```

Export modes:

- `ExportMode::Portable` -- fails if the schema uses non-portable features. This is the safe default.
- `ExportMode::Extended` -- emits the core schema plus language-specific extension namespaces.

### Importing Schemas

Schemas can be imported from a JSON document produced by any AnyVali SDK:

```cpp
// Import from a portable JSON document
auto imported = import_schema(doc);

// The imported schema works exactly like a locally-built one
json user = imported->parse({{"id", 1}, {"name", "Bob"}});
```

This enables cross-language schema sharing. A schema authored in TypeScript, Python, or any other supported language can be exported, stored, and imported back into C++.

## Practical Examples

### REST API Request Validation

```cpp
#include <anyvali/anyvali.hpp>
#include <iostream>

using namespace anyvali;

auto create_order_schema() {
    return object()
        ->prop("customer_id", int64())
        ->prop("items", array(
            object()
                ->prop("product_id", int64())
                ->prop("quantity", uint16()->min(1))
                ->prop("unit_price", float64()->min(0)->multiple_of(0.01))
                ->required({"product_id", "quantity", "unit_price"})
        )->min_items(1))
        ->prop("notes", optional_(nullable(string_()->max_length(1000))))
        ->required({"customer_id", "items"});
}

void handle_create_order(const json& body) {
    static auto schema = create_order_schema();

    auto result = schema->safe_parse(body);
    if (!result.success) {
        json errors = json::array();
        for (const auto& issue : result.issues) {
            errors.push_back({
                {"field", issue.path_string()},
                {"message", issue.message},
            });
        }
        // respond with 400 and errors
        std::cout << json({{"errors", errors}}).dump(2) << "\n";
        return;
    }

    json order = result.value;
    // proceed with validated data
}
```

### Configuration File Validation

```cpp
#include <anyvali/anyvali.hpp>
#include <fstream>

using namespace anyvali;

auto config_schema() {
    return object()
        ->prop("database", object()
            ->prop("host", string_())
            ->prop("port", uint16()->min(1)->max(65535))
            ->prop("name", string_()->min_length(1))
            ->prop("user", string_())
            ->prop("password", string_())
            ->required({"host", "port", "name", "user", "password"})
        )
        ->prop("cache", object()
            ->prop("driver", enum_({"redis", "memcached", "file"}))
            ->prop("ttl", int_()->min(0))
            ->required({"driver", "ttl"})
        )
        ->prop("debug", optional_(bool_()))
        ->required({"database", "cache"});
}

int main() {
    std::ifstream file("config.json");
    json raw;
    file >> raw;

    static auto schema = config_schema();

    try {
        json config = schema->parse(raw);
        std::string db_host = config["database"]["host"];
        uint16_t db_port = config["database"]["port"];
        // use validated config
    } catch (const ValidationError& e) {
        std::cerr << "Invalid configuration:\n";
        for (const auto& issue : e.issues()) {
            std::cerr << "  " << issue.path_string()
                      << ": " << issue.message << "\n";
        }
        return 1;
    }

    return 0;
}
```

### Message Protocol Validation

```cpp
#include <anyvali/anyvali.hpp>

using namespace anyvali;

// Define message types for a protocol
auto text_message = object()
    ->prop("type", literal("text"))
    ->prop("content", string_()->min_length(1)->max_length(4096))
    ->prop("sender_id", int64())
    ->required({"type", "content", "sender_id"});

auto image_message = object()
    ->prop("type", literal("image"))
    ->prop("url", string_()->format("uri"))
    ->prop("width", uint32()->min(1))
    ->prop("height", uint32()->min(1))
    ->prop("sender_id", int64())
    ->required({"type", "url", "width", "height", "sender_id"});

auto message_schema = union_({text_message, image_message});

void process_message(const json& raw) {
    auto result = message_schema->safe_parse(raw);
    if (!result.success) {
        // handle invalid message
        return;
    }

    json msg = result.value;
    std::string type = msg["type"];
    if (type == "text") {
        std::string content = msg["content"];
        // handle text message
    } else if (type == "image") {
        std::string url = msg["url"];
        // handle image message
    }
}
```

### Cross-Language Schema Sharing

```cpp
#include <anyvali/anyvali.hpp>
#include <fstream>

using namespace anyvali;

int main() {
    // Import a schema produced by another SDK (e.g., Python, TypeScript)
    std::ifstream file("schemas/user.anyvali.json");
    json doc;
    file >> doc;
    auto user_schema = import_schema(doc);

    // Validate data against the imported schema
    json input = {{"id", 1}, {"name", "Bob"}, {"email", "bob@test.com"}};
    json user = user_schema->parse(input);

    // Build a local schema and export it for other services
    auto audit_event = object()
        ->prop("action", enum_({"create", "update", "delete"}))
        ->prop("actor_id", int64())
        ->prop("timestamp", string_()->format("date-time"))
        ->prop("payload", record(any()))
        ->required({"action", "actor_id", "timestamp"});

    json exported = export_schema(audit_event, ExportMode::Portable);

    std::ofstream out("schemas/audit_event.anyvali.json");
    out << exported.dump(2);

    return 0;
}
```

### Typed Parsing with Structs

```cpp
#include <anyvali/anyvali.hpp>

using namespace anyvali;

struct User {
    int64_t id;
    std::string name;
    std::string email;
};

// nlohmann::json conversion (standard nlohmann pattern)
void from_json(const json& j, User& u) {
    j.at("id").get_to(u.id);
    j.at("name").get_to(u.name);
    j.at("email").get_to(u.email);
}

int main() {
    auto schema = object()
        ->prop("id", int64())
        ->prop("name", string_()->min_length(1))
        ->prop("email", string_()->format("email"))
        ->required({"id", "name", "email"});

    json input = {{"id", 42}, {"name", "Alice"}, {"email", "alice@example.com"}};

    // parse_as<T> validates then converts to your struct
    User user = parse_as<User>(schema, input);
    // user.id == 42, user.name == "Alice", user.email == "alice@example.com"

    // Safe version
    auto result = safe_parse_as<User>(schema, input);
    if (result.success) {
        User u = result.value;
        // use typed struct
    }

    return 0;
}
```

## API Reference

### Builder Functions (anyvali namespace)

| Function | Returns | Description |
|---|---|---|
| `string_()` | `shared_ptr<StringSchema>` | String schema |
| `number()` | `shared_ptr<NumberSchema>` | IEEE 754 float64 schema |
| `int_()` | `shared_ptr<IntSchema>` | Signed int64 schema |
| `int8()` | `shared_ptr<IntSchema>` | Signed 8-bit integer |
| `int16()` | `shared_ptr<IntSchema>` | Signed 16-bit integer |
| `int32()` | `shared_ptr<IntSchema>` | Signed 32-bit integer |
| `int64()` | `shared_ptr<IntSchema>` | Signed 64-bit integer |
| `uint8()` | `shared_ptr<IntSchema>` | Unsigned 8-bit integer |
| `uint16()` | `shared_ptr<IntSchema>` | Unsigned 16-bit integer |
| `uint32()` | `shared_ptr<IntSchema>` | Unsigned 32-bit integer |
| `uint64()` | `shared_ptr<IntSchema>` | Unsigned 64-bit integer |
| `float32()` | `shared_ptr<NumberSchema>` | 32-bit float |
| `float64()` | `shared_ptr<NumberSchema>` | 64-bit float (same as `number()`) |
| `bool_()` | `shared_ptr<BoolSchema>` | Boolean schema |
| `null()` | `shared_ptr<NullSchema>` | Null schema |
| `any()` | `shared_ptr<AnySchema>` | Accepts any value |
| `unknown()` | `shared_ptr<UnknownSchema>` | Accepts any value (explicit handling) |
| `never()` | `shared_ptr<NeverSchema>` | Rejects all values |
| `literal(value)` | `shared_ptr<LiteralSchema>` | Matches a single exact value |
| `enum_(values)` | `shared_ptr<EnumSchema>` | Matches one of the given values |
| `array(items)` | `shared_ptr<ArraySchema>` | Homogeneous array |
| `tuple(elements)` | `shared_ptr<TupleSchema>` | Fixed-position typed array |
| `object()` | `shared_ptr<ObjectSchema>` | Object schema (use builder to add properties) |
| `record(values)` | `shared_ptr<RecordSchema>` | String-keyed map with uniform value type |
| `union_(variants)` | `shared_ptr<UnionSchema>` | Value must match at least one variant |
| `intersection(schemas)` | `shared_ptr<IntersectionSchema>` | Value must match all schemas |
| `optional_(inner)` | `shared_ptr<OptionalSchema>` | Value may be absent |
| `nullable(inner)` | `shared_ptr<NullableSchema>` | Value may be null |
| `ref(ref_path)` | `shared_ptr<RefSchema>` | Reference to a definition |

### Free Functions

| Function | Returns | Description |
|---|---|---|
| `export_schema(schema, mode)` | `nlohmann::json` | Export schema to portable JSON document |
| `import_schema(doc)` | `shared_ptr<Schema>` | Import schema from portable JSON document |
| `parse_as<T>(schema, input)` | `T` | Parse, validate, and convert to type `T`; throws on failure |
| `safe_parse_as<T>(schema, input)` | `TypedParseResult<T>` | Parse, validate, and convert to `T`; never throws |

### Schema Instance Methods

| Method | Returns | Description |
|---|---|---|
| `schema->parse(input)` | `nlohmann::json` | Parse and validate; throws `ValidationError` on failure |
| `schema->safe_parse(input)` | `ParseResult` | Parse and validate; never throws |

### ObjectSchema Builder Methods

| Method | Returns | Description |
|---|---|---|
| `->prop(name, schema)` | `shared_ptr<ObjectSchema>` | Add a property |
| `->required(names)` | `shared_ptr<ObjectSchema>` | Set required property names |
| `->unknown_keys(mode)` | `shared_ptr<ObjectSchema>` | Set unknown key handling mode |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->min_length(n)` | `size_t` | Minimum character count |
| `->max_length(n)` | `size_t` | Maximum character count |
| `->pattern(p)` | `std::string` | Regex pattern the value must match |
| `->starts_with(s)` | `std::string` | Required prefix |
| `->ends_with(s)` | `std::string` | Required suffix |
| `->includes(s)` | `std::string` | Required substring |
| `->format(f)` | `std::string` | Named format (`email`, `uri`, `uuid`, `date`, `date-time`, etc.) |

### NumberSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->min(n)` | `double` | Minimum value (inclusive) |
| `->max(n)` | `double` | Maximum value (inclusive) |
| `->exclusive_min(n)` | `double` | Minimum value (exclusive) |
| `->exclusive_max(n)` | `double` | Maximum value (exclusive) |
| `->multiple_of(n)` | `double` | Value must be a multiple of this |

### ArraySchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->min_items(n)` | `size_t` | Minimum number of elements |
| `->max_items(n)` | `size_t` | Maximum number of elements |

### UnknownKeyMode (Enum)

| Value | Description |
|---|---|
| `UnknownKeyMode::Reject` | Error on unexpected keys (default) |
| `UnknownKeyMode::Strip` | Silently remove unexpected keys |
| `UnknownKeyMode::Allow` | Keep unexpected keys as-is |

### ExportMode (Enum)

| Value | Description |
|---|---|
| `ExportMode::Portable` | Fail on non-portable features (safe default) |
| `ExportMode::Extended` | Include language-specific extensions |

### ParseResult

| Field | Type | Description |
|---|---|---|
| `success` | `bool` | Whether parsing succeeded |
| `value` | `nlohmann::json` | The parsed value (only meaningful when `success` is `true`) |
| `issues` | `std::vector<ValidationIssue>` | Validation issues (empty when `success` is `true`) |

### TypedParseResult\<T\>

| Field | Type | Description |
|---|---|---|
| `success` | `bool` | Whether parsing succeeded |
| `value` | `T` | The parsed value (only meaningful when `success` is `true`) |
| `issues` | `std::vector<ValidationIssue>` | Validation issues (empty when `success` is `true`) |

### ValidationIssue

| Field | Type | Description |
|---|---|---|
| `code` | `std::string` | Machine-readable error code (e.g., `invalid_type`, `too_small`) |
| `message` | `std::string` | Human-readable error description |
| `path` | `std::vector<PathSegment>` | Path to the invalid value |
| `expected` | `std::string` | What was expected (e.g., `string`, `int64`) |
| `received` | `std::string` | What was received (e.g., `null`, `bool`) |
| `path_string()` | `std::string` | Dot-separated path (e.g., `"users.0.email"`) |

### ValidationError

Inherits from `std::runtime_error`. Thrown by `parse()` and `parse_as<T>()` on validation failure.

| Method | Returns | Description |
|---|---|---|
| `what()` | `const char*` | Summary error message |
| `issues()` | `const std::vector<ValidationIssue>&` | List of validation issues |

# AnyVali Ruby SDK Reference

## Installation

Install via RubyGems:

```bash
gem install anyvali
```

Or add to your Gemfile:

```ruby
gem 'anyvali'
```

Then run:

```bash
bundle install
```

Requires Ruby 3.1 or later.

## Quick Start

```ruby
require 'anyvali'

# Define a schema
user_schema = AnyVali.object(
  properties: {
    id:    AnyVali.int64,
    name:  AnyVali.string.min_length(1).max_length(100),
    email: AnyVali.string.format("email"),
    age:   AnyVali.optional(AnyVali.int_.min(0).max(150)),
    role:  AnyVali.enum_("admin", "user", "guest"),
  },
  required: [:id, :name, :email]
)

# Parse input (raises ValidationError on failure)
user = user_schema.parse({
  id:    42,
  name:  "Alice",
  email: "alice@example.com",
  role:  "admin",
})
# => { id: 42, name: "Alice", email: "alice@example.com", role: "admin" }

# Safe parse (never raises)
result = user_schema.safe_parse({ id: "not-a-number", name: "" })
if result.failure?
  result.issues.each do |issue|
    puts "#{issue.path.join('.')}: [#{issue.code}] #{issue.message}"
  end
end
```

## Schema Builders

All schemas are created through module methods on `AnyVali`. Each method returns a schema instance that can be refined with chained constraints.

### Special Types

```ruby
any_schema     = AnyVali.any       # accepts any value
unknown_schema = AnyVali.unknown   # accepts any value, requires explicit handling
never_schema   = AnyVali.never     # rejects all values
```

### Primitives

```ruby
str  = AnyVali.string
num  = AnyVali.number    # IEEE 754 float64
bool = AnyVali.bool
null = AnyVali.null
```

### Numeric Types

AnyVali provides sized integer and float types for cross-language safety:

```ruby
# Default integer (int64)
int = AnyVali.int_

# Signed integers
i8  = AnyVali.int8     # -128 to 127
i16 = AnyVali.int16    # -32,768 to 32,767
i32 = AnyVali.int32    # -2^31 to 2^31-1
i64 = AnyVali.int64    # -2^63 to 2^63-1

# Unsigned integers
u8  = AnyVali.uint8    # 0 to 255
u16 = AnyVali.uint16   # 0 to 65,535
u32 = AnyVali.uint32   # 0 to 2^32-1
u64 = AnyVali.uint64   # 0 to 2^64-1

# Floats
f32 = AnyVali.float32
f64 = AnyVali.float64  # same as number
```

Note that `int_` uses a trailing underscore to avoid conflicting with Ruby's built-in `Integer` conversion method.

### Literal and Enum

```ruby
# Literal matches a single exact value
active = AnyVali.literal("active")
zero   = AnyVali.literal(0)
yes    = AnyVali.literal(true)

# Enum matches one of several allowed values
status = AnyVali.enum_("pending", "active", "disabled")
```

Note that `enum_` uses a trailing underscore because `enum` is not a Ruby keyword but avoids potential naming conflicts.

### Arrays and Tuples

```ruby
# Homogeneous array of strings
tags = AnyVali.array(AnyVali.string)

# Array with length constraints
top_three = AnyVali.array(AnyVali.string)
  .min_items(1)
  .max_items(3)

# Tuple with fixed positional types
coordinate = AnyVali.tuple(
  AnyVali.float64,
  AnyVali.float64,
)
```

### Objects

```ruby
# Basic object with required and optional fields
user = AnyVali.object(
  properties: {
    name:  AnyVali.string,
    email: AnyVali.string.format("email"),
    age:   AnyVali.int_,
  },
  required: [:name, :email]
)

# Unknown key handling
strict = AnyVali.object(
  properties: { name: AnyVali.string },
  required: [:name],
  unknown_keys: "reject"    # error on unexpected keys (default)
)

stripped = AnyVali.object(
  properties: { name: AnyVali.string },
  required: [:name],
  unknown_keys: "strip"     # silently remove unexpected keys
)

passthrough = AnyVali.object(
  properties: { name: AnyVali.string },
  required: [:name],
  unknown_keys: "allow"     # keep unexpected keys as-is
)
```

### Records

A record validates a hash where all values conform to a single schema:

```ruby
# Map of string keys to integer values
scores = AnyVali.record(AnyVali.int_)

scores.parse({ "alice" => 95, "bob" => 87 })
# => { "alice" => 95, "bob" => 87 }
```

### Composition

```ruby
# Union: value must match at least one variant
string_or_int = AnyVali.union(
  AnyVali.string,
  AnyVali.int_,
)

# Intersection: value must match all schemas
named_and_aged = AnyVali.intersection(
  AnyVali.object(properties: { name: AnyVali.string }, required: [:name]),
  AnyVali.object(properties: { age: AnyVali.int_ }, required: [:age]),
)
```

### Modifiers

```ruby
# Optional: value may be absent (field can be omitted)
maybe_age = AnyVali.optional(AnyVali.int_)

# Nullable: value may be nil
nullable_name = AnyVali.nullable(AnyVali.string)

# Combine both: field can be omitted, or present as nil or string
optional_nullable_name = AnyVali.optional(
  AnyVali.nullable(AnyVali.string)
)
```

### References

References allow recursive and reusable schema definitions:

```ruby
tree_node = AnyVali.object(
  properties: {
    value:    AnyVali.string,
    children: AnyVali.array(AnyVali.ref("#/definitions/TreeNode")),
  },
  required: [:value]
)
```

## String Constraints

All string constraints return the schema instance for chaining:

```ruby
schema = AnyVali.string
  .min_length(1)          # minimum character count
  .max_length(255)        # maximum character count
  .pattern(/^[A-Z]/)      # regex pattern the value must match
  .starts_with("Hello")   # value must start with this prefix
  .ends_with("!")          # value must end with this suffix
  .includes("world")      # value must contain this substring
  .format("email")        # named format (email, uri, uuid, date, etc.)
```

Each constraint can be used independently:

```ruby
slug  = AnyVali.string
  .pattern(/\A[a-z0-9]+(?:-[a-z0-9]+)*\z/)
  .min_length(1)
  .max_length(100)

email = AnyVali.string.format("email")
url   = AnyVali.string.format("uri")
uuid  = AnyVali.string.format("uuid")
```

## Number Constraints

Number constraints apply to all numeric types (`number`, `int_`, `int8`--`int64`, `uint8`--`uint64`, `float32`, `float64`):

```ruby
price = AnyVali.float64
  .min(0)                # value >= 0
  .max(999_999.99)       # value <= 999_999.99
  .multiple_of(0.01)     # must be a multiple of 0.01

rating = AnyVali.int_
  .min(1)
  .max(5)

temperature = AnyVali.float64
  .exclusive_min(-273.15)    # value > -273.15
  .exclusive_max(1000.0)     # value < 1000.0
```

## Array Constraints

```ruby
tags = AnyVali.array(AnyVali.string)
  .min_items(1)      # at least 1 element
  .max_items(10)     # at most 10 elements
```

## Coercion

Coercion transforms the input value before validation. It runs only when the value is present. Call `.coerce(config)` on any schema to enable coercion.

### Available Coercions

#### String to Integer

```ruby
age = AnyVali.int_.coerce(from: "string")
age.parse("42")   # => 42 (string coerced to integer)
age.parse(42)     # => 42 (already an integer, no coercion needed)
```

#### String to Number

```ruby
price = AnyVali.number.coerce(from: "string")
price.parse("3.14")  # => 3.14
```

#### String to Boolean

```ruby
flag = AnyVali.bool.coerce(from: "string")
flag.parse("true")   # => true
flag.parse("false")  # => false
flag.parse("1")      # => true
flag.parse("0")      # => false
```

#### Trim Whitespace

```ruby
trimmed = AnyVali.string.coerce(trim: true)
trimmed.parse("  hello  ")  # => "hello"
```

#### Lowercase / Uppercase

```ruby
lower = AnyVali.string.coerce(lower: true)
lower.parse("HELLO")  # => "hello"

upper = AnyVali.string.coerce(upper: true)
upper.parse("hello")  # => "HELLO"
```

Transformations can be combined:

```ruby
normalized = AnyVali.string.coerce(trim: true, lower: true)
normalized.parse("  Hello World  ")  # => "hello world"
```

## Defaults

Defaults fill in missing (absent) values. They run after coercion and before validation. Call `.default(value)` on any schema.

The default only applies when the value is absent -- if a value is present, it is validated normally. Defaults must be static values (for portability across SDKs).

```ruby
role = AnyVali.string.default("user")
role.parse(nil)      # => "user" (absent value filled)
role.parse("admin")  # => "admin"

tags = AnyVali.array(AnyVali.string).default([])
```

Defaults work with optional fields in objects:

```ruby
config = AnyVali.object(
  properties: {
    theme:    AnyVali.optional(AnyVali.string.default("light")),
    language: AnyVali.optional(AnyVali.string.default("en")),
  },
  required: []
)

config.parse({})
# => { "theme" => "light", "language" => "en" }
```

If the default value itself fails validation, a `default_invalid` issue is produced.

## Parsing

### Throwing Parse

`parse` returns the validated and coerced value. If validation fails, it raises `AnyVali::ValidationError`:

```ruby
schema = AnyVali.string.min_length(1)

begin
  value = schema.parse("hello")
  # => "hello"
rescue AnyVali::ValidationError => e
  puts e.message
  e.issues.each do |issue|
    puts "[#{issue.code}] #{issue.message}"
  end
end
```

### Safe Parse

`safe_parse` never raises. It returns a `ParseResult` with structured success/failure information:

```ruby
schema = AnyVali.object(
  properties: {
    name:  AnyVali.string.min_length(1),
    email: AnyVali.string.format("email"),
  },
  required: [:name, :email]
)

result = schema.safe_parse({ name: "", email: "not-an-email" })

if result.success?
  # result.value contains the parsed data
  user = result.value
else
  # result.issues contains all validation errors
  result.issues.each do |issue|
    puts "#{issue.path.join('.')}: [#{issue.code}] #{issue.message}"
    # "name: [too_small] String must have at least 1 character"
    # "email: [invalid_format] Invalid email format"
  end
end
```

### Issue Structure

Each issue in a parse result contains:

| Field | Type | Description |
|---|---|---|
| `code` | `String` | Machine-readable error code (e.g., `invalid_type`, `too_small`) |
| `message` | `String` | Human-readable error description |
| `path` | `Array` | Path to the invalid value (e.g., `["users", 0, "email"]`) |
| `expected` | `String` | What was expected (e.g., `string`, `int64`) |
| `received` | `String` | What was received (e.g., `nil`, `true`) |

## Export and Import

### Exporting Schemas

Any schema can be exported to AnyVali's portable JSON format:

```ruby
schema = AnyVali.object(
  properties: {
    id:   AnyVali.int64,
    name: AnyVali.string.min_length(1).max_length(100),
  },
  required: [:id, :name]
)

# Export to portable JSON (returns a Hash)
doc = AnyVali.export(schema, mode: :portable)

require 'json'
puts JSON.pretty_generate(doc)
# {
#   "anyvaliVersion": "1.0",
#   "schemaVersion": "1",
#   "root": {
#     "kind": "object",
#     "properties": {
#       "id": { "kind": "int64" },
#       "name": { "kind": "string", "minLength": 1, "maxLength": 100 }
#     },
#     "required": ["id", "name"],
#     "unknownKeys": "reject"
#   },
#   "definitions": {},
#   "extensions": {}
# }
```

Export modes:

- `:portable` -- fails if the schema uses non-portable features. This is the safe default.
- `:extended` -- emits the core schema plus language-specific extension namespaces.

### Importing Schemas

Schemas can be imported from a JSON document produced by any AnyVali SDK:

```ruby
# Import from a portable JSON document (Hash or JSON string)
imported = AnyVali.import_schema(doc)

# The imported schema works exactly like a locally-built one
user = imported.parse({ id: 1, name: "Bob" })
```

This enables cross-language schema sharing. A schema authored in TypeScript, Go, or any other supported language can be exported, stored, and imported back into Ruby.

## Type Signatures with RBS

The Ruby SDK ships RBS type signature files in `sig/anyvali.rbs` for use with Steep, Sorbet, and other Ruby type checkers.

The signatures provide type information for schema builders and parse results:

```ruby
# With RBS, type checkers understand that:
#   AnyVali.string  => StringSchema
#   AnyVali.int_    => IntSchema
#   schema.parse(x) => validated type
#   result.value    => validated type when result.success? is true

schema = AnyVali.string.min_length(1)
value = schema.parse("hello")   # Steep/Sorbet knows this is String
```

To use the type signatures with Steep, add to your `Steepfile`:

```ruby
target :app do
  signature "sig"
  check "lib"
  library "anyvali"
end
```

## Practical Examples

### Rails Controller Validation

```ruby
class UsersController < ApplicationController
  CREATE_SCHEMA = AnyVali.object(
    properties: {
      name:     AnyVali.string.min_length(1).max_length(200),
      email:    AnyVali.string.format("email"),
      role:     AnyVali.enum_("admin", "member", "viewer"),
      settings: AnyVali.optional(AnyVali.object(
        properties: {
          notifications: AnyVali.bool,
          theme:         AnyVali.enum_("light", "dark"),
        },
        required: [:notifications, :theme]
      )),
    },
    required: [:name, :email, :role]
  )

  def create
    result = CREATE_SCHEMA.safe_parse(params.to_unsafe_h[:user])
    if result.failure?
      render json: {
        errors: result.issues.map { |i|
          { field: i.path.join("."), message: i.message }
        }
      }, status: :unprocessable_entity
      return
    end

    user = User.create!(result.value)
    render json: user, status: :created
  end
end
```

### API Client Response Validation

```ruby
require 'net/http'
require 'json'
require 'anyvali'

PRODUCT_SCHEMA = AnyVali.object(
  properties: {
    id:          AnyVali.int64,
    name:        AnyVali.string.min_length(1),
    price_cents: AnyVali.uint32.min(0),
    currency:    AnyVali.enum_("USD", "EUR", "GBP"),
    tags:        AnyVali.array(AnyVali.string).max_items(20),
    metadata:    AnyVali.optional(AnyVali.record(AnyVali.any)),
  },
  required: [:id, :name, :price_cents, :currency, :tags]
)

PRODUCT_LIST_SCHEMA = AnyVali.array(PRODUCT_SCHEMA).min_items(0)

uri = URI("https://api.example.com/products")
response = Net::HTTP.get(uri)
data = JSON.parse(response)

products = PRODUCT_LIST_SCHEMA.parse(data)
products.each do |product|
  puts "#{product['name']}: #{product['price_cents']} #{product['currency']}"
end
```

### Configuration File Validation

```ruby
require 'yaml'
require 'anyvali'

CONFIG_SCHEMA = AnyVali.object(
  properties: {
    database: AnyVali.object(
      properties: {
        host:     AnyVali.string,
        port:     AnyVali.uint16.min(1).max(65535),
        name:     AnyVali.string.min_length(1),
        user:     AnyVali.string,
        password: AnyVali.string,
        pool:     AnyVali.optional(AnyVali.uint8.min(1).max(100)),
      },
      required: [:host, :port, :name, :user, :password]
    ),
    redis: AnyVali.optional(AnyVali.object(
      properties: {
        url: AnyVali.string.format("uri"),
        ttl: AnyVali.int_.min(0),
      },
      required: [:url, :ttl]
    )),
    log_level: AnyVali.enum_("debug", "info", "warn", "error", "fatal"),
  },
  required: [:database, :log_level]
)

raw = YAML.safe_load_file("config/app.yml")
config = CONFIG_SCHEMA.parse(raw)
```

### Cross-Language Schema Sharing

```ruby
require 'net/http'
require 'json'
require 'anyvali'

# Fetch a schema published by a Go service
uri = URI("https://api.example.com/schemas/event")
doc = JSON.parse(Net::HTTP.get(uri))
event_schema = AnyVali.import_schema(doc)

# Validate incoming webhook data
event_schema.parse(webhook_payload)

# Build a local schema and share it
audit_log = AnyVali.object(
  properties: {
    action:    AnyVali.enum_("create", "update", "delete"),
    actor_id:  AnyVali.int64,
    timestamp: AnyVali.string.format("date-time"),
    changes:   AnyVali.record(AnyVali.any),
  },
  required: [:action, :actor_id, :timestamp]
)

exported = AnyVali.export(audit_log, mode: :portable)
# Store in a schema registry or send to other services
```

## Common Patterns

### Validating Environment Variables

Use `unknown_keys: "strip"` when parsing hashes that contain many extra keys you don't care about, like `ENV`:

```ruby
env_schema = AnyVali.object(
  properties: { "DATABASE_URL" => AnyVali.string },
  required: ["DATABASE_URL"],
  unknown_keys: "strip"
)
```

Without `"strip"`, parse would raise with `unknown_key` issues for every other variable in the environment (PATH, HOME, etc.) because the default mode is `"reject"`.

| Mode | What happens with extra keys |
|---|---|
| `"reject"` (default) | Parse fails with `unknown_key` issues |
| `"strip"` | Extra keys silently removed from output |
| `"allow"` | Extra keys passed through to output |

### Eagerly Evaluated vs Lazy Defaults

`.default()` accepts any value of the correct type. Expressions like `Dir.pwd` are evaluated immediately when the schema is created and stored as a static value -- this works fine. What AnyVali does not support is lazy proc/lambda defaults that re-evaluate on each parse call. If you need a fresh value on every parse, apply it after:

```ruby
config_schema = AnyVali.object(
  properties: {
    "profile" => AnyVali.optional(AnyVali.string.default("default")),
    "app_dir" => AnyVali.optional(AnyVali.string),
  },
  required: ["profile"],
  unknown_keys: "strip"
)

config = config_schema.parse(data)
config["app_dir"] ||= Dir.pwd
```

This keeps the schema fully portable -- the same JSON document can be imported in Go, Python, or any other SDK without relying on language-specific function calls.

## API Reference

### AnyVali (Module Methods)

| Method | Returns | Description |
|---|---|---|
| `AnyVali.string` | `StringSchema` | String schema |
| `AnyVali.number` | `NumberSchema` | IEEE 754 float64 schema |
| `AnyVali.int_` | `IntSchema` | Signed int64 schema |
| `AnyVali.int8` | `IntSchema` | Signed 8-bit integer |
| `AnyVali.int16` | `IntSchema` | Signed 16-bit integer |
| `AnyVali.int32` | `IntSchema` | Signed 32-bit integer |
| `AnyVali.int64` | `IntSchema` | Signed 64-bit integer |
| `AnyVali.uint8` | `IntSchema` | Unsigned 8-bit integer |
| `AnyVali.uint16` | `IntSchema` | Unsigned 16-bit integer |
| `AnyVali.uint32` | `IntSchema` | Unsigned 32-bit integer |
| `AnyVali.uint64` | `IntSchema` | Unsigned 64-bit integer |
| `AnyVali.float32` | `NumberSchema` | 32-bit float |
| `AnyVali.float64` | `NumberSchema` | 64-bit float (same as `number`) |
| `AnyVali.bool` | `BoolSchema` | Boolean schema |
| `AnyVali.null` | `NullSchema` | Null schema |
| `AnyVali.any` | `AnySchema` | Accepts any value |
| `AnyVali.unknown` | `UnknownSchema` | Accepts any value (explicit handling) |
| `AnyVali.never` | `NeverSchema` | Rejects all values |
| `AnyVali.literal(value)` | `LiteralSchema` | Matches a single exact value |
| `AnyVali.enum_(*values)` | `EnumSchema` | Matches one of the given values |
| `AnyVali.array(items)` | `ArraySchema` | Homogeneous array |
| `AnyVali.tuple(*elements)` | `TupleSchema` | Fixed-position typed array |
| `AnyVali.object(properties:, required:, unknown_keys:)` | `ObjectSchema` | Object/hash schema |
| `AnyVali.record(values)` | `RecordSchema` | String-keyed hash with uniform value type |
| `AnyVali.union(*variants)` | `UnionSchema` | Value must match at least one variant |
| `AnyVali.intersection(*schemas)` | `IntersectionSchema` | Value must match all schemas |
| `AnyVali.optional(schema)` | `OptionalSchema` | Value may be absent |
| `AnyVali.nullable(schema)` | `NullableSchema` | Value may be nil |
| `AnyVali.ref(ref_path)` | `RefSchema` | Reference to a definition |
| `AnyVali.export(schema, mode:)` | `Hash` | Export schema to portable JSON |
| `AnyVali.import_schema(doc)` | `Schema` | Import schema from portable JSON |

### Schema Instance Methods

| Method | Returns | Description |
|---|---|---|
| `schema.parse(input)` | `Object` | Parse and validate; raises `ValidationError` on failure |
| `schema.safe_parse(input)` | `ParseResult` | Parse and validate; never raises |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min_length(n)` | `Integer` | Minimum character count |
| `.max_length(n)` | `Integer` | Maximum character count |
| `.pattern(p)` | `Regexp` | Regex pattern the value must match |
| `.starts_with(s)` | `String` | Required prefix |
| `.ends_with(s)` | `String` | Required suffix |
| `.includes(s)` | `String` | Required substring |
| `.format(f)` | `String` | Named format (`email`, `uri`, `uuid`, `date`, `date-time`, etc.) |

### NumberSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `Numeric` | Minimum value (inclusive) |
| `.max(n)` | `Numeric` | Maximum value (inclusive) |
| `.exclusive_min(n)` | `Numeric` | Minimum value (exclusive) |
| `.exclusive_max(n)` | `Numeric` | Maximum value (exclusive) |
| `.multiple_of(n)` | `Numeric` | Value must be a multiple of this |

### ArraySchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min_items(n)` | `Integer` | Minimum number of elements |
| `.max_items(n)` | `Integer` | Maximum number of elements |

### Object Unknown Keys

| Value | Description |
|---|---|
| `"reject"` | Error on unexpected keys (default) |
| `"strip"` | Silently remove unexpected keys |
| `"allow"` | Keep unexpected keys as-is |

### ParseResult

| Method / Property | Type | Description |
|---|---|---|
| `result.success?` | `Boolean` | Whether parsing succeeded |
| `result.failure?` | `Boolean` | Whether parsing failed |
| `result.value` | `Object` | The parsed value (only meaningful when `success?` is `true`) |
| `result.issues` | `Array` | List of validation issues (empty when `success?` is `true`) |

### ValidationError

Inherits from `StandardError`. Raised by `parse` on validation failure.

| Method / Property | Type | Description |
|---|---|---|
| `e.issues` | `Array` | List of validation issues |
| `e.message` | `String` | Summary error message |

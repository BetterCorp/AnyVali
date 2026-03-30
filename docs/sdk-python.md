# Python SDK Reference

## Installation

```bash
pip install anyvali
```

AnyVali requires Python 3.10 or later.

### Import Convention

The recommended import style uses a short namespace alias:

```python
import anyvali as v
```

All builder functions, types, and utilities are available directly from the `anyvali` module.

## Quick Start

```python
import anyvali as v

# 1. Define a schema
user_schema = v.object_({
    "name": v.string().min_length(1).max_length(100),
    "email": v.string().format("email"),
    "age": v.int_().min(0).max(150),
    "role": v.enum_(["admin", "user", "guest"]).default("user"),
})

# 2. Parse input -- raises ValidationError on failure
user = user_schema.parse({
    "name": "Alice",
    "email": "alice@example.com",
    "age": 30,
})
# => {"name": "Alice", "email": "alice@example.com", "age": 30, "role": "user"}

# 3. Safe parse -- returns a ParseResult
result = user_schema.safe_parse({"name": "", "email": "bad"})
if not result.success:
    for issue in result.issues:
        path = ".".join(str(p) for p in issue.path)
        print(f"{path}: [{issue.code}] {issue.message}")
```

## Type Inference

The Python SDK uses `Generic[T]` to provide type annotations on schemas and parse results. Type checkers like mypy and pyright can infer the output type from schema definitions.

### Basic Usage

```python
import anyvali as v

name_schema: v.BaseSchema[str] = v.string()
age_schema: v.BaseSchema[int] = v.int_()
active_schema: v.BaseSchema[bool] = v.bool_()
```

Each builder function returns a schema class that extends `BaseSchema[T]` with the appropriate type parameter:

- `v.string()` returns `StringSchema` which extends `BaseSchema[str]`
- `v.int_()` returns `IntSchema` which extends `BaseSchema[int]`
- `v.number()` returns `NumberSchema` which extends `BaseSchema[float]`
- `v.bool_()` returns `BoolSchema` which extends `BaseSchema[bool]`
- `v.null()` returns `NullSchema` which extends `BaseSchema[None]`

### ParseResult Typing

`ParseResult[T]` carries the type parameter through to the `.data` attribute:

```python
result: v.ParseResult[str] = v.string().safe_parse("hello")
if result.success:
    name: str = result.data  # type checker knows this is str
```

### Optional and Nullable Typing

```python
# OptionalSchema[str] -> BaseSchema[str | None]
opt = v.optional(v.string())

# NullableSchema[str] -> BaseSchema[str | None]
nul = v.nullable(v.string())
```

## Schema Types

### Primitives

#### string()

Creates a schema that validates string values. Supports chained constraint methods.

```python
s = v.string()
s.parse("hello")  # => "hello"
s.parse(42)        # raises: Expected string, received integer
```

#### number()

Creates a schema for floating-point numbers. Defaults to float64 (IEEE 754 double precision).

```python
n = v.number()
n.parse(3.14)   # => 3.14
n.parse("3")    # raises: Expected number, received string
```

#### int\_()

Creates a schema for integer values. Defaults to int64. Named `int_()` to avoid shadowing the built-in `int`.

```python
i = v.int_()
i.parse(42)    # => 42
i.parse(3.5)   # raises: Expected integer, received float
```

#### bool\_()

Creates a schema for boolean values. Named `bool_()` to avoid shadowing the built-in `bool`.

```python
b = v.bool_()
b.parse(True)    # => True
b.parse("yes")   # raises: Expected bool, received string
```

#### null()

Creates a schema that only accepts `None` (the Python equivalent of JSON `null`).

```python
n = v.null()
n.parse(None)  # => None
n.parse(0)     # raises: Expected null, received integer
```

#### any\_()

Accepts any value without validation. Named `any_()` to avoid shadowing the built-in `any`.

```python
a = v.any_()
a.parse("hello")  # => "hello"
a.parse(42)        # => 42
a.parse(None)      # => None
```

#### unknown()

Accepts any value without validation. Semantically identical to `any_()` at runtime.

```python
u = v.unknown()
u.parse("hello")  # => "hello"
```

#### never()

Always fails validation.

```python
n = v.never()
n.parse("anything")  # raises: Expected never
```

### Numeric Widths

AnyVali provides width-specific numeric schemas that enforce range constraints automatically.

#### Float Types

```python
v.float32()  # IEEE 754 single precision
v.float64()  # IEEE 754 double precision (same semantics as number())
```

#### Signed Integer Types

```python
v.int8()     # -128 to 127
v.int16()    # -32,768 to 32,767
v.int32()    # -2,147,483,648 to 2,147,483,647
v.int64()    # full Python int range (clamped to safe bounds)
```

#### Unsigned Integer Types

```python
v.uint8()    # 0 to 255
v.uint16()   # 0 to 65,535
v.uint32()   # 0 to 4,294,967,295
v.uint64()   # 0 to safe integer max
```

All integer types reject non-integer values (floats):

```python
port = v.uint16()
port.parse(8080)    # => 8080
port.parse(-1)      # raises: Value -1 is below the minimum for uint16
port.parse(100000)  # raises: Value 100000 is above the maximum for uint16
```

All numeric width schemas inherit the same constraint methods as `number()` and `int_()` (`.min()`, `.max()`, `.exclusive_min()`, `.exclusive_max()`, `.multiple_of()`).

### Values

#### literal(value)

Creates a schema that matches exactly one specific value.

```python
admin = v.literal("admin")
admin.parse("admin")  # => "admin"
admin.parse("user")   # raises: Expected literal admin, received user

forty_two = v.literal(42)
forty_two.parse(42)  # => 42
```

#### enum\_(values)

Creates a schema that matches any value in the given list.

```python
status = v.enum_(["active", "inactive", "pending"])
status.parse("active")   # => "active"
status.parse("deleted")  # raises: Expected one of enum(active,inactive,pending)
```

### Collections

#### array(item\_schema)

Creates a schema for lists where every element must match the item schema.

```python
tags = v.array(v.string())
tags.parse(["a", "b", "c"])  # => ["a", "b", "c"]
tags.parse([1, 2, 3])        # raises: Expected string, received integer (at index 0)
tags.parse("not a list")     # raises: Expected array, received string
```

#### tuple\_(schemas)

Creates a schema for fixed-length lists where each element matches the schema at its position.

```python
point = v.tuple_([v.number(), v.number()])
point.parse([1.5, 2.5])       # => [1.5, 2.5]
point.parse([1])               # raises: Tuple must have exactly 2 element(s)

mixed = v.tuple_([v.string(), v.int_(), v.bool_()])
mixed.parse(["hello", 42, True])  # => ["hello", 42, True]
```

#### object\_(properties, \*, required=None, unknown\_keys="reject")

Creates a schema for dicts with named properties.

```python
user = v.object_({
    "name": v.string(),
    "age": v.int_(),
})

user.parse({"name": "Alice", "age": 30})            # => {"name": "Alice", "age": 30}
user.parse({"name": "Alice"})                        # raises: Required field 'age' is missing
user.parse({"name": "Alice", "age": 30, "extra": 1}) # raises: Unknown key 'extra'
```

By default, all properties are required. Use `required` to specify which fields are required:

```python
user = v.object_(
    {"name": v.string(), "age": v.int_(), "bio": v.string()},
    required=["name"],
)
user.parse({"name": "Alice"})  # => {"name": "Alice"}
```

Or use the `.optional()` method on individual field schemas:

```python
user = v.object_({
    "name": v.string(),
    "bio": v.string().optional(),
})
user.parse({"name": "Alice"})  # => {"name": "Alice"}
```

Control unknown key handling:

```python
# Strip unknown keys silently
loose = v.object_({"name": v.string()}, unknown_keys="strip")
loose.parse({"name": "Alice", "extra": 1})  # => {"name": "Alice"}

# Allow unknown keys in the output
open_ = v.object_({"name": v.string()}, unknown_keys="allow")
open_.parse({"name": "Alice", "extra": 1})  # => {"name": "Alice", "extra": 1}
```

#### record(value\_schema)

Creates a schema for dicts with arbitrary string keys where all values must match a single schema.

```python
scores = v.record(v.int_())
scores.parse({"alice": 100, "bob": 95})  # => {"alice": 100, "bob": 95}
scores.parse({"alice": "high"})          # raises: Expected integer, received string
```

### Composition

#### union(schemas)

Creates a schema that accepts any value matching at least one of the given schemas. Variants are tried in order; the first match wins.

```python
id_schema = v.union([v.string(), v.int_()])
id_schema.parse("abc")  # => "abc"
id_schema.parse(42)     # => 42
id_schema.parse(True)   # raises: Input did not match any variant of the union
```

#### intersection(schemas)

Creates a schema that requires the value to match all given schemas. For objects, the results are merged.

```python
with_name = v.object_({"name": v.string()}, unknown_keys="strip")
with_age = v.object_({"age": v.int_()}, unknown_keys="strip")
person = v.intersection([with_name, with_age])

person.parse({"name": "Alice", "age": 30})  # => {"name": "Alice", "age": 30}
person.parse({"name": "Alice"})              # raises: Required field 'age' is missing
```

### Modifiers

#### optional(schema)

Wraps a schema so that absent values are accepted. Available both as a standalone function and as a method on any schema.

```python
# As a standalone function
maybe_age = v.optional(v.int_())

# As a method
maybe_age = v.int_().optional()
```

#### nullable(schema)

Wraps a schema so that `None` values are accepted. Available both as a standalone function and as a method on any schema.

```python
# As a standalone function
nullable_name = v.nullable(v.string())

# As a method
nullable_name = v.string().nullable()

nullable_name.parse("Alice")  # => "Alice"
nullable_name.parse(None)     # => None
nullable_name.parse(42)       # raises: Expected string, received integer
```

### ref(reference)

Creates a reference to a named definition. Used for recursive or shared schemas within an interchange document.

```python
node = v.ref("#/definitions/TreeNode")
```

## Constraints

### String Constraints

All string constraint methods return a new schema instance (immutable builder pattern).

```python
v.string().min_length(1)               # at least 1 character
v.string().max_length(255)             # at most 255 characters
v.string().pattern(r"^[a-z]+$")        # must match regex pattern
v.string().starts_with("https://")     # must start with prefix
v.string().ends_with(".json")          # must end with suffix
v.string().includes("@")              # must contain substring
v.string().format("email")            # must match a named format
```

Chain multiple constraints:

```python
email = (
    v.string()
    .min_length(5)
    .max_length(254)
    .format("email")
)
```

#### String Formats

The following format names are supported:

| Format | Description |
|---|---|
| `"email"` | Email address |
| `"url"` | URL |
| `"uuid"` | UUID (v1-v5) |
| `"ipv4"` | IPv4 address |
| `"ipv6"` | IPv6 address |
| `"date"` | ISO 8601 date (YYYY-MM-DD) |
| `"date-time"` | ISO 8601 date-time |

### Numeric Constraints

All numeric constraint methods apply to `number()`, `int_()`, and all width-specific variants.

```python
v.number().min(0)              # value >= 0
v.number().max(100)            # value <= 100
v.number().exclusive_min(0)    # value > 0
v.number().exclusive_max(100)  # value < 100
v.number().multiple_of(5)      # value must be divisible by 5
```

Chain multiple constraints:

```python
percentage = (
    v.number()
    .min(0)
    .max(100)
    .multiple_of(0.01)
)
```

### Array Constraints

```python
v.array(v.string()).min_items(1)   # at least 1 element
v.array(v.string()).max_items(10)  # at most 10 elements

# Combined
tags = v.array(v.string()).min_items(1).max_items(5)
```

### Object Options

The `unknown_keys` parameter controls how keys not declared in the properties are handled:

| Mode | Behavior |
|---|---|
| `"reject"` (default) | Produces an `unknown_key` issue for each extra key |
| `"strip"` | Silently removes extra keys from the output |
| `"allow"` | Passes extra keys through to the output |

## Coercion

Coercion transforms the input value before validation. It runs in step 2 of the parse pipeline, only when the value is present.

### Usage

Call `.coerce(...)` on any schema with keyword arguments:

```python
age = v.int_().coerce(to_int=True)
age.parse("42")  # => 42
age.parse(42)    # => 42
```

### Coercion Parameters

| Parameter | Type | Description |
|---|---|---|
| `to_int` | `bool` | Coerce string to integer |
| `to_number` | `bool` | Coerce string to float |
| `to_bool` | `bool` | Coerce string to boolean |
| `trim` | `bool` | Strip whitespace from string values |
| `lower` | `bool` | Convert string to lowercase |
| `upper` | `bool` | Convert string to uppercase |

### Available Coercions

#### String to Number

```python
n = v.number().coerce(to_number=True)
n.parse("3.14")  # => 3.14
n.parse("")      # raises: coercion_failed
```

#### String to Integer

```python
i = v.int_().coerce(to_int=True)
i.parse("42")   # => 42
i.parse("3.5")  # raises: coercion_failed
```

#### String to Boolean

```python
b = v.bool_().coerce(to_bool=True)
b.parse("true")   # => True
b.parse("false")  # => False
b.parse("1")      # => True
b.parse("0")      # => False
b.parse("yes")    # => True
b.parse("no")     # => False
```

#### String Transformations

```python
trimmed = v.string().coerce(trim=True)
trimmed.parse("  hello  ")  # => "hello"

lower = v.string().coerce(lower=True)
lower.parse("HELLO")  # => "hello"

upper = v.string().coerce(upper=True)
upper.parse("hello")  # => "HELLO"
```

Transformations can be combined:

```python
normalized = v.string().coerce(trim=True, lower=True)
normalized.parse("  Hello World  ")  # => "hello world"
```

## Defaults

Defaults fill in missing (absent) values. They run in step 3 of the parse pipeline, after coercion and before validation.

### Usage

Call `.default(value)` on any schema:

```python
role = v.string().default("user")
role.parse(None)     # => "user" (absent value filled)
role.parse("admin")  # => "admin"

tags = v.array(v.string()).default([])
```

Defaults work with optional fields in objects:

```python
config = v.object_({
    "theme": v.string().default("light").optional(),
    "language": v.string().default("en").optional(),
})

config.parse({})                    # => {"theme": "light", "language": "en"}
config.parse({"theme": "dark"})     # => {"theme": "dark", "language": "en"}
```

If the default value itself fails validation, a `default_invalid` issue is produced.

## Export and Import

AnyVali schemas can be exported to a portable JSON document and imported back in any supported SDK.

### export\_schema(schema, \*, mode="portable")

Exports a schema to a dict in the AnyVali interchange format.

```python
import anyvali as v

schema = v.object_({
    "name": v.string().min_length(1),
    "age": v.int_().min(0),
})

# Portable mode (default)
doc = v.export_schema(schema)

# Extended mode
doc = v.export_schema(schema, mode="extended")
```

You can also call `.export(mode)` directly on a schema instance:

```python
doc = schema.export("portable")
```

### export\_schema\_json(schema, \*, mode="portable", indent=2)

Convenience function that exports directly to a JSON string:

```python
json_str = v.export_schema_json(schema)
print(json_str)
```

### import\_schema(doc)

Imports an interchange document dict into a live schema that can be used for parsing.

```python
import json
import anyvali as v

doc = {
    "anyvaliVersion": "1.0",
    "schemaVersion": "1",
    "root": {
        "kind": "object",
        "properties": {
            "name": {"kind": "string", "minLength": 1},
            "age": {"kind": "int", "min": 0},
        },
        "required": ["name", "age"],
        "unknownKeys": "reject",
    },
    "definitions": {},
    "extensions": {},
}

schema = v.import_schema(doc)
user = schema.parse({"name": "Alice", "age": 30})
```

### Export Modes

| Mode | Behavior |
|---|---|
| `"portable"` | Only emits portable schema features. Safe default. |
| `"extended"` | Emits the core schema plus language-specific extensions in the `extensions` field. |

## Error Handling

### ValidationError

When `parse()` fails, it raises a `ValidationError` containing a list of issues:

```python
import anyvali as v

try:
    v.string().min_length(5).parse("hi")
except v.ValidationError as err:
    print(err)
    # "Validation failed: String must have at least 5 character(s)"

    for issue in err.issues:
        print(issue.code)     # "too_small"
        print(issue.message)  # "String must have at least 5 character(s)"
        print(issue.path)     # []
```

### ParseResult

`safe_parse()` never raises. It returns a `ParseResult[T]` dataclass:

```python
@dataclass(frozen=True)
class ParseResult(Generic[T]):
    success: bool
    data: T | None = None
    issues: list[ValidationIssue] = field(default_factory=list)
```

```python
result = v.string().safe_parse(42)

if result.success:
    print(result.data)  # str
else:
    for issue in result.issues:
        print(issue)
```

### ValidationIssue

Each issue is a frozen dataclass:

```python
@dataclass(frozen=True)
class ValidationIssue:
    code: str                           # Machine-readable issue code
    message: str                        # Human-readable description
    path: list[str | int] = []          # Path to the failing value
    expected: Any = None                # What was expected
    received: Any = None                # What was received
    meta: dict[str, Any] | None = None  # Additional metadata
```

The `path` list describes the location of the error within nested structures:

```python
schema = v.object_({
    "users": v.array(v.object_({
        "email": v.string().format("email"),
    })),
})

result = schema.safe_parse({
    "users": [{"email": "not-an-email"}],
})

# result.issues[0].path => ["users", 0, "email"]
```

### Issue Codes

All issue codes are available as module-level constants:

```python
import anyvali as v

v.INVALID_TYPE
v.REQUIRED
v.UNKNOWN_KEY
v.TOO_SMALL
v.TOO_LARGE
v.INVALID_STRING
v.INVALID_NUMBER
v.INVALID_LITERAL
v.INVALID_UNION
v.COERCION_FAILED
v.DEFAULT_INVALID
v.CUSTOM_VALIDATION_NOT_PORTABLE
v.UNSUPPORTED_EXTENSION
v.UNSUPPORTED_SCHEMA_KIND
```

| Code | When |
|---|---|
| `invalid_type` | Value is the wrong type |
| `required` | Required property is missing |
| `unknown_key` | Object has an undeclared key (when `unknown_keys` is `"reject"`) |
| `too_small` | String too short, number too low, or array too few items |
| `too_large` | String too long, number too high, or array too many items |
| `invalid_string` | String fails a pattern, starts_with, ends_with, includes, or format check |
| `invalid_number` | Number fails a multiple_of check or is non-finite |
| `invalid_literal` | Value does not match the expected literal |
| `invalid_union` | Value does not match any variant in a union |
| `coercion_failed` | Coercion could not convert the input |
| `default_invalid` | The materialized default value failed validation |
| `custom_validation_not_portable` | Non-portable custom validation was encountered |
| `unsupported_extension` | Unknown extension namespace in an imported document |
| `unsupported_schema_kind` | Unknown schema kind in an imported document |

## API Reference

### Builder Functions

| Function | Returns | Description |
|---|---|---|
| `v.string()` | `StringSchema` | String values |
| `v.number()` | `NumberSchema` | Float64 numbers |
| `v.float32()` | `Float32Schema` | Float32 numbers |
| `v.float64()` | `Float64Schema` | Float64 numbers |
| `v.int_()` | `IntSchema` | Int64 integers |
| `v.int8()` | `Int8Schema` | Int8 integers |
| `v.int16()` | `Int16Schema` | Int16 integers |
| `v.int32()` | `Int32Schema` | Int32 integers |
| `v.int64()` | `Int64Schema` | Int64 integers |
| `v.uint8()` | `Uint8Schema` | Uint8 integers |
| `v.uint16()` | `Uint16Schema` | Uint16 integers |
| `v.uint32()` | `Uint32Schema` | Uint32 integers |
| `v.uint64()` | `Uint64Schema` | Uint64 integers |
| `v.bool_()` | `BoolSchema` | Boolean values |
| `v.null()` | `NullSchema` | None/null only |
| `v.any_()` | `AnySchema` | Any value |
| `v.unknown()` | `UnknownSchema` | Any value |
| `v.never()` | `NeverSchema` | Always fails |
| `v.literal(value)` | `LiteralSchema` | Exact value match |
| `v.enum_(values)` | `EnumSchema` | One of the listed values |
| `v.array(items)` | `ArraySchema` | List of uniform type |
| `v.tuple_(items)` | `TupleSchema` | Fixed-length typed list |
| `v.object_(properties, *, required=None, unknown_keys="reject")` | `ObjectSchema` | Dict with named properties |
| `v.record(value_schema)` | `RecordSchema` | Dict with uniform value type |
| `v.union(schemas)` | `UnionSchema` | First-match union |
| `v.intersection(schemas)` | `IntersectionSchema` | All-match intersection |
| `v.optional(schema)` | `OptionalSchema` | Allows absent values |
| `v.nullable(schema)` | `NullableSchema` | Allows `None` |
| `v.ref(reference)` | `RefSchema` | Reference to a definition |

### BaseSchema Methods

Available on all schema instances:

| Method | Returns | Description |
|---|---|---|
| `.parse(input)` | `T` | Parse or raise `ValidationError` |
| `.safe_parse(input)` | `ParseResult[T]` | Parse and return result object |
| `.default(value)` | `BaseSchema[T]` | Set a default for absent values |
| `.coerce(...)` | `BaseSchema[T]` | Configure coercion behavior (keyword args) |
| `.optional()` | `OptionalSchema[T]` | Wrap as optional |
| `.nullable()` | `NullableSchema[T]` | Wrap as nullable |
| `.export(mode="portable")` | `dict` | Export to interchange document dict |

### StringSchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.min_length(n)` | `int` | Minimum string length |
| `.max_length(n)` | `int` | Maximum string length |
| `.pattern(p)` | `str` | Regex pattern to match |
| `.starts_with(s)` | `str` | Required prefix |
| `.ends_with(s)` | `str` | Required suffix |
| `.includes(s)` | `str` | Required substring |
| `.format(f)` | `str` | Named format validation |

### NumberSchema / IntSchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.min(v)` | `float \| int` | Minimum value (inclusive) |
| `.max(v)` | `float \| int` | Maximum value (inclusive) |
| `.exclusive_min(v)` | `float \| int` | Minimum value (exclusive) |
| `.exclusive_max(v)` | `float \| int` | Maximum value (exclusive) |
| `.multiple_of(v)` | `float \| int` | Value must be divisible by `v` |

### ArraySchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.min_items(n)` | `int` | Minimum number of elements |
| `.max_items(n)` | `int` | Maximum number of elements |

### Top-Level Functions

| Function | Signature | Description |
|---|---|---|
| `v.parse` | `parse(schema, input) -> T` | Parse using a schema reference |
| `v.safe_parse` | `safe_parse(schema, input) -> ParseResult[T]` | Safe parse using a schema reference |
| `v.export_schema` | `export_schema(schema, *, mode="portable") -> dict` | Export to interchange format |
| `v.export_schema_json` | `export_schema_json(schema, *, mode="portable", indent=2) -> str` | Export to JSON string |
| `v.import_schema` | `import_schema(doc) -> BaseSchema` | Import from interchange format |

### Types

| Type | Description |
|---|---|
| `BaseSchema[T]` | Abstract base for all schema types |
| `ParseResult[T]` | Dataclass with `.success`, `.data`, `.issues` |
| `ValidationIssue` | Frozen dataclass with `.code`, `.message`, `.path`, `.expected`, `.received`, `.meta` |
| `ValidationError` | Exception with `.issues: list[ValidationIssue]` |
| `AnyValiDocument` | Dataclass with `.anyvali_version`, `.schema_version`, `.root`, `.definitions`, `.extensions` |
| `ExportMode` | `Literal["portable", "extended"]` |
| `UnknownKeyMode` | `Literal["reject", "strip", "allow"]` |
| `CoercionConfig` | Dataclass with `.to_int`, `.to_number`, `.to_bool`, `.trim`, `.lower`, `.upper` |

### AnyValiDocument Methods

The `AnyValiDocument` dataclass provides convenience methods for dict/JSON conversion:

| Method | Returns | Description |
|---|---|---|
| `.to_dict()` | `dict` | Convert to a plain dict with camelCase keys |
| `AnyValiDocument.from_dict(d)` | `AnyValiDocument` | Construct from a plain dict |

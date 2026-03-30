# AnyVali Go SDK Reference

The Go SDK provides a native, idiomatic API for schema validation with full portable interchange support. Schemas are defined using exported builder functions, validated at runtime, and can be exported to or imported from AnyVali's canonical JSON format.

## Installation

```bash
go get github.com/BetterCorp/AnyVali/sdk/go
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    v "github.com/BetterCorp/AnyVali/sdk/go"
)

func main() {
    // Define a schema
    userSchema := v.Object(map[string]v.Schema{
        "id":    v.Int64(),
        "name":  v.String().MinLength(1).MaxLength(100),
        "email": v.String().Format("email"),
        "age":   v.Optional(v.Int().Min(0).Max(150)),
        "role":  v.Enum("admin", "user", "guest").Default("user"),
    })

    // Parse input (returns error on failure)
    user, err := userSchema.Parse(map[string]any{
        "id":    42,
        "name":  "Alice",
        "email": "alice@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user)
    // map[age:<nil> email:alice@example.com id:42 name:Alice role:user]

    // SafeParse returns a result struct instead of an error
    result := userSchema.SafeParse(map[string]any{
        "id":   "not-a-number",
        "name": "",
    })
    if !result.Success {
        for _, issue := range result.Issues {
            fmt.Printf("%v: [%s] %s\n", issue.Path, issue.Code, issue.Message)
        }
    }
}
```

## Type Inference

Go interfaces are not generic, so the `Schema` interface returns `any` from `Parse`. The SDK provides generic helper functions that parse and cast in one step.

### TypedParse

```go
schema := v.String().MinLength(1)

// Without TypedParse -- requires manual assertion
raw, err := schema.Parse("hello")
str := raw.(string)

// With TypedParse -- returns the correct type directly
str, err := v.TypedParse[string](schema, "hello")
// str is string, no cast needed
```

### TypedSafeParse

```go
schema := v.Int().Min(0)

result := v.TypedSafeParse[int64](schema, 42)
if result.Success {
    fmt.Println(result.Data) // result.Data is int64, not any
}
```

### Typed result struct

`TypedParseResult[T]` mirrors `ParseResult` but with a typed `Data` field:

```go
type TypedParseResult[T any] struct {
    Success bool
    Data    T
    Issues  []ValidationIssue
}
```

### Complex typed parsing

```go
schema := v.Object(map[string]v.Schema{
    "name":  v.String(),
    "score": v.Number(),
})

result, err := v.TypedParse[map[string]any](schema, map[string]any{
    "name":  "Alice",
    "score": 98.5,
})
// result is map[string]any with validated data
```

If the validated output cannot be cast to `T`, `TypedParse` returns an error and `TypedSafeParse` returns a result with a `type_assertion_failed` issue.

## Schema Types

### Primitives

```go
// String
s := v.String()

// Boolean
b := v.Bool()

// Null -- accepts only nil
n := v.Null()
```

### Numeric Types

```go
// Number (float64, the safe cross-language default)
num := v.Number()

// Explicit float widths
f64 := v.Float64()
f32 := v.Float32()   // enforces float32 range

// Signed integers
i   := v.Int()       // int64
i8  := v.Int8()      // int8 range
i16 := v.Int16()     // int16 range
i32 := v.Int32()     // int32 range
i64 := v.Int64()     // int64 range

// Unsigned integers
u8  := v.Uint8()     // uint8 range
u16 := v.Uint16()    // uint16 range
u32 := v.Uint32()    // uint32 range
u64 := v.Uint64()    // uint64 range
```

`Number()` and `Float64()` both produce a `Float64Schema`. `Int()` and `Int64()` both produce an `IntSchema` with int64 range. Narrower widths (Int8, Uint16, etc.) produce the same `IntSchema` type with range enforcement built in.

### Special Types

```go
// Any -- accepts any value, passes it through
a := v.Any()

// Unknown -- accepts any value (semantically "not yet validated")
u := v.Unknown()

// Never -- rejects all values
n := v.Never()
```

### Literal and Enum

```go
// Literal -- matches a single exact value
lit := v.Literal("active")
lit2 := v.Literal(42)
lit3 := v.Literal(true)

// Enum -- matches one of several allowed values
role := v.Enum("admin", "user", "guest")
status := v.Enum(1, 2, 3)
```

### Array and Tuple

```go
// Array -- all items must match the given schema
tags := v.Array(v.String())

// Array with length constraints
scores := v.Array(v.Number()).MinItems(1).MaxItems(100)

// Tuple -- fixed-length array with per-position schemas
pair := v.Tuple(v.String(), v.Int())
```

### Object

```go
// Object -- all declared properties are required by default
user := v.Object(map[string]v.Schema{
    "name":  v.String().MinLength(1),
    "email": v.String().Format("email"),
    "age":   v.Int().Min(0),
})

// Make specific fields optional using the Optional wrapper
profile := v.Object(map[string]v.Schema{
    "name": v.String(),
    "bio":  v.Optional(v.String().MaxLength(500)),
})

// Control which fields are required explicitly
config := v.Object(map[string]v.Schema{
    "host": v.String(),
    "port": v.Int(),
    "tls":  v.Bool(),
}).Required("host", "port") // tls is now optional
```

### Record

```go
// Record -- validates a map where all values match a schema
headers := v.Record(v.String())

// Accepts: map[string]any{"Content-Type": "text/html", "Accept": "application/json"}
```

### Composition

```go
// Union -- value must match at least one of the schemas
strOrNum := v.Union(v.String(), v.Number())

// Intersection -- value must match all schemas
// Useful for combining object schemas
base := v.Object(map[string]v.Schema{
    "id": v.Int64(),
})
named := v.Object(map[string]v.Schema{
    "name": v.String(),
})
entity := v.Intersection(base, named)
```

### Modifiers

```go
// Optional -- absent values are accepted (returns nil)
opt := v.Optional(v.String())

// Nullable -- null values are accepted
nul := v.Nullable(v.String())

// Optional with default -- absent values get the default
optWithDefault := v.Optional(v.String()).Default("unknown")
```

## Constraints

### String Constraints

```go
s := v.String().
    MinLength(1).            // minimum character count
    MaxLength(255).          // maximum character count
    Pattern(`^\w+$`).        // regex pattern (Go RE2 syntax)
    StartsWith("hello").     // must start with prefix
    EndsWith(".com").        // must end with suffix
    Includes("@").           // must contain substring
    Format("email")          // built-in format validator
```

Supported format values: `"email"`, `"url"`, `"uuid"`, `"ipv4"`, `"ipv6"`, `"date"`, `"date-time"`.

### Numeric Constraints

All numeric schemas (Number, Float64, Float32, Int, Int8--Int64, Uint8--Uint64) support:

```go
// Float schemas use float64 parameters
price := v.Number().
    Min(0.0).                // value >= 0
    Max(999.99).             // value <= 999.99
    ExclusiveMin(0.0).       // value > 0
    ExclusiveMax(1000.0).    // value < 1000
    MultipleOf(0.01)         // must be a multiple of 0.01

// Int schemas use int64 parameters
age := v.Int().
    Min(0).                  // value >= 0
    Max(150).                // value <= 150
    ExclusiveMin(0).         // value > 0
    ExclusiveMax(200).       // value < 200
    MultipleOf(1)            // must be a multiple of 1
```

### Array Constraints

```go
items := v.Array(v.String()).
    MinItems(1).             // at least 1 item
    MaxItems(50)             // at most 50 items
```

### Object Unknown Key Handling

The `UnknownKeys` option controls how keys not declared in the shape are handled:

| Mode | Behavior |
|---|---|
| `Reject` (default) | Produces an `unknown_key` issue for each extra key |
| `Strip` | Silently removes extra keys from the output |
| `Allow` | Passes extra keys through to the output |

```go
// Reject unknown keys (default)
strict := v.Object(map[string]v.Schema{
    "name": v.String(),
}).UnknownKeys(v.Reject)

// Strip unknown keys silently
stripped := v.Object(map[string]v.Schema{
    "name": v.String(),
}).UnknownKeys(v.Strip)

// Allow unknown keys to pass through
loose := v.Object(map[string]v.Schema{
    "name": v.String(),
}).UnknownKeys(v.Allow)
```

## Coercion and Defaults

### Coercion

Coercions transform values before validation. They run on present values only.

```go
// Coerce string to integer
port := v.Int().Coerce(v.CoerceToInt)
result, _ := port.Parse("8080")  // => int64(8080)

// Coerce string to float
temp := v.Number().Coerce(v.CoerceToNumber)
result, _ := temp.Parse("36.6")  // => float64(36.6)

// Coerce string to bool
flag := v.Bool().Coerce(v.CoerceToBool)
result, _ := flag.Parse("true")  // => true

// String coercions
name := v.String().Coerce(v.CoerceTrim)   // trims whitespace
tag := v.String().Coerce(v.CoerceLower)    // lowercases
code := v.String().Coerce(v.CoerceUpper)   // uppercases
```

Available coercion types:

| Constant | Effect |
|---|---|
| `CoerceToInt` | Parse string to int64 |
| `CoerceToNumber` | Parse string to float64 |
| `CoerceToBool` | Parse string to bool |
| `CoerceTrim` | Trim whitespace from string |
| `CoerceLower` | Lowercase string |
| `CoerceUpper` | Uppercase string |

### Defaults

Defaults fill in missing (absent) values. The default value must pass validation. Defaults must be static values (for portability across SDKs); use the Common Patterns section below for computed defaults.

Call `.Default(value)` on any schema:

```go
role := v.Enum("admin", "user", "guest").Default("user")
age := v.Int().Min(0).Default(int64(0))
active := v.Bool().Default(true)
name := v.String().Default("Anonymous")
tags := v.Array(v.String()).Default([]any{})
```

Optional schemas also support defaults:

```go
bio := v.Optional(v.String().MaxLength(500)).Default("No bio provided")
```

## Export and Import

### Export

Convert a schema to the portable AnyVali JSON document:

```go
schema := v.Object(map[string]v.Schema{
    "name":  v.String().MinLength(1),
    "email": v.String().Format("email"),
})

// Export to Document struct
doc, err := v.Export(schema, v.Portable)
if err != nil {
    log.Fatal(err)
}

// Export directly to JSON bytes
jsonBytes, err := v.ExportJSON(schema, v.Portable)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(jsonBytes))
```

Output:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": {
    "kind": "object",
    "properties": {
      "name": { "kind": "string", "minLength": 1 },
      "email": { "kind": "string", "format": "email" }
    },
    "required": ["name", "email"],
    "unknownKeys": "reject"
  },
  "definitions": {},
  "extensions": {}
}
```

Export modes:

| Mode | Constant | Behavior |
|---|---|---|
| Portable | `v.Portable` | Fails if schema uses non-portable features |
| Extended | `v.Extended` | Includes language-specific extensions |

### Import

Reconstruct a schema from a JSON document:

```go
// Import from Document struct
schema, err := v.Import(doc)
if err != nil {
    log.Fatal(err)
}

// Import from raw JSON bytes
schema, err := v.ImportJSON(jsonBytes)
if err != nil {
    log.Fatal(err)
}

// Use the imported schema normally
result, err := schema.Parse(map[string]any{
    "name":  "Bob",
    "email": "bob@test.com",
})
```

## Error Handling

### Parse (throwing)

`Parse` returns `(any, error)`. The error is always a `*ValidationError`:

```go
result, err := schema.Parse(input)
if err != nil {
    var ve *v.ValidationError
    if errors.As(err, &ve) {
        for _, issue := range ve.Issues {
            fmt.Printf("[%s] %s at %v\n", issue.Code, issue.Message, issue.Path)
        }
    }
}
```

### SafeParse (non-throwing)

`SafeParse` returns a `ParseResult` struct that never errors:

```go
result := schema.SafeParse(input)
if result.Success {
    fmt.Println("Valid:", result.Data)
} else {
    for _, issue := range result.Issues {
        fmt.Printf("[%s] %s at %v\n", issue.Code, issue.Message, issue.Path)
    }
}
```

### ValidationIssue

Each issue has a consistent structure:

```go
type ValidationIssue struct {
    Code     string         // machine-readable issue code
    Message  string         // human-readable description
    Path     []any          // location in the input (strings for keys, ints for indices)
    Expected string         // what was expected
    Received string         // what was received
    Meta     map[string]any // optional additional metadata
}
```

### Issue Codes

| Constant | Code String | Meaning |
|---|---|---|
| `IssueInvalidType` | `"invalid_type"` | Value has wrong type |
| `IssueRequired` | `"required"` | Required field is missing |
| `IssueUnknownKey` | `"unknown_key"` | Object has an undeclared key |
| `IssueTooSmall` | `"too_small"` | Below minimum (length, value, items) |
| `IssueTooLarge` | `"too_large"` | Above maximum (length, value, items) |
| `IssueInvalidString` | `"invalid_string"` | String constraint failed (pattern, format, etc.) |
| `IssueInvalidNumber` | `"invalid_number"` | Numeric constraint failed (multipleOf, range) |
| `IssueInvalidLiteral` | `"invalid_literal"` | Literal/enum value mismatch |
| `IssueInvalidUnion` | `"invalid_union"` | No union variant matched |
| `IssueCoercionFailed` | `"coercion_failed"` | Coercion could not convert the value |
| `IssueDefaultInvalid` | `"default_invalid"` | Default value fails validation |
| `IssueUnsupportedSchemaKind` | `"unsupported_schema_kind"` | Unknown or unresolved schema kind |
| `IssueUnsupportedExtension` | `"unsupported_extension"` | Non-portable extension encountered |

## Common Patterns

### Validating Environment Variables

Use `UnknownKeysStrip` when parsing maps that contain many extra keys you don't care about, like environment variables:

```go
envSchema := av.Object(map[string]av.Schema{
    "NODE_ENV":     av.Optional(av.String()).WithDefault("development"),
    "PORT":         av.Optional(av.Int()),
    "DATABASE_URL": av.String(),
}).UnknownKeys(av.UnknownKeysStrip)
```

Without `Strip`, parse would fail with `unknown_key` issues for every other variable in the environment (PATH, HOME, etc.) because the default mode is `Reject`.

### Eagerly Evaluated vs Lazy Defaults

`.Default()` accepts any value of the correct type. Expressions like `os.Getwd()` are evaluated immediately when the schema is created and stored as a static value -- this works fine. What AnyVali does not support is lazy function defaults that re-evaluate on each parse call. If you need a fresh value on every parse, apply it after:

```go
configSchema := av.Object(map[string]av.Schema{
    "profile": av.Optional(av.String()).Default("default"),
    "appDir":  av.Optional(av.String()),
}).UnknownKeys(av.Strip)

config, err := configSchema.Parse(envMap)
if err != nil {
    log.Fatal(err)
}
if config.(map[string]any)["appDir"] == nil {
    wd, _ := os.Getwd()
    config.(map[string]any)["appDir"] = wd
}
```

This keeps the schema fully portable -- the same JSON document can be imported in Python, JavaScript, or any other SDK without relying on language-specific function calls.

## API Reference

### Builder Functions

| Function | Returns | Description |
|---|---|---|
| `String()` | `*StringSchema` | String validator |
| `Number()` | `*Float64Schema` | Float64 validator (alias: `number`) |
| `Float64()` | `*Float64Schema` | Float64 validator |
| `Float32()` | `*Float32Schema` | Float32 validator with range check |
| `Int()` | `*IntSchema` | Int64 validator (alias: `int`) |
| `Int8()` | `*IntSchema` | Int8-range validator |
| `Int16()` | `*IntSchema` | Int16-range validator |
| `Int32()` | `*IntSchema` | Int32-range validator |
| `Int64()` | `*IntSchema` | Int64-range validator |
| `Uint8()` | `*IntSchema` | Uint8-range validator |
| `Uint16()` | `*IntSchema` | Uint16-range validator |
| `Uint32()` | `*IntSchema` | Uint32-range validator |
| `Uint64()` | `*IntSchema` | Uint64-range validator |
| `Bool()` | `*BoolSchema` | Boolean validator |
| `Null()` | `*NullSchema` | Null validator |
| `Any()` | `*AnySchema` | Accepts any value |
| `Unknown()` | `*UnknownSchema` | Accepts any value (semantically unvalidated) |
| `Never()` | `*NeverSchema` | Rejects all values |
| `Literal(v any)` | `*LiteralSchema` | Exact value match |
| `Enum(values ...any)` | `*EnumSchema` | One of allowed values |
| `Array(item Schema)` | `*ArraySchema` | Homogeneous array |
| `Tuple(items ...Schema)` | `*TupleSchema` | Fixed-length typed array |
| `Object(props map[string]Schema)` | `*ObjectSchema` | Structured object |
| `Record(value Schema)` | `*RecordSchema` | String-keyed map with uniform values |
| `Union(schemas ...Schema)` | `*UnionSchema` | Matches any one variant |
| `Intersection(schemas ...Schema)` | `*IntersectionSchema` | Matches all schemas |
| `Optional(s Schema)` | `*OptionalSchema` | Allows absent values |
| `Nullable(s Schema)` | `*NullableSchema` | Allows null values |

### Schema Interface

```go
type Schema interface {
    Parse(input any) (any, error)
    SafeParse(input any) ParseResult
    ToNode() map[string]any
}
```

### Generic Helpers

```go
func TypedParse[T any](s Schema, input any) (T, error)
func TypedSafeParse[T any](s Schema, input any) TypedParseResult[T]
```

### Constraint Methods

**StringSchema** -- all return `*StringSchema` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `MinLength(n)` | `int` | Minimum character count |
| `MaxLength(n)` | `int` | Maximum character count |
| `Pattern(p)` | `string` | Regex pattern (RE2 syntax) |
| `StartsWith(s)` | `string` | Required prefix |
| `EndsWith(s)` | `string` | Required suffix |
| `Includes(s)` | `string` | Required substring |
| `Format(f)` | `string` | Built-in format check |
| `Default(v)` | `string` | Default value |
| `Coerce(c)` | `CoercionType` | Add coercion |

**Float64Schema / NumberSchema / Float32Schema** -- all return `*Float64Schema` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `Min(n)` | `float64` | Inclusive minimum |
| `Max(n)` | `float64` | Inclusive maximum |
| `ExclusiveMin(n)` | `float64` | Exclusive minimum |
| `ExclusiveMax(n)` | `float64` | Exclusive maximum |
| `MultipleOf(n)` | `float64` | Divisibility constraint |
| `Default(v)` | `float64` | Default value |
| `Coerce(c)` | `CoercionType` | Add coercion |

**IntSchema** -- all return `*IntSchema` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `Min(n)` | `int64` | Inclusive minimum |
| `Max(n)` | `int64` | Inclusive maximum |
| `ExclusiveMin(n)` | `int64` | Exclusive minimum |
| `ExclusiveMax(n)` | `int64` | Exclusive maximum |
| `MultipleOf(n)` | `int64` | Divisibility constraint |
| `Default(v)` | `int64` | Default value |
| `Coerce(c)` | `CoercionType` | Add coercion |

**ArraySchema** -- all return `*ArraySchema` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `MinItems(n)` | `int` | Minimum item count |
| `MaxItems(n)` | `int` | Maximum item count |
| `Default(v)` | `[]any` | Default value |

**ObjectSchema** -- all return `*ObjectSchema` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `Required(fields...)` | `...string` | Set which fields are required (overrides default) |
| `UnknownKeys(mode)` | `UnknownKeyMode` | How to handle undeclared keys |
| `Default(v)` | `map[string]any` | Default value |

**BoolSchema**, **EnumSchema**, **LiteralSchema**, **OptionalSchema**, **NullableSchema** -- each supports `Default(v)` and (where applicable) `Coerce(c)`.

### Export/Import Functions

```go
func Export(schema Schema, mode ExportMode) (*Document, error)
func ExportJSON(schema Schema, mode ExportMode) ([]byte, error)
func Import(doc *Document) (Schema, error)
func ImportJSON(data []byte) (Schema, error)
```

### Core Types

```go
type ParseResult struct {
    Success bool
    Data    any
    Issues  []ValidationIssue
}

type TypedParseResult[T any] struct {
    Success bool
    Data    T
    Issues  []ValidationIssue
}

type ValidationIssue struct {
    Code     string
    Message  string
    Path     []any
    Expected string
    Received string
    Meta     map[string]any
}

type ValidationError struct {
    Issues []ValidationIssue
}

type Document struct {
    AnyvaliVersion string                    `json:"anyvaliVersion"`
    SchemaVersion  string                    `json:"schemaVersion"`
    Root           map[string]any            `json:"root"`
    Definitions    map[string]map[string]any `json:"definitions,omitempty"`
    Extensions     map[string]any            `json:"extensions,omitempty"`
}

type ExportMode string   // Portable, Extended
type UnknownKeyMode string // Reject, Strip, Allow
type CoercionType string  // CoerceToInt, CoerceToNumber, CoerceToBool, CoerceTrim, CoerceLower, CoerceUpper
```

# C# SDK Reference

## Installation

```bash
dotnet add package AnyVali
```

## Quick Start

```csharp
using AnyVali;

// Define a schema
var userSchema = V.Object(new Dictionary<string, Schema>
{
    ["name"] = V.String().MinLength(1).MaxLength(100),
    ["email"] = V.String().Format("email"),
    ["age"] = V.Optional(V.Int().Min(0).Max(150)),
});

// Parse input (throws on failure)
var user = userSchema.Parse(new Dictionary<string, object?>
{
    ["name"] = "Alice",
    ["email"] = "alice@example.com",
    ["age"] = 30,
});

// Safe parse (returns result object)
var result = userSchema.SafeParse(new Dictionary<string, object?>
{
    ["name"] = "",
    ["email"] = "bad",
});

if (!result.Success)
{
    foreach (var issue in result.Issues)
    {
        Console.WriteLine($"{string.Join(".", issue.Path)}: [{issue.Code}] {issue.Message}");
    }
}
```

## Schema Builders

All schemas are created through static methods on the `V` class.

### Primitive Types

```csharp
using AnyVali;

var str = V.String();
var num = V.Number();    // float64
var integer = V.Int();   // int64
var flag = V.Bool();
var nil = V.Null();
```

### Numeric Types

AnyVali provides explicit numeric schemas for cross-language safety.

```csharp
// Floating point
var f32 = V.Float32();
var f64 = V.Float64();
var number = V.Number();   // alias for float64

// Signed integers
var i = V.Int();           // int64 (default)
var i8 = V.Int8();         // -128 to 127
var i16 = V.Int16();       // -32,768 to 32,767
var i32 = V.Int32();       // -2^31 to 2^31-1
var i64 = V.Int64();       // -2^63 to 2^63-1

// Unsigned integers
var u8 = V.Uint8();        // 0 to 255
var u16 = V.Uint16();      // 0 to 65,535
var u32 = V.Uint32();      // 0 to 4,294,967,295
var u64 = V.Uint64();      // 0 to int64 max (capped for safety)
```

### Special Types

```csharp
var any = V.Any();         // accepts any value
var unknown = V.Unknown(); // accepts any value (stricter intent)
var never = V.Never();     // always fails
```

### Literal and Enum

```csharp
// Literal: must be exactly this value
var status = V.Literal("active");

// Enum: must be one of these values
var role = V.Enum("admin", "user", "guest");
```

### Collections

```csharp
// Array of strings
var tags = V.Array(V.String());

// Tuple: fixed-length, each position typed
var point = V.Tuple(V.Number(), V.Number());

// Record: string keys, all values share a schema
var scores = V.Record(V.Int());
```

### Object

```csharp
var user = V.Object(new Dictionary<string, Schema>
{
    ["name"] = V.String().MinLength(1),
    ["email"] = V.String().Format("email"),
    ["age"] = V.Optional(V.Int().Min(0)),
});
```

By default, unknown keys are rejected. Pass `UnknownKeyMode` to change this:

```csharp
// Strip unknown keys silently
var loose = V.Object(
    new Dictionary<string, Schema>
    {
        ["id"] = V.Int(),
    },
    UnknownKeyMode.Strip
);

// Allow unknown keys in output
var passthrough = V.Object(
    new Dictionary<string, Schema>
    {
        ["id"] = V.Int(),
    },
    UnknownKeyMode.Allow
);
```

### Composition

```csharp
// Union: value must match at least one variant
var stringOrInt = V.Union(V.String(), V.Int());

// Intersection: value must match all schemas
var named = V.Intersection(
    V.Object(new Dictionary<string, Schema> { ["name"] = V.String() }),
    V.Object(new Dictionary<string, Schema> { ["age"] = V.Int() })
);
```

### Modifiers

```csharp
// Optional: field may be absent
var maybeAge = V.Optional(V.Int());

// Nullable: field may be null
var nullableName = V.Nullable(V.String());
```

## String Constraints

```csharp
var schema = V.String()
    .MinLength(1)          // minimum character count
    .MaxLength(255)        // maximum character count
    .Pattern(@"^\w+$")    // regex pattern
    .StartsWith("hello")   // must start with prefix
    .EndsWith(".com")      // must end with suffix
    .Includes("@")         // must contain substring
    .Format("email");      // built-in format validator
```

Each constraint method returns a new schema instance -- schemas are immutable.

### Built-in Formats

The `.Format()` method supports these built-in format validators:

- `"email"` -- RFC 5322 email address
- `"uri"` -- absolute URI
- `"uuid"` -- UUID v4
- `"iso8601"` -- ISO 8601 date-time
- `"ipv4"` -- IPv4 address
- `"ipv6"` -- IPv6 address

```csharp
var email = V.String().Format("email");
var uri = V.String().Format("uri");
var id = V.String().Format("uuid");
```

## Numeric Constraints

Constraints apply to all numeric schemas (`NumberSchema`, `IntSchema`, and their subtypes).

```csharp
var price = V.Number()
    .Min(0)                // value >= 0
    .Max(10000)            // value <= 10000
    .ExclusiveMin(0)       // value > 0
    .ExclusiveMax(10000)   // value < 10000
    .MultipleOf(0.01);     // must be a multiple of 0.01

var port = V.Int()
    .Min(1)
    .Max(65535);

var evenNumber = V.Int().MultipleOf(2);
```

## Array Constraints

```csharp
var tags = V.Array(V.String())
    .MinItems(1)    // at least 1 element
    .MaxItems(10);  // at most 10 elements
```

## Parsing

### Parse (Throwing)

`Parse()` returns the validated value or throws a `ValidationError`.

```csharp
try
{
    // Untyped parse: returns object?
    object? value = V.String().Parse("hello");

    // Typed parse via Schema<T>: returns T
    string name = V.String().Parse("Alice");
    long age = V.Int().Parse(30);
    double price = V.Number().Parse(9.99);
    bool flag = V.Bool().Parse(true);
}
catch (ValidationError ex)
{
    foreach (var issue in ex.Issues)
    {
        Console.WriteLine($"[{issue.Code}] {issue.Message}");
    }
}
```

### SafeParse (Non-Throwing)

`SafeParse()` returns a `ParseResult` that never throws.

```csharp
ParseResult result = V.String().SafeParse(42);

if (result.Success)
{
    Console.WriteLine($"Parsed: {result.Data}");
}
else
{
    foreach (var issue in result.Issues)
    {
        Console.WriteLine($"[{issue.Code}] {issue.Message}");
        // [invalid_type] Expected string, received number
    }
}
```

### SafeParseTyped (Strongly Typed)

`SafeParseTyped()` is available on `Schema<T>` and returns `ParseResult<T>` with typed `Data`.

```csharp
ParseResult<string> result = V.String().SafeParseTyped("hello");

if (result.Success)
{
    string value = result.Data!;   // no cast needed
    Console.WriteLine(value.ToUpper());
}
```

```csharp
ParseResult<Dictionary<string, object?>> userResult = V.Object(
    new Dictionary<string, Schema>
    {
        ["name"] = V.String(),
        ["age"] = V.Int(),
    }
).SafeParseTyped(new Dictionary<string, object?>
{
    ["name"] = "Alice",
    ["age"] = 30,
});

if (userResult.Success)
{
    Dictionary<string, object?> user = userResult.Data!;
    Console.WriteLine(user["name"]);
}
```

## Defaults and Coercion

### Defaults

Set a default value that is materialized when the input is absent.

```csharp
var role = V.String().Default("user");
var count = V.Int().Default(0);
```

Defaults are applied inside the parse pipeline: if the input is absent and a default exists, the default is validated against the schema constraints before being returned.

### Coercion

Enable automatic type coercion during parsing.

```csharp
var schema = V.String().Coerce();
var result = schema.Parse(42);  // coerces 42 to "42"
```

## Export and Import

Schemas can be exported to a canonical JSON document and imported back, enabling cross-language schema sharing.

### Export

```csharp
var schema = V.Object(new Dictionary<string, Schema>
{
    ["name"] = V.String().MinLength(1),
    ["age"] = V.Int().Min(0),
});

// Export to AnyValiDocument
AnyValiDocument doc = schema.Export();

// Or via the static helper
AnyValiDocument doc2 = V.Export(schema);

// With explicit mode
AnyValiDocument portable = schema.Export(ExportMode.Portable);
AnyValiDocument extended = schema.Export(ExportMode.Extended);
```

The exported `AnyValiDocument` contains:

- `AnyvaliVersion` -- protocol version (`"1.0"`)
- `SchemaVersion` -- schema format version (`"1"`)
- `Root` -- the schema node tree
- `Definitions` -- named reusable schema definitions
- `Extensions` -- language-specific extension metadata

### Import

```csharp
// Import an AnyValiDocument back into a live schema
Schema imported = V.Import(doc);

// Use it like any other schema
var result = imported.SafeParse(new Dictionary<string, object?>
{
    ["name"] = "Bob",
    ["age"] = 25,
});
```

### Cross-Language Example

Export a schema in C#, use it in another language:

```csharp
// C# side: define and export
var schema = V.Object(new Dictionary<string, Schema>
{
    ["id"] = V.Int64(),
    ["email"] = V.String().Format("email"),
});

AnyValiDocument doc = schema.Export(ExportMode.Portable);
// Serialize doc to JSON and send to another service
```

The resulting JSON document can be imported by any AnyVali SDK (JavaScript, Python, Go, Java, Kotlin, etc.).

## Type Inference

The C# SDK uses generic base classes to provide static type inference. `Schema<T>` subclasses carry their output type:

| Schema Class | Output Type (`T`) |
|---|---|
| `StringSchema` | `string` |
| `NumberSchema` | `double` |
| `IntSchema` | `long` |
| `BoolSchema` | `bool` |
| `ObjectSchema` | `Dictionary<string, object?>` |
| `ArraySchema` | `List<object?>` |

When you call `Parse()` on a typed schema, the return type is `T` -- no cast required:

```csharp
string name = V.String().Parse("Alice");          // returns string
long age = V.Int().Parse(30);                      // returns long
double price = V.Number().Parse(9.99);             // returns double
bool active = V.Bool().Parse(true);                // returns bool
Dictionary<string, object?> obj = V.Object(
    new Dictionary<string, Schema>
    {
        ["x"] = V.Int(),
    }
).Parse(new Dictionary<string, object?> { ["x"] = 1 });  // returns Dictionary<string, object?>
```

## Validation Issues

When parsing fails, you get a list of `ValidationIssue` objects with structured error information.

```csharp
var schema = V.Object(new Dictionary<string, Schema>
{
    ["name"] = V.String().MinLength(1),
    ["age"] = V.Int().Min(0).Max(150),
});

var result = schema.SafeParse(new Dictionary<string, object?>
{
    ["name"] = "",
    ["age"] = 200,
});

foreach (var issue in result.Issues)
{
    Console.WriteLine($"Path: {string.Join(".", issue.Path)}");
    Console.WriteLine($"Code: {issue.Code}");
    Console.WriteLine($"Message: {issue.Message}");
    Console.WriteLine($"Expected: {issue.Expected}");
    Console.WriteLine($"Received: {issue.Received}");
    Console.WriteLine();
}
// Path: name
// Code: too_small
// Message: String must have at least 1 character(s)
// Expected: 1
// Received: 0
//
// Path: age
// Code: too_large
// Message: Number must be <= 150
// Expected: 150
// Received: 200
```

### Issue Codes

| Code | Meaning |
|---|---|
| `invalid_type` | Value is the wrong type |
| `too_small` | Value is below minimum |
| `too_large` | Value is above maximum |
| `invalid_string` | String constraint violated (pattern, format, etc.) |
| `invalid_number` | Numeric constraint violated (multipleOf, etc.) |
| `required` | Required object property is missing |
| `unknown_key` | Object has an unrecognized key |
| `coercion_failed` | Coercion could not convert the value |
| `default_invalid` | Default value failed schema validation |

## API Reference

### V (Static Factory)

| Method | Returns | Description |
|---|---|---|
| `V.String()` | `StringSchema` | String schema |
| `V.Number()` | `NumberSchema` | Float64 number schema |
| `V.Float32()` | `Float32Schema` | Float32 schema |
| `V.Float64()` | `Float64Schema` | Float64 schema |
| `V.Int()` | `IntSchema` | Int64 integer schema |
| `V.Int8()` | `Int8Schema` | Signed 8-bit integer |
| `V.Int16()` | `Int16Schema` | Signed 16-bit integer |
| `V.Int32()` | `Int32Schema` | Signed 32-bit integer |
| `V.Int64()` | `Int64Schema` | Signed 64-bit integer |
| `V.Uint8()` | `Uint8Schema` | Unsigned 8-bit integer |
| `V.Uint16()` | `Uint16Schema` | Unsigned 16-bit integer |
| `V.Uint32()` | `Uint32Schema` | Unsigned 32-bit integer |
| `V.Uint64()` | `Uint64Schema` | Unsigned 64-bit integer |
| `V.Bool()` | `BoolSchema` | Boolean schema |
| `V.Null()` | `NullSchema` | Null schema |
| `V.Any()` | `AnySchema` | Accepts any value |
| `V.Unknown()` | `UnknownSchema` | Accepts any value (stricter intent) |
| `V.Never()` | `NeverSchema` | Always fails |
| `V.Literal(value)` | `LiteralSchema` | Exact value match |
| `V.Enum(values)` | `EnumSchema` | One of the listed values |
| `V.Array(items)` | `ArraySchema` | Array with item schema |
| `V.Tuple(items)` | `TupleSchema` | Fixed-length typed array |
| `V.Object(shape, unknownKeys?)` | `ObjectSchema` | Object with property schemas |
| `V.Record(valueSchema)` | `RecordSchema` | String-keyed record |
| `V.Union(variants)` | `UnionSchema` | At least one variant must match |
| `V.Intersection(schemas)` | `IntersectionSchema` | All schemas must match |
| `V.Optional(inner)` | `OptionalSchema` | Value may be absent |
| `V.Nullable(inner)` | `NullableSchema` | Value may be null |
| `V.Parse(schema, input)` | `object?` | Parse via static helper |
| `V.SafeParse(schema, input)` | `ParseResult` | Safe parse via static helper |
| `V.Export(schema, mode?)` | `AnyValiDocument` | Export schema to document |
| `V.Import(doc)` | `Schema` | Import document to live schema |

### Schema (Base Class)

| Member | Signature | Description |
|---|---|---|
| `Parse` | `object? Parse(object? input)` | Parse or throw `ValidationError` |
| `SafeParse` | `ParseResult SafeParse(object? input)` | Parse without throwing |
| `Default` | `Schema Default(object? value)` | Set a default value |
| `Coerce` | `Schema Coerce(CoercionConfig? config)` | Enable coercion |
| `Export` | `AnyValiDocument Export(ExportMode mode)` | Export to document |

### Schema&lt;T&gt; (Generic Base)

| Member | Signature | Description |
|---|---|---|
| `Parse` | `T Parse(object? input)` | Typed parse or throw |
| `SafeParseTyped` | `ParseResult<T> SafeParseTyped(object? input)` | Typed safe parse |

### ParseResult

| Property | Type | Description |
|---|---|---|
| `Success` | `bool` | Whether parsing succeeded |
| `Data` | `object?` | Parsed value (null on failure) |
| `Issues` | `IReadOnlyList<ValidationIssue>` | List of issues (empty on success) |

### ParseResult&lt;T&gt;

| Property | Type | Description |
|---|---|---|
| `Success` | `bool` | Whether parsing succeeded |
| `Data` | `T?` | Typed parsed value (default on failure) |
| `Issues` | `IReadOnlyList<ValidationIssue>` | List of issues (empty on success) |

### ValidationIssue

| Property | Type | Description |
|---|---|---|
| `Code` | `string` | Issue code (e.g. `"invalid_type"`) |
| `Message` | `string` | Human-readable message |
| `Path` | `List<object>` | Path to the issue (strings and ints) |
| `Expected` | `string?` | What was expected |
| `Received` | `string?` | What was received |
| `Meta` | `Dictionary<string, object>?` | Optional metadata |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.MinLength(n)` | `int` | Minimum string length |
| `.MaxLength(n)` | `int` | Maximum string length |
| `.Pattern(p)` | `string` | Regex pattern |
| `.StartsWith(s)` | `string` | Required prefix |
| `.EndsWith(s)` | `string` | Required suffix |
| `.Includes(s)` | `string` | Required substring |
| `.Format(f)` | `string` | Built-in format validator |

### NumberSchema / IntSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.Min(n)` | `double` | Minimum value (inclusive) |
| `.Max(n)` | `double` | Maximum value (inclusive) |
| `.ExclusiveMin(n)` | `double` | Minimum value (exclusive) |
| `.ExclusiveMax(n)` | `double` | Maximum value (exclusive) |
| `.MultipleOf(n)` | `double` | Must be a multiple of this value |

### ArraySchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.MinItems(n)` | `int` | Minimum number of elements |
| `.MaxItems(n)` | `int` | Maximum number of elements |

### UnknownKeyMode Enum

| Value | Description |
|---|---|
| `UnknownKeyMode.Reject` | Reject unknown keys (default) |
| `UnknownKeyMode.Strip` | Remove unknown keys from output |
| `UnknownKeyMode.Allow` | Pass unknown keys through |

### ExportMode Enum

| Value | Description |
|---|---|
| `ExportMode.Portable` | Fails if schema contains non-portable features |
| `ExportMode.Extended` | Emits core schema plus extension namespaces |

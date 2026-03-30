# Java SDK Reference

## Installation

### Maven

```xml
<dependency>
    <groupId>com.anyvali</groupId>
    <artifactId>anyvali</artifactId>
    <version>0.0.1</version>
</dependency>
```

### Gradle

```groovy
implementation 'com.anyvali:anyvali:0.0.1'
```

## Quick Start

```java
import static com.anyvali.AnyVali.*;
import com.anyvali.*;

import java.util.Map;
import java.util.Set;

// Define a schema
ObjectSchema userSchema = object_(Map.of(
    "name", string().minLength(1).maxLength(100),
    "email", string().format("email"),
    "age", int_().min(0).max(150)
));

// Parse input (throws on failure)
Map<String, Object> user = userSchema.parse(Map.of(
    "name", "Alice",
    "email", "alice@example.com",
    "age", 30
));

// Safe parse (returns result object)
ParseResult<Map<String, Object>> result = userSchema.safeParse(Map.of(
    "name", "",
    "email", "bad"
));

if (!result.success()) {
    for (ValidationIssue issue : result.issues()) {
        System.out.printf("%s: [%s] %s%n",
            issue.path(), issue.code(), issue.message());
    }
}
```

## Schema Builders

All schemas are created through static methods on the `AnyVali` class. Use `import static com.anyvali.AnyVali.*` for concise builder syntax.

### Primitive Types

```java
import static com.anyvali.AnyVali.*;

var str = string();
var num = number();    // float64
var integer = int_();  // int64
var flag = bool_();
var nil = null_();
```

Note: `int_()`, `bool_()`, `null_()`, and `any_()` use trailing underscores to avoid conflicts with Java reserved words.

### Numeric Types

```java
// Floating point
var f32 = float32();
var f64 = float64();
var number = number();   // alias for float64

// Signed integers
var i = int_();          // int64 (default)
var i8 = int8();         // -128 to 127
var i16 = int16();       // -32,768 to 32,767
var i32 = int32();       // -2^31 to 2^31-1
var i64 = int64();       // -2^63 to 2^63-1

// Unsigned integers
var u8 = uint8();        // 0 to 255
var u16 = uint16();      // 0 to 65,535
var u32 = uint32();      // 0 to 4,294,967,295
var u64 = uint64();      // 0 to int64 max (capped for safety)
```

### Special Types

```java
var any = any_();        // accepts any value
var unk = unknown();     // accepts any value (stricter intent)
var nev = never();       // always fails
```

### Literal and Enum

```java
import java.util.List;

// Literal: must be exactly this value
var status = literal("active");

// Enum: must be one of these values
var role = enum_(List.of("admin", "user", "guest"));
```

### Collections

```java
// Array of strings
var tags = array(string());

// Tuple: fixed-length, each position typed
var point = tuple_(List.of(number(), number()));

// Record: string keys, all values share a schema
var scores = record(int_());
```

### Object

```java
// All properties required by default
var user = object_(Map.of(
    "name", string().minLength(1),
    "email", string().format("email"),
    "age", int_().min(0)
));

// Specify which fields are required
var flexUser = object_(
    Map.of(
        "name", string().minLength(1),
        "email", string().format("email"),
        "age", int_().min(0)
    ),
    Set.of("name", "email")  // age is optional
);

// Control unknown key handling
var looseUser = object_(
    Map.of("name", string()),
    Set.of("name"),
    UnknownKeyMode.STRIP     // silently remove unknown keys
);
```

### Composition

```java
// Union: value must match at least one variant
var stringOrInt = union(List.of(string(), int_()));

// Intersection: value must match all schemas
var named = intersection(List.of(
    object_(Map.of("name", string())),
    object_(Map.of("age", int_()))
));
```

### Modifiers

```java
// Optional: field may be absent
var maybeAge = optional(int_());

// Nullable: field may be null
var nullableName = nullable(string());

// Chain .optional() or .nullable() on any schema
var schema = string().optional();
var schema2 = int_().nullable();
```

### References

```java
// Reference a named definition (used with export/import)
var ref = ref("User");
```

## String Constraints

```java
var schema = string()
    .minLength(1)          // minimum character count
    .maxLength(255)        // maximum character count
    .pattern("^\\w+$")    // regex pattern
    .startsWith("hello")   // must start with prefix
    .endsWith(".com")      // must end with suffix
    .includes("@")         // must contain substring
    .format("email");      // built-in format validator
```

Each constraint method returns a new schema instance -- schemas are immutable.

### Built-in Formats

- `"email"` -- RFC 5322 email address
- `"uri"` -- absolute URI
- `"uuid"` -- UUID v4
- `"iso8601"` -- ISO 8601 date-time
- `"ipv4"` -- IPv4 address
- `"ipv6"` -- IPv6 address

```java
var email = string().format("email");
var uri = string().format("uri");
var id = string().format("uuid");
```

## Numeric Constraints

Constraints apply to all numeric schemas (`NumberSchema`, `IntSchema`, `Float32Schema`, `Float64Schema`).

```java
var price = number()
    .min(0)                // value >= 0
    .max(10000)            // value <= 10000
    .exclusiveMin(0)       // value > 0
    .exclusiveMax(10000)   // value < 10000
    .multipleOf(0.01);     // must be a multiple of 0.01

var port = int_()
    .min(1)
    .max(65535);

var evenNumber = int_().multipleOf(2);
```

## Parsing

### parse (Throwing)

`parse()` returns the validated and typed value or throws `ValidationError`.

```java
try {
    String name = string().parse("Alice");
    Long age = int_().parse(30);
    Double price = number().parse(9.99);
    Boolean active = bool_().parse(true);
    Map<String, Object> user = object_(Map.of(
        "name", string()
    )).parse(Map.of("name", "Alice"));
} catch (ValidationError e) {
    for (ValidationIssue issue : e.issues()) {
        System.out.printf("[%s] %s%n", issue.code(), issue.message());
    }
}
```

### safeParse (Non-Throwing)

`safeParse()` returns a `ParseResult<T>` record that never throws.

```java
ParseResult<String> result = string().safeParse(42);

if (result.success()) {
    System.out.println("Parsed: " + result.data());
} else {
    for (ValidationIssue issue : result.issues()) {
        System.out.printf("[%s] %s%n", issue.code(), issue.message());
        // [invalid_type] Expected string, received integer
    }
}
```

```java
ParseResult<Map<String, Object>> userResult = object_(Map.of(
    "name", string(),
    "age", int_()
)).safeParse(Map.of("name", "Alice", "age", 30));

if (userResult.success()) {
    Map<String, Object> user = userResult.data();
    System.out.println(user.get("name"));
}
```

## Defaults and Coercion

### Defaults

Set a default value that is materialized when input is absent.

```java
var role = string().withDefault("user");
var count = int_().withDefault(0);
```

When a field with a default is missing from object input, the default is validated against the schema before being returned.

### Coercion

Enable automatic type coercion during parsing.

```java
import com.anyvali.parse.CoercionConfig;

var schema = string().coerce(new CoercionConfig());
```

## Export and Import

### Export

```java
import static com.anyvali.AnyVali.*;
import java.util.Map;

var schema = object_(Map.of(
    "name", string().minLength(1),
    "age", int_().min(0)
));

// Export to Map
Map<String, Object> doc = exportSchema(schema);

// Export to JSON string
String json = exportSchemaJson(schema);
System.out.println(json);
// {
//   "anyvaliVersion": "1.0",
//   "schemaVersion": "1",
//   "root": {
//     "kind": "object",
//     "properties": { ... },
//     ...
//   },
//   "definitions": {},
//   "extensions": {}
// }

// Export via schema instance
Map<String, Object> doc2 = schema.export();
```

### Import

```java
// Import from a document map or JSON string
Schema<?> imported = importSchema(doc);

// Use it like any other schema
var result = imported.safeParse(Map.of("name", "Bob", "age", 25));
```

### Cross-Language Example

```java
// Java side: define and export
var schema = object_(Map.of(
    "id", int64(),
    "email", string().format("email")
));

String json = exportSchemaJson(schema);
// Send JSON to another service using Python, Go, C#, etc.
```

## Type Inference

The Java SDK uses `Schema<T>` generics to carry output types through the parse pipeline:

| Schema Class | Output Type (`T`) |
|---|---|
| `StringSchema` | `String` |
| `NumberSchema` | `Double` |
| `IntSchema` | `Long` |
| `BoolSchema` | `Boolean` |
| `ObjectSchema` | `Map<String, Object>` |
| `ArraySchema` | `List<Object>` |

When you call `parse()`, the return type is fully typed:

```java
String name = string().parse("Alice");          // returns String
Long age = int_().parse(30);                    // returns Long
Double price = number().parse(9.99);            // returns Double
Boolean active = bool_().parse(true);           // returns Boolean
Map<String, Object> obj = object_(Map.of(
    "x", int_()
)).parse(Map.of("x", 1));                       // returns Map<String, Object>
```

## Validation Issues

When parsing fails, structured `ValidationIssue` records describe each problem.

```java
var schema = object_(Map.of(
    "name", string().minLength(1),
    "age", int_().min(0).max(150)
));

var result = schema.safeParse(Map.of("name", "", "age", 200));

for (ValidationIssue issue : result.issues()) {
    System.out.println("Path: " + issue.path());
    System.out.println("Code: " + issue.code());
    System.out.println("Message: " + issue.message());
    System.out.println("Expected: " + issue.expected());
    System.out.println("Received: " + issue.received());
    System.out.println();
}
// Path: [name]
// Code: too_small
// Message: String must have at least 1 character(s)
// Expected: 1
// Received: 0
//
// Path: [age]
// Code: too_large
// Message: Number must be <= 150.0
// Expected: 150.0
// Received: 200.0
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

### AnyVali (Static Factory)

| Method | Returns | Description |
|---|---|---|
| `string()` | `StringSchema` | String schema |
| `number()` | `NumberSchema` | Float64 number schema |
| `float32()` | `Float32Schema` | Float32 schema |
| `float64()` | `Float64Schema` | Float64 schema |
| `int_()` | `IntSchema` | Int64 integer schema |
| `int8()` | `IntSchema.Int8Schema` | Signed 8-bit integer |
| `int16()` | `IntSchema.Int16Schema` | Signed 16-bit integer |
| `int32()` | `IntSchema.Int32Schema` | Signed 32-bit integer |
| `int64()` | `IntSchema.Int64Schema` | Signed 64-bit integer |
| `uint8()` | `IntSchema.Uint8Schema` | Unsigned 8-bit integer |
| `uint16()` | `IntSchema.Uint16Schema` | Unsigned 16-bit integer |
| `uint32()` | `IntSchema.Uint32Schema` | Unsigned 32-bit integer |
| `uint64()` | `IntSchema.Uint64Schema` | Unsigned 64-bit integer |
| `bool_()` | `BoolSchema` | Boolean schema |
| `null_()` | `NullSchema` | Null schema |
| `any_()` | `AnySchema` | Accepts any value |
| `unknown()` | `UnknownSchema` | Accepts any value (stricter intent) |
| `never()` | `NeverSchema` | Always fails |
| `literal(value)` | `LiteralSchema` | Exact value match |
| `enum_(values)` | `EnumSchema` | One of the listed values |
| `array(items)` | `ArraySchema` | Array with item schema |
| `tuple_(items)` | `TupleSchema` | Fixed-length typed list |
| `object_(properties)` | `ObjectSchema` | Object (all keys required) |
| `object_(properties, required)` | `ObjectSchema` | Object with explicit required set |
| `object_(properties, required, unknownKeys)` | `ObjectSchema` | Object with unknown key mode |
| `record(valueSchema)` | `RecordSchema` | String-keyed record |
| `union(schemas)` | `UnionSchema` | At least one schema must match |
| `intersection(schemas)` | `IntersectionSchema` | All schemas must match |
| `optional(schema)` | `OptionalSchema` | Value may be absent |
| `nullable(schema)` | `NullableSchema` | Value may be null |
| `ref(reference)` | `RefSchema` | Reference to a named definition |
| `exportSchema(schema)` | `Map<String, Object>` | Export to document map |
| `exportSchemaJson(schema)` | `String` | Export to JSON string |
| `importSchema(source)` | `Schema<?>` | Import document to live schema |

### Schema&lt;T&gt; (Base Class)

| Member | Signature | Description |
|---|---|---|
| `parse` | `T parse(Object input)` | Parse or throw `ValidationError` |
| `safeParse` | `ParseResult<T> safeParse(Object input)` | Parse without throwing |
| `optional` | `OptionalSchema optional()` | Wrap as optional |
| `nullable` | `NullableSchema nullable()` | Wrap as nullable |
| `withDefault` | `Schema<T> withDefault(Object value)` | Set a default value |
| `coerce` | `Schema<T> coerce(CoercionConfig config)` | Enable coercion |
| `export` | `Map<String, Object> export()` | Export to document map |
| `export` | `Map<String, Object> export(ExportMode mode)` | Export with explicit mode |

### ParseResult&lt;T&gt; (Record)

| Accessor | Type | Description |
|---|---|---|
| `success()` | `boolean` | Whether parsing succeeded |
| `data()` | `T` | Parsed value (null on failure) |
| `issues()` | `List<ValidationIssue>` | List of issues (empty on success) |

### ValidationIssue (Record)

| Accessor | Type | Description |
|---|---|---|
| `code()` | `String` | Issue code (e.g. `"invalid_type"`) |
| `message()` | `String` | Human-readable message |
| `path()` | `List<Object>` | Path to the issue (strings and ints) |
| `expected()` | `Object` | What was expected |
| `received()` | `Object` | What was received |
| `meta()` | `Map<String, Object>` | Optional metadata |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.minLength(n)` | `int` | Minimum string length |
| `.maxLength(n)` | `int` | Maximum string length |
| `.pattern(p)` | `String` | Regex pattern |
| `.startsWith(s)` | `String` | Required prefix |
| `.endsWith(s)` | `String` | Required suffix |
| `.includes(s)` | `String` | Required substring |
| `.format(f)` | `String` | Built-in format validator |

### NumberSchema / IntSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `double` | Minimum value (inclusive) |
| `.max(n)` | `double` | Maximum value (inclusive) |
| `.exclusiveMin(n)` | `double` | Minimum value (exclusive) |
| `.exclusiveMax(n)` | `double` | Maximum value (exclusive) |
| `.multipleOf(n)` | `double` | Must be a multiple of this value |

### UnknownKeyMode Enum

| Value | Description |
|---|---|
| `UnknownKeyMode.REJECT` | Reject unknown keys (default) |
| `UnknownKeyMode.STRIP` | Remove unknown keys from output |
| `UnknownKeyMode.ALLOW` | Pass unknown keys through |

### ExportMode Enum

| Value | Description |
|---|---|
| `ExportMode.PORTABLE` | Fails if schema contains non-portable features |
| `ExportMode.EXTENDED` | Emits core schema plus extension namespaces |

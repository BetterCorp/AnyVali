# Kotlin SDK Reference

## Installation

### Gradle (Kotlin DSL)

```kotlin
dependencies {
    implementation("com.anyvali:anyvali:0.1.0")
}
```

### Gradle (Groovy)

```groovy
dependencies {
    implementation 'com.anyvali:anyvali:0.1.0'
}
```

The Kotlin SDK depends on `kotlinx.serialization.json` for JSON interchange.

## Quick Start

```kotlin
import com.anyvali.*

// Define a schema
val userSchema = obj(
    properties = mapOf(
        "name" to string().minLength(1).maxLength(100),
        "email" to string().format("email"),
        "age" to int_().min(0).max(150)
    ),
    required = setOf("name", "email")
)

// Parse input (throws on failure)
val user = userSchema.parse(mapOf(
    "name" to "Alice",
    "email" to "alice@example.com",
    "age" to 30
))

// Safe parse (returns sealed result)
val result = userSchema.safeParse(mapOf(
    "name" to "",
    "email" to "bad"
))

when (result) {
    is ParseResult.Success -> println("Parsed: ${result.value}")
    is ParseResult.Failure -> {
        for (issue in result.issues) {
            println("${issue.path.joinToString(".")}: [${issue.code}] ${issue.message}")
        }
    }
}
```

## Schema Builders

All schemas are created through top-level functions in the `com.anyvali` package.

### Primitive Types

```kotlin
import com.anyvali.*

val str = string()
val num = number()    // float64
val integer = int_()  // int64
val flag = bool()
val nil = null_()
```

Note: `int_()`, `null_()`, and `any_()` use trailing underscores to avoid conflicts with Kotlin keywords and built-in names.

### Numeric Types

```kotlin
// Floating point
val f32 = float32()
val f64 = float64()
val number = number()   // alias for float64

// Signed integers
val i = int_()          // int64 (default)
val i8 = int8()         // -128 to 127
val i16 = int16()       // -32,768 to 32,767
val i32 = int32()       // -2^31 to 2^31-1
val i64 = int64()       // -2^63 to 2^63-1

// Unsigned integers
val u8 = uint8()        // 0 to 255
val u16 = uint16()      // 0 to 65,535
val u32 = uint32()      // 0 to 4,294,967,295
val u64 = uint64()      // 0 to Long.MAX_VALUE (capped for safety)
```

### Special Types

```kotlin
val any = any_()        // accepts any value
val unk = unknown()     // accepts any value (stricter intent)
val nev = never()       // always fails
```

### Literal and Enum

```kotlin
// Literal: must be exactly this value
val status = literal("active")

// Enum: must be one of these values (vararg)
val role = enum_("admin", "user", "guest")
```

### Collections

```kotlin
// Array of strings
val tags = array(string())

// Tuple: fixed-length, each position typed (vararg)
val point = tuple(number(), number())

// Record: string keys, all values share a schema
val scores = record(int_())
```

### Object

```kotlin
// Object with explicit required set
val user = obj(
    properties = mapOf(
        "name" to string().minLength(1),
        "email" to string().format("email"),
        "age" to int_().min(0)
    ),
    required = setOf("name", "email")   // age is optional
)

// All parameters have defaults
val minimal = obj(
    properties = mapOf("id" to int_()),
    required = setOf("id")
)
```

Control how unknown keys are handled. The `unknownKeys` parameter controls how keys not declared in the properties are handled:

| Mode | Behavior |
|---|---|
| `UnknownKeyMode.REJECT` (default) | Produces an `unknown_key` issue for each extra key |
| `UnknownKeyMode.STRIP` | Silently removes extra keys from the output |
| `UnknownKeyMode.ALLOW` | Passes extra keys through to the output |

```kotlin
// Strip unknown keys silently
val loose = obj(
    properties = mapOf("id" to int_()),
    required = setOf("id"),
    unknownKeys = UnknownKeyMode.STRIP
)

// Allow unknown keys in output
val passthrough = obj(
    properties = mapOf("id" to int_()),
    required = setOf("id"),
    unknownKeys = UnknownKeyMode.ALLOW
)
```

### Composition

```kotlin
// Union: value must match at least one variant (vararg)
val stringOrInt = union(string(), int_())

// Intersection: value must match all schemas (vararg)
val named = intersection(
    obj(mapOf("name" to string()), setOf("name")),
    obj(mapOf("age" to int_()), setOf("age"))
)
```

### Modifiers

```kotlin
// Optional: field may be absent
val maybeAge = optional(int_())

// Nullable: field may be null
val nullableName = nullable(string())
```

### References

```kotlin
// Reference a named definition (used with export/import)
val userRef = ref("User")
```

## String Constraints

```kotlin
val schema = string()
    .minLength(1)          // minimum character count
    .maxLength(255)        // maximum character count
    .pattern("^\\w+$")    // regex pattern
    .startsWith("hello")   // must start with prefix
    .endsWith(".com")      // must end with suffix
    .includes("@")         // must contain substring
    .format("email")       // built-in format validator
```

Each constraint method returns a new `StringSchema` instance -- schemas are immutable data classes that use Kotlin's `copy()`.

### Built-in Formats

- `"email"` -- RFC 5322 email address
- `"uri"` -- absolute URI
- `"uuid"` -- UUID v4
- `"iso8601"` -- ISO 8601 date-time
- `"ipv4"` -- IPv4 address
- `"ipv6"` -- IPv6 address

```kotlin
val email = string().format("email")
val uri = string().format("uri")
val id = string().format("uuid")
```

## Numeric Constraints

Constraints apply to all numeric schemas (`NumberSchema`, `IntSchema`).

```kotlin
val price = number()
    .min(0)                // value >= 0
    .max(10000)            // value <= 10000
    .exclusiveMin(0)       // value > 0
    .exclusiveMax(10000)   // value < 10000
    .multipleOf(0.01)      // must be a multiple of 0.01

val port = int_()
    .min(1)
    .max(65535)

val evenNumber = int_().multipleOf(2)
```

`NumberSchema` constraints accept `Number` parameters (converted to `Double` internally). `IntSchema` constraints accept `Long` parameters.

## Array Constraints

```kotlin
val tags = array(string())
    .minItems(1)    // at least 1 element
    .maxItems(10)   // at most 10 elements
```

## Parsing

### parse (Throwing)

`parse()` returns the validated and typed value or throws `ValidationError`.

```kotlin
try {
    val name: String = string().parse("Alice")
    val age: Long = int_().parse(30)
    val price: Double = number().parse(9.99)
    val active: Boolean = bool().parse(true)
    val user: Map<String, Any?> = obj(
        mapOf("name" to string()),
        setOf("name")
    ).parse(mapOf("name" to "Alice"))
} catch (e: ValidationError) {
    for (issue in e.issues) {
        println("[${issue.code}] ${issue.message}")
    }
}
```

You can also parse with named definitions for schemas that use `ref()`:

```kotlin
val result = schema.parse(input, definitions = mapOf("User" to userSchema))
```

### safeParse (Non-Throwing)

`safeParse()` returns a `ParseResult<T>` sealed class that never throws.

```kotlin
val result = string().safeParse(42)

when (result) {
    is ParseResult.Success -> println("Parsed: ${result.value}")
    is ParseResult.Failure -> {
        for (issue in result.issues) {
            println("[${issue.code}] ${issue.message}")
            // [invalid_type] Expected string, received number
        }
    }
}
```

### ParseResult Utilities

```kotlin
val result = string().safeParse("hello")

// Check status
result.isSuccess   // true
result.isFailure   // false

// Get value or null
val value: String? = result.getOrNull()

// Get value or throw
val value2: String = result.getOrThrow()

// Get issues (empty list on success)
val issues: List<ValidationIssue> = result.issuesOrEmpty()
```

## Defaults and Coercion

### Defaults

Defaults fill in missing (absent) values. They run after coercion and before validation. Call `.default(value)` on any schema.

The default only applies when the value is absent -- if a value is present, it is validated normally. Defaults must be static values (for portability across SDKs).

```kotlin
val role = string().default("user")
role.parse(null)     // => "user" (absent value filled)
role.parse("admin")  // => "admin"

val count = int_().default(0L)
```

If the default value itself fails validation, a `default_invalid` issue is produced.

### Coercion

Enable string-to-type coercion during parsing.

```kotlin
val schema = int_().coerce("string->int")
val result = schema.parse("42")  // coerces "42" to 42L

val numSchema = number().coerce("string->number")
val result2 = numSchema.parse("3.14")  // coerces "3.14" to 3.14

val strSchema = string().coerce("trim", "lower")
val result3 = strSchema.parse("  HELLO  ")  // coerces to "hello"
```

String coercion supports: `"trim"`, `"lower"`, `"upper"`.

## Export and Import

### Export

```kotlin
import com.anyvali.interchange.Exporter

val schema = obj(
    mapOf(
        "name" to string().minLength(1),
        "age" to int_().min(0)
    ),
    setOf("name", "age")
)

// Export to AnyValiDocument
val doc = schema.export()

// Export with explicit mode
val portable = schema.export(ExportMode.PORTABLE)

// Export to JSON string
val json = Exporter.exportToJson(schema)
println(json)
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

// Export with definitions and extensions
val fullDoc = Exporter.exportDocument(
    schema = schema,
    definitions = mapOf("Address" to addressSchema),
    extensions = emptyMap()
)
```

The exported `AnyValiDocument` contains:

- `anyvaliVersion` -- protocol version (`"1.0"`)
- `schemaVersion` -- schema format version (`"1"`)
- `root` -- the root schema node as `JsonObject`
- `definitions` -- named reusable schema definitions
- `extensions` -- language-specific extension metadata

### Import

```kotlin
// Import from a JSON string
val (imported, definitions) = import_(jsonString)

// Use the imported schema
val result = imported.safeParse(mapOf("name" to "Bob", "age" to 25))

// Parse with imported definitions (for ref schemas)
val result2 = imported.parse(input, definitions)
```

The `import_()` function returns a `Pair<Schema<*>, Map<String, Schema<*>>>` containing the root schema and any named definitions.

### Cross-Language Example

```kotlin
// Kotlin side: define and export
val schema = obj(
    mapOf(
        "id" to int64(),
        "email" to string().format("email")
    ),
    setOf("id", "email")
)

val json = Exporter.exportToJson(schema)
// Send JSON to another service using Python, Go, C#, Java, etc.
```

## Type Inference

The Kotlin SDK uses `Schema<out T>` generics with covariant type parameters:

| Schema Class | Output Type (`T`) |
|---|---|
| `StringSchema` | `String` |
| `NumberSchema` | `Double` |
| `IntSchema` | `Long` |
| `BoolSchema` | `Boolean` |
| `ObjectSchema` | `Map<String, Any?>` |
| `ArraySchema` | `List<Any?>` |

When you call `parse()`, the return type is fully inferred:

```kotlin
val name: String = string().parse("Alice")          // returns String
val age: Long = int_().parse(30)                     // returns Long
val price: Double = number().parse(9.99)             // returns Double
val active: Boolean = bool().parse(true)             // returns Boolean
val obj: Map<String, Any?> = obj(
    mapOf("x" to int_()),
    setOf("x")
).parse(mapOf("x" to 1))                            // returns Map<String, Any?>
```

`safeParse()` returns `ParseResult<T>` where `T` is the schema's output type:

```kotlin
val result: ParseResult<String> = string().safeParse("hello")
```

## Validation Issues

When parsing fails, structured `ValidationIssue` data classes describe each problem.

```kotlin
val schema = obj(
    mapOf(
        "name" to string().minLength(1),
        "age" to int_().min(0).max(150)
    ),
    setOf("name", "age")
)

val result = schema.safeParse(mapOf("name" to "", "age" to 200))

if (result is ParseResult.Failure) {
    for (issue in result.issues) {
        println("Path: ${issue.path}")
        println("Code: ${issue.code}")
        println("Message: ${issue.message}")
        println("Expected: ${issue.expected}")
        println("Received: ${issue.received}")
        println()
    }
}
// Path: [name]
// Code: too_small
// Message:
// Expected: 1
// Received: 0
//
// Path: [age]
// Code: too_large
// Message:
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
| `custom_validation_not_portable` | Schema uses non-portable custom validators |

## Common Patterns

### Validating Environment Variables

Use `UnknownKeyMode.STRIP` when parsing maps that contain many extra keys you don't care about, like environment variables:

```kotlin
val envSchema = obj(
    mapOf("DATABASE_URL" to string()),
    required = setOf("DATABASE_URL"),
    unknownKeys = UnknownKeyMode.STRIP
)
```

Without `STRIP`, parse would fail with `unknown_key` issues for every other variable in the environment (PATH, HOME, etc.) because the default mode is `REJECT`.

| Mode | What happens with extra keys |
|---|---|
| `REJECT` (default) | Parse fails with `unknown_key` issues |
| `STRIP` | Extra keys silently removed from output |
| `ALLOW` | Extra keys passed through to output |

### Eagerly Evaluated vs Lazy Defaults

`.default()` accepts any value of the correct type. Expressions like `System.getProperty("user.dir")` are evaluated immediately when the schema is created and stored as a static value -- this works fine. What AnyVali does not support is lazy lambda defaults that re-evaluate on each parse call. If you need a fresh value on every parse, apply it after:

```kotlin
val configSchema = obj(
    mapOf(
        "profile" to string().default("default"),
        "appDir" to optional(string())
    ),
    required = setOf("profile"),
    unknownKeys = UnknownKeyMode.STRIP
)

val config = configSchema.parse(data).toMutableMap()
if (config["appDir"] == null) {
    config["appDir"] = System.getProperty("user.dir")
}
```

This keeps the schema fully portable -- the same JSON document can be imported in Go, Python, or any other SDK without relying on language-specific function calls.

## API Reference

### Top-Level Builder Functions

| Function | Returns | Description |
|---|---|---|
| `string()` | `StringSchema` | String schema |
| `number()` | `NumberSchema` | Float64 number schema |
| `float32()` | `NumberSchema` | Float32 schema |
| `float64()` | `NumberSchema` | Float64 schema |
| `int_()` | `IntSchema` | Int64 integer schema |
| `int8()` | `IntSchema` | Signed 8-bit integer |
| `int16()` | `IntSchema` | Signed 16-bit integer |
| `int32()` | `IntSchema` | Signed 32-bit integer |
| `int64()` | `IntSchema` | Signed 64-bit integer |
| `uint8()` | `IntSchema` | Unsigned 8-bit integer |
| `uint16()` | `IntSchema` | Unsigned 16-bit integer |
| `uint32()` | `IntSchema` | Unsigned 32-bit integer |
| `uint64()` | `IntSchema` | Unsigned 64-bit integer |
| `bool()` | `BoolSchema` | Boolean schema |
| `null_()` | `NullSchema` | Null schema |
| `any_()` | `AnySchema` | Accepts any value |
| `unknown()` | `UnknownSchema` | Accepts any value (stricter intent) |
| `never()` | `NeverSchema` | Always fails |
| `literal(value)` | `LiteralSchema` | Exact value match |
| `enum_(values)` | `EnumSchema` | One of the listed values (vararg) |
| `array(items)` | `ArraySchema` | Array with item schema |
| `tuple(elements)` | `TupleSchema` | Fixed-length typed list (vararg) |
| `obj(properties, required, unknownKeys)` | `ObjectSchema` | Object with property schemas |
| `record(values)` | `RecordSchema` | String-keyed record |
| `union(variants)` | `UnionSchema` | At least one variant must match (vararg) |
| `intersection(schemas)` | `IntersectionSchema` | All schemas must match (vararg) |
| `optional(schema)` | `OptionalSchema` | Value may be absent |
| `nullable(schema)` | `NullableSchema` | Value may be null |
| `ref(ref)` | `RefSchema` | Reference to a named definition |
| `import_(jsonStr)` | `Pair<Schema<*>, Map<String, Schema<*>>>` | Import JSON to schema + definitions |

### Schema&lt;out T&gt; (Base Class)

| Member | Signature | Description |
|---|---|---|
| `kind` | `val kind: String` | Schema kind identifier |
| `parse` | `fun parse(input: Any?): T` | Parse or throw `ValidationError` |
| `parse` | `fun parse(input: Any?, definitions: Map<String, Schema<*>>): T` | Parse with definitions |
| `safeParse` | `fun safeParse(input: Any?): ParseResult<T>` | Parse without throwing |
| `safeParse` | `fun safeParse(input: Any?, definitions: Map<String, Schema<*>>): ParseResult<T>` | Parse with definitions |
| `export` | `fun export(mode: ExportMode = ExportMode.PORTABLE): AnyValiDocument` | Export to document |

### ParseResult&lt;out T&gt; (Sealed Class)

| Subclass | Properties | Description |
|---|---|---|
| `ParseResult.Success<T>` | `val value: T` | Successful parse result |
| `ParseResult.Failure` | `val issues: List<ValidationIssue>` | Failed parse result |

| Property / Method | Type | Description |
|---|---|---|
| `isSuccess` | `Boolean` | Whether parsing succeeded |
| `isFailure` | `Boolean` | Whether parsing failed |
| `getOrNull()` | `T?` | Value or null on failure |
| `getOrThrow()` | `T` | Value or throw `ValidationError` |
| `issuesOrEmpty()` | `List<ValidationIssue>` | Issues (empty on success) |

### ValidationIssue (Data Class)

| Property | Type | Description |
|---|---|---|
| `code` | `String` | Issue code (e.g. `"invalid_type"`) |
| `message` | `String` | Human-readable message (default `""`) |
| `path` | `List<Any>` | Path to the issue (strings and ints) |
| `expected` | `String` | What was expected (default `""`) |
| `received` | `String` | What was received (default `""`) |
| `meta` | `Map<String, Any?>` | Optional metadata (default empty) |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.minLength(n)` | `Int` | Minimum string length |
| `.maxLength(n)` | `Int` | Maximum string length |
| `.pattern(p)` | `String` | Regex pattern |
| `.startsWith(s)` | `String` | Required prefix |
| `.endsWith(s)` | `String` | Required suffix |
| `.includes(s)` | `String` | Required substring |
| `.format(f)` | `String` | Built-in format validator |
| `.default(v)` | `String` | Set default value |
| `.coerce(c)` | `vararg String` | Enable coercions (e.g. `"trim"`, `"lower"`) |

### NumberSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `Number` | Minimum value (inclusive) |
| `.max(n)` | `Number` | Maximum value (inclusive) |
| `.exclusiveMin(n)` | `Number` | Minimum value (exclusive) |
| `.exclusiveMax(n)` | `Number` | Maximum value (exclusive) |
| `.multipleOf(n)` | `Number` | Must be a multiple of this value |
| `.default(v)` | `Number` | Set default value |
| `.coerce(c)` | `String` | Enable coercion (e.g. `"string->number"`) |

### IntSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `Long` | Minimum value (inclusive) |
| `.max(n)` | `Long` | Maximum value (inclusive) |
| `.exclusiveMin(n)` | `Long` | Minimum value (exclusive) |
| `.exclusiveMax(n)` | `Long` | Maximum value (exclusive) |
| `.multipleOf(n)` | `Long` | Must be a multiple of this value |
| `.default(v)` | `Long` | Set default value |
| `.coerce(c)` | `String` | Enable coercion (e.g. `"string->int"`) |

### ArraySchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `.minItems(n)` | `Int` | Minimum number of elements |
| `.maxItems(n)` | `Int` | Maximum number of elements |

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

### Exporter (Object)

| Method | Signature | Description |
|---|---|---|
| `exportToJson` | `fun exportToJson(schema: Schema<*>, mode: ExportMode = ExportMode.PORTABLE): String` | Export schema to JSON string |
| `exportDocument` | `fun exportDocument(schema: Schema<*>, definitions: Map<String, Schema<*>>, extensions: Map<String, JsonElement>, mode: ExportMode): AnyValiDocument` | Export with definitions and extensions |
| `toJsonString` | `fun toJsonString(doc: AnyValiDocument): String` | Serialize document to JSON |

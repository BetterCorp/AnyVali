# AnyVali Rust SDK Reference

The Rust SDK provides a native, idiomatic API for schema validation with full portable interchange support. Schemas are defined using builder functions, validated against `serde_json::Value` inputs, and can be exported to or imported from AnyVali's canonical JSON format.

## Installation

Add the dependency to your `Cargo.toml`:

```toml
[dependencies]
anyvali = "0.1"
serde_json = "1"
```

## Quick Start

```rust
use anyvali::*;
use serde_json::json;

fn main() {
    // Define a schema
    let user_schema = object()
        .field("id", Box::new(int64()))
        .field("name", Box::new(string().min_length(1).max_length(100)))
        .field("email", Box::new(string().format("email")))
        .field("age", Box::new(optional(Box::new(int().min(0.0).max(150.0)))))
        .field("role", Box::new(enum_(vec![
            json!("admin"), json!("user"), json!("guest"),
        ])))
        .required(vec!["id", "name", "email", "role"]);

    // Parse input (returns Result)
    let input = json!({
        "id": 42,
        "name": "Alice",
        "email": "alice@example.com",
        "role": "user"
    });

    match user_schema.parse(&input) {
        Ok(user) => println!("Valid: {}", user),
        Err(e) => {
            for issue in &e.issues {
                println!("[{}] {} at {:?}", issue.code, issue.expected, issue.path);
            }
        }
    }

    // SafeParse returns a ParseResult struct
    let bad_input = json!({
        "id": "not-a-number",
        "name": "",
        "role": "user"
    });

    let result = user_schema.safe_parse(&bad_input);
    if !result.success {
        for issue in &result.issues {
            let path: Vec<String> = issue.path.iter().map(|p| p.to_string()).collect();
            println!("{}: [{}]", path.join("."), issue.code);
        }
    }
}
```

## Type Inference

Rust's type system allows the SDK to provide compile-time type safety through the `TypedSchema` trait and the `parse_as` helper function.

### TypedSchema Trait

Concrete schema types implement `TypedSchema` with an associated `Output` type:

```rust
use anyvali::*;
use serde_json::json;

let schema = string().min_length(1);

// parse_typed returns Result<String, ValidationError> -- no cast needed
let name: String = schema.parse_typed(&json!("Alice")).unwrap();

let int_schema = int().min(0.0);
let age: i64 = int_schema.parse_typed(&json!(25)).unwrap();

let num_schema = number().min(0.0);
let score: f64 = num_schema.parse_typed(&json!(98.5)).unwrap();

let bool_schema = bool_();
let active: bool = bool_schema.parse_typed(&json!(true)).unwrap();
```

Built-in `TypedSchema` implementations:

| Schema | Output Type |
|---|---|
| `StringSchema` | `String` |
| `BoolSchema` | `bool` |
| `NumberSchema` | `f64` |
| `IntSchema` | `i64` |
| `NullSchema` | `()` |
| `AnySchema` | `serde_json::Value` |
| `UnknownSchema` | `serde_json::Value` |
| `ArraySchema` | `Vec<serde_json::Value>` |

### parse_as for Dynamic Schemas

When working with `dyn Schema` trait objects (e.g., after import), use `parse_as`:

```rust
use anyvali::*;
use serde_json::json;
use serde::Deserialize;

#[derive(Deserialize, Debug)]
struct User {
    name: String,
    email: String,
}

let schema = object()
    .field("name", Box::new(string()))
    .field("email", Box::new(string().format("email")))
    .required(vec!["name", "email"]);

let input = json!({"name": "Alice", "email": "alice@test.com"});

// Deserialize validated output directly to a Rust struct
let user: User = parse_as(&schema, &input).unwrap();
println!("{:?}", user);
```

`parse_as<T>` works with any type that implements `serde::de::DeserializeOwned`. It first validates through the schema, then deserializes the resulting `Value` into `T`.

## Schema Types

### Primitives

```rust
use anyvali::*;

// String
let s = string();

// Boolean (note trailing underscore to avoid keyword conflict)
let b = bool_();

// Null -- accepts only null
let n = null();
```

### Numeric Types

```rust
use anyvali::*;

// Number (float64, the safe cross-language default)
let num = number();

// Explicit float widths
let f64_schema = float64();
let f32_schema = float32();   // enforces float32 range

// Signed integers
let i   = int();       // int64
let i8  = int8();      // int8 range
let i16 = int16();     // int16 range
let i32_schema = int32();     // int32 range
let i64_schema = int64();     // int64 range

// Unsigned integers
let u8_schema  = uint8();     // uint8 range
let u16_schema = uint16();    // uint16 range
let u32_schema = uint32();    // uint32 range
let u64_schema = uint64();    // uint64 range
```

`number()` and `float64()` both produce a `NumberSchema`. `int()` and `int64()` both produce an `IntSchema` with int64 range. Narrower widths enforce their range automatically.

### Special Types

```rust
use anyvali::*;

// Any -- accepts any value, passes it through
let a = any();

// Unknown -- accepts any value (semantically "not yet validated")
let u = unknown();

// Never -- rejects all values
let n = never();
```

### Literal and Enum

```rust
use anyvali::*;
use serde_json::json;

// Literal -- matches a single exact value
let lit = literal(json!("active"));
let lit2 = literal(json!(42));
let lit3 = literal(json!(true));

// Enum -- matches one of several allowed values
let role = enum_(vec![json!("admin"), json!("user"), json!("guest")]);
let status = enum_(vec![json!(1), json!(2), json!(3)]);
```

### Array and Tuple

```rust
use anyvali::*;

// Array -- all items must match the given schema
let tags = array(Box::new(string()));

// Array with length constraints
let scores = array(Box::new(number())).min_items(1).max_items(100);

// Tuple -- fixed-length array with per-position schemas
let pair = tuple(vec![Box::new(string()), Box::new(int())]);
```

### Object

Objects use a builder pattern with `.field()` and `.required()`:

```rust
use anyvali::*;
use serde_json::json;

// All fields listed in required() are required; others are optional
let user = object()
    .field("name", Box::new(string().min_length(1)))
    .field("email", Box::new(string().format("email")))
    .field("age", Box::new(int().min(0.0)))
    .required(vec!["name", "email"]);

// Field with a default value
let config = object()
    .field("host", Box::new(string()))
    .field_with_default("port", Box::new(int()), json!(8080))
    .required(vec!["host"]);
```

### Record

```rust
use anyvali::*;

// Record -- validates a map where all values match a schema
let headers = record(Box::new(string()));

// Accepts: {"Content-Type": "text/html", "Accept": "application/json"}
```

### Composition

```rust
use anyvali::*;

// Union -- value must match at least one variant
let str_or_num = union(vec![Box::new(string()), Box::new(number())]);

// Intersection -- value must match all schemas
let base = object()
    .field("id", Box::new(int64()))
    .required(vec!["id"]);
let named = object()
    .field("name", Box::new(string()))
    .required(vec!["name"]);
let entity = intersection(vec![Box::new(base), Box::new(named)]);
```

### Modifiers

```rust
use anyvali::*;
use serde_json::json;

// Optional -- absent/null values are accepted
let opt = optional(Box::new(string()));

// Nullable -- null values are accepted
let nul = nullable(Box::new(string()));

// Ref -- reference to a named definition (for recursion)
let r = ref_("#/definitions/Node");
```

## Constraints

### String Constraints

```rust
let s = string()
    .min_length(1)           // minimum byte length
    .max_length(255)         // maximum byte length
    .pattern(r"^\w+$")      // regex pattern
    .starts_with("hello")   // must start with prefix
    .ends_with(".com")      // must end with suffix
    .includes("@")          // must contain substring
    .format("email");       // built-in format validator
```

Supported format values: `"email"`, `"url"`, `"uuid"`, `"ipv4"`, `"ipv6"`, `"date"`, `"date-time"`.

### Numeric Constraints

All numeric schemas (number, float64, float32, int variants) accept `f64` constraint values:

```rust
// Float schemas
let price = number()
    .min(0.0)                // value >= 0
    .max(999.99)             // value <= 999.99
    .exclusive_min(0.0)      // value > 0
    .exclusive_max(1000.0)   // value < 1000
    .multiple_of(0.01);      // must be a multiple of 0.01

// Int schemas also use f64 for constraint values
let age = int()
    .min(0.0)                // value >= 0
    .max(150.0)              // value <= 150
    .exclusive_min(-1.0)     // value > -1
    .exclusive_max(200.0)    // value < 200
    .multiple_of(1.0);       // must be a whole number multiple of 1
```

### Array Constraints

```rust
let items = array(Box::new(string()))
    .min_items(1)            // at least 1 item
    .max_items(50);          // at most 50 items
```

### Object Unknown Key Handling

```rust
use anyvali::*;

// Reject unknown keys (default)
let strict = object()
    .field("name", Box::new(string()))
    .required(vec!["name"])
    .unknown_keys(UnknownKeyMode::Reject);

// Strip unknown keys silently
let stripped = object()
    .field("name", Box::new(string()))
    .required(vec!["name"])
    .unknown_keys(UnknownKeyMode::Strip);

// Allow unknown keys to pass through
let loose = object()
    .field("name", Box::new(string()))
    .required(vec!["name"])
    .unknown_keys(UnknownKeyMode::Allow);
```

## Coercion and Defaults

### Coercion

Coercions transform values before validation. They are specified as string identifiers.

```rust
use anyvali::*;
use serde_json::json;

// Coerce strings before validation
let name = string().coerce(vec!["trim".to_string()]);
let tag = string().coerce(vec!["lower".to_string()]);
let code = string().coerce(vec!["upper".to_string()]);

// Multiple coercions in sequence
let normalized = string()
    .coerce(vec!["trim".to_string(), "lower".to_string()]);
```

Available coercion strings:

| Coercion | Effect |
|---|---|
| `"trim"` | Trim whitespace from string |
| `"lower"` | Lowercase string |
| `"upper"` | Uppercase string |
| `"string->number"` | Parse string to number |
| `"string->int"` | Parse string to integer |

### Defaults

Defaults fill in missing values. Specified as `serde_json::Value`:

```rust
use anyvali::*;
use serde_json::json;

let role = enum_(vec![json!("admin"), json!("user")])
    .default(json!("user"));

let active = bool_().default(json!(true));
let name = string().default(json!("Anonymous"));

// Object fields with defaults
let config = object()
    .field("host", Box::new(string()))
    .field_with_default("port", Box::new(int()), json!(8080))
    .field_with_default("tls", Box::new(bool_()), json!(false))
    .required(vec!["host"]);
```

## Export and Import

### Export

Convert a schema to the portable AnyVali JSON document:

```rust
use anyvali::*;
use std::collections::HashMap;

let schema = object()
    .field("name", Box::new(string().min_length(1)))
    .field("email", Box::new(string().format("email")))
    .required(vec!["name", "email"]);

// Export to AnyValiDocument
let definitions: HashMap<String, Box<dyn Schema>> = HashMap::new();
let doc = export(&schema, ExportMode::Portable, &definitions).unwrap();

// Serialize to JSON string
let json_str = serde_json::to_string_pretty(&doc).unwrap();
println!("{}", json_str);
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

| Mode | Variant | Behavior |
|---|---|---|
| Portable | `ExportMode::Portable` | Fails if schema uses non-portable features |
| Extended | `ExportMode::Extended` | Includes language-specific extensions |

### Import

Reconstruct a schema from a JSON document:

```rust
use anyvali::*;
use serde_json::json;

// Import from JSON string
let json_str = r#"{
    "anyvaliVersion": "1.0",
    "schemaVersion": "1",
    "root": { "kind": "string", "minLength": 1 },
    "definitions": {},
    "extensions": {}
}"#;

let (schema, ctx) = import(json_str).unwrap();

// Import from serde_json::Value
let doc_value = json!({
    "anyvaliVersion": "1.0",
    "schemaVersion": "1",
    "root": { "kind": "int", "min": 0 },
    "definitions": {},
    "extensions": {}
});
let (schema, ctx) = import_value(&doc_value).unwrap();

// Use the imported schema -- pass ctx for ref resolution
let result = schema.parse_with_context(&json!(42), &ctx).unwrap();
```

The import functions return a tuple of `(Box<dyn Schema>, ParseContext)`. The `ParseContext` carries resolved definitions needed for `ref` schemas. Use `parse_with_context` or `safe_parse_with_context` when working with imported schemas that may contain references.

## Error Handling

### parse (returning Result)

`parse` returns `Result<Value, ValidationError>`:

```rust
use anyvali::*;
use serde_json::json;

let schema = int().min(0.0);

match schema.parse(&json!(-5)) {
    Ok(value) => println!("Valid: {}", value),
    Err(e) => {
        println!("Validation failed with {} issue(s)", e.issues.len());
        for issue in &e.issues {
            let path: Vec<String> = issue.path.iter().map(|p| p.to_string()).collect();
            println!(
                "  [{}] expected={}, received={} at {}",
                issue.code, issue.expected, issue.received,
                path.join(".")
            );
        }
    }
}
```

### safe_parse (returning ParseResult)

`safe_parse` returns a `ParseResult` struct:

```rust
use anyvali::*;
use serde_json::json;

let schema = string().format("email");

let result = schema.safe_parse(&json!("not-an-email"));
if result.success {
    println!("Valid: {:?}", result.value);
} else {
    for issue in &result.issues {
        println!("[{}] {}", issue.code, issue.expected);
    }
}
```

### ValidationIssue

Each issue has a consistent structure:

```rust
pub struct ValidationIssue {
    pub code: String,              // machine-readable issue code
    pub path: Vec<PathSegment>,    // location in the input
    pub expected: String,          // what was expected
    pub received: String,          // what was received
    pub meta: Option<Value>,       // optional additional metadata
}

pub enum PathSegment {
    Key(String),    // object field key
    Index(usize),   // array index
}
```

### Issue Codes

| Constant | Code String | Meaning |
|---|---|---|
| `INVALID_TYPE` | `"invalid_type"` | Value has wrong type |
| `REQUIRED` | `"required"` | Required field is missing |
| `UNKNOWN_KEY` | `"unknown_key"` | Object has an undeclared key |
| `TOO_SMALL` | `"too_small"` | Below minimum (length, value, items) |
| `TOO_LARGE` | `"too_large"` | Above maximum (length, value, items) |
| `INVALID_STRING` | `"invalid_string"` | String constraint failed (pattern, format, etc.) |
| `INVALID_NUMBER` | `"invalid_number"` | Numeric constraint failed (multipleOf, range) |
| `INVALID_LITERAL` | `"invalid_literal"` | Literal/enum value mismatch |
| `INVALID_UNION` | `"invalid_union"` | No union variant matched |
| `COERCION_FAILED` | `"coercion_failed"` | Coercion could not convert the value |
| `DEFAULT_INVALID` | `"default_invalid"` | Default value fails validation |
| `UNSUPPORTED_SCHEMA_KIND` | `"unsupported_schema_kind"` | Unknown or unresolved schema kind |
| `UNSUPPORTED_EXTENSION` | `"unsupported_extension"` | Non-portable extension encountered |

## API Reference

### Builder Functions

| Function | Returns | Description |
|---|---|---|
| `string()` | `StringSchema` | String validator |
| `number()` | `NumberSchema` | Float64 validator (alias: `number`) |
| `float64()` | `NumberSchema` | Float64 validator |
| `float32()` | `NumberSchema` | Float32 validator with range check |
| `int()` | `IntSchema` | Int64 validator (alias: `int`) |
| `int8()` | `IntSchema` | Int8-range validator |
| `int16()` | `IntSchema` | Int16-range validator |
| `int32()` | `IntSchema` | Int32-range validator |
| `int64()` | `IntSchema` | Int64-range validator |
| `uint8()` | `IntSchema` | Uint8-range validator |
| `uint16()` | `IntSchema` | Uint16-range validator |
| `uint32()` | `IntSchema` | Uint32-range validator |
| `uint64()` | `IntSchema` | Uint64-range validator |
| `bool_()` | `BoolSchema` | Boolean validator |
| `null()` | `NullSchema` | Null validator |
| `any()` | `AnySchema` | Accepts any value |
| `unknown()` | `UnknownSchema` | Accepts any value (semantically unvalidated) |
| `never()` | `NeverSchema` | Rejects all values |
| `literal(value)` | `LiteralSchema` | Exact value match (`Value`) |
| `enum_(values)` | `EnumSchema` | One of allowed values (`Vec<Value>`) |
| `array(items)` | `ArraySchema` | Homogeneous array (`Box<dyn Schema>`) |
| `tuple(elements)` | `TupleSchema` | Fixed-length typed array (`Vec<Box<dyn Schema>>`) |
| `object()` | `ObjectSchema` | Structured object (use `.field()` to add properties) |
| `record(values)` | `RecordSchema` | String-keyed map with uniform values (`Box<dyn Schema>`) |
| `union(variants)` | `UnionSchema` | Matches any one variant (`Vec<Box<dyn Schema>>`) |
| `intersection(all_of)` | `IntersectionSchema` | Matches all schemas (`Vec<Box<dyn Schema>>`) |
| `optional(schema)` | `OptionalSchema` | Allows absent/null values (`Box<dyn Schema>`) |
| `nullable(schema)` | `NullableSchema` | Allows null values (`Box<dyn Schema>`) |
| `ref_(path)` | `RefSchema` | Reference to a definition (`&str`) |

### Schema Trait

```rust
pub trait Schema: SchemaClone + std::fmt::Debug + Send + Sync {
    fn kind(&self) -> &str;
    fn parse_value(&self, input: &Value, path: &[PathSegment], ctx: &ParseContext)
        -> Result<Value, Vec<ValidationIssue>>;
    fn parse(&self, input: &Value) -> Result<Value, ValidationError>;
    fn parse_with_context(&self, input: &Value, ctx: &ParseContext)
        -> Result<Value, ValidationError>;
    fn safe_parse(&self, input: &Value) -> ParseResult;
    fn safe_parse_with_context(&self, input: &Value, ctx: &ParseContext) -> ParseResult;
    fn export_node(&self) -> Value;
    fn has_custom_validators(&self) -> bool;
}
```

### TypedSchema Trait

```rust
pub trait TypedSchema: Schema {
    type Output: DeserializeOwned;
    fn parse_typed(&self, input: &Value) -> Result<Self::Output, ValidationError>;
}
```

### Generic Helper

```rust
pub fn parse_as<T: DeserializeOwned>(
    schema: &dyn Schema,
    input: &Value,
) -> Result<T, ValidationError>;
```

### Constraint Methods

**StringSchema** -- all consume `self` and return `Self` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `.min_length(n)` | `usize` | Minimum byte length |
| `.max_length(n)` | `usize` | Maximum byte length |
| `.pattern(p)` | `&str` | Regex pattern |
| `.starts_with(s)` | `&str` | Required prefix |
| `.ends_with(s)` | `&str` | Required suffix |
| `.includes(s)` | `&str` | Required substring |
| `.format(f)` | `&str` | Built-in format check |
| `.default(v)` | `Value` | Default value |
| `.coerce(c)` | `Vec<String>` | Coercion list |

**NumberSchema** -- all consume `self` and return `Self` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `f64` | Inclusive minimum |
| `.max(n)` | `f64` | Inclusive maximum |
| `.exclusive_min(n)` | `f64` | Exclusive minimum |
| `.exclusive_max(n)` | `f64` | Exclusive maximum |
| `.multiple_of(n)` | `f64` | Divisibility constraint |
| `.default(v)` | `Value` | Default value |
| `.coerce(c)` | `Vec<String>` | Coercion list |

**IntSchema** -- all consume `self` and return `Self` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `f64` | Inclusive minimum |
| `.max(n)` | `f64` | Inclusive maximum |
| `.exclusive_min(n)` | `f64` | Exclusive minimum |
| `.exclusive_max(n)` | `f64` | Exclusive maximum |
| `.multiple_of(n)` | `f64` | Divisibility constraint |
| `.default(v)` | `Value` | Default value |
| `.coerce(c)` | `Vec<String>` | Coercion list |

**ArraySchema** -- all consume `self` and return `Self` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `.min_items(n)` | `usize` | Minimum item count |
| `.max_items(n)` | `usize` | Maximum item count |

**ObjectSchema** -- all consume `self` and return `Self` for chaining:

| Method | Parameter | Description |
|---|---|---|
| `.field(name, schema)` | `&str`, `Box<dyn Schema>` | Add a property |
| `.field_with_default(name, schema, default)` | `&str`, `Box<dyn Schema>`, `Value` | Add a property with a default value |
| `.required(fields)` | `Vec<&str>` | Set which fields are required |
| `.unknown_keys(mode)` | `UnknownKeyMode` | How to handle undeclared keys |

### Export/Import Functions

```rust
pub fn export(
    schema: &dyn Schema,
    mode: ExportMode,
    definitions: &HashMap<String, Box<dyn Schema>>,
) -> Result<AnyValiDocument, String>;

pub fn import(json_str: &str) -> Result<(Box<dyn Schema>, ParseContext), String>;
pub fn import_value(value: &Value) -> Result<(Box<dyn Schema>, ParseContext), String>;
```

### Core Types

```rust
pub struct ParseResult {
    pub success: bool,
    pub value: Option<Value>,
    pub issues: Vec<ValidationIssue>,
}

pub struct ValidationIssue {
    pub code: String,
    pub path: Vec<PathSegment>,
    pub expected: String,
    pub received: String,
    pub meta: Option<Value>,
}

pub struct ValidationError {
    pub issues: Vec<ValidationIssue>,
}

pub enum PathSegment {
    Key(String),
    Index(usize),
}

pub struct AnyValiDocument {
    pub anyvali_version: String,
    pub schema_version: String,
    pub root: Value,
    pub definitions: serde_json::Map<String, Value>,
    pub extensions: serde_json::Map<String, Value>,
}

pub enum ExportMode {
    Portable,
    Extended,
}

pub enum UnknownKeyMode {
    Reject,   // default
    Strip,
    Allow,
}

pub struct ParseContext {
    pub definitions: HashMap<String, Box<dyn Schema>>,
}
```

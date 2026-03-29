# AnyVali Documentation

> Write schemas in your language. Share them everywhere.

This guide gets you from zero to validating in minutes. For deeper topics, see the linked guides at the bottom.

---

## Install

Pick your language:

```bash
# JavaScript / TypeScript
npm install anyvali

# Python
pip install anyvali

# Go
go get github.com/BetterCorp/AnyVali/sdk/go

# Java (Maven)
<dependency>
  <groupId>com.anyvali</groupId>
  <artifactId>anyvali</artifactId>
  <version>0.0.1</version>
</dependency>

# C# (.NET)
dotnet add package AnyVali

# Rust
cargo add anyvali

# PHP
composer require anyvali/anyvali

# Ruby
gem install anyvali

# Kotlin (Gradle)
implementation("com.anyvali:anyvali:0.0.1")

# C++ (CMake)
FetchContent_Declare(anyvali GIT_REPOSITORY https://github.com/BetterCorp/AnyVali)
```

---

## 1. Define a Schema

Schemas are built with simple builder functions. Every SDK uses the same concepts with language-appropriate naming.

### JavaScript / TypeScript

```typescript
import { string, number, int, object, array, optional } from 'anyvali';

const UserSchema = object({
  name: string().minLength(1).maxLength(100),
  email: string().format('email'),
  age: int().min(0).max(150).optional(),
  tags: array(string()).maxItems(10),
});
```

### Python

```python
import anyvali as v

user_schema = v.object_({
    "name": v.string().min_length(1).max_length(100),
    "email": v.string().format("email"),
    "age": v.int_().min(0).max(150).optional(),
    "tags": v.array(v.string()).max_items(10),
})
```

### Go

```go
import av "github.com/BetterCorp/AnyVali/sdk/go"

userSchema := av.Object(map[string]av.Schema{
    "name":  av.String().MinLength(1).MaxLength(100),
    "email": av.String().Format("email"),
    "age":   av.Optional(av.Int().Min(0).Max(150)),
    "tags":  av.Array(av.String()).MaxItems(10),
})
```

---

## 2. Parse Input

Two options: throw on failure, or get a result object.

### Throwing parse

```typescript
// JS — throws ValidationError if invalid
const user = UserSchema.parse(inputData);
```

```python
# Python — raises ValidationError if invalid
user = user_schema.parse(input_data)
```

```go
// Go — returns (value, error)
user, err := userSchema.Parse(inputData)
```

### Safe parse (recommended)

Returns a result with `success` + `data`, or `success: false` + `issues`.

```typescript
// JS
const result = UserSchema.safeParse(inputData);
if (result.success) {
  console.log(result.data);       // parsed value
} else {
  console.log(result.issues);     // array of issues
}
```

```python
# Python
result = user_schema.safe_parse(input_data)
if result.success:
    print(result.data)
else:
    for issue in result.issues:
        print(f"{issue.path}: [{issue.code}] {issue.message}")
```

```go
// Go
result := userSchema.SafeParse(inputData)
if result.Success {
    fmt.Println(result.Data)
} else {
    for _, issue := range result.Issues {
        fmt.Printf("%v: [%s] %s\n", issue.Path, issue.Code, issue.Message)
    }
}
```

---

## 3. Available Schema Types

| Type | Description | JS Builder | Python Builder |
|------|-------------|------------|----------------|
| `string` | Text | `string()` | `v.string()` |
| `number` | Float64 | `number()` | `v.number()` |
| `int` | Int64 | `int()` | `v.int_()` |
| `bool` | Boolean | `bool()` | `v.bool_()` |
| `null` | Null only | `null_()` | `v.null()` |
| `any` | Accepts anything | `any()` | `v.any_()` |
| `unknown` | Accepts anything (safer) | `unknown()` | `v.unknown()` |
| `never` | Rejects everything | `never()` | `v.never()` |
| `literal` | Exact value | `literal("admin")` | `v.literal("admin")` |
| `enum` | Set of values | `enum_(["a","b"])` | `v.enum_(["a","b"])` |
| `array` | Typed list | `array(string())` | `v.array(v.string())` |
| `tuple` | Fixed-length list | `tuple([string(), int()])` | `v.tuple_([v.string(), v.int_()])` |
| `object` | Typed object | `object({...})` | `v.object_({...})` |
| `record` | String-keyed map | `record(number())` | `v.record(v.number())` |
| `union` | One of several types | `union([string(), int()])` | `v.union([v.string(), v.int_()])` |
| `intersection` | All of several types | `intersection([...])` | `v.intersection([...])` |
| `optional` | Can be absent | `optional(string())` | `v.optional(v.string())` |
| `nullable` | Can be null | `nullable(string())` | `v.nullable(v.string())` |

### Numeric widths

Need a specific size? Use explicit types:

| Type | Range |
|------|-------|
| `int8` | -128 to 127 |
| `int16` | -32,768 to 32,767 |
| `int32` | -2.1B to 2.1B |
| `int64` | -9.2 quintillion to 9.2 quintillion |
| `uint8` | 0 to 255 |
| `uint16` | 0 to 65,535 |
| `uint32` | 0 to 4.2B |
| `uint64` | 0 to 18.4 quintillion |
| `float32` | ~7 decimal digits precision |
| `float64` | ~15 decimal digits precision |

```typescript
int8()    // JS
v.int8()  // Python
av.Int8() // Go
```

---

## 4. Constraints

### String constraints

```typescript
string()
  .minLength(1)           // at least 1 character
  .maxLength(100)         // at most 100 characters
  .pattern(/^[A-Z]+$/)   // must match regex
  .startsWith("hello")   // must start with
  .endsWith("!")          // must end with
  .includes("world")     // must contain
  .format("email")       // must be valid email
```

Available formats: `email`, `url`, `uuid`, `ipv4`, `ipv6`, `date`, `date-time`

### Numeric constraints

```typescript
number()
  .min(0)              // >= 0
  .max(100)            // <= 100
  .exclusiveMin(0)     // > 0
  .exclusiveMax(100)   // < 100
  .multipleOf(5)       // must be divisible by 5
```

### Array constraints

```typescript
array(string())
  .minItems(1)    // at least 1 item
  .maxItems(10)   // at most 10 items
```

### Object options

```typescript
// All fields are required by default
object({
  name: string(),
  age: int(),
})

// Make fields optional
object({
  name: string(),
  age: int().optional(),
})

// Unknown key handling (default: "reject")
object({ name: string() })                        // rejects unknown keys
object({ name: string() }, { unknownKeys: "strip" })  // silently removes unknown keys
object({ name: string() }, { unknownKeys: "allow" })  // keeps unknown keys
```

---

## 5. Defaults

Set a default value for absent fields. Defaults are applied before validation.

```typescript
// JS
const schema = object({
  role: string().default("user"),
  score: int().min(0).default(0),
});

schema.parse({});
// => { role: "user", score: 0 }

schema.parse({ role: "admin" });
// => { role: "admin", score: 0 }
```

```python
# Python
schema = v.object_({
    "role": v.string().default("user"),
    "score": v.int_().min(0).default(0),
})

schema.parse({})
# => {"role": "user", "score": 0}
```

Rules:
- Defaults are **only applied to absent values** (present values are never overwritten)
- The defaulted value **must still pass validation**
- Defaults must be **plain data** (no functions in portable schemas)

---

## 6. Coercions

Automatically convert input types before validation. Must be explicitly enabled.

```typescript
// JS — coerce string input to int
const schema = int().coerce({ toInt: true });
schema.parse("42");  // => 42
schema.parse("abc"); // => error: coercion_failed

// Trim whitespace
const trimmed = string().coerce({ trim: true });
trimmed.parse("  hello  "); // => "hello"

// Case normalization
const lower = string().coerce({ lower: true });
lower.parse("HELLO"); // => "hello"
```

```python
# Python
schema = v.int_().coerce(to_int=True)
schema.parse("42")  # => 42

trimmed = v.string().coerce(trim=True)
trimmed.parse("  hello  ")  # => "hello"
```

Available coercions:
- `string -> int` — parse string as integer
- `string -> number` — parse string as float
- `string -> bool` — "true"/"1"/"yes" -> True, "false"/"0"/"no" -> False
- `trim` — remove leading/trailing whitespace
- `lower` — convert to lowercase
- `upper` — convert to uppercase

Processing order: **coerce** (present values) -> **defaults** (absent values) -> **validate**

---

## 7. Export and Import (Cross-Language Sharing)

This is AnyVali's superpower. Define a schema in one language, use it in another.

### Export a schema to JSON

```typescript
// JS — export to portable JSON
const doc = UserSchema.export();

// Save to file, database, API response, etc.
const json = JSON.stringify(doc, null, 2);
```

```python
# Python
doc = user_schema.export()
json_str = json.dumps(doc, indent=2)
```

The exported JSON looks like:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": {
    "kind": "object",
    "properties": {
      "name": { "kind": "string", "minLength": 1, "maxLength": 100 },
      "email": { "kind": "string", "format": "email" }
    },
    "required": ["name", "email"],
    "unknownKeys": "reject"
  },
  "definitions": {},
  "extensions": {}
}
```

### Import a schema from JSON

```typescript
// JS — import from JSON
import { importSchema } from 'anyvali';
const schema = importSchema(doc);
schema.parse({ name: "Alice", email: "alice@test.com" });
```

```python
# Python — import from JSON
schema = v.import_schema(doc)
schema.parse({"name": "Alice", "email": "alice@test.com"})
```

```go
// Go — import from JSON
schema, err := av.ImportJSON(jsonBytes)
result := schema.SafeParse(input)
```

### Real-world example: Frontend to Backend

```typescript
// Frontend (TypeScript) — define and export
const FormSchema = object({
  username: string().minLength(3).maxLength(30).pattern(/^[a-zA-Z0-9_]+$/),
  email: string().format('email'),
  age: int().min(13).max(120),
});
const schemaJSON = JSON.stringify(FormSchema.export());
// Send schemaJSON to your API or save to a shared config
```

```python
# Backend (Python) — import and validate
import json, anyvali as v

schema_doc = json.loads(schema_json_from_frontend)
form_schema = v.import_schema(schema_doc)

# Same validation rules, no duplication
result = form_schema.safe_parse(request.json)
if not result.success:
    return {"errors": [{"path": i.path, "message": i.message} for i in result.issues]}, 400
```

---

## 8. Error Handling

When validation fails, you get structured issues with machine-readable codes.

### Issue structure

```json
{
  "code": "too_small",
  "message": "String must have at least 1 character",
  "path": ["users", 0, "name"],
  "expected": 1,
  "received": 0
}
```

- `code` — machine-readable error type (see table below)
- `message` — human-readable description
- `path` — where the error occurred (keys and array indexes)
- `expected` — what was expected
- `received` — what was actually provided

### Issue codes

| Code | When it happens |
|------|----------------|
| `invalid_type` | Wrong type (e.g., string where int expected) |
| `required` | Required field is missing |
| `unknown_key` | Object has a key not in the schema |
| `too_small` | Below min length/value/items |
| `too_large` | Above max length/value/items |
| `invalid_string` | String constraint failed (pattern, format, etc.) |
| `invalid_number` | Numeric constraint failed (multipleOf, etc.) |
| `invalid_literal` | Doesn't match expected literal value |
| `invalid_union` | Matches none of the union variants |
| `coercion_failed` | Type coercion didn't work |
| `default_invalid` | Default value fails validation |
| `custom_validation_not_portable` | Can't export custom validator |
| `unsupported_extension` | Required extension not available |
| `unsupported_schema_kind` | Schema kind not supported |

---

## 9. Recursive Schemas

Use `ref` and `definitions` for recursive or reusable types:

```typescript
// JS — recursive tree
import { object, string, array, ref } from 'anyvali';

// In exported JSON:
// {
//   "root": { "kind": "ref", "ref": "#/definitions/TreeNode" },
//   "definitions": {
//     "TreeNode": {
//       "kind": "object",
//       "properties": {
//         "value": { "kind": "string" },
//         "children": {
//           "kind": "array",
//           "items": { "kind": "ref", "ref": "#/definitions/TreeNode" }
//         }
//       }
//     }
//   }
// }
```

---

## 10. Portability Tiers

Not everything can cross language boundaries. AnyVali has three tiers:

| Tier | What | Can export? | Can import everywhere? |
|------|------|-------------|----------------------|
| **Portable core** | Built-in types, constraints, coercions, formats | Yes | Yes |
| **Core + extensions** | Core plus language-specific metadata | Yes (extended mode) | Only where extension is understood |
| **Local-only** | Custom validators, function defaults | No (portable mode fails) | N/A |

### Export modes

```typescript
// Portable mode (default) — fails if schema has non-portable features
schema.export();                      // or
schema.export({ mode: 'portable' });

// Extended mode — includes language-specific extensions
schema.export({ mode: 'extended' });
```

---

## API Quick Reference

### All builder functions

| Category | Builders |
|----------|----------|
| **Primitives** | `string()`, `bool()`, `null()` |
| **Numbers** | `number()`, `int()`, `float32()`, `float64()` |
| **Sized ints** | `int8()`, `int16()`, `int32()`, `int64()`, `uint8()`, `uint16()`, `uint32()`, `uint64()` |
| **Special** | `any()`, `unknown()`, `never()` |
| **Values** | `literal(value)`, `enum([values])` |
| **Collections** | `array(schema)`, `tuple([schemas])`, `object({fields})`, `record(valueSchema)` |
| **Composition** | `union([schemas])`, `intersection([schemas])` |
| **Modifiers** | `optional(schema)` or `.optional()`, `nullable(schema)` or `.nullable()` |

### All constraint methods

| On | Method | What it does |
|----|--------|-------------|
| string | `.minLength(n)` | Min character count |
| string | `.maxLength(n)` | Max character count |
| string | `.pattern(regex)` | Must match regex |
| string | `.startsWith(s)` | Must start with prefix |
| string | `.endsWith(s)` | Must end with suffix |
| string | `.includes(s)` | Must contain substring |
| string | `.format(fmt)` | Must match format (email, url, uuid, etc.) |
| number/int | `.min(n)` | Minimum value (inclusive) |
| number/int | `.max(n)` | Maximum value (inclusive) |
| number/int | `.exclusiveMin(n)` | Minimum value (exclusive) |
| number/int | `.exclusiveMax(n)` | Maximum value (exclusive) |
| number/int | `.multipleOf(n)` | Must be divisible by n |
| array | `.minItems(n)` | Min array length |
| array | `.maxItems(n)` | Max array length |

### All schema methods

| Method | What it does |
|--------|-------------|
| `.parse(input)` | Validate and return value, or throw |
| `.safeParse(input)` | Validate and return result object |
| `.export(options?)` | Export to JSON document |
| `.default(value)` | Set default for absent values |
| `.coerce(config)` | Enable type coercion |
| `.optional()` | Make the value optional |
| `.nullable()` | Allow null |

---

## Further Reading

- **[Numeric Semantics](https://docs.anyvali.com/docs/numeric-semantics)** — Why `number` = float64 and `int` = int64, and how to choose the right numeric type
- **[Portability Guide](https://docs.anyvali.com/docs/portability-guide)** — Design schemas that work across all 10 languages
- **[SDK Authors Guide](https://docs.anyvali.com/docs/sdk-authors-guide)** — How to implement a new AnyVali SDK
- **[Product Overview](https://docs.anyvali.com/docs/overview)** — Architecture, design decisions, and comparison with Zod/JSON Schema/etc.
- **[Canonical Spec](https://docs.anyvali.com/spec/spec)** — The normative specification
- **[JSON Format Spec](https://docs.anyvali.com/spec/json-format)** — The interchange format in detail
- **[Development Guide](https://docs.anyvali.com/docs/development)** — Building, testing, and contributing

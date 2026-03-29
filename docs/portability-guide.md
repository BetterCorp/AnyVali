# AnyVali Portability Guide

## Overview

AnyVali's core value proposition is that schemas move between languages. This guide explains how the portability system works, how to design schemas that transfer cleanly, and what to do when they cannot.

## The Three Portability Tiers

Every feature in AnyVali falls into one of three tiers.

### Tier 1: Portable Core

**Guaranteed import and export compatibility across all AnyVali SDKs.**

The portable core includes:

- All schema kinds (`string`, `int64`, `object`, `array`, `union`, etc.)
- All declarative constraints (`minLength`, `max`, `pattern`, etc.)
- All portable format validators (`email`, `url`, `uuid`, `ipv4`, `ipv6`, `date`, `date-time`)
- All portable coercions (`string -> int`, `string -> number`, `string -> bool`, whitespace trim, case normalization)
- Defaults (when the default value is representable in JSON)
- Definitions and references
- Unknown key modes (`reject`, `strip`, `allow`)

If your schema uses only Tier 1 features, it will export successfully in portable mode and import correctly in every AnyVali SDK.

### Tier 2: Portable Core + Extensions

**Compatible only where the relevant extension namespace is understood.**

Extensions allow SDKs to attach language-specific metadata to schemas. Examples:

- A Go SDK might add struct tag information
- A TypeScript SDK might add branded type markers
- A Java SDK might add annotation metadata

Extensions are namespaced in the exported JSON:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": { "kind": "string", "minLength": 1 },
  "definitions": {},
  "extensions": {
    "go": {
      "structTags": { "name": "json:\"name\" validate:\"required\"" }
    },
    "js": {
      "brandedType": "UserId"
    }
  }
}
```

A Tier 2 schema exports successfully in extended mode. It imports successfully in SDKs that understand its extension namespaces. SDKs that do not understand the extensions may still import the schema if the extensions are informational (safe to ignore).

### Tier 3: Local-Only

**Never part of canonical interchange.**

Local-only features include:

- Custom validators (arbitrary functions attached to schemas)
- Function-based defaults
- Async validation hooks
- Any feature that requires executable code

These features work within a single SDK but cannot be serialized. Attempting to export a schema with Tier 3 features in portable mode will fail. In extended mode, the schema exports with the Tier 3 features omitted (or the export fails if the feature is required for correctness).

## Designing Portable Schemas

### Use only portable schema kinds

Stick to the kinds defined in the AnyVali spec:

```typescript
// Portable -- these work everywhere
v.string()
v.int()          // int64
v.number()       // float64
v.bool()
v.object({...})
v.array(v.string())
v.union([v.string(), v.int()])
v.enum(["a", "b", "c"])
v.literal("active")
v.nullable(v.string())
v.optional(v.int())
v.tuple([v.string(), v.int()])
v.record(v.string(), v.int())
```

### Use only portable constraints

All constraints in the portable core are declarative and data-only:

```python
# Portable constraints
name = v.string().min_length(1).max_length(255)
age = v.int().min(0).max(150)
email = v.string().format("email")
code = v.string().pattern(r"^[A-Z]{3}-\d{4}$")
items = v.array(v.string()).min_items(1).max_items(100)
price = v.number().min(0).exclusive_min(True)
```

### Use JSON-representable defaults

Defaults must be pure data that can be serialized to JSON:

```go
// Portable defaults
role := v.String().Default("user")
tags := v.Array(v.String()).Default([]string{})
config := v.Object(v.Fields{
    "retries": v.Int().Default(3),
    "timeout": v.Number().Default(30.0),
})
```

Do not use function-based defaults:

```typescript
// NOT PORTABLE -- function default
const createdAt = v.string().default(() => new Date().toISOString());
// This will fail portable export
```

### Avoid custom validators for shared schemas

Custom validators are Tier 3 (local-only):

```python
# NOT PORTABLE -- custom validator
schema = v.string().custom(lambda s: s.startswith("usr_"))

# PORTABLE equivalent -- use pattern constraint
schema = v.string().pattern(r"^usr_")
```

## Export Modes

### Portable Mode

Portable mode is the strict default. It ensures the exported schema uses only Tier 1 features.

```typescript
// Portable export -- fails if schema uses non-portable features
const doc = schema.export({ mode: "portable" });
```

If the schema contains any non-portable features, the export fails with a descriptive error:

```
ExportError: Schema contains non-portable features:
  - Custom validator at path "user.email" (custom validators are local-only)
  - Function default at path "user.createdAt" (function defaults are not serializable)
```

### Extended Mode

Extended mode emits the core schema plus any extension namespaces:

```typescript
// Extended export -- includes extension metadata
const doc = schema.export({ mode: "extended" });
```

The resulting document contains the full portable core in `root` and language-specific metadata in `extensions`. SDKs that understand the extensions can use them; others can ignore informational extensions or reject the schema if semantic extensions are missing.

## The Extension Model

### Informational Extensions

Informational extensions provide metadata that is safe to ignore. The schema remains correct without them.

Examples:

- Documentation strings
- Struct tags for code generation
- UI display hints
- Source location metadata

When importing a schema with informational extensions for an unknown namespace, the SDK silently ignores them. The schema imports and validates correctly.

### Semantic Extensions

Semantic extensions change the meaning or behavior of the schema. The schema may not validate correctly without them.

Examples:

- Custom coercion rules specific to a language
- Additional validation constraints beyond the portable core
- Type system features that alter parsing behavior

When importing a schema with semantic extensions for an unknown namespace:

1. If the target SDK's namespace has the required extension, use it.
2. If the target SDK's namespace is missing but a `default` extension is provided, use the default.
3. If neither the target namespace nor a default can satisfy the requirement, the import fails.

```json
{
  "extensions": {
    "default": {
      "customCoerce": { "type": "trimAndLower" }
    },
    "js": {
      "customCoerce": { "type": "trimAndLower", "locale": "en-US" }
    },
    "go": {
      "customCoerce": { "type": "trimAndLower" }
    }
  }
}
```

In this example, a Python SDK importing the schema would find no `python` namespace. It would check for a `default` namespace, find the `customCoerce` instruction, and apply it if the SDK understands that extension. If the extension is semantic and the Python SDK does not understand `customCoerce`, the import fails.

## What Happens When Import Fails

Import can fail for several reasons. Each produces a specific error.

### Missing Required Extension

```
ImportError: Schema requires semantic extension "customCoerce"
in namespace "go". No handler registered for this extension
and no "default" fallback provided.
Code: unsupported_extension
```

### Unsupported Schema Kind

If a future version of AnyVali adds new schema kinds that an older SDK does not understand:

```
ImportError: Unknown schema kind "branded" at path "root.userId".
Code: unsupported_schema_kind
```

### Custom Validator Referenced

If a schema somehow references a custom validator (this should not happen with portable export, but could with hand-edited JSON):

```
ImportError: Schema references custom validator at path
"root.email". Custom validators are not portable.
Code: custom_validation_not_portable
```

### Recovery Strategies

When import fails, consumers can:

1. **Remove the offending feature** -- edit the JSON document to remove the non-portable extension or custom validator, then retry.
2. **Register a local handler** -- implement the missing extension in the target SDK, then retry.
3. **Use a different export mode** -- re-export the schema in portable mode from the source SDK, removing non-portable features.

## Custom Validators and Why They Are Local-Only

Custom validators are arbitrary functions:

```typescript
const positiveEven = v.int().custom((val) => {
  if (val % 2 !== 0) return { code: "not_even", message: "Must be even" };
  return true;
});
```

These cannot be made portable because:

- Functions cannot be serialized to JSON
- Executing foreign code would be a security risk
- Different languages have different capabilities and standard libraries
- Behavioral equivalence across languages is not verifiable

Instead, express validation rules using portable constraints wherever possible:

| Custom Validator | Portable Alternative |
|---|---|
| `val => val.startsWith("usr_")` | `.startsWith("usr_")` or `.pattern("^usr_")` |
| `val => val % 2 === 0` | `.multipleOf(2)` |
| `val => val >= 0 && val <= 100` | `.min(0).max(100)` |
| `val => isEmail(val)` | `.format("email")` |
| `val => val.length >= 3` | `.minLength(3)` |

When no portable constraint can express your rule, the custom validator is necessarily local-only.

## Coercion Portability

AnyVali v1 defines a small set of portable coercions. These are the only coercions that are guaranteed to work identically across all SDKs.

### Portable Coercions

| From | To | Behavior |
|---|---|---|
| `string` | `int` | Parse string as integer (e.g., `"42"` becomes `42`). Fail on non-numeric strings. |
| `string` | `number` | Parse string as float (e.g., `"3.14"` becomes `3.14`). Fail on non-numeric strings. |
| `string` | `bool` | `"true"` / `"1"` become `true`, `"false"` / `"0"` become `false`. All other strings fail. |
| `string` | `string` (trim) | Remove leading and trailing whitespace. |
| `string` | `string` (lower) | Convert to lowercase. |
| `string` | `string` (upper) | Convert to uppercase. |

### Usage

```typescript
// JS/TS
const port = v.int().coerce("string");    // "8080" -> 8080
const name = v.string().coerce("trim");   // "  Alice  " -> "Alice"
const flag = v.bool().coerce("string");   // "true" -> true
```

```python
# Python
port = v.int().coerce("string")
name = v.string().coerce("trim")
flag = v.bool().coerce("string")
```

```go
// Go
port := v.Int().Coerce(v.FromString)
name := v.String().Coerce(v.Trim)
flag := v.Bool().Coerce(v.FromString)
```

### Coercion Ordering

Coercions run early in the parse pipeline:

1. Detect presence or absence
2. **If present and coercion configured, attempt coercion**
3. If absent and default exists, materialize default
4. Validate resulting value
5. Return parsed output or structured error

If coercion fails, the SDK returns an issue with code `coercion_failed`:

```json
{
  "code": "coercion_failed",
  "message": "Cannot coerce \"abc\" to int",
  "path": ["port"],
  "expected": "int",
  "received": "\"abc\""
}
```

## Format Validator Portability

AnyVali v1 defines seven portable format validators for strings:

| Format | Description | Example |
|---|---|---|
| `email` | Email address (RFC 5321/5322 simplified) | `user@example.com` |
| `url` | URL (RFC 3986) | `https://example.com/path` |
| `uuid` | UUID (RFC 4122) | `550e8400-e29b-41d4-a716-446655440000` |
| `ipv4` | IPv4 address | `192.168.1.1` |
| `ipv6` | IPv6 address | `::1` |
| `date` | ISO 8601 date | `2025-01-15` |
| `date-time` | ISO 8601 date-time | `2025-01-15T09:30:00Z` |

All SDKs must implement these format validators and produce consistent accept/reject decisions for the same input. The conformance test suite includes test cases for each format.

Format values outside this list are not portable. SDKs may support additional format values as local extensions, but they will not export in portable mode.

## Step-by-Step Migration Guide

### Migrating from Zod (JavaScript)

#### Step 1: Map Zod types to AnyVali types

| Zod | AnyVali |
|---|---|
| `z.string()` | `v.string()` |
| `z.number()` | `v.number()` |
| `z.number().int()` | `v.int()` |
| `z.boolean()` | `v.bool()` |
| `z.object({...})` | `v.object({...})` |
| `z.array(...)` | `v.array(...)` |
| `z.enum([...])` | `v.enum([...])` |
| `z.union([...])` | `v.union([...])` |
| `z.literal(...)` | `v.literal(...)` |
| `z.optional(...)` | `v.optional(...)` |
| `z.nullable(...)` | `v.nullable(...)` |

#### Step 2: Replace transforms with portable alternatives

Zod transforms are not portable. Replace them with coercions or remove them:

```typescript
// Zod (non-portable)
const schema = z.string().transform((s) => s.trim().toLowerCase());

// AnyVali (portable)
const schema = v.string().coerce("trim").coerce("lower");
```

#### Step 3: Replace refinements with constraints

```typescript
// Zod (non-portable)
const schema = z.number().refine((n) => n % 2 === 0, "Must be even");

// AnyVali (portable)
const schema = v.number().multipleOf(2);
```

#### Step 4: Replace function defaults with data defaults

```typescript
// Zod (non-portable)
const schema = z.string().default(() => crypto.randomUUID());

// AnyVali -- keep as local-only, or use a static default
const schema = v.string(); // generate UUID at call site, not in schema
```

#### Step 5: Export and verify

```typescript
const doc = schema.export({ mode: "portable" });
// If this succeeds, your schema is fully portable
```

### Migrating from JSON Schema

#### Step 1: Import the JSON Schema as a starting point

AnyVali's canonical JSON is different from JSON Schema, so there is no direct import. Map types manually:

| JSON Schema | AnyVali |
|---|---|
| `{ "type": "string" }` | `{ "kind": "string" }` |
| `{ "type": "integer" }` | `{ "kind": "int64" }` |
| `{ "type": "number" }` | `{ "kind": "float64" }` |
| `{ "type": "boolean" }` | `{ "kind": "bool" }` |
| `{ "type": "array", "items": ... }` | `{ "kind": "array", "items": ... }` |
| `{ "type": "object", "properties": ... }` | `{ "kind": "object", "properties": ... }` |
| `{ "$ref": "#/$defs/X" }` | `{ "kind": "ref", "ref": "#/definitions/X" }` |

#### Step 2: Map constraints

| JSON Schema | AnyVali |
|---|---|
| `minLength` | `minLength` |
| `maxLength` | `maxLength` |
| `pattern` | `pattern` |
| `minimum` | `min` |
| `maximum` | `max` |
| `exclusiveMinimum` | `exclusiveMin` |
| `exclusiveMaximum` | `exclusiveMax` |
| `multipleOf` | `multipleOf` |
| `minItems` | `minItems` |
| `maxItems` | `maxItems` |
| `format` | `format` (only portable values) |

#### Step 3: Handle features AnyVali does not support

JSON Schema features without AnyVali equivalents in v1:

- `oneOf` -- use `union` (closest equivalent)
- `allOf` -- use `intersection`
- `anyOf` -- use `union`
- `not` -- no equivalent; model differently
- `if/then/else` -- no equivalent; use unions or restructure
- `patternProperties` -- no equivalent in v1
- `additionalProperties` -- use `unknownKeys` mode

### Migrating from Pydantic (Python)

#### Step 1: Map Pydantic fields to AnyVali

```python
# Pydantic
class User(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    age: int = Field(ge=0, le=150)
    email: EmailStr

# AnyVali
user_schema = v.object({
    "name": v.string().min_length(1).max_length(100),
    "age": v.int().min(0).max(150),
    "email": v.string().format("email"),
})
```

#### Step 2: Replace validators with constraints

```python
# Pydantic (non-portable)
@validator("name")
def name_must_be_capitalized(cls, v):
    return v.title()

# AnyVali -- use coercion or handle at application level
```

#### Step 3: Export and test

```python
doc = user_schema.export(mode="portable")
# Import in another language to verify
```

## FAQ and Common Pitfalls

### Can I use regex patterns portably?

Yes, but with caveats. The `pattern` constraint accepts regular expressions, but regex syntax varies across languages. Stick to basic regex features that are common across all target languages:

- Character classes: `[a-z]`, `[0-9]`, `\d`, `\w`, `\s`
- Quantifiers: `*`, `+`, `?`, `{n}`, `{n,m}`
- Anchors: `^`, `$`
- Alternation: `|`
- Groups: `(...)`, non-capturing `(?:...)`

Avoid language-specific regex features like lookbehinds (not supported in all JS engines), possessive quantifiers, or Unicode property escapes.

### What happens if I export from a newer SDK and import into an older one?

The `anyvaliVersion` and `schemaVersion` fields in the document allow the importing SDK to detect version mismatches. If the document uses a newer schema version with features the SDK does not understand, the import fails with a descriptive error.

### Can I mix portable and local-only features in one schema?

Yes. A schema can have both portable constraints and local custom validators. The local features work during runtime validation. When you export:

- **Portable mode**: the export fails because local features are present.
- **Extended mode**: the export includes the portable core; local features are either omitted or represented as extensions (depending on SDK implementation).

### Why does portable export fail instead of silently dropping non-portable features?

Failing loudly prevents subtle bugs. If a custom validator is silently dropped during export, the importing SDK would accept values that the authoring SDK would reject. This leads to data quality issues that are hard to diagnose. Explicit failure forces the developer to make a conscious decision about portability.

### How do I share schemas between frontend (JS) and backend (Go)?

1. Define the schema in either language.
2. Export it in portable mode: `schema.export({ mode: "portable" })`.
3. Store the JSON document (in a file, database, or API response).
4. Import it in the other language: `v.ImportSchema(doc)`.
5. Both sides now validate identically.

For this to work, the schema must use only Tier 1 features.

### Can I use `int64` safely with JavaScript?

Values up to `Number.MAX_SAFE_INTEGER` (9,007,199,254,740,991 or 2^53 - 1) work without issues. For larger values, the JS SDK will use `BigInt` or reject the value. If your integer values never exceed 2^53, you can use `int64` freely. If you need the full `int64` range in a JavaScript context, be aware of the `BigInt` requirement.

### What is the `default` extension namespace?

The `default` namespace in extensions provides fallback behavior for SDKs that do not have their own namespace. When an SDK imports a schema with a semantic extension:

1. It looks for its own namespace (e.g., `python`).
2. If not found, it looks for `default`.
3. If neither exists, the import fails.

This allows schema authors to provide reasonable default behavior for SDKs they have not explicitly targeted.

### How do I handle unknown keys in portable schemas?

AnyVali defaults to rejecting unknown keys. You can change this per object:

```typescript
// Reject unknown keys (default)
v.object({ name: v.string() })

// Strip unknown keys
v.object({ name: v.string() }).unknownKeys("strip")

// Allow unknown keys (pass through)
v.object({ name: v.string() }).unknownKeys("allow")
```

All three modes are portable and behave identically across SDKs.

### My schema uses a custom format. Is it portable?

Only the seven built-in format values are portable: `email`, `url`, `uuid`, `ipv4`, `ipv6`, `date`, `date-time`. Custom format values are local-only. If you need portable validation for a custom format, express it using `pattern` or other portable constraints.

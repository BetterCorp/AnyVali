# JavaScript / TypeScript SDK Reference

## Installation

```bash
npm install @anyvali/js
```

AnyVali requires Node.js 22 or later.

### Imports

```typescript
import {
  // Builder functions
  string, number, int, bool, null_, any, unknown, never,
  float32, float64,
  int8, int16, int32, int64,
  uint8, uint16, uint32, uint64,
  literal, enum_,
  array, tuple, object, record,
  union, intersection,
  optional, nullable,

  // Parse
  parse, safeParse,

  // Interchange
  exportSchema, importSchema,

  // Types
  type Infer,
  type InferInput,
  type ParseResult,
  type ValidationIssue,
  type AnyValiDocument,
  type ExportMode,
  type CoercionConfig,

  // Errors
  ValidationError,
  ISSUE_CODES,
} from "@anyvali/js";
```

The forms layer is available from a separate entry point:

```typescript
import { initForm, createFormBindings } from "@anyvali/js/forms";
```

## Quick Start

```typescript
import { object, string, int, enum_, type Infer } from "@anyvali/js";

// 1. Define a schema
const UserSchema = object({
  name: string().minLength(1).maxLength(100),
  email: string().format("email"),
  age: int().min(0).max(150),
  role: enum_(["admin", "user", "guest"]).default("user"),
});

// 2. Infer the TypeScript type
type User = Infer<typeof UserSchema>;
// => { name: string; email: string; age: number; role: "admin" | "user" | "guest" }

// 3. Parse input -- throws ValidationError on failure
const user = UserSchema.parse({
  name: "Alice",
  email: "alice@example.com",
  age: 30,
});
// => { name: "Alice", email: "alice@example.com", age: 30, role: "user" }

// 4. Safe parse -- returns a result object
const result = UserSchema.safeParse({ name: "", email: "bad" });
if (!result.success) {
  for (const issue of result.issues) {
    console.log(`${issue.path.join(".")}: [${issue.code}] ${issue.message}`);
  }
}
```

## Type Inference

AnyVali provides full static type inference via the `Infer<T>` utility type. This is the flagship feature of the TypeScript SDK -- every parsed value carries the correct type without manual annotations or casts.

### Basic Usage

```typescript
import { string, int, bool, type Infer } from "@anyvali/js";

const NameSchema = string();
type Name = Infer<typeof NameSchema>; // string

const AgeSchema = int();
type Age = Infer<typeof AgeSchema>; // number

const ActiveSchema = bool();
type Active = Infer<typeof ActiveSchema>; // boolean
```

### Object Inference

Object schemas infer a full object type. Required fields become required properties; optional fields become optional properties with `| undefined`.

```typescript
import { object, string, int, optional, type Infer } from "@anyvali/js";

const UserSchema = object({
  name: string(),
  email: string().format("email"),
  age: optional(int().min(0)),
});

type User = Infer<typeof UserSchema>;
// => { name: string; email: string; age?: number | undefined }

const user = UserSchema.parse({ name: "Alice", email: "a@b.com" });
// user.name  -> string (no cast needed)
// user.age   -> number | undefined
```

### Array and Tuple Inference

```typescript
import { array, tuple, string, int, bool, type Infer } from "@anyvali/js";

const TagsSchema = array(string());
type Tags = Infer<typeof TagsSchema>; // string[]

const CoordSchema = tuple([int(), int(), string()]);
type Coord = Infer<typeof CoordSchema>; // [number, number, string]
```

### Union and Intersection Inference

```typescript
import {
  union, intersection, object, string, int, literal,
  type Infer,
} from "@anyvali/js";

// Union infers to a union type
const IdSchema = union([string(), int()]);
type Id = Infer<typeof IdSchema>; // string | number

// Intersection merges object shapes
const Named = object({ name: string() });
const Aged = object({ age: int() });
const Person = intersection([Named, Aged]);
type Person = Infer<typeof Person>; // { name: string } & { age: number }
```

### Literal and Enum Inference

```typescript
import { literal, enum_, type Infer } from "@anyvali/js";

const StatusSchema = literal("active");
type Status = Infer<typeof StatusSchema>; // "active"

const RoleSchema = enum_(["admin", "user", "guest"] as const);
type Role = Infer<typeof RoleSchema>; // "admin" | "user" | "guest"
```

Pass `as const` to `enum_()` to infer narrow literal union types instead of `string`.

### Nullable and Optional Inference

```typescript
import { nullable, optional, string, type Infer } from "@anyvali/js";

const NullableName = nullable(string());
type NullableName = Infer<typeof NullableName>; // string | null

const OptionalName = optional(string());
type OptionalName = Infer<typeof OptionalName>; // string | undefined
```

### InferInput

`InferInput<T>` extracts the input type (what goes into `parse()`), which may differ from the output type when coercion or defaults are in play.

```typescript
import { string, type Infer, type InferInput } from "@anyvali/js";

const Schema = string().default("anonymous");
type Output = Infer<typeof Schema>;      // string
type Input = InferInput<typeof Schema>;   // string (the input is still expected to be a string)
```

### Record Inference

```typescript
import { record, int, type Infer } from "@anyvali/js";

const ScoresSchema = record(int());
type Scores = Infer<typeof ScoresSchema>; // Record<string, number>
```

## Schema Types

### Primitives

#### string()

Creates a schema that validates string values. Supports chained constraint methods.

```typescript
const s = string();
s.parse("hello"); // => "hello"
s.parse(42);      // throws: Expected string, received number
```

#### number()

Creates a schema for floating-point numbers. Defaults to float64 (IEEE 754 double precision).

```typescript
const n = number();
n.parse(3.14); // => 3.14
n.parse("3");  // throws: Expected number, received string
```

#### int()

Creates a schema for integer values. Defaults to int64 (safe integer range). Rejects floats.

```typescript
const i = int();
i.parse(42);   // => 42
i.parse(3.5);  // throws: Expected integer, received float
```

#### bool()

Creates a schema for boolean values.

```typescript
const b = bool();
b.parse(true);  // => true
b.parse("yes"); // throws: Expected bool, received string
```

#### null\_()

Creates a schema that only accepts `null`. Named `null_` to avoid conflicting with the JavaScript keyword.

```typescript
const n = null_();
n.parse(null); // => null
n.parse(0);    // throws: Expected null, received number
```

#### any()

Accepts any value without validation.

```typescript
const a = any();
a.parse("hello"); // => "hello"
a.parse(42);      // => 42
a.parse(null);    // => null
```

#### unknown()

Accepts any value. Semantically identical to `any()` at runtime, but infers to `unknown` instead of `any` in TypeScript, forcing downstream code to narrow the type.

```typescript
const u = unknown();
u.parse("hello"); // => "hello" (typed as unknown)
```

#### never()

Always fails validation. Useful for marking branches of a union that should be unreachable.

```typescript
const n = never();
n.parse("anything"); // throws: Expected never
```

### Numeric Widths

AnyVali provides width-specific numeric schemas that enforce range constraints automatically.

#### Float Types

```typescript
float32()  // IEEE 754 single precision
float64()  // IEEE 754 double precision (same as number())
```

#### Signed Integer Types

```typescript
int8()     // -128 to 127
int16()    // -32,768 to 32,767
int32()    // -2,147,483,648 to 2,147,483,647
int64()    // Number.MIN_SAFE_INTEGER to Number.MAX_SAFE_INTEGER
```

#### Unsigned Integer Types

```typescript
uint8()    // 0 to 255
uint16()   // 0 to 65,535
uint32()   // 0 to 4,294,967,295
uint64()   // 0 to Number.MAX_SAFE_INTEGER
```

All integer types also reject non-integer values (floats).

```typescript
const port = uint16();
port.parse(8080);   // => 8080
port.parse(-1);     // throws: Value -1 is below the minimum for uint16
port.parse(100000); // throws: Value 100000 is above the maximum for uint16
port.parse(3.5);    // throws: Expected integer, received float
```

All numeric width schemas inherit the same constraint methods as `number()` and `int()` (`.min()`, `.max()`, `.exclusiveMin()`, `.exclusiveMax()`, `.multipleOf()`).

### Values

#### literal(value)

Creates a schema that matches exactly one specific value. Accepts `string`, `number`, `boolean`, or `null`.

```typescript
const admin = literal("admin");
admin.parse("admin"); // => "admin"
admin.parse("user");  // throws: Expected literal admin, received user

const fortyTwo = literal(42);
fortyTwo.parse(42); // => 42

const yes = literal(true);
yes.parse(true); // => true
```

#### enum\_(values)

Creates a schema that matches any value in the given array. Accepts strings and numbers.

```typescript
const status = enum_(["active", "inactive", "pending"]);
status.parse("active");  // => "active"
status.parse("deleted"); // throws: Expected one of enum(active,inactive,pending)

// Use `as const` for narrow type inference
const role = enum_(["admin", "user"] as const);
type Role = Infer<typeof role>; // "admin" | "user"
```

### Collections

#### array(itemSchema)

Creates a schema for arrays where every element must match the item schema.

```typescript
const tags = array(string());
tags.parse(["a", "b", "c"]); // => ["a", "b", "c"]
tags.parse([1, 2, 3]);       // throws: Expected string, received number (at index 0)
tags.parse("not an array");  // throws: Expected array, received string
```

#### tuple(schemas)

Creates a schema for fixed-length arrays where each element matches the schema at its position.

```typescript
const point = tuple([number(), number()]);
point.parse([1.5, 2.5]);     // => [1.5, 2.5]
point.parse([1]);             // throws: Tuple must have exactly 2 element(s)
point.parse([1, 2, 3]);      // throws: Tuple must have exactly 2 element(s)

const mixed = tuple([string(), int(), bool()]);
mixed.parse(["hello", 42, true]); // => ["hello", 42, true]
```

#### object(shape, options?)

Creates a schema for objects with named properties. By default, all properties are required and unknown keys are rejected.

```typescript
const user = object({
  name: string(),
  age: int(),
});

user.parse({ name: "Alice", age: 30 });          // => { name: "Alice", age: 30 }
user.parse({ name: "Alice" });                    // throws: Required property "age" is missing
user.parse({ name: "Alice", age: 30, extra: 1 }); // throws: Unknown key "extra"
```

Mark fields as optional using the `optional()` wrapper:

```typescript
const user = object({
  name: string(),
  age: optional(int()),
});
user.parse({ name: "Alice" }); // => { name: "Alice" }
```

Control unknown key handling with the `unknownKeys` option:

```typescript
// Strip unknown keys silently
const loose = object({ name: string() }, { unknownKeys: "strip" });
loose.parse({ name: "Alice", extra: 1 }); // => { name: "Alice" }

// Allow unknown keys in the output
const open = object({ name: string() }, { unknownKeys: "allow" });
open.parse({ name: "Alice", extra: 1 }); // => { name: "Alice", extra: 1 }
```

You can also change the mode after construction:

```typescript
const schema = object({ name: string() }).unknownKeys("strip");
```

#### record(valueSchema)

Creates a schema for objects with arbitrary string keys where all values must match a single schema. Think of it as `Record<string, T>` in TypeScript.

```typescript
const scores = record(int());
scores.parse({ alice: 100, bob: 95 }); // => { alice: 100, bob: 95 }
scores.parse({ alice: "high" });       // throws: Expected integer, received string
```

### Composition

#### union(schemas)

Creates a schema that accepts any value matching at least one of the given schemas. Variants are tried in order; the first match wins.

```typescript
const id = union([string(), int()]);
id.parse("abc"); // => "abc"
id.parse(42);    // => 42
id.parse(true);  // throws: Input did not match any variant of the union
```

#### intersection(schemas)

Creates a schema that requires the value to match all given schemas. For objects, the results are merged.

```typescript
const withName = object({ name: string() }, { unknownKeys: "strip" });
const withAge = object({ age: int() }, { unknownKeys: "strip" });
const person = intersection([withName, withAge]);

person.parse({ name: "Alice", age: 30 }); // => { name: "Alice", age: 30 }
person.parse({ name: "Alice" });           // throws: Required property "age" is missing
```

### Modifiers

#### optional(schema)

Wraps a schema so that `undefined` (absent) values are accepted. When the value is present, it must match the inner schema.

`optional()` is a standalone wrapper function, not a chainable method.

```typescript
const maybeAge = optional(int());
maybeAge.parse(42);        // => 42
maybeAge.parse(undefined); // => undefined

// Typically used inside object()
const user = object({
  name: string(),
  bio: optional(string()),
});
```

#### nullable(schema)

Wraps a schema so that `null` values are accepted. When the value is non-null, it must match the inner schema.

`nullable()` is a standalone wrapper function, not a chainable method.

```typescript
const nullableName = nullable(string());
nullableName.parse("Alice"); // => "Alice"
nullableName.parse(null);    // => null
nullableName.parse(42);      // throws: Expected string, received number
```

## Constraints

### String Constraints

All string constraint methods return a new schema instance (immutable builder pattern).

```typescript
string().minLength(1)               // at least 1 character
string().maxLength(255)             // at most 255 characters
string().pattern("^[a-z]+$")       // must match regex pattern
string().startsWith("https://")    // must start with prefix
string().endsWith(".json")         // must end with suffix
string().includes("@")            // must contain substring
string().format("email")          // must match a named format
```

Chain multiple constraints:

```typescript
const email = string()
  .minLength(5)
  .maxLength(254)
  .format("email");
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

All numeric constraint methods apply to `number()`, `int()`, and all width-specific variants.

```typescript
number().min(0)              // value >= 0
number().max(100)            // value <= 100
number().exclusiveMin(0)     // value > 0
number().exclusiveMax(100)   // value < 100
number().multipleOf(5)       // value must be divisible by 5
```

Chain multiple constraints:

```typescript
const percentage = number()
  .min(0)
  .max(100)
  .multipleOf(0.01);
```

### Array Constraints

```typescript
array(string()).minItems(1)   // at least 1 element
array(string()).maxItems(10)  // at most 10 elements

// Combined
const tags = array(string()).minItems(1).maxItems(5);
```

### Object Options

The `unknownKeys` option controls how keys not declared in the shape are handled:

| Mode | Behavior |
|---|---|
| `"reject"` (default) | Produces an `unknown_key` issue for each extra key |
| `"strip"` | Silently removes extra keys from the output |
| `"allow"` | Passes extra keys through to the output |

```typescript
const strict = object({ name: string() }); // unknownKeys defaults to "reject"
const lenient = object({ name: string() }, { unknownKeys: "strip" });
const passthrough = object({ name: string() }, { unknownKeys: "allow" });
```

## Coercion

Coercion transforms the input value before validation. It runs in step 2 of the parse pipeline, only when the value is present.

### Usage

Call `.coerce(config)` on any schema to enable coercion:

```typescript
const age = int().coerce({ from: "string" });
age.parse("42"); // => 42 (string coerced to integer)
age.parse(42);   // => 42 (already an integer, no coercion needed)
```

### CoercionConfig

The `CoercionConfig` object supports the following properties:

| Property | Type | Description |
|---|---|---|
| `from` | `string` | Source type for type conversion. Currently `"string"` is supported. |
| `trim` | `boolean` | Trim whitespace from string values |
| `lower` | `boolean` | Convert string to lowercase |
| `upper` | `boolean` | Convert string to uppercase |

### Available Coercions

#### String to Number

```typescript
const n = number().coerce({ from: "string" });
n.parse("3.14"); // => 3.14
n.parse("");     // throws: coercion_failed
```

#### String to Integer

```typescript
const i = int().coerce({ from: "string" });
i.parse("42");   // => 42
i.parse("3.5");  // throws: coercion_failed (not an integer string)
```

#### String to Boolean

```typescript
const b = bool().coerce({ from: "string" });
b.parse("true");  // => true
b.parse("false"); // => false
b.parse("1");     // => true
b.parse("0");     // => false
b.parse("yes");   // throws: coercion_failed
```

#### String Transformations

```typescript
const trimmed = string().coerce({ trim: true });
trimmed.parse("  hello  "); // => "hello"

const lower = string().coerce({ lower: true });
lower.parse("HELLO"); // => "hello"

const upper = string().coerce({ upper: true });
upper.parse("hello"); // => "HELLO"
```

Transformations can be combined:

```typescript
const normalized = string().coerce({ trim: true, lower: true });
normalized.parse("  Hello World  "); // => "hello world"
```

## Defaults

Defaults fill in missing (absent) values. They run in step 3 of the parse pipeline, after coercion and before validation.

### Usage

Call `.default(value)` on any schema:

```typescript
const role = string().default("user");
role.parse(undefined); // => "user"
role.parse("admin");   // => "admin"

const tags = array(string()).default([]);
tags.parse(undefined); // => []
tags.parse(["a"]);     // => ["a"]
```

Defaults work with optional fields in objects:

```typescript
const config = object({
  theme: optional(string().default("light")),
  language: optional(string().default("en")),
});

config.parse({}); // => { theme: "light", language: "en" }
config.parse({ theme: "dark" }); // => { theme: "dark", language: "en" }
```

If the default value itself fails validation, a `default_invalid` issue is produced.

## Export and Import

AnyVali schemas can be exported to a portable JSON document and imported back in any supported SDK.

### exportSchema(schema, mode?)

Exports a schema to an `AnyValiDocument` object.

```typescript
import { object, string, int, exportSchema } from "@anyvali/js";

const schema = object({
  name: string().minLength(1),
  age: int().min(0),
});

// Portable mode (default) -- fails if schema uses non-portable features
const doc = exportSchema(schema);
// Or: const doc = exportSchema(schema, "portable");

// Extended mode -- includes language-specific extension namespaces
const extended = exportSchema(schema, "extended");

console.log(JSON.stringify(doc, null, 2));
```

You can also call `.export(mode?)` directly on a schema instance:

```typescript
const doc = schema.export("portable");
```

The document structure:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": { ... },
  "definitions": {},
  "extensions": {}
}
```

### importSchema(doc)

Imports an `AnyValiDocument` into a live schema that can be used for parsing.

```typescript
import { importSchema } from "@anyvali/js";

const doc = {
  anyvaliVersion: "1.0",
  schemaVersion: "1",
  root: {
    kind: "object",
    properties: {
      name: { kind: "string", minLength: 1 },
      age: { kind: "int", min: 0 },
    },
    required: ["name", "age"],
    unknownKeys: "reject",
  },
  definitions: {},
  extensions: {},
};

const schema = importSchema(doc);
const user = schema.parse({ name: "Alice", age: 30 });
```

### Export Modes

| Mode | Behavior |
|---|---|
| `"portable"` | Only emits portable schema features. Fails if the schema depends on non-portable constructs. |
| `"extended"` | Emits the core schema plus language-specific extensions in the `extensions` field. |

## Error Handling

### ValidationError

When `parse()` fails, it throws a `ValidationError` containing a list of issues:

```typescript
import { string, ValidationError } from "@anyvali/js";

try {
  string().minLength(5).parse("hi");
} catch (err) {
  if (err instanceof ValidationError) {
    console.log(err.message);
    // "[too_small] String must have at least 5 character(s)"

    for (const issue of err.issues) {
      console.log(issue.code);    // "too_small"
      console.log(issue.message); // "String must have at least 5 character(s)"
      console.log(issue.path);    // []
    }
  }
}
```

### ParseResult

`safeParse()` never throws. It returns a discriminated union:

```typescript
type ParseResult<T> =
  | { success: true; data: T }
  | { success: false; issues: ValidationIssue[] };
```

```typescript
const result = string().safeParse(42);

if (result.success) {
  console.log(result.data); // string
} else {
  for (const issue of result.issues) {
    console.log(issue);
  }
}
```

### ValidationIssue

Each issue has the following structure:

```typescript
interface ValidationIssue {
  code: string;                   // Machine-readable issue code
  message: string;                // Human-readable description
  path: (string | number)[];     // Path to the failing value
  expected?: string;              // What was expected
  received?: string;              // What was received
  meta?: Record<string, unknown>; // Additional metadata
}
```

The `path` array describes the location of the error within nested structures:

```typescript
const schema = object({
  users: array(object({
    email: string().format("email"),
  })),
});

const result = schema.safeParse({
  users: [{ email: "not-an-email" }],
});

// result.issues[0].path => ["users", 0, "email"]
```

### Issue Codes

All issue codes are available as constants via `ISSUE_CODES`:

```typescript
import { ISSUE_CODES } from "@anyvali/js";
```

| Code | When |
|---|---|
| `invalid_type` | Value is the wrong type |
| `required` | Required property is missing |
| `unknown_key` | Object has an undeclared key (when `unknownKeys` is `"reject"`) |
| `too_small` | String too short, number too low, or array too few items |
| `too_large` | String too long, number too high, or array too many items |
| `invalid_string` | String fails a pattern, startsWith, endsWith, includes, or format check |
| `invalid_number` | Number fails a multipleOf check |
| `invalid_literal` | Value does not match the expected literal |
| `invalid_union` | Value does not match any variant in a union |
| `coercion_failed` | Coercion could not convert the input |
| `default_invalid` | The materialized default value failed validation |
| `custom_validation_not_portable` | Non-portable custom validation was encountered |
| `unsupported_extension` | Unknown extension namespace in an imported document |
| `unsupported_schema_kind` | Unknown schema kind in an imported document |

## Common Patterns

### Validating Environment Variables

Use `unknownKeys: "strip"` when parsing objects that contain many extra keys you don't care about, like `process.env`:

```typescript
import { object, string, int, optional, type Infer } from "@anyvali/js";

const EnvSchema = object({
  NODE_ENV: optional(string()).default("development"),
  PORT: optional(int().coerce({ from: "string" })).default(3000),
  DATABASE_URL: string(),
}, { unknownKeys: "strip" });

type Env = Infer<typeof EnvSchema>;
// => { NODE_ENV?: string | undefined; PORT?: number | undefined; DATABASE_URL: string }

const env = EnvSchema.parse(process.env);
// Returns only { NODE_ENV, PORT, DATABASE_URL } -- all other env vars are stripped
```

Without `"strip"`, parse would fail with `unknown_key` issues for every other variable in `process.env` (PATH, HOME, etc.) because the default mode is `"reject"`.

| Mode | What happens with extra keys |
|---|---|
| `"reject"` (default) | Parse fails with `unknown_key` issues |
| `"strip"` | Extra keys silently removed from output |
| `"allow"` | Extra keys passed through to output |

### Function-Based Defaults

AnyVali v1 only supports static data defaults (for portability). If you need a computed default like `process.cwd()`, apply it after parsing:

```typescript
const ConfigSchema = object({
  profile: optional(string()).default("default"),
  appDir: optional(string()),
}, { unknownKeys: "strip" });

const config = ConfigSchema.parse(process.env);
config.appDir ??= process.cwd(); // apply function-based default manually
```

This keeps the schema fully portable -- the same JSON document can be imported in Go, Python, or any other SDK without relying on language-specific function calls.

## Forms

The `@anyvali/js/forms` entry point provides a browser-side forms integration layer that connects AnyVali schemas to HTML forms, including native constraint validation and htmx support.

### createFormBindings(options)

Generates HTML attribute objects for form fields and error slots based on a schema.

```typescript
import { createFormBindings } from "@anyvali/js/forms";
import { object, string, int } from "@anyvali/js";

const schema = object({
  name: string().minLength(1).maxLength(100),
  age: int().min(0).max(150),
});

const form = createFormBindings({ schema });

// Get attributes for a field (name, type, required, min, max, etc.)
const nameAttrs = form.field("name");
// => { name: "name", required: true, minLength: 1, maxLength: 100, ... }

const ageAttrs = form.field("age");
// => { name: "age", required: true, type: "number", min: 0, max: 150, step: 1, ... }

// Get attributes for an error display slot
const nameError = form.errorSlot("name");
// => { id: "anyvali-error-name", "data-anyvali-error-for": "name", "aria-live": "polite" }
```

### initForm(target, options)

Attaches live validation to a `<form>` element. Returns a `FormController`.

```typescript
import { initForm } from "@anyvali/js/forms";

const controller = initForm("#my-form", {
  schema: UserSchema,
  validateOn: ["blur", "submit"],  // when to validate (default: ["blur", "submit"])
  nativeValidation: true,          // apply native HTML constraints (default: true)
  reportValidity: true,            // call reportValidity() on fields (default: true)
  htmx: true,                     // integrate with htmx (default: true)
});

controller.validate();   // manually trigger validation, returns boolean
controller.getValues();  // read current form values as an object
controller.getResult();  // run safeParse and return ParseResult
controller.destroy();    // remove all event listeners
```

## API Reference

### Builder Functions

| Function | Returns | Description |
|---|---|---|
| `string()` | `StringSchema` | String values |
| `number()` | `NumberSchema` | Float64 numbers |
| `float32()` | `Float32Schema` | Float32 numbers |
| `float64()` | `Float64Schema` | Float64 numbers |
| `int()` | `IntSchema` | Int64 integers |
| `int8()` | `Int8Schema` | Int8 integers |
| `int16()` | `Int16Schema` | Int16 integers |
| `int32()` | `Int32Schema` | Int32 integers |
| `int64()` | `Int64Schema` | Int64 integers |
| `uint8()` | `Uint8Schema` | Uint8 integers |
| `uint16()` | `Uint16Schema` | Uint16 integers |
| `uint32()` | `Uint32Schema` | Uint32 integers |
| `uint64()` | `Uint64Schema` | Uint64 integers |
| `bool()` | `BoolSchema` | Boolean values |
| `null_()` | `NullSchema` | Null only |
| `any()` | `AnySchema` | Any value (typed as `any`) |
| `unknown()` | `UnknownSchema` | Any value (typed as `unknown`) |
| `never()` | `NeverSchema` | Always fails |
| `literal(value)` | `LiteralSchema<T>` | Exact value match |
| `enum_(values)` | `EnumSchema<T>` | One of the listed values |
| `array(items)` | `ArraySchema<T>` | Array of uniform type |
| `tuple(schemas)` | `TupleSchema<T>` | Fixed-length typed array |
| `object(shape, options?)` | `ObjectSchema<T>` | Object with named properties |
| `record(valueSchema)` | `RecordSchema<T>` | Object with uniform value type |
| `union(variants)` | `UnionSchema<T>` | First-match union |
| `intersection(schemas)` | `IntersectionSchema<T>` | All-match intersection |
| `optional(schema)` | `OptionalSchema<T>` | Allows `undefined` |
| `nullable(schema)` | `NullableSchema<T>` | Allows `null` |

### BaseSchema Methods

Available on all schema instances:

| Method | Returns | Description |
|---|---|---|
| `.parse(input)` | `T` | Parse or throw `ValidationError` |
| `.safeParse(input)` | `ParseResult<T>` | Parse and return result object |
| `.default(value)` | `this` | Set a default for absent values |
| `.coerce(config)` | `this` | Configure coercion behavior |
| `.export(mode?)` | `AnyValiDocument` | Export to interchange document |

### StringSchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.minLength(n)` | `number` | Minimum string length |
| `.maxLength(n)` | `number` | Maximum string length |
| `.pattern(p)` | `string` | Regex pattern to match |
| `.startsWith(s)` | `string` | Required prefix |
| `.endsWith(s)` | `string` | Required suffix |
| `.includes(s)` | `string` | Required substring |
| `.format(f)` | `StringFormat` | Named format validation |

### NumberSchema / IntSchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.min(n)` | `number` | Minimum value (inclusive) |
| `.max(n)` | `number` | Maximum value (inclusive) |
| `.exclusiveMin(n)` | `number` | Minimum value (exclusive) |
| `.exclusiveMax(n)` | `number` | Maximum value (exclusive) |
| `.multipleOf(n)` | `number` | Value must be divisible by `n` |

### ArraySchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.minItems(n)` | `number` | Minimum number of elements |
| `.maxItems(n)` | `number` | Maximum number of elements |

### ObjectSchema Methods

| Method | Parameter | Description |
|---|---|---|
| `.unknownKeys(mode)` | `UnknownKeyMode` | Set unknown key handling (`"reject"`, `"strip"`, `"allow"`) |

### Top-Level Functions

| Function | Signature | Description |
|---|---|---|
| `parse` | `parse<T>(schema, input): T` | Parse using a schema reference |
| `safeParse` | `safeParse<T>(schema, input): ParseResult<T>` | Safe parse using a schema reference |
| `exportSchema` | `exportSchema(schema, mode?): AnyValiDocument` | Export a schema to interchange format |
| `importSchema` | `importSchema(doc): BaseSchema` | Import a schema from interchange format |

### Types

| Type | Description |
|---|---|
| `Infer<T>` | Extract the output type from a schema |
| `InferInput<T>` | Extract the input type from a schema |
| `ParseResult<T>` | `{ success: true; data: T } \| { success: false; issues: ValidationIssue[] }` |
| `ValidationIssue` | `{ code, message, path, expected?, received?, meta? }` |
| `ValidationError` | Error class with `.issues: ValidationIssue[]` |
| `AnyValiDocument` | `{ anyvaliVersion, schemaVersion, root, definitions, extensions }` |
| `ExportMode` | `"portable" \| "extended"` |
| `UnknownKeyMode` | `"reject" \| "strip" \| "allow"` |
| `CoercionConfig` | `{ from?, trim?, lower?, upper? }` |
| `StringFormat` | `"email" \| "url" \| "uuid" \| "ipv4" \| "ipv6" \| "date" \| "date-time"` |
| `SchemaKind` | Union of all schema kind strings |
| `IssueCode` | Union of all issue code strings |

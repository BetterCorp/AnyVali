# AnyVali Product Overview

## What Is AnyVali?

AnyVali is a family of native validation libraries that share a single portable schema model across 10 programming languages. Unlike language-neutral schema DSLs, AnyVali lets you write schemas directly in your host language using a small, idiomatic API. Each SDK can then export those schemas to a canonical JSON document and import them into any other supported SDK.

The supported languages are:

- JavaScript / TypeScript
- Python
- Go
- Java
- C#
- Rust
- PHP
- Ruby
- Kotlin
- C++

AnyVali occupies a unique position: it gives you the ergonomics of native schema builders (like Zod for JS or pydantic for Python) with the cross-language portability of JSON Schema.

## Core Concepts

### Native-First Authoring

Schemas are authored in the host language, not in a separate DSL. Each SDK exposes builder functions that feel natural to developers in that ecosystem. The API surface is small and consistent across SDKs while respecting language idioms.

For example, a JS developer uses camelCase methods, a Python developer uses snake_case, and a Go developer uses exported PascalCase functions. The semantics are identical; the naming adapts.

### Portable Export and Import

Any schema built with an AnyVali SDK can be exported to a canonical JSON document. That document can then be imported into any other AnyVali SDK. This enables:

- Sharing validation rules between a TypeScript frontend and a Go backend
- Storing schemas in a database and loading them at runtime in any language
- Building schema registries that serve multiple services in different languages
- Generating documentation from schemas regardless of the authoring language

Export operates in two modes:

- **Portable mode** -- fails if the schema depends on non-portable features. This is the safe default.
- **Extended mode** -- emits the core schema plus language-specific extension namespaces.

### Safe Numeric Defaults

AnyVali defaults `number` to IEEE 754 `float64` and `int` to signed `int64`. This is a deliberate choice that prioritizes cross-language safety over memory efficiency. See the [Numeric Semantics Guide](numeric-semantics.md) for the full rationale.

## Quick Start

### JavaScript / TypeScript

```typescript
import { object, string, int64, int, enum_, importSchema } from "@anyvali/js";

// 1. Define a schema
const UserSchema = object({
  id: int64(),
  name: string().minLength(1).maxLength(100),
  email: string().format("email"),
  age: int().min(0).max(150).optional(),
  role: enum_(["admin", "user", "guest"]).default("user"),
});

// 2. Parse input (throws on failure)
const user = UserSchema.parse({
  id: 42,
  name: "Alice",
  email: "alice@example.com",
});
// => { id: 42, name: "Alice", email: "alice@example.com", role: "user" }

// 3. Handle errors with safeParse
const result = UserSchema.safeParse({ id: "not-a-number", name: "" });
if (!result.success) {
  for (const issue of result.issues) {
    console.log(`${issue.path.join(".")}: [${issue.code}] ${issue.message}`);
    // "id: [invalid_type] Expected int64, received string"
    // "name: [too_small] String must have at least 1 character"
  }
}

// 4. Export to portable JSON
const doc = UserSchema.export({ mode: "portable" });
console.log(JSON.stringify(doc, null, 2));
// {
//   "anyvaliVersion": "1.0",
//   "schemaVersion": "1",
//   "root": {
//     "kind": "object",
//     "properties": {
//       "id": { "kind": "int64" },
//       "name": { "kind": "string", "minLength": 1, "maxLength": 100 },
//       ...
//     },
//     ...
//   },
//   "definitions": {},
//   "extensions": {}
// }

// 5. Import from JSON
const imported = importSchema(doc);
const parsed = imported.parse({ id: 1, name: "Bob", email: "bob@test.com" });
```

### Python

```python
import anyvali as v

# 1. Define a schema
user_schema = v.object({
    "id": v.int64(),
    "name": v.string().min_length(1).max_length(100),
    "email": v.string().format("email"),
    "age": v.int().min(0).max(150).optional(),
    "role": v.enum(["admin", "user", "guest"]).default("user"),
})

# 2. Parse input (raises on failure)
user = user_schema.parse({
    "id": 42,
    "name": "Alice",
    "email": "alice@example.com",
})
# => {"id": 42, "name": "Alice", "email": "alice@example.com", "role": "user"}

# 3. Handle errors with safe_parse
result = user_schema.safe_parse({"id": "not-a-number", "name": ""})
if not result.success:
    for issue in result.issues:
        print(f"{'.'.join(str(p) for p in issue.path)}: [{issue.code}] {issue.message}")

# 4. Export to portable JSON
import json

doc = user_schema.export(mode="portable")
print(json.dumps(doc, indent=2))

# 5. Import from JSON
imported = v.import_schema(doc)
parsed = imported.parse({"id": 1, "name": "Bob", "email": "bob@test.com"})
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    v "github.com/BetterCorp/AnyVali-go"
)

func main() {
    // 1. Define a schema
    userSchema := v.Object(v.Fields{
        "id":    v.Int64(),
        "name":  v.String().MinLength(1).MaxLength(100),
        "email": v.String().Format("email"),
        "age":   v.Int().Min(0).Max(150).Optional(),
        "role":  v.Enum("admin", "user", "guest").Default("user"),
    })

    // 2. Parse input
    input := map[string]any{
        "id":    42,
        "name":  "Alice",
        "email": "alice@example.com",
    }
    user, err := userSchema.Parse(input)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user)

    // 3. Handle errors with SafeParse
    result := userSchema.SafeParse(map[string]any{
        "id":   "not-a-number",
        "name": "",
    })
    if !result.Success {
        for _, issue := range result.Issues {
            fmt.Printf("%s: [%s] %s\n",
                issue.Path, issue.Code, issue.Message)
        }
    }

    // 4. Export to portable JSON
    doc, err := userSchema.Export(v.ExportPortable)
    if err != nil {
        log.Fatal(err)
    }
    jsonBytes, _ := json.MarshalIndent(doc, "", "  ")
    fmt.Println(string(jsonBytes))

    // 5. Import from JSON
    imported, err := v.ImportSchema(doc)
    if err != nil {
        log.Fatal(err)
    }
    parsed, err := imported.Parse(map[string]any{
        "id": 1, "name": "Bob", "email": "bob@test.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(parsed)
}
```

## Architecture Overview

### Schema Kinds

AnyVali v1 defines the following portable schema kinds:

| Category | Kinds |
|---|---|
| Special | `any`, `unknown`, `never` |
| Primitives | `null`, `bool`, `string` |
| Numeric | `number` (float64), `int` (int64), `float32`, `float64`, `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64` |
| Literal / Enum | `literal`, `enum` |
| Collections | `array`, `tuple`, `object`, `record` |
| Composition | `union`, `intersection` |
| Modifiers | `optional`, `nullable` |
| Reference | `ref` |

### Validation Constraints

Constraints are declarative and data-only. No executable code is serialized.

- **String**: `minLength`, `maxLength`, `pattern`, `startsWith`, `endsWith`, `includes`, `format`
- **Numeric**: `min`, `max`, `exclusiveMin`, `exclusiveMax`, `multipleOf`
- **Array**: `minItems`, `maxItems`
- **Object**: required/optional fields, unknown key mode (`reject`, `strip`, `allow`)

### Parse Pipeline

Every SDK follows the same five-step parse pipeline:

```
Input
  |
  v
1. Detect presence or absence
  |
  v
2. If present and coercion configured, attempt coercion
  |
  v
3. If absent and default exists, materialize default
  |
  v
4. Validate resulting value against schema constraints
  |
  v
5. Return parsed output or structured error
```

This ordering is deterministic across all SDKs. Coercions run on present values before defaults fill in missing ones. The final value -- whether original, coerced, or defaulted -- must pass all validation constraints.

### Parse Result

Every SDK provides two parse APIs:

- **Throwing/panicking parse**: returns the parsed value or raises an error.
- **Safe parse**: returns a result object containing either the parsed value or a list of issues.

Issues follow a standard structure:

```json
{
  "code": "invalid_type",
  "message": "Expected int64, received string",
  "path": ["users", 0, "id"],
  "expected": "int64",
  "received": "string"
}
```

### Portable JSON Document

Schemas export to a versioned JSON document:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": { ... },
  "definitions": { ... },
  "extensions": { ... }
}
```

- `root` contains the top-level schema node.
- `definitions` holds named schema nodes for reuse and recursion (referenced via `{ "kind": "ref", "ref": "#/definitions/Name" }`).
- `extensions` contains namespaced, language-specific metadata.

## Comparison with Similar Tools

### vs Zod (JavaScript)

Zod is a TypeScript-first schema validation library. AnyVali shares Zod's builder-style API design but differs in key ways:

- **Zod** is JS/TS only. **AnyVali** targets 10 languages.
- **Zod** schemas are not serializable. **AnyVali** schemas export to portable JSON.
- **Zod** supports arbitrary transforms and refinements. **AnyVali** v1 limits transforms to portable coercions.
- **Zod** has no cross-language interop story. **AnyVali** is built for it.

### vs Valibot (JavaScript)

Valibot is a modular, tree-shakeable JS validation library. Compared to AnyVali:

- **Valibot** optimizes for JS bundle size via modular imports. **AnyVali** optimizes for cross-language portability.
- **Valibot** is JS/TS only. **AnyVali** spans 10 languages.
- **Valibot** schemas are not serializable to a portable format.

### vs ArkType (TypeScript)

ArkType uses TypeScript's type system for schema inference. Compared to AnyVali:

- **ArkType** is deeply coupled to TypeScript's type system. **AnyVali** is language-agnostic.
- **ArkType** offers powerful type-level inference. **AnyVali** focuses on runtime validation and portability.
- **ArkType** schemas are not portable across languages.

### vs JSON Schema

JSON Schema is a specification for describing JSON data formats. Compared to AnyVali:

- **JSON Schema** is authoring-language-neutral (you write JSON or YAML). **AnyVali** lets you write in your host language.
- **JSON Schema** has a very large specification surface. **AnyVali** has a deliberately small core.
- **JSON Schema** does not define a parse pipeline or result format. **AnyVali** specifies exact parse semantics.
- **JSON Schema** does not include coercions or defaults as first-class concepts. **AnyVali** does.
- **JSON Schema** validators vary in behavior across implementations. **AnyVali** mandates a conformance test suite.
- **AnyVali's** portable JSON format is simpler and more opinionated than JSON Schema, which makes cross-SDK consistency achievable.

## Next Steps

- [Numeric Semantics Guide](numeric-semantics.md) -- understand AnyVali's numeric type system
- [Portability Guide](portability-guide.md) -- design schemas that work across languages
- [SDK Authors Guide](sdk-authors-guide.md) -- implement a new AnyVali SDK

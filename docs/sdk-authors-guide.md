# AnyVali SDK Authors Guide

## Overview

This guide covers everything you need to implement a new AnyVali SDK. It is intended for developers building an AnyVali library for a language not yet supported, or for contributors extending an existing SDK.

An AnyVali SDK must:

1. Expose a native builder API for defining schemas
2. Implement the five-step parse pipeline
3. Support schema export to the canonical JSON format
4. Support schema import from the canonical JSON format
5. Pass the shared conformance test suite

## Required Operations

Every SDK must implement these five conceptual operations. The naming should follow the conventions of the target language.

### Define Schema

Provide builder functions or methods for every portable schema kind:

| Kind | Example API |
|---|---|
| `any` | `v.any()` |
| `unknown` | `v.unknown()` |
| `never` | `v.never()` |
| `null` | `v.null()` |
| `bool` | `v.bool()` |
| `string` | `v.string()` |
| `number` | `v.number()` |
| `int` | `v.int()` |
| `float32` | `v.float32()` |
| `float64` | `v.float64()` |
| `int8` - `int64` | `v.int8()` ... `v.int64()` |
| `uint8` - `uint64` | `v.uint8()` ... `v.uint64()` |
| `literal` | `v.literal("active")` |
| `enum` | `v.enum(["a", "b", "c"])` |
| `array` | `v.array(v.string())` |
| `tuple` | `v.tuple([v.string(), v.int()])` |
| `object` | `v.object({ name: v.string() })` |
| `record` | `v.record(v.string(), v.int())` |
| `union` | `v.union([v.string(), v.int()])` |
| `intersection` | `v.intersection([schemaA, schemaB])` |
| `optional` | `v.optional(v.string())` or `v.string().optional()` |
| `nullable` | `v.nullable(v.string())` or `v.string().nullable()` |

Each schema kind should support method chaining for constraints:

```
v.string().minLength(1).maxLength(100).pattern("^[a-z]+$")
v.int().min(0).max(1000)
v.array(v.string()).minItems(1).maxItems(50)
```

### Parse (Throwing)

```
schema.parse(input) -> output
```

- Accepts an input value
- Returns the parsed and validated output
- Throws/panics/raises on validation failure
- The error must contain the full list of validation issues

### Safe Parse (Non-Throwing)

```
schema.safeParse(input) -> ParseResult
```

- Accepts an input value
- Returns a result object, never throws
- On success: `{ success: true, data: output }`
- On failure: `{ success: false, issues: [...] }`

The result type should be idiomatic. In Go, use the `(result, error)` pattern. In Rust, use `Result<T, Vec<Issue>>`. In Java, return a `ParseResult<T>` object.

### Export

```
schema.export(options) -> AnyValiDocument
```

- Serializes the schema to the canonical JSON format
- Supports two modes: `portable` and `extended`
- In portable mode, fails if the schema uses non-portable features
- In extended mode, emits core schema plus extension namespaces

The returned document has this shape:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": { ... },
  "definitions": { ... },
  "extensions": { ... }
}
```

### Import

```
importSchema(document) -> Schema
```

- Accepts a canonical JSON document (or parsed object)
- Returns a usable schema instance
- Fails if the document uses unknown schema kinds, unsupported semantic extensions, or references custom validators

## Parse Pipeline Implementation

Every SDK must implement the parse pipeline in exactly this order. Deviating from this order breaks cross-SDK determinism.

### Step 1: Detect Presence or Absence

Determine whether the input value is present or absent.

- For object fields: the key exists in the input object
- For array items: the index is within bounds
- Distinguish between "absent" and "present but null" -- these are different

```
if field is absent:
    go to step 3 (default materialization)
else:
    go to step 2 (coercion)
```

### Step 2: Coerce (If Configured)

If the schema has coercion configured and the value is present, attempt coercion.

```
if coercion is configured:
    try to coerce the value
    if coercion fails:
        return issue with code "coercion_failed"
    replace value with coerced result
```

Portable coercions:

| From | To | Rules |
|---|---|---|
| string | int | Parse as integer. Reject non-numeric strings, floats-as-strings. |
| string | number | Parse as float. Reject non-numeric strings. |
| string | bool | Accept `"true"`, `"1"` as true. Accept `"false"`, `"0"` as false. Reject all others. |
| string | string (trim) | Remove leading and trailing whitespace. |
| string | string (lower) | Convert all characters to lowercase. |
| string | string (upper) | Convert all characters to uppercase. |

### Step 3: Materialize Defaults

If the value is absent and a default is configured, materialize the default.

```
if value is absent and default exists:
    value = deep_clone(default)
```

Important: defaults must be deep-cloned to prevent mutation of the default value across multiple parse calls.

### Step 4: Validate

Validate the resulting value (whether original, coerced, or defaulted) against all schema constraints.

Validation must check:

1. **Type correctness** -- is the value the right kind? (issue code: `invalid_type`)
2. **Required fields** -- are all required object fields present? (issue code: `required`)
3. **Unknown keys** -- does the object have keys not in the schema? (issue code: `unknown_key`)
4. **Constraints** -- does the value satisfy all constraints?
   - String: `minLength`, `maxLength`, `pattern`, `startsWith`, `endsWith`, `includes`, `format`
   - Numeric: `min`, `max`, `exclusiveMin`, `exclusiveMax`, `multipleOf`
   - Array: `minItems`, `maxItems`
5. **Nested schemas** -- recursively validate nested objects, arrays, tuples, unions, intersections

Collect all issues rather than stopping at the first one. Users expect to see all validation errors at once.

### Step 5: Return Result

Return either the parsed value or the list of issues.

For throwing parse: return the value or throw with the issue list.
For safe parse: return a result object containing either the value or the issue list.

## Issue Codes

Every SDK must use these issue codes consistently. The conformance test suite validates that the correct codes are returned for each scenario.

### `invalid_type`

The value is not the expected type.

```json
{
  "code": "invalid_type",
  "message": "Expected string, received number",
  "path": ["name"],
  "expected": "string",
  "received": "number"
}
```

Use when: a string schema receives a number, an object schema receives an array, etc.

### `required`

A required field is missing from an object.

```json
{
  "code": "required",
  "message": "Required field \"email\" is missing",
  "path": ["email"],
  "expected": "string",
  "received": "undefined"
}
```

Use when: an object schema's required field is absent and has no default.

### `unknown_key`

An object has a key that is not defined in the schema, and the unknown key mode is `reject`.

```json
{
  "code": "unknown_key",
  "message": "Unknown key \"foo\" is not allowed",
  "path": ["foo"],
  "expected": "never",
  "received": "\"bar\""
}
```

Use when: unknown key mode is `reject` and the input contains unrecognized keys.

### `too_small`

A value is below the minimum constraint.

```json
{
  "code": "too_small",
  "message": "Value must be at least 0",
  "path": ["age"],
  "expected": ">= 0",
  "received": "-5"
}
```

Use when: `min`, `exclusiveMin`, `minLength`, or `minItems` constraint is violated.

### `too_large`

A value exceeds the maximum constraint.

```json
{
  "code": "too_large",
  "message": "Value must be at most 100",
  "path": ["score"],
  "expected": "<= 100",
  "received": "150"
}
```

Use when: `max`, `exclusiveMax`, `maxLength`, or `maxItems` constraint is violated.

### `invalid_string`

A string fails a format, pattern, or other string-specific constraint.

```json
{
  "code": "invalid_string",
  "message": "Invalid email format",
  "path": ["email"],
  "expected": "email",
  "received": "\"not-an-email\""
}
```

Use when: `format`, `pattern`, `startsWith`, `endsWith`, or `includes` constraint is violated.

### `invalid_number`

A numeric value fails a numeric-specific constraint (beyond range).

```json
{
  "code": "invalid_number",
  "message": "Value must be a multiple of 5",
  "path": ["quantity"],
  "expected": "multiple of 5",
  "received": "7"
}
```

Use when: `multipleOf` constraint is violated, or the value is `NaN`/`Infinity` where not allowed.

### `invalid_literal`

A literal value does not match.

```json
{
  "code": "invalid_literal",
  "message": "Expected \"active\", received \"inactive\"",
  "path": ["status"],
  "expected": "\"active\"",
  "received": "\"inactive\""
}
```

Use when: a literal schema receives a different value.

### `invalid_union`

No branch of a union matched.

```json
{
  "code": "invalid_union",
  "message": "Value does not match any union member",
  "path": ["value"],
  "expected": "string | number",
  "received": "boolean"
}
```

Use when: a union schema's input fails validation against all branches.

### `custom_validation_not_portable`

A schema references a custom validator that cannot be serialized.

Use when: exporting in portable mode and the schema has custom validators, or importing a schema that references custom validators.

### `unsupported_extension`

A required semantic extension is not available.

Use when: importing a schema with semantic extensions that the SDK does not understand, and no default fallback is provided.

### `unsupported_schema_kind`

A schema kind in the document is not recognized.

Use when: importing a document that contains a `kind` value the SDK does not implement.

### `coercion_failed`

A coercion attempt did not succeed.

```json
{
  "code": "coercion_failed",
  "message": "Cannot coerce \"abc\" to int",
  "path": ["port"],
  "expected": "int",
  "received": "\"abc\""
}
```

Use when: a configured coercion could not convert the input value.

### `default_invalid`

A default value does not pass validation.

Use when: the materialized default value fails the schema's own constraints. This indicates a schema authoring error.

## Conformance Test Runner Implementation

### Test Corpus Structure

The shared conformance test suite is a collection of JSON files in `spec/tests/`. Each file contains test cases in this format:

```json
{
  "suite": "string-constraints",
  "tests": [
    {
      "description": "minLength rejects short strings",
      "schema": {
        "kind": "string",
        "minLength": 3
      },
      "input": "ab",
      "expected": {
        "success": false,
        "issues": [
          {
            "code": "too_small",
            "path": []
          }
        ]
      }
    },
    {
      "description": "minLength accepts valid strings",
      "schema": {
        "kind": "string",
        "minLength": 3
      },
      "input": "abc",
      "expected": {
        "success": true,
        "data": "abc"
      }
    }
  ]
}
```

### Building the Runner

Each SDK must implement a test runner that:

1. **Reads all test fixture files** from the conformance corpus directory.
2. **Imports each test schema** using the SDK's `importSchema` function.
3. **Runs `safeParse`** with the provided input.
4. **Asserts the result** matches the expected outcome:
   - For success cases: verify `success` is `true` and `data` matches.
   - For failure cases: verify `success` is `false`, the expected issue codes are present, and the paths match.
5. **Reports results** in a standard format.

### Example Runner (Go)

```go
package conformance

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    v "github.com/BetterCorp/AnyVali-go"
)

type TestSuite struct {
    Suite string     `json:"suite"`
    Tests []TestCase `json:"tests"`
}

type TestCase struct {
    Description string          `json:"description"`
    Schema      json.RawMessage `json:"schema"`
    Input       any             `json:"input"`
    Expected    ExpectedResult  `json:"expected"`
}

type ExpectedResult struct {
    Success bool            `json:"success"`
    Data    any             `json:"data,omitempty"`
    Issues  []ExpectedIssue `json:"issues,omitempty"`
}

type ExpectedIssue struct {
    Code string `json:"code"`
    Path []any  `json:"path"`
}

func TestConformance(t *testing.T) {
    files, _ := filepath.Glob("../../spec/tests/*.json")
    for _, file := range files {
        data, _ := os.ReadFile(file)
        var suite TestSuite
        json.Unmarshal(data, &suite)

        t.Run(suite.Suite, func(t *testing.T) {
            for _, tc := range suite.Tests {
                t.Run(tc.Description, func(t *testing.T) {
                    // Import the schema from JSON
                    var schemaNode map[string]any
                    json.Unmarshal(tc.Schema, &schemaNode)
                    schema, err := v.ImportSchemaNode(schemaNode)
                    if err != nil {
                        t.Fatalf("Failed to import schema: %v", err)
                    }

                    // Run safeParse
                    result := schema.SafeParse(tc.Input)

                    // Assert
                    if result.Success != tc.Expected.Success {
                        t.Errorf("Expected success=%v, got %v",
                            tc.Expected.Success, result.Success)
                    }

                    if tc.Expected.Success {
                        // Compare parsed data
                        assertDeepEqual(t, tc.Expected.Data, result.Data)
                    } else {
                        // Compare issue codes and paths
                        assertIssuesMatch(t, tc.Expected.Issues, result.Issues)
                    }
                })
            }
        })
    }
}
```

### Example Runner (Python)

```python
import json
import glob
import pytest
import anyvali as v


def load_test_suites():
    suites = []
    for path in glob.glob("../../spec/tests/*.json"):
        with open(path) as f:
            suites.append(json.load(f))
    return suites


@pytest.mark.parametrize("suite", load_test_suites(), ids=lambda s: s["suite"])
def test_conformance(suite):
    for tc in suite["tests"]:
        schema = v.import_schema_node(tc["schema"])
        result = schema.safe_parse(tc["input"])

        assert result.success == tc["expected"]["success"], (
            f"[{tc['description']}] "
            f"Expected success={tc['expected']['success']}, "
            f"got {result.success}"
        )

        if tc["expected"]["success"]:
            assert result.data == tc["expected"]["data"]
        else:
            expected_codes = {issue["code"] for issue in tc["expected"]["issues"]}
            actual_codes = {issue.code for issue in result.issues}
            assert expected_codes <= actual_codes, (
                f"[{tc['description']}] Missing issue codes: "
                f"{expected_codes - actual_codes}"
            )
```

### What the Runner Must Verify

The conformance test corpus covers these categories:

- **Core typing**: valid and invalid primitive values, explicit numeric widths
- **Defaults**: missing field gets default, present field is not overwritten, invalid defaults are rejected
- **Coercion**: successful and failed coercions, correct ordering with validation
- **Objects**: unknown key handling in all three modes
- **Composition**: unions, intersections, recursive refs
- **Portability**: round-trip export/import, portable mode rejection, extension handling
- **Numeric safety**: `number` as float64, `int` as int64, narrowing rejection

## Extension Handling

### Registering Extensions

An SDK may register extension handlers for its own namespace:

```typescript
// JS SDK registering a JS-specific extension handler
v.registerExtension("js", "brandedType", {
  criticality: "informational",
  onImport: (value, schemaNode) => {
    // Attach brand metadata to the schema
    schemaNode.meta.brand = value;
  },
  onExport: (schemaNode) => {
    // Emit brand metadata
    return schemaNode.meta?.brand;
  },
});
```

### Extension Criticality

Every extension must declare its criticality:

- **informational**: safe to ignore. The schema validates correctly without it.
- **semantic**: required for correct behavior. Import must fail if the handler is missing.

### Import Behavior

When importing a schema with extensions:

```
for each extension namespace in the document:
    if namespace matches this SDK's language:
        apply all extensions using registered handlers
    elif namespace is "default":
        attempt to apply using default handlers
    else:
        for each extension in the namespace:
            if criticality is "informational":
                skip (safe to ignore)
            elif criticality is "semantic":
                if "default" namespace has this extension:
                    apply default
                else:
                    fail import with "unsupported_extension"
```

### Export Behavior

When exporting with extended mode:

```
emit portable core schema
for each registered extension in this SDK:
    add to extensions under this SDK's namespace
    if default fallback behavior exists:
        also add to extensions under "default" namespace
```

## Naming Conventions

Each SDK should adapt API names to the conventions of its language. The spec requires semantic equivalence, not naming equivalence.

### Language-Specific Conventions

| Concept | JS/TS | Python | Go | Java | C# | Rust |
|---|---|---|---|---|---|---|
| String schema | `v.string()` | `v.string()` | `v.String()` | `V.string()` | `V.String()` | `v::string()` |
| Min length | `.minLength(n)` | `.min_length(n)` | `.MinLength(n)` | `.minLength(n)` | `.MinLength(n)` | `.min_length(n)` |
| Safe parse | `.safeParse(x)` | `.safe_parse(x)` | `.SafeParse(x)` | `.safeParse(x)` | `.SafeParse(x)` | `.safe_parse(x)` |
| Export | `.export(opts)` | `.export(opts)` | `.Export(opts)` | `.export(opts)` | `.Export(opts)` | `.export(opts)` |
| Import | `importSchema(d)` | `import_schema(d)` | `ImportSchema(d)` | `importSchema(d)` | `ImportSchema(d)` | `import_schema(d)` |
| Unknown keys | `.unknownKeys("strip")` | `.unknown_keys("strip")` | `.UnknownKeys(Strip)` | `.unknownKeys(STRIP)` | `.UnknownKeys(Strip)` | `.unknown_keys(Strip)` |

### General Rules

- Use the language's standard casing: camelCase for JS/Java/Kotlin, snake_case for Python/Rust/Ruby/PHP, PascalCase for Go/C#
- Use the language's error handling idiom: exceptions for Java/Python/C#/Ruby/PHP/Kotlin, Result types for Rust, error returns for Go, thrown errors for JS
- Use the language's module/package system naturally
- Use the language's type system: generics in Java/C#/Rust/Kotlin/C++, type parameters in Go, type hints in Python, TypeScript generics in JS/TS

### Type Inference and Generics

Where possible, SDKs should provide type-level output inference:

```typescript
// TypeScript: inferred output type
const UserSchema = v.object({
  name: v.string(),
  age: v.int(),
});
type User = v.Infer<typeof UserSchema>;
// => { name: string; age: number }
```

```rust
// Rust: derive output type
let user_schema = v::object(fields![
    "name" => v::string(),
    "age" => v::int(),
]);
// user_schema.parse(input) returns a typed struct or map
```

This is a language-specific ergonomic feature and is not part of the portable contract.

## Implementation Checklist

Use this checklist when building a new SDK:

### Phase 1: Core

- [ ] Implement all 25 schema kinds
- [ ] Implement all string constraints (minLength, maxLength, pattern, startsWith, endsWith, includes, format)
- [ ] Implement all numeric constraints (min, max, exclusiveMin, exclusiveMax, multipleOf)
- [ ] Implement all array constraints (minItems, maxItems)
- [ ] Implement object schemas (required/optional fields, unknown key modes)
- [ ] Implement the five-step parse pipeline
- [ ] Implement throwing parse
- [ ] Implement safe parse
- [ ] Implement all 14 issue codes
- [ ] Implement default materialization
- [ ] Implement all portable coercions

### Phase 2: Interchange

- [ ] Implement schema export (portable mode)
- [ ] Implement schema export (extended mode)
- [ ] Implement schema import
- [ ] Implement definitions and refs
- [ ] Implement extension import logic (informational vs semantic)
- [ ] Implement export failure for non-portable features

### Phase 3: Conformance

- [ ] Build the conformance test runner
- [ ] Pass all core typing tests
- [ ] Pass all default tests
- [ ] Pass all coercion tests
- [ ] Pass all object tests
- [ ] Pass all composition tests
- [ ] Pass all portability tests
- [ ] Pass all numeric safety tests

### Phase 4: Polish

- [ ] Add type inference / generics where the language supports it
- [ ] Write SDK-specific documentation
- [ ] Ensure idiomatic error messages
- [ ] Performance testing and optimization
- [ ] Publish package to the language's package registry

## Common Implementation Pitfalls

### Default mutation

If the default value is a mutable object (array, map), failing to deep-clone it will cause mutations from one parse call to leak into subsequent calls. Always clone defaults.

### Coercion ordering

Coercions must run before defaults. If you apply defaults first, coercion never gets a chance to convert present values. If you validate before coercing, you will reject valid coercible inputs.

### Collecting all issues

The parse pipeline must collect all validation issues, not stop at the first one. For objects, validate every field. For arrays, validate every item. For unions, collect issues from all branches (to report which ones failed and why).

### Unknown key mode default

The default unknown key mode is `reject`. If your SDK defaults to a permissive mode, you will fail conformance tests. Reject is the safe default.

### Numeric alias resolution

`number` and `int` are aliases for `float64` and `int64` respectively. During export, you may preserve the alias for readability, but import must treat them as semantically identical to their canonical forms.

### Path representation

Issue paths are ordered lists of keys and indexes. Object keys are strings, array indexes are integers. Maintain this format consistently:

```json
["users", 0, "email"]
```

Not:

```
"users.0.email"
"users[0].email"
```

### Recursive schemas

Definitions and refs enable recursive schemas. Your import logic must handle circular references without infinite loops. Typically this means a two-pass approach: first register all definition names, then resolve references.

### Float precision in JSON

JSON numbers have arbitrary precision in the specification, but most parsers use float64. When exporting numeric constraints (like `min: 0.1`), the value may not be exactly representable in float64. Use the host language's standard JSON serialization and document any precision limitations.

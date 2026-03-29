# AnyVali v1.0 Normative Specification

This document is the normative specification for AnyVali v1.0. All conforming SDKs
MUST implement the semantics described here. The key words "MUST", "MUST NOT",
"REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY",
and "OPTIONAL" are to be interpreted as described in RFC 2119.

---

## 1. Schema Kinds

AnyVali v1.0 defines exactly 24 schema kinds. Every schema node carries a `kind`
discriminator that determines its validation semantics.

### 1.1 Special Kinds

| Kind | Semantics |
|-----------|-----------|
| `any` | Accepts every value including `null` and `undefined`/absent. Never produces a type error. The output value is the input value unchanged. |
| `unknown` | Identical to `any` at validation time: accepts every value. Differs only in static type intent -- SDKs with static type systems SHOULD map `unknown` to their "unknown" or top type rather than an unchecked "any". |
| `never` | Rejects every value unconditionally. Always produces an `invalid_type` issue with `expected: "never"`. |

### 1.2 Null and Boolean

| Kind | Semantics |
|--------|-----------|
| `null` | Accepts only the JSON `null` value. Rejects everything else with `invalid_type`. |
| `bool` | Accepts only JSON booleans (`true`, `false`). Rejects everything else with `invalid_type`. |

### 1.3 String

| Kind | Semantics |
|----------|-----------|
| `string` | Accepts only JSON string values. Rejects non-strings with `invalid_type`. Supports string constraints (see Section 3.1). |

### 1.4 Numeric Kinds

AnyVali defines 12 numeric kinds organized in two families.

#### Float family

| Kind | Semantics |
|-----------|-----------|
| `number` | Alias for `float64`. Accepts any JSON number representable as an IEEE 754 binary64 value. |
| `float32` | Accepts any JSON number representable as an IEEE 754 binary32 value without loss of precision beyond the binary32 range. The SDK MUST reject values outside the binary32 representable range. |
| `float64` | Accepts any JSON number representable as an IEEE 754 binary64 value. This is the canonical float kind. |

Float kinds reject non-numeric input with `invalid_type`. They reject NaN and
Infinity (these are not representable in JSON).

#### Integer family

| Kind | Semantics |
|----------|-----------|
| `int` | Alias for `int64`. Accepts any JSON number that is a mathematical integer within the int64 range. |
| `int8` | Accepts integers in [-128, 127]. |
| `int16` | Accepts integers in [-32768, 32767]. |
| `int32` | Accepts integers in [-2147483648, 2147483647]. |
| `int64` | Accepts integers in [-9223372036854775808, 9223372036854775807]. |
| `uint8` | Accepts integers in [0, 255]. |
| `uint16` | Accepts integers in [0, 65535]. |
| `uint32` | Accepts integers in [0, 4294967295]. |
| `uint64` | Accepts integers in [0, 18446744073709551615]. |

Integer kinds MUST reject:
- Non-numeric input with `invalid_type`.
- Numbers with a fractional part with `invalid_type` (expected: the integer kind).
- Integers outside the kind's range with `too_small` or `too_large`.

### 1.5 Literal

| Kind | Semantics |
|-----------|-----------|
| `literal` | Accepts only a single constant value. The literal value MUST be a JSON primitive: string, number, boolean, or null. Comparison is by value equality. Rejects mismatches with `invalid_literal`. |

Node shape:
```json
{ "kind": "literal", "value": "exact-value" }
```

### 1.6 Enum

| Kind | Semantics |
|--------|-----------|
| `enum` | Accepts any value that matches one of the listed constant values. Each value MUST be a JSON primitive. Rejects mismatches with `invalid_type` and expected set. |

Node shape:
```json
{ "kind": "enum", "values": ["A", "B", "C"] }
```

### 1.7 Array

| Kind | Semantics |
|---------|-----------|
| `array` | Accepts JSON arrays where every element validates against the `items` schema. Rejects non-arrays with `invalid_type`. Supports array constraints (see Section 3.3). |

Node shape:
```json
{ "kind": "array", "items": { ... }, "minItems": 0, "maxItems": 100 }
```

### 1.8 Tuple

| Kind | Semantics |
|---------|-----------|
| `tuple` | Accepts JSON arrays with a fixed length matching `elements`. Each element at index `i` MUST validate against `elements[i]`. Extra elements are rejected. Fewer elements are rejected. Rejects non-arrays with `invalid_type`. |

Node shape:
```json
{ "kind": "tuple", "elements": [ { ... }, { ... } ] }
```

### 1.9 Object

| Kind | Semantics |
|----------|-----------|
| `object` | Accepts JSON objects. Each key listed in `properties` is validated against the corresponding schema. Keys listed in `required` MUST be present; absent required keys produce a `required` issue. Unknown keys are handled according to `unknownKeys` mode (default: `reject`). |

Node shape:
```json
{
  "kind": "object",
  "properties": { "name": { "kind": "string" } },
  "required": ["name"],
  "unknownKeys": "reject"
}
```

### 1.10 Record

| Kind | Semantics |
|----------|-----------|
| `record` | Accepts JSON objects where every key is a string and every value validates against the `values` schema. The set of keys is unconstrained. Rejects non-objects with `invalid_type`. |

Node shape:
```json
{ "kind": "record", "values": { "kind": "string" } }
```

### 1.11 Union

| Kind | Semantics |
|---------|-----------|
| `union` | Accepts a value if it validates against at least one of the `variants`. Variants are tried in order. The first successful variant determines the output. If no variant matches, an `invalid_union` issue is produced. |

Node shape:
```json
{ "kind": "union", "variants": [ { ... }, { ... } ] }
```

### 1.12 Intersection

| Kind | Semantics |
|----------------|-----------|
| `intersection` | Accepts a value only if it validates against every schema in `allOf`. All schemas are evaluated. All issues from all failing schemas are collected. The output is the merged result of all schemas. For object intersections, properties are merged; conflicting constraints are conjunctive (all must pass). |

Node shape:
```json
{ "kind": "intersection", "allOf": [ { ... }, { ... } ] }
```

### 1.13 Optional

| Kind | Semantics |
|------------|-----------|
| `optional` | Wraps an inner schema. Accepts absent/undefined values (produces no output and no issue). If a value is present, it MUST validate against the `schema`. `null` is NOT considered absent -- only truly missing values are absent. |

Node shape:
```json
{ "kind": "optional", "schema": { ... } }
```

### 1.14 Nullable

| Kind | Semantics |
|------------|-----------|
| `nullable` | Wraps an inner schema. Accepts `null` (output is `null`). If the value is not `null`, it MUST validate against the `schema`. |

Node shape:
```json
{ "kind": "nullable", "schema": { ... } }
```

### 1.15 Ref

| Kind | Semantics |
|-------|-----------|
| `ref` | References a named definition. The `ref` field MUST be a JSON pointer of the form `#/definitions/<name>`. Validation delegates to the referenced schema. Enables recursion and reuse. |

Node shape:
```json
{ "kind": "ref", "ref": "#/definitions/User" }
```

---

## 2. Parse Pipeline

All conforming SDKs MUST implement parsing as a deterministic 5-step pipeline
applied to every schema node. The steps MUST execute in the following order.

### Step 1: Presence

Determine whether the input value is present or absent.

- "Absent" means the key does not exist in the parent object. For root-level
  parsing, the value is always considered present.
- `null` is present, not absent.

If the schema is `optional` and the value is absent, parsing succeeds with no
output. Skip remaining steps.

### Step 2: Coercion

If the value is present AND the schema node has coercion configured, attempt
the configured coercion.

- If coercion succeeds, the coerced value replaces the input for subsequent steps.
- If coercion fails, produce a `coercion_failed` issue and stop (do not proceed
  to validation).

See Section 5 for coercion rules.

### Step 3: Defaults

If the value is absent AND the schema node has a `default` configured, materialize
the default value.

- The default value replaces the absent input for subsequent steps.
- Defaults MUST be pure JSON data. Function defaults are not portable.

### Step 4: Validate

Validate the resulting value (original, coerced, or defaulted) against the
schema's kind and constraints.

- Kind validation: check that the value matches the expected type.
- Constraint validation: check all applicable constraints (see Section 3).
- All failing constraints produce issues. Validation collects ALL issues, it
  does not short-circuit on the first failure within a node.

### Step 5: Result

Return the parse result.

- On success: the parsed (possibly coerced/defaulted) value.
- On failure: a list of `ValidationIssue` objects.

---

## 3. Validation Constraints

### 3.1 String Constraints

All string constraints are OPTIONAL on `string` schema nodes. When present,
they are evaluated against the string value after coercion.

| Constraint | Type | Semantics |
|-------------|---------|-----------|
| `minLength` | integer | String length (in Unicode code points) MUST be >= value. Issue: `too_small`. |
| `maxLength` | integer | String length (in Unicode code points) MUST be <= value. Issue: `too_large`. |
| `pattern` | string | String MUST match the regular expression. The pattern follows ECMA-262 (JavaScript) regex syntax as the portable baseline. Issue: `invalid_string`. |
| `startsWith` | string | String MUST begin with the given prefix. Issue: `invalid_string`. |
| `endsWith` | string | String MUST end with the given suffix. Issue: `invalid_string`. |
| `includes` | string | String MUST contain the given substring. Issue: `invalid_string`. |
| `format` | string | String MUST match the named format. See Section 3.2. Issue: `invalid_string`. |

### 3.2 Format Validators

The following format identifiers are portable in v1.0. SDKs MUST implement
validation for each.

#### `email`
The string MUST match a simplified email format: one or more characters, then
`@`, then one or more characters containing at least one `.` after the `@`.
Formally: `/^[^\s@]+@[^\s@]+\.[^\s@]+$/`. This is intentionally simplified
and does not attempt full RFC 5322 compliance.

#### `url`
The string MUST begin with `http://` or `https://` and be followed by at least
one character. Formally: `/^https?:\/\/.+$/`.

#### `uuid`
The string MUST be a UUID in the canonical 8-4-4-4-12 hex format (case
insensitive). Formally: `/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i`.

#### `ipv4`
The string MUST be a dotted-decimal IPv4 address with four octets, each in
[0, 255], no leading zeros. Formally: four groups of 1-3 digits separated
by `.`, each parsed as a decimal integer in [0, 255], and no group has a
leading zero unless the group is exactly `0`.

#### `ipv6`
The string MUST be a valid IPv6 address per RFC 5952 simplified rules: 8
groups of 1-4 hex digits separated by `:`, or valid use of `::` for
zero-group compression, or IPv4-mapped suffix. SDKs SHOULD accept the full
range of valid representations.

#### `date`
The string MUST be a date in `YYYY-MM-DD` format where the date is
calendrically valid (e.g., `2024-02-29` is valid, `2023-02-29` is not).

#### `date-time`
The string MUST be an ISO 8601 date-time with timezone offset or `Z`.
The minimum accepted format is `YYYY-MM-DDTHH:MM:SSZ` or
`YYYY-MM-DDTHH:MM:SS+HH:MM`. Fractional seconds are OPTIONAL and accepted
if present.

### 3.3 Numeric Constraints

All numeric constraints are OPTIONAL on numeric schema kinds (`number`,
`float32`, `float64`, `int`, `int8`-`int64`, `uint8`-`uint64`).

| Constraint | Type | Semantics |
|----------------|--------|-----------|
| `min` | number | Value MUST be >= `min`. Issue: `too_small`. |
| `max` | number | Value MUST be <= `max`. Issue: `too_large`. |
| `exclusiveMin` | number | Value MUST be > `exclusiveMin`. Issue: `too_small`. |
| `exclusiveMax` | number | Value MUST be < `exclusiveMax`. Issue: `too_large`. |
| `multipleOf` | number | Value MUST be an exact multiple of `multipleOf`. For floats, SDKs MUST use a tolerance of at most 1e-10 for the remainder check. Issue: `invalid_number`. |

### 3.4 Array Constraints

| Constraint | Type | Semantics |
|------------|---------|-----------|
| `minItems` | integer | Array length MUST be >= `minItems`. Issue: `too_small`. |
| `maxItems` | integer | Array length MUST be <= `maxItems`. Issue: `too_large`. |

---

## 4. Issue Codes

AnyVali v1.0 defines exactly 14 issue codes. SDKs MUST use these codes for
the described conditions and MUST NOT repurpose them.

### 4.1 Issue Object Shape

```json
{
  "code": "invalid_type",
  "path": ["user", "age"],
  "message": "Expected int, received string",
  "expected": "int",
  "received": "string",
  "meta": {}
}
```

- `code` (string, REQUIRED): One of the 14 codes below.
- `path` (array, REQUIRED): Ordered list of string keys and integer indices
  from root to the failing value.
- `message` (string, REQUIRED): Human-readable description.
- `expected` (string, OPTIONAL): What was expected.
- `received` (string, OPTIONAL): What was received.
- `meta` (object, OPTIONAL): Additional structured data.

### 4.2 Code Definitions

| Code | When Produced |
|-------------------------------------|---------------|
| `invalid_type` | The value's runtime type does not match the schema kind. E.g., string where int expected, object where array expected, non-integer where integer kind expected. |
| `required` | A required object property is absent. |
| `unknown_key` | An object contains a key not listed in `properties` and `unknownKeys` is `reject`. |
| `too_small` | A numeric value is below `min` or `exclusiveMin`; a string is shorter than `minLength`; an array has fewer than `minItems` elements; an integer is below the kind's minimum range. |
| `too_large` | A numeric value exceeds `max` or `exclusiveMax`; a string exceeds `maxLength`; an array exceeds `maxItems` elements; an integer exceeds the kind's maximum range. |
| `invalid_string` | A string fails a `pattern`, `startsWith`, `endsWith`, `includes`, or `format` constraint. |
| `invalid_number` | A numeric value fails a `multipleOf` constraint. |
| `invalid_literal` | A value does not match the expected `literal` value. |
| `invalid_union` | A value does not match any variant of a `union`. |
| `custom_validation_not_portable` | A schema depends on a custom (non-portable) validator and is being exported in `portable` mode, or imported by a runtime that does not have the required handler. |
| `unsupported_extension` | A schema requires a semantic extension namespace that the importing runtime does not support and no `default` fallback is available. |
| `unsupported_schema_kind` | A schema uses a `kind` value not recognized by the importing runtime. |
| `coercion_failed` | A configured coercion could not convert the input value. |
| `default_invalid` | A materialized default value does not pass validation against the schema. |

---

## 5. Coercions

Coercions are configured per schema node via the `coerce` property. Coercions
are OPTIONAL -- if not configured, no coercion is attempted.

### 5.1 Portable Coercions

| Coercion | Input Type | Target Kind | Rule |
|-----------------|------------|-------------|------|
| `string->int` | string | any integer kind | Parse string as decimal integer. Leading/trailing whitespace is trimmed. Reject if non-numeric, fractional, or out of range. |
| `string->number` | string | any float kind | Parse string as decimal floating-point. Leading/trailing whitespace is trimmed. Reject NaN, Infinity. |
| `string->bool` | string | `bool` | `"true"`, `"1"` => `true`; `"false"`, `"0"` => `false` (case-insensitive). All other strings are rejected. |
| `trim` | string | `string` | Remove leading and trailing whitespace. |
| `lower` | string | `string` | Convert to lowercase (Unicode-aware where feasible, ASCII minimum). |
| `upper` | string | `string` | Convert to uppercase (Unicode-aware where feasible, ASCII minimum). |

### 5.2 Coercion Node Shape

```json
{ "kind": "int", "coerce": "string->int" }
```

For string transforms that can be combined:

```json
{ "kind": "string", "coerce": ["trim", "lower"] }
```

When `coerce` is an array, transforms are applied left to right.

### 5.3 Coercion Failure

If the input cannot be coerced, the SDK MUST produce a `coercion_failed` issue
and MUST NOT proceed to validation. The original (pre-coercion) value is
reported in the issue's `received` field.

---

## 6. Defaults

### 6.1 Default Node Shape

```json
{ "kind": "string", "default": "hello" }
```

### 6.2 Rules

1. Defaults are applied only when the value is absent (Step 3 of the pipeline).
2. A present value (including `null`) is never overwritten by a default.
3. The materialized default MUST be validated (Step 4). If the default fails
   validation, a `default_invalid` issue is produced.
4. Defaults MUST be representable as portable JSON. Function-based defaults
   are a Tier 3 (local-only) feature and MUST NOT appear in exported schemas.

---

## 7. Unknown Key Modes

Object schemas support three modes for handling keys not listed in `properties`.

| Mode | Semantics |
|----------|-----------|
| `reject` | Unknown keys cause an `unknown_key` issue for each. This is the DEFAULT. |
| `strip` | Unknown keys are silently removed from the output. No issue is produced. |
| `allow` | Unknown keys are passed through to the output unchanged. No issue is produced. |

The mode is specified via `unknownKeys` on the object node. If omitted, `reject`
is used.

---

## 8. Portability Tiers

### Tier 1: Portable Core

Schemas using only the 24 standard kinds, portable constraints, portable
coercions, portable defaults, and refs to definitions are Tier 1.

Guarantees:
- Export succeeds in any SDK.
- Import succeeds in any SDK.
- Validation produces identical results across SDKs for the same input.

### Tier 2: Portable Core + Extensions

Schemas that use semantic extension namespaces in addition to the portable core.

Guarantees:
- Export succeeds in `extended` mode.
- Import succeeds only in SDKs that support the required extension namespaces
  (or have a suitable `default` extension fallback).
- Portable portions validate identically.

### Tier 3: Local-Only

Features that never leave the SDK boundary: custom validators, function-based
defaults, async validation, branded types.

Guarantees:
- Cannot be exported in `portable` mode.
- May be exported in `extended` mode as extension metadata, but importing
  SDKs are not required to support them.

---

## 9. Export Modes

### 9.1 `portable` Mode

- The SDK MUST inspect the schema graph for any non-portable features.
- If any are found, export MUST fail with a `custom_validation_not_portable`
  or `unsupported_extension` issue.
- The output document uses only portable core features.

### 9.2 `extended` Mode

- The SDK emits the portable core schema plus any extension namespaces.
- Extension data is placed under `extensions.<namespace>` in the document.
- Node-level extensions are placed under `extensions` on individual nodes.

---

## 10. Numeric Semantics

### 10.1 Why `number` = `float64`

Many mainstream languages have different default numeric types:
- JavaScript: all numbers are float64.
- Python: `int` is arbitrary precision; `float` is float64.
- Go: `int` varies by platform (32 or 64 bit).
- Java: `int` is int32; `long` is int64.
- C#: `int` is int32; `long` is int64.

Using `float64` as the default for `number` ensures:
- Maximum range for cross-language interchange.
- No silent truncation when importing schemas from JS (where all numbers are float64).
- Explicit narrowing via `float32` when needed.

### 10.2 Why `int` = `int64`

Using `int64` as the default for `int` ensures:
- Values up to 2^63-1 are representable without overflow.
- No platform-dependent behavior (unlike Go's `int` or C's `int`).
- Unsigned values have explicit kinds (`uint8`-`uint64`).
- Silent narrowing from `int64` to `int32` is prevented -- schema authors
  must explicitly choose `int32` when they want the smaller range.

### 10.3 Narrowing Rules

When a schema specifies a specific numeric width (e.g., `int32`), the importing
SDK MUST either:
1. Validate within the exact range, OR
2. Reject the schema kind as unsupported with `unsupported_schema_kind`.

Silently widening is acceptable (e.g., storing an `int32` value in a 64-bit
variable). Silently narrowing is NEVER acceptable.

---

## 11. Format Validator Reference

See Section 3.2 for the complete format validator specifications. SDKs MUST
implement all 7 format validators identically.

Summary of portable formats:
1. `email` -- simplified email check
2. `url` -- HTTP/HTTPS URL check
3. `uuid` -- canonical UUID hex format
4. `ipv4` -- dotted-decimal, no leading zeros
5. `ipv6` -- RFC 5952 representations
6. `date` -- `YYYY-MM-DD`, calendrically valid
7. `date-time` -- ISO 8601 with timezone

---

## 12. References and Definitions

### 12.1 Definition Map

The top-level `definitions` object maps string names to schema nodes:

```json
{
  "definitions": {
    "User": { "kind": "object", "properties": { ... }, "required": [...] },
    "Address": { "kind": "object", "properties": { ... }, "required": [...] }
  }
}
```

### 12.2 Ref Resolution

A `ref` node MUST use a JSON pointer of the form `#/definitions/<name>`.
Resolution is performed within the same document. Forward references and
circular references are permitted.

When validating, a `ref` node delegates entirely to the referenced schema.
Infinite recursion MUST be handled by SDKs (e.g., via stack depth limits
or cycle detection).

---

## 13. Versioning

### 13.1 `anyvaliVersion`

The `anyvaliVersion` field identifies the specification family version. For
this specification: `"1.0"`.

SDKs MUST reject documents with an `anyvaliVersion` they do not support.

### 13.2 `schemaVersion`

The `schemaVersion` field identifies the interchange format version within
the specification family. For this specification: `"1"`.

SDKs MUST reject documents with a `schemaVersion` they do not support.

# AnyVali v1.0 JSON Interchange Format Specification

This document specifies the canonical JSON document structure used to export
and import AnyVali schemas across SDKs.

---

## 1. Top-Level Document Structure

Every AnyVali canonical JSON document MUST be a JSON object with exactly these
top-level properties:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": { ... },
  "definitions": { ... },
  "extensions": { ... }
}
```

### 1.1 Required Properties

| Property | Type | Description |
|------------------|--------|-------------|
| `anyvaliVersion` | string | The specification family version. MUST be `"1.0"` for this version. |
| `schemaVersion` | string | The interchange format version. MUST be `"1"` for this version. |
| `root` | object | The root schema node. MUST be a valid schema node (see Section 2). |
| `definitions` | object | A map of definition names to schema nodes. MAY be empty (`{}`). |
| `extensions` | object | A map of extension namespace names to extension data. MAY be empty (`{}`). |

All five properties are REQUIRED. SDKs MUST reject documents missing any of them.

### 1.2 No Additional Properties

The top-level document MUST NOT contain properties other than those listed above.

---

## 2. Schema Node Shapes

Every schema node is a JSON object with a `kind` property that acts as the
discriminator. Additional properties depend on the kind.

### 2.1 Common Optional Properties

The following properties MAY appear on any schema node:

| Property | Type | Description |
|------------|-----------------|-------------|
| `default` | any JSON value | Default value materialized when input is absent. |
| `coerce` | string or array | Coercion identifier or ordered list of identifiers. |
| `extensions` | object | Node-level extension data (keyed by namespace). |

### 2.2 Node Shapes by Kind

#### `any`
```json
{ "kind": "any" }
```
No additional required properties.

#### `unknown`
```json
{ "kind": "unknown" }
```
No additional required properties.

#### `never`
```json
{ "kind": "never" }
```
No additional required properties.

#### `null`
```json
{ "kind": "null" }
```
No additional required properties.

#### `bool`
```json
{ "kind": "bool" }
```
No additional required properties.

#### `string`
```json
{
  "kind": "string",
  "minLength": 1,
  "maxLength": 255,
  "pattern": "^[a-z]+$",
  "startsWith": "prefix",
  "endsWith": "suffix",
  "includes": "substring",
  "format": "email"
}
```
All constraint properties are OPTIONAL.

| Property | Type | Description |
|-------------|---------|-------------|
| `minLength` | integer | Minimum string length in Unicode code points. |
| `maxLength` | integer | Maximum string length in Unicode code points. |
| `pattern` | string | ECMA-262 regular expression pattern. |
| `startsWith` | string | Required prefix. |
| `endsWith` | string | Required suffix. |
| `includes` | string | Required substring. |
| `format` | string | Named format identifier. One of: `email`, `url`, `uuid`, `ipv4`, `ipv6`, `date`, `date-time`. |

#### `number` (alias for `float64`)
```json
{
  "kind": "number",
  "min": 0,
  "max": 100,
  "exclusiveMin": -1,
  "exclusiveMax": 101,
  "multipleOf": 0.5
}
```

#### `float32`
```json
{
  "kind": "float32",
  "min": 0,
  "max": 100
}
```

#### `float64`
```json
{
  "kind": "float64",
  "min": 0,
  "max": 100
}
```

#### `int` (alias for `int64`)
```json
{
  "kind": "int",
  "min": 0,
  "max": 1000
}
```

#### `int8`, `int16`, `int32`, `int64`
```json
{
  "kind": "int32",
  "min": 0,
  "max": 100
}
```
Same shape for `int8`, `int16`, `int32`, `int64`. All numeric constraint
properties are OPTIONAL.

#### `uint8`, `uint16`, `uint32`, `uint64`
```json
{
  "kind": "uint8",
  "min": 0,
  "max": 200
}
```
Same shape for all unsigned integer kinds.

**Numeric constraint properties (shared by all numeric kinds):**

| Property | Type | Description |
|----------------|--------|-------------|
| `min` | number | Inclusive minimum. |
| `max` | number | Inclusive maximum. |
| `exclusiveMin` | number | Exclusive minimum (value must be strictly greater). |
| `exclusiveMax` | number | Exclusive maximum (value must be strictly less). |
| `multipleOf` | number | Value must be an exact multiple. |

#### `literal`
```json
{ "kind": "literal", "value": "hello" }
```

| Property | Type | Description |
|----------|-------------|-------------|
| `value` | JSON primitive | REQUIRED. The exact value to match (string, number, boolean, or null). |

#### `enum`
```json
{ "kind": "enum", "values": ["A", "B", "C"] }
```

| Property | Type | Description |
|----------|-------|-------------|
| `values` | array | REQUIRED. Non-empty array of JSON primitive values. |

#### `array`
```json
{
  "kind": "array",
  "items": { "kind": "string" },
  "minItems": 1,
  "maxItems": 50
}
```

| Property | Type | Description |
|------------|---------|-------------|
| `items` | object | REQUIRED. Schema node for array elements. |
| `minItems` | integer | OPTIONAL. Minimum array length. |
| `maxItems` | integer | OPTIONAL. Maximum array length. |

#### `tuple`
```json
{
  "kind": "tuple",
  "elements": [
    { "kind": "string" },
    { "kind": "int" }
  ]
}
```

| Property | Type | Description |
|------------|-------|-------------|
| `elements` | array | REQUIRED. Ordered array of schema nodes. |

#### `object`
```json
{
  "kind": "object",
  "properties": {
    "name": { "kind": "string" },
    "age": { "kind": "int" }
  },
  "required": ["name", "age"],
  "unknownKeys": "reject"
}
```

| Property | Type | Description |
|--------------|--------|-------------|
| `properties` | object | REQUIRED. Map of property names to schema nodes. |
| `required` | array | REQUIRED. Array of property names that must be present. |
| `unknownKeys` | string | OPTIONAL. One of `"reject"`, `"strip"`, `"allow"`. Default: `"reject"`. |

#### `record`
```json
{
  "kind": "record",
  "values": { "kind": "int" }
}
```

| Property | Type | Description |
|----------|--------|-------------|
| `values` | object | REQUIRED. Schema node for all values. |

#### `union`
```json
{
  "kind": "union",
  "variants": [
    { "kind": "string" },
    { "kind": "int" }
  ]
}
```

| Property | Type | Description |
|------------|-------|-------------|
| `variants` | array | REQUIRED. Non-empty array of schema nodes (tried in order). |

#### `intersection`
```json
{
  "kind": "intersection",
  "allOf": [
    { "kind": "object", "properties": { "a": { "kind": "string" } }, "required": ["a"], "unknownKeys": "allow" },
    { "kind": "object", "properties": { "b": { "kind": "int" } }, "required": ["b"], "unknownKeys": "allow" }
  ]
}
```

| Property | Type | Description |
|----------|-------|-------------|
| `allOf` | array | REQUIRED. Array of schema nodes (all must validate). |

#### `optional`
```json
{
  "kind": "optional",
  "schema": { "kind": "string" }
}
```

| Property | Type | Description |
|----------|--------|-------------|
| `schema` | object | REQUIRED. The inner schema to validate present values against. |

#### `nullable`
```json
{
  "kind": "nullable",
  "schema": { "kind": "string" }
}
```

| Property | Type | Description |
|----------|--------|-------------|
| `schema` | object | REQUIRED. The inner schema to validate non-null values against. |

#### `ref`
```json
{ "kind": "ref", "ref": "#/definitions/User" }
```

| Property | Type | Description |
|----------|--------|-------------|
| `ref` | string | REQUIRED. JSON pointer to a definition. MUST match `#/definitions/<name>`. |

---

## 3. Definitions

The `definitions` object is a flat map from string names to schema nodes.

```json
{
  "definitions": {
    "User": {
      "kind": "object",
      "properties": {
        "name": { "kind": "string" },
        "friend": { "kind": "ref", "ref": "#/definitions/User" }
      },
      "required": ["name"],
      "unknownKeys": "reject"
    }
  }
}
```

### 3.1 Naming Rules

Definition names MUST be non-empty strings consisting of ASCII alphanumeric
characters, underscores, and hyphens. Pattern: `^[A-Za-z_][A-Za-z0-9_-]*$`.

### 3.2 Circular References

Circular references between definitions are permitted. SDKs MUST handle them
without infinite loops (e.g., via lazy resolution or depth limits).

---

## 4. Ref Resolution

### 4.1 Pointer Format

All `ref` values use JSON Pointer syntax relative to the document root:

```
#/definitions/<name>
```

No other pointer targets are valid in v1.0. A `ref` pointing outside
`definitions` MUST be rejected.

### 4.2 Resolution Semantics

1. Parse the pointer to extract `<name>`.
2. Look up `<name>` in the document's `definitions` map.
3. If found, the `ref` node is semantically replaced by the definition's
   schema node for validation purposes.
4. If not found, the document is invalid. SDKs MUST reject it at import time.

### 4.3 Ref Does Not Merge

A `ref` node does not merge with or inherit properties from surrounding context.
It delegates entirely to the referenced definition. Properties like `default` or
`coerce` on a `ref` node are applied to the ref wrapper, not the referenced
definition.

---

## 5. Extensions

### 5.1 Document-Level Extensions

```json
{
  "extensions": {
    "js": {
      "brandedTypes": true
    },
    "go": {
      "structTags": { "User": "json" }
    },
    "default": {
      "description": "Fallback extension semantics"
    }
  }
}
```

### 5.2 Extension Namespaces

Each key under `extensions` is a namespace. Reserved namespaces:

| Namespace | Purpose |
|-----------|---------|
| `default` | Fallback semantics when a language-specific namespace is absent. |

Language-specific namespaces SHOULD use the SDK identifier: `js`, `python`,
`go`, `java`, `csharp`, `rust`, `php`, `ruby`, `kotlin`, `cpp`.

### 5.3 Extension Criticality

Extensions MUST declare their criticality:

| Level | Semantics |
|----------------|-----------|
| `informational` | Safe to ignore. Does not affect validation semantics. |
| `semantic` | Required for correct behavior. Import MUST fail if the importing SDK cannot satisfy the extension. |

Criticality is indicated by a `_criticality` key within the namespace object:

```json
{
  "extensions": {
    "js": {
      "_criticality": "informational",
      "description": "JS-specific type hints"
    }
  }
}
```

If `_criticality` is absent, `informational` is assumed.

### 5.4 Node-Level Extensions

Individual schema nodes MAY carry an `extensions` property:

```json
{
  "kind": "string",
  "minLength": 1,
  "extensions": {
    "js": { "customValidator": "isSlug" }
  }
}
```

Node-level extension criticality follows the same rules. If a node-level
semantic extension is not satisfiable, validation of that node MUST produce
an `unsupported_extension` issue.

### 5.5 Import Behavior

When importing a document:

1. Parse and validate the document structure.
2. For each extension namespace used with `semantic` criticality:
   a. If the importing SDK's language namespace is present and supported, use it.
   b. Else if a `default` namespace is present and sufficient, use it.
   c. Else import MUST fail with `unsupported_extension`.
3. `informational` extensions are silently ignored by SDKs that do not
   recognize them.

---

## 6. Export Behavior

### 6.1 Portable Mode

In `portable` mode, the SDK MUST:
1. Walk the entire schema graph.
2. Verify every node uses only portable core features.
3. If any non-portable feature is found, fail with `custom_validation_not_portable`.
4. Emit the canonical JSON document with empty `extensions`.

### 6.2 Extended Mode

In `extended` mode, the SDK MUST:
1. Emit the portable core schema.
2. Attach all extension namespaces to the `extensions` object.
3. Attach node-level extensions to individual nodes.
4. Set `_criticality` appropriately for each namespace.

---

## 7. Canonical Encoding Rules

### 7.1 Encoding

Documents MUST be encoded as UTF-8 JSON.

### 7.2 Property Order

Property order is not significant. SDKs SHOULD emit properties in a stable
order for human readability, but MUST NOT depend on order for correctness.

### 7.3 Numeric Encoding

- Integer values MUST NOT have fractional parts (e.g., `42`, not `42.0`).
- Float constraint values MAY have fractional parts.
- No leading zeros.
- No positive sign prefix.

### 7.4 No Comments

JSON does not support comments. Do not emit or expect them.

---

## 8. Document Validation Summary

A valid AnyVali v1.0 document MUST:

1. Be valid JSON (RFC 8259).
2. Have `anyvaliVersion` equal to `"1.0"`.
3. Have `schemaVersion` equal to `"1"`.
4. Have a `root` that is a valid schema node.
5. Have a `definitions` object where every value is a valid schema node.
6. Have an `extensions` object.
7. Resolve all `ref` pointers to existing definitions.
8. Contain no unrecognized `kind` values (in portable mode).

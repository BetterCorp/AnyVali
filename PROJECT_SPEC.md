# AnyVali v1 Specification

## Summary

AnyVali is a family of native validation libraries for multiple languages that share one portable schema model.

The product is not a single language-neutral authoring DSL. Authors define schemas in their host language using a small native API, similar in spirit to Zod, Valibot, and ArkType. Each SDK can then:

- validate and parse input locally
- apply defaults during parsing
- run portable validations
- export schemas to a canonical JSON document
- import that JSON document into another SDK
- reject schemas that depend on unsupported language-specific extensions

The design target for v1 is:

- minimal surface area
- portable by default
- safe numeric defaults
- deterministic behavior across languages
- support for the first 10 SDKs:
  - JS/TS
  - Python
  - Go
  - Java
  - C#
  - Rust
  - PHP
  - Ruby
  - Kotlin
  - C++

## Product Shape

### Repository / project structure

Specify the project as a mono-repo with:

- `spec/`
  - canonical AnyVali schema specification
  - JSON interchange format
  - validation semantics
  - test corpus / fixtures
- `sdk/js`
- `sdk/python`
- `sdk/go`
- `sdk/java`
- `sdk/csharp`
- `sdk/rust`
- `sdk/php`
- `sdk/ruby`
- `sdk/kotlin`
- `sdk/cpp`

The spec is the source of truth for interchange and semantics. Each SDK is the source of truth for its native authoring ergonomics.

### Design principles

- Native-first authoring: write schemas in the host language.
- Portable core: exported schemas only guarantee the standard AnyVali feature set.
- Explicit non-portability: language-specific behavior must live in namespaced extensions.
- Small core: no arbitrary transform pipeline in v1.
- Deterministic parsing: defaults run before validation.
- Safe defaults for numbers:
  - `number` means `float64`
  - `int` means `int64`

## Canonical Model

### Core schema kinds

The portable v1 core should include:

- `any`
- `unknown`
- `never`
- `null`
- `bool`
- `string`
- `number`
- `int`
- `float32`
- `float64`
- `int8`
- `int16`
- `int32`
- `int64`
- `uint8`
- `uint16`
- `uint32`
- `uint64`
- `literal`
- `enum`
- `array`
- `tuple`
- `object`
- `record`
- `union`
- `intersection`
- `optional`
- `nullable`

Do not include in v1:

- arbitrary executable transforms
- serialized custom validator logic
- branded or opaque types as portable core semantics
- async validation as part of the portable contract

### Validation constraints

Portable constraints must be declarative and data-only.

#### String constraints

- `minLength`
- `maxLength`
- `pattern`
- `startsWith`
- `endsWith`
- `includes`
- `format`

Portable `format` values in v1:

- `email`
- `url`
- `uuid`
- `ipv4`
- `ipv6`
- `date`
- `date-time`

#### Numeric constraints

- `min`
- `max`
- `exclusiveMin`
- `exclusiveMax`
- `multipleOf`

#### Array constraints

- `minItems`
- `maxItems`

#### Object constraints

- field required or optional state
- unknown key mode

Default unknown key mode:

- `reject`

Allowed explicit modes:

- `reject`
- `strip`
- `allow`

### Defaults

Defaults are supported on fields and schemas where the default value is representable in portable JSON.

Rules:

- defaults are applied only when the input value is absent
- default application happens before validation
- the resulting materialized value must still validate
- defaults must be pure data, not code
- defaults are exportable only if serializable in the canonical JSON format

Do not support function-based defaults in the portable schema.

SDKs may support local function defaults, but those must not export as portable defaults.

### Coercions

v1 supports a small set of portable coercions only.

Portable coercions:

- `string -> int`
- `string -> number`
- `string -> bool`
- `string -> date-time string` normalization is out of scope unless exact rules are defined
- whitespace trim for strings
- case normalization:
  - `lower`
  - `upper`

Rules:

- coercion must be explicitly enabled on a schema node
- coercions happen before defaults only for present input values
- defaults are then applied to missing values
- final result is validated after coercion and default materialization

Recommended parse order:

1. detect presence or absence
2. if present and coercion configured, attempt coercion
3. if absent and default exists, materialize default
4. validate resulting value
5. return parsed output or structured error

## Native SDK Contract

Each SDK should expose native builder APIs, but the spec should require the following conceptual API surface.

### Required operations

- define schema
- parse
- safe parse or result-returning parse
- export portable schema
- import portable schema

### Native authoring examples

Every SDK may choose idiomatic naming, but the conceptual API should be:

- `string()`
- `number()`
- `int()`
- `int32()`
- `int64()`
- `object({...})`
- `array(schema)`
- `optional(schema)` or modifier form
- `default(value)`
- validation methods like `min()`, `max()`, `pattern()`
- `coerce(...)`
- `export()`
- `import(schemaJson)`

The spec should explicitly allow each SDK to adapt naming to language idioms while preserving semantics.

### Parse result contract

Each SDK must provide:

- throwing parse API
- non-throwing parse API

Portable result shape:

- success:
  - parsed value
- failure:
  - list of issues

Issue fields:

- `code`
- `message`
- `path`
- `expected`
- `received`
- optional `meta`

Path representation should be an ordered list of object keys and array indexes.

### Custom validators

Custom validators are allowed only as local SDK features.

Rules:

- they must be marked non-portable
- exporting a schema that depends on them must fail unless the SDK is asked to export with extensions
- importing runtimes must reject them unless the relevant local extension handler is available

## Portable JSON Interchange Format

### Top-level document

Define a versioned canonical JSON document:

```json
{
  "anyvaliVersion": "1.0",
  "schemaVersion": "1",
  "root": {},
  "definitions": {},
  "extensions": {}
}
```

### Required properties

- `anyvaliVersion`: spec/runtime family version
- `schemaVersion`: interchange schema version
- `root`: root schema node
- `definitions`: reusable named schema nodes for recursion and deduplication
- `extensions`: namespaced non-portable metadata

### Node shape

Each schema node should be data-only and use explicit discriminators:

```json
{
  "kind": "object",
  "properties": {
    "id": { "kind": "int64" },
    "name": { "kind": "string", "minLength": 1 }
  },
  "required": ["id", "name"],
  "unknownKeys": "reject"
}
```

For aliases:

- `number` maps to `float64`
- `int` maps to `int64`

The exported form may preserve the alias for readability, but imported semantics must be identical to the widened canonical meaning.

### Definitions and references

v1 should support named definitions and references:

- `definitions.<name>`
- `{ "kind": "ref", "ref": "#/definitions/User" }`

This is required for recursion and schema reuse.

### Extensions model

Language-specific metadata belongs under explicit namespaces.

Example:

```json
{
  "extensions": {
    "default": {
      "...": "fallback extension behavior"
    },
    "js": {
      "...": "js-specific"
    },
    "go": {
      "...": "go-specific"
    }
  }
}
```

Import rules:

- if a schema uses only portable core features, import succeeds everywhere
- if a schema requires a language extension for the importing runtime and that namespace is missing, import fails
- if the runtime namespace is missing but `default` extension semantics are sufficient, runtime may use `default`
- if neither runtime namespace nor `default` can satisfy the requirement, import fails as invalid or unsupported

The spec must distinguish:

- informational extensions: safe to ignore
- semantic extensions: required for correct behavior

Only semantic extensions can block import.

## Numeric Semantics

### Default aliases

Portable aliases:

- `number` => `float64`
- `int` => `int64`

This is the recommended default because it is safer across languages even when it uses more memory.

The spec should explicitly justify this:

- many mainstream languages differ on native integer width and numeric semantics
- JS and Python blur or abstract some width details
- cross-language export and import must preserve intent
- using wide defaults reduces silent narrowing and overflow risk

### Narrower numeric types

Portable explicit types must exist for:

- signed ints: 8, 16, 32, 64
- unsigned ints: 8, 16, 32, 64
- floats: 32, 64

Import rules:

- if a target language cannot represent a type natively, the SDK must still preserve schema semantics
- if exact runtime validation is not feasible without helper types, the SDK must ship helpers or reject that schema kind
- approximation without explicit opt-in is not allowed for portable core types

## Cross-Language Support Policy

### Portability tiers

The spec should define three tiers:

1. Portable core
2. Portable core plus extensions
3. Local-only SDK features

#### Tier 1

Guaranteed import and export compatibility across all AnyVali SDKs.

#### Tier 2

Compatible only where the relevant extension namespace is understood.

#### Tier 3

Never part of canonical interchange.

### Export modes

Require two export modes:

- `portable`
- `extended`

`portable`:

- fails if the schema depends on non-portable features

`extended`:

- emits core schema plus extension namespaces

This keeps the default strict and predictable.

## Public APIs / Interfaces / Types

The spec should explicitly define these public concepts.

### SDK-level concepts

- `Schema<TInput, TOutput>` or closest language equivalent
- `parse(input): output`
- `safeParse(input): ParseResult`
- `export(options): AnyValiDocument`
- `import(document): Schema`

### Portable document types

- `AnyValiDocument`
- `SchemaNode`
- `ValidationIssue`
- `ExtensionNamespace`
- `DefinitionMap`

### Important semantic enums

- schema `kind`
- unknown key mode
- issue `code`
- export mode
- extension criticality

### Required issue codes

At minimum:

- `invalid_type`
- `required`
- `unknown_key`
- `too_small`
- `too_large`
- `invalid_string`
- `invalid_number`
- `invalid_literal`
- `invalid_union`
- `custom_validation_not_portable`
- `unsupported_extension`
- `unsupported_schema_kind`
- `coercion_failed`
- `default_invalid`

## Testing And Acceptance Criteria

### Spec conformance suite

The spec should require a shared cross-language test corpus in JSON with:

- input payload
- schema document
- expected success or failure
- expected parsed output
- expected issue codes and paths

Every SDK must run the same corpus.

### Required test scenarios

#### Core typing

- parse valid primitive values
- reject invalid primitive values
- validate explicit widths: `int32`, `int64`, `uint32`, `float32`, `float64`

#### Defaults

- missing field gets default
- present field does not get overwritten by default
- defaulted value is validated
- invalid default causes schema, export, or import failure as specified

#### Coercion

- allowed coercion succeeds
- failed coercion returns `coercion_failed`
- coercion plus validation ordering is deterministic

#### Objects

- unknown keys rejected by default
- unknown keys stripped when configured
- unknown keys allowed when configured

#### Composition

- unions validate correctly
- intersections validate correctly
- recursive refs import and export correctly

#### Portability

- schema exported in JS imports in Python, Go, Java, and other SDKs
- `portable` export fails on local-only custom validators
- `extended` export includes extension namespaces
- import fails when required language extension is missing
- import uses `default` extension when allowed by schema semantics

#### Numeric safety

- `number` round-trips as `float64`
- `int` round-trips as `int64`
- narrowing mismatch is rejected, not silently approximated

### Acceptance criteria

v1 is complete when:

- the canonical JSON format is fully specified
- all 10 SDKs implement the portable core
- all 10 SDKs pass the shared conformance suite for portable features
- extension failure modes are standardized
- import and export round-trips are deterministic for portable schemas

## Documentation Requirements

The spec should require three doc tracks:

- product overview:
  - what AnyVali is
  - native-first authoring
  - portable export and import
- numeric semantics:
  - why `number = float64`
  - why `int = int64`
  - memory vs safety tradeoff
- portability guide:
  - portable core vs extensions vs local-only
  - how to design schemas that move cleanly across languages

## Implementation Sequencing

### Phase 1

- write canonical spec
- define JSON document format
- define parse semantics
- build shared conformance corpus

### Phase 2

- deliver reference SDKs:
  - JS/TS
  - Python
  - Go

### Phase 3

- deliver remaining SDKs:
  - Java
  - C#
  - Rust
  - PHP
  - Ruby
  - Kotlin
  - C++

### Phase 4

- stabilize extension contract
- add versioning and migration guidance for future spec versions

## Assumptions And Defaults Chosen

- Schemas are authored natively in each language, not in one universal authoring DSL.
- The interchange format is canonical JSON.
- Defaults are applied before validation.
- v1 includes portable coercions only, not arbitrary transforms.
- `number` defaults to `float64`.
- `int` defaults to `int64`.
- Unknown object keys are rejected by default.
- Custom validators are local-only and non-exportable in portable mode.
- Initial first-class SDK matrix is:
  - JS/TS
  - Python
  - Go
  - Java
  - C#
  - Rust
  - PHP
  - Ruby
  - Kotlin
  - C++
- Language-specific behavior is represented via namespaced extensions plus optional `default` fallback behavior.
- Import must fail when required semantic extensions for the target runtime are unavailable.
- The portable contract is synchronous and declarative in v1.

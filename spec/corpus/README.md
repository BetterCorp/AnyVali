# AnyVali Conformance Test Corpus

This directory contains the shared cross-language conformance test corpus for
AnyVali v1.0. Every conforming SDK MUST pass all test cases in this corpus.

## File Format

Each `.json` file in this corpus is a test suite with the following structure:

```json
{
  "suite": "suite-name",
  "cases": [
    {
      "description": "Human-readable description of what this case tests",
      "schema": {
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string" },
        "definitions": {},
        "extensions": {}
      },
      "input": "the input value to parse",
      "valid": true,
      "output": "the expected parsed output (or null if validation fails)",
      "issues": []
    }
  ]
}
```

## Field Definitions

### Suite Level

- `suite` (string): A unique identifier for the test suite.
- `cases` (array): Ordered list of test case objects.

### Case Level

- `description` (string): Human-readable description of the test scenario.
- `schema` (object): A complete AnyVali canonical JSON document with all five
  required top-level properties (`anyvaliVersion`, `schemaVersion`, `root`,
  `definitions`, `extensions`).
- `input`: The value to be parsed/validated. Can be any JSON value. Use the
  special string `"__ABSENT__"` to represent a truly absent value (e.g., a
  missing object property at the root level).
- `valid` (boolean): Whether the input should pass validation.
- `output`: The expected parsed output value. `null` when `valid` is `false`.
  For successful parses, this is the value after coercion and default
  materialization.
- `issues` (array): Expected validation issues. Empty array `[]` when `valid`
  is `true`. Each issue object has:
  - `code` (string): One of the 14 AnyVali issue codes.
  - `path` (array): Path from root to the failing value (strings for object
    keys, integers for array indices).
  - `expected` (string): What was expected (optional in test, but present when
    the spec defines it).
  - `received` (string): What was received (optional in test, but present when
    the spec defines it).

## Directory Structure

```
corpus/
  primitives/       - Tests for primitive schema kinds
  constraints/      - Tests for validation constraints
  objects/          - Tests for object schemas
  composition/      - Tests for union, intersection, optional, nullable, etc.
  defaults/         - Tests for default value materialization
  coercions/        - Tests for coercion pipeline
  refs/             - Tests for definition references
  numeric-safety/   - Tests for numeric width and safety semantics
```

## Running the Corpus

SDK test runners should:

1. Read each `.json` file.
2. For each case, import the `schema` as an AnyVali document.
3. Parse `input` against the imported schema.
4. Assert that the result matches `valid`, `output`, and `issues`.

For issue matching, the test runner should verify:
- The `code` matches exactly.
- The `path` matches exactly.
- If `expected` is present in the test case, it matches.
- If `received` is present in the test case, it matches.

The `message` field of issues is NOT tested (it is human-readable and may
vary across SDKs).

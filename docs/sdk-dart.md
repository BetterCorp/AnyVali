# AnyVali for Dart and Flutter

Pure Dart validation with immutable builders, typed results, structured issue paths,
portable JSON schema import/export, coercion, defaults, and sensitive-data helpers.
Requires Dart 3.4 or later. No runtime dependencies or Flutter SDK dependency.

## Installation

This package is available from source and has not yet been published to pub.dev.
Within a checkout of this repository, add a path dependency to your app:

```yaml
dependencies:
  anyvali:
    path: /path/to/AnyVali/sdk/dart
```

Then run `dart pub get`, or `flutter pub get` in a Flutter project.

## Usage

```dart
import 'package:anyvali/anyvali.dart' as av;

final user = av.object({
  'name': av.string().coerce(trim: true).minLength(1),
  'email': av.string().format('email'),
  'age': av.int_().coerce().min(0).optional(),
});
final result = user.safeParse({
  'name': 'Alice', 'email': 'alice@example.com', 'age': '30',
});
if (result.success) {
  print(result.data);
} else {
  for (final issue in result.issues) {
    print('${issue.path}: ${issue.message}');
  }
}
```

`schema.parse(input)` throws `ValidationError` on validation failure.
`safeParse(input)` returns `ParseResult<T>`. Omit the argument or pass `av.absent`
to represent absence; explicit null is a present value. At the top level, `any`
and `unknown` preserve the absence sentinel. The native `.optional()` decoder
represents omitted values as null for its nullable Dart type; imported
`Schema<Object?>` results retain the sentinel. Builders return new schemas.

## Flutter forms

Pass the validator directly to a `TextFormField`:

```dart
final email = av.string().minLength(1).format('email');
final age = av.int_().coerce().min(18);

TextFormField(validator: email.validator());
TextFormField(
  keyboardType: TextInputType.number,
  validator: age.validator(),
);
```

The callback returns null for valid input and the first issue message otherwise.
Empty strings remain present values. Normalize empty optional form fields to
`av.absent` when validating your assembled object if that is your intended behavior.
No Flutter import is required in the validation package, so the same schemas run
on mobile, desktop, servers, and Flutter web.

## Schema API

- Primitives: `string`, `bool_`, `null_`, `any`, `unknown`, `never`.
- Numbers: `number`, `float32`, `float64`, `int_`, `int8`–`int64`, `uint8`–`uint64`.
- Values: `literal(value)`, `enum_(values)`.
- Collections: `array(items)`, `tuple(elements)`, `object(properties)`, `record(values)`.
- Composition: `union(variants)`, `intersection(schemas)`, `.optional()`, `.nullable()`.
- Constraints: `.min()`, `.max()`, `.exclusiveMin()`, `.exclusiveMax()`, `.multipleOf()`,
  `.minLength()`, `.maxLength()`, `.pattern()`, `.startsWith()`, `.endsWith()`,
  `.includes()`, `.format()`, `.minItems()`, `.maxItems()`.
- Behavior: `.defaultValue(value)`, `.coerce(trim: true, lower: true)`,
  `.unknownKeys('strip' | 'allow' | 'reject')`.
- Reserved metadata: `.describe('Description', title: 'Title', sensitive: true,
  deprecated: true, deprecatedMessage: 'Use next', since: '1.0',
  notStable: true, readonly: true, writeonly: false, examples: ['value'])`.
- Custom metadata: `.metadata({...})`; use `replace: true` to replace custom
  keys while preserving reserved fields. Reserved names must use `describe()`.

A parent object using `strip` or `reject` applies that policy recursively to nested
objects, including through arrays, wrappers, and references. A parent using
`allow` leaves the nested object policy in effect.

Dart keywords use a trailing underscore (`int_`, `bool_`, `null_`, `enum_`).
`Schema<T>` carries primitive and collection output types; heterogeneous objects
return `Map<String, Object?>`, and imported documents return `Schema<Object?>`.

In Dart, a parent object's `strip` or `reject` unknown-key mode propagates
unchanged to nested objects, including union variants. A parent set to `allow`
leaves each nested object's own mode in effect. JavaScript and Go currently reset
the inherited mode at union boundaries. An `allow` object inside a union can
therefore retain extra keys in those SDKs that Dart strips or rejects according
to the parent's mode.

## Interchange

```dart
final document = user.exportSchema();
final imported = av.importSchema(document);
```

Use `jsonEncode`/`jsonDecode` from `dart:convert` to transmit documents.
Schema defaults, metadata, and extensions must contain JSON values: finite
numbers, strings, booleans, null, lists, and maps with string keys. Cycles and
custom Dart objects are rejected; custom `toJson()` methods are not invoked.
`exportSchema` avoids Dart's reserved `export` keyword. Documents use schemaVersion
`1.1`; imports also accept the existing `1` and `1.0` forms and legacy node aliases.
Definitions and metadata survive export and native composition. Conflicting
imported definition names are rejected. Unknown semantic extensions and custom
validator references are rejected. Informational extensions are retained in
`exportSchema(mode: 'extended')`; portable exports remove extensions from the document and every schema node,
while preserving defaults and metadata as data. Portable exports use canonical
child keys and ordered coercion token lists; imports also accept older short
coercion names (`int`, `number`, and `bool`).

## Pattern limits

Patterns use bounded state-machine matching and support literals, character
classes, dot, anchors, word boundaries, ordinary and noncapturing groups,
alternatives, positive and negative lookahead/lookbehind, and `*`, `+`, `?`,
`{m,n}` repetitions (including lazy forms). Nested lookarounds share the same
compilation and execution limits. Numeric and named backreferences and named
capturing groups are supported. Legacy octal escapes outside character classes
and Unicode property escapes are rejected during construction or import with
`FormatException`.
Escape literal braces rather than relying on permissive regex parsing.

Compilation allows at most 1,024 UTF-16 code units, 64 nested groups, repetition
counts of 1,000, and 2,048 states. A match accepts at most 100,000 UTF-16 input
code units and 2,000,000 matching operations. Backreference patterns additionally
allow at most 64 capture groups and 16,384 explored capture states. Exceeding a
match limit produces an
`invalid_string` issue with an explicit limit diagnostic. These bounds apply to
native and JavaScript targets.

## Numeric limits

Native Dart integers are signed 64-bit. On JavaScript targets, numeric integer
inputs beyond ±(2^53−1) are rejected to prevent silently accepting rounded values.
`uint64()` accepts `BigInt` for the full unsigned range; string coercion can
produce a `BigInt` when needed. Other integer schemas return `int` and reject
values that cannot be represented exactly on the target platform. JavaScript
targets also reject decoded integer-schema constraints outside the safe-integer
range. This check runs after JSON decoding and cannot detect precision already
lost in a numeric token, including fractional tokens rounded into the safe range.
Validate numeric tokens before decoding when exact bounds matter.

`BigInt` is not directly JSON serializable. Exchange wide integer data using an
explicit string transport contract and coercion, rather than JSON numeric tokens
that a host parser may already have rounded. Schema defaults and literal/enum
values must be JSON-compatible for export. `float32` checks the IEEE 754 binary32 range
and rejects overflow without rounding the parsed value; `number` and `float64` require finite numeric values.

## Sensitive data

```dart
final credentials = av.object({
  'password': av.string().minLength(12).describe('Password', sensitive: true),
});
final stored = av.encrypt(credentials, input,
    (path, value) => 'encrypted:${encryptWithYourKms(value)}');
final check = av.safeParseEncrypted(credentials, stored);
final clear = av.decrypt(credentials, stored,
    (path, value) => decryptWithYourKms((value as String).substring(10)));
```

Supply your own transformation and key management. Sensitive composites transform
as opaque values. Encryption and decryption validate before and after the callback.
Normal parsing always validates plaintext. Encryption rejects output when union branch selection could bypass decryption
of a sensitive value. Add a literal discriminator
or mark the entire union sensitive when encrypted branches would be ambiguous.

## Development

From the repository root, run `bash runner.sh ci dart`. The VM suite exercises
both shared conformance corpora, including export/reimport and sensitive paths.
Run `dart test --platform node test/schema_test.dart test/robustness_test.dart test/pattern_test.dart` within `sdk/dart` for web checks.

For SDK installation, first publication, publisher ownership, and GitHub OIDC configuration, see the [Dart setup and publishing guide](https://docs.anyvali.com/docs/dart-setup-publishing/).

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
to represent absence; explicit null is a present value. Builders return new schemas.

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
- Metadata: `.describe('Description', sensitive: true)`, `.metadata({...})`.

Dart keywords use a trailing underscore (`int_`, `bool_`, `null_`, `enum_`).
`Schema<T>` carries primitive and collection output types; heterogeneous objects
return `Map<String, Object?>`, and imported documents return `Schema<Object?>`.

## Interchange

```dart
final document = user.exportSchema();
final imported = av.importSchema(document);
```

Use `jsonEncode`/`jsonDecode` from `dart:convert` to transmit documents.
`exportSchema` avoids Dart's reserved `export` keyword. Documents use schemaVersion
`1.1`; imports also accept the existing `1` and `1.0` forms and legacy node aliases.
Definitions and metadata survive export and native composition. Conflicting
imported definition names are rejected. Unknown semantic extensions and custom
validator references are rejected. Informational extensions are retained in
`exportSchema(mode: 'extended')`; portable exports omit document extensions.

## Numeric limits

Native Dart integers are signed 64-bit. On JavaScript targets, numeric integer
inputs beyond ±(2^53−1) are rejected to prevent silently accepting rounded values.
`uint64()` accepts `BigInt` for the full unsigned range; string coercion can
produce a `BigInt` when needed. Other integer schemas return `int` and reject
values that cannot be represented exactly on the target platform.

`BigInt` is not directly JSON serializable. Exchange wide integer data using an
explicit string transport contract and coercion, rather than JSON numeric tokens
that a host parser may already have rounded. Schema defaults and literal/enum
values must be JSON-compatible for export. `float32` rounds to IEEE 754 binary32
and rejects overflow; `number` and `float64` require finite numeric values.

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
Normal parsing always validates plaintext.

## Development

From the repository root, run `bash runner.sh ci dart`. The VM suite exercises
both shared conformance corpora, including export/reimport and sensitive paths.
Run `dart test --platform node test/schema_test.dart` within `sdk/dart` for web checks.

For SDK installation, first publication, publisher ownership, and GitHub OIDC configuration, see the [Dart setup and publishing guide](https://docs.anyvali.com/docs/dart-setup-publishing/).

# AnyVali PHP SDK Reference

## Installation

Install via Composer:

```bash
composer require anyvali/anyvali
```

Requires PHP 8.1 or later.

## Quick Start

```php
<?php

use AnyVali\AnyVali;
use AnyVali\Schema;

// Define a schema
$userSchema = AnyVali::object([
    'id'    => AnyVali::int64(),
    'name'  => AnyVali::string()->minLength(1)->maxLength(100),
    'email' => AnyVali::string()->format('email'),
    'age'   => AnyVali::optional(AnyVali::int()->min(0)->max(150)),
    'role'  => AnyVali::enum(['admin', 'user', 'guest']),
], required: ['id', 'name', 'email']);

// Parse input (throws ValidationError on failure)
$user = $userSchema->parse([
    'id'    => 42,
    'name'  => 'Alice',
    'email' => 'alice@example.com',
    'role'  => 'admin',
]);
// => ['id' => 42, 'name' => 'Alice', 'email' => 'alice@example.com', 'role' => 'admin']

// Safe parse (never throws)
$result = $userSchema->safeParse(['id' => 'not-a-number', 'name' => '']);
if (!$result->success) {
    foreach ($result->issues as $issue) {
        echo implode('.', $issue->path) . ": [{$issue->code}] {$issue->message}\n";
    }
}
```

## Schema Builders

All schemas are created through static factory methods on the `AnyVali` class. Each method returns a `Schema` instance that can be further refined with chained constraints.

### Special Types

```php
use AnyVali\AnyVali;

$any     = AnyVali::any();       // accepts any value
$unknown = AnyVali::unknown();   // accepts any value, requires explicit handling
$never   = AnyVali::never();     // rejects all values
```

### Primitives

```php
$str  = AnyVali::string();
$num  = AnyVali::number();   // IEEE 754 float64
$bool = AnyVali::bool();
$null = AnyVali::null();
```

### Numeric Types

AnyVali provides sized integer and float types for cross-language safety:

```php
// Default integer (int64)
$int = AnyVali::int();

// Signed integers
$i8  = AnyVali::int8();    // -128 to 127
$i16 = AnyVali::int16();   // -32,768 to 32,767
$i32 = AnyVali::int32();   // -2^31 to 2^31-1
$i64 = AnyVali::int64();   // -2^63 to 2^63-1

// Unsigned integers
$u8  = AnyVali::uint8();   // 0 to 255
$u16 = AnyVali::uint16();  // 0 to 65,535
$u32 = AnyVali::uint32();  // 0 to 2^32-1
$u64 = AnyVali::uint64();  // 0 to 2^64-1

// Floats
$f32 = AnyVali::float32();
$f64 = AnyVali::float64(); // same as number()
```

### Literal and Enum

```php
// Literal matches a single exact value
$active = AnyVali::literal('active');
$zero   = AnyVali::literal(0);
$yes    = AnyVali::literal(true);

// Enum matches one of several allowed values
$status = AnyVali::enum(['pending', 'active', 'disabled']);
```

### Arrays and Tuples

```php
// Homogeneous array of strings
$tags = AnyVali::array(AnyVali::string());

// Array with length constraints
$topThree = AnyVali::array(AnyVali::string())
    ->minItems(1)
    ->maxItems(3);

// Tuple with fixed positional types
$coordinate = AnyVali::tuple([
    AnyVali::float64(),
    AnyVali::float64(),
]);
```

### Objects

```php
use AnyVali\UnknownKeyMode;

// Basic object with required and optional fields
$user = AnyVali::object(
    properties: [
        'name'  => AnyVali::string(),
        'email' => AnyVali::string()->format('email'),
        'age'   => AnyVali::int(),
    ],
    required: ['name', 'email'],
);

// Unknown key handling
$strict = AnyVali::object(
    properties: ['name' => AnyVali::string()],
    required: ['name'],
    unknownKeys: UnknownKeyMode::Reject,   // error on unexpected keys (default)
);

$stripped = AnyVali::object(
    properties: ['name' => AnyVali::string()],
    required: ['name'],
    unknownKeys: UnknownKeyMode::Strip,    // silently remove unexpected keys
);

$passthrough = AnyVali::object(
    properties: ['name' => AnyVali::string()],
    required: ['name'],
    unknownKeys: UnknownKeyMode::Allow,    // keep unexpected keys as-is
);
```

### Records

A record validates a dictionary where all values conform to a single schema:

```php
// Map of string keys to integer values
$scores = AnyVali::record(AnyVali::int());

$scores->parse(['alice' => 95, 'bob' => 87]);
// => ['alice' => 95, 'bob' => 87]
```

### Composition

```php
// Union: value must match at least one variant
$stringOrInt = AnyVali::union([
    AnyVali::string(),
    AnyVali::int(),
]);

// Intersection: value must match all schemas
$namedAndAged = AnyVali::intersection([
    AnyVali::object(['name' => AnyVali::string()], required: ['name']),
    AnyVali::object(['age' => AnyVali::int()], required: ['age']),
]);
```

### Modifiers

```php
// Optional: value may be absent (field can be omitted)
$maybeAge = AnyVali::optional(AnyVali::int());

// Nullable: value may be null
$nullableName = AnyVali::nullable(AnyVali::string());

// Combine both: field can be omitted, or present as null or string
$optionalNullableName = AnyVali::optional(
    AnyVali::nullable(AnyVali::string())
);
```

### References

References allow recursive and reusable schema definitions:

```php
$treeNode = AnyVali::object(
    properties: [
        'value'    => AnyVali::string(),
        'children' => AnyVali::array(AnyVali::ref('#/definitions/TreeNode')),
    ],
    required: ['value'],
);
```

## String Constraints

All string constraints return the schema instance for chaining:

```php
$schema = AnyVali::string()
    ->minLength(1)           // minimum character count
    ->maxLength(255)         // maximum character count
    ->pattern('/^[A-Z]/u')   // regex pattern the value must match
    ->startsWith('Hello')    // value must start with this prefix
    ->endsWith('!')          // value must end with this suffix
    ->includes('world')      // value must contain this substring
    ->format('email');       // named format (email, uri, uuid, date, etc.)
```

Each constraint can be used independently:

```php
$slug = AnyVali::string()
    ->pattern('/^[a-z0-9]+(?:-[a-z0-9]+)*$/')
    ->minLength(1)
    ->maxLength(100);

$email = AnyVali::string()->format('email');
$url   = AnyVali::string()->format('uri');
$uuid  = AnyVali::string()->format('uuid');
```

## Number Constraints

Number constraints apply to all numeric types (`number`, `int`, `int8`--`int64`, `uint8`--`uint64`, `float32`, `float64`):

```php
$price = AnyVali::float64()
    ->min(0)               // value >= 0
    ->max(999999.99)       // value <= 999999.99
    ->multipleOf(0.01);    // must be a multiple of 0.01

$rating = AnyVali::int()
    ->min(1)
    ->max(5);

$temperature = AnyVali::float64()
    ->exclusiveMin(-273.15)    // value > -273.15
    ->exclusiveMax(1000.0);    // value < 1000.0
```

## Array Constraints

```php
$tags = AnyVali::array(AnyVali::string())
    ->minItems(1)      // at least 1 element
    ->maxItems(10);    // at most 10 elements
```

## Coercion

Coercion transforms the input value before validation. It runs only when the value is present. Call `->coerce($config)` on any schema to enable coercion.

### Available Coercions

#### String to Integer

```php
$age = AnyVali::int()->coerce(['from' => 'string']);
$age->parse("42");   // => 42 (string coerced to integer)
$age->parse(42);     // => 42 (already an integer, no coercion needed)
```

#### String to Number

```php
$price = AnyVali::number()->coerce(['from' => 'string']);
$price->parse("3.14");  // => 3.14
```

#### String to Boolean

```php
$flag = AnyVali::bool()->coerce(['from' => 'string']);
$flag->parse("true");   // => true
$flag->parse("false");  // => false
$flag->parse("1");      // => true
$flag->parse("0");      // => false
```

#### Trim Whitespace

```php
$trimmed = AnyVali::string()->coerce(['trim' => true]);
$trimmed->parse("  hello  ");  // => "hello"
```

#### Lowercase / Uppercase

```php
$lower = AnyVali::string()->coerce(['lower' => true]);
$lower->parse("HELLO");  // => "hello"

$upper = AnyVali::string()->coerce(['upper' => true]);
$upper->parse("hello");  // => "HELLO"
```

Transformations can be combined:

```php
$normalized = AnyVali::string()->coerce(['trim' => true, 'lower' => true]);
$normalized->parse("  Hello World  ");  // => "hello world"
```

## Defaults

Defaults fill in missing (absent) values. They run after coercion and before validation. Call `->default($value)` on any schema.

The default only applies when the value is absent -- if a value is present, it is validated normally. Defaults must be static values (for portability across SDKs).

```php
$role = AnyVali::string()->default('user');
$role->parse(null);     // => "user" (absent value filled)
$role->parse('admin');  // => "admin"

$tags = AnyVali::array(AnyVali::string())->default([]);
```

Defaults work with optional fields in objects:

```php
$config = AnyVali::object([
    'theme'    => AnyVali::optional(AnyVali::string()->default('light')),
    'language' => AnyVali::optional(AnyVali::string()->default('en')),
], required: []);

$config->parse([]);
// => ['theme' => 'light', 'language' => 'en']
```

If the default value itself fails validation, a `default_invalid` issue is produced.

## Parsing

### Throwing Parse

`parse()` returns the validated and coerced value. If validation fails, it throws a `ValidationError`:

```php
use AnyVali\AnyVali;
use AnyVali\ValidationError;

$schema = AnyVali::string()->minLength(1);

try {
    $value = $schema->parse('hello');
    // => 'hello'
} catch (ValidationError $e) {
    echo $e->getMessage();
    foreach ($e->issues as $issue) {
        echo "[{$issue->code}] {$issue->message}\n";
    }
}
```

### Safe Parse

`safeParse()` never throws. It returns a `ParseResult` with structured success/failure information:

```php
$schema = AnyVali::object([
    'name'  => AnyVali::string()->minLength(1),
    'email' => AnyVali::string()->format('email'),
], required: ['name', 'email']);

$result = $schema->safeParse([
    'name'  => '',
    'email' => 'not-an-email',
]);

if ($result->success) {
    // $result->value contains the parsed data
    $user = $result->value;
} else {
    // $result->issues contains all validation errors
    foreach ($result->issues as $issue) {
        echo implode('.', $issue->path) . ": [{$issue->code}] {$issue->message}\n";
        // "name: [too_small] String must have at least 1 character"
        // "email: [invalid_format] Invalid email format"
    }
}
```

### Issue Structure

Each issue in a parse result contains:

| Field | Type | Description |
|---|---|---|
| `code` | `string` | Machine-readable error code (e.g., `invalid_type`, `too_small`) |
| `message` | `string` | Human-readable error description |
| `path` | `array` | Path to the invalid value (e.g., `['users', 0, 'email']`) |
| `expected` | `string` | What was expected (e.g., `string`, `int64`) |
| `received` | `string` | What was received (e.g., `null`, `bool`) |

## Export and Import

### Exporting Schemas

Any schema can be exported to AnyVali's portable JSON format:

```php
$schema = AnyVali::object([
    'id'   => AnyVali::int64(),
    'name' => AnyVali::string()->minLength(1)->maxLength(100),
], required: ['id', 'name']);

// Export to portable JSON (associative array)
$doc = $schema->export();

echo json_encode($doc, JSON_PRETTY_PRINT);
// {
//     "anyvaliVersion": "1.0",
//     "schemaVersion": "1",
//     "root": {
//         "kind": "object",
//         "properties": {
//             "id": { "kind": "int64" },
//             "name": { "kind": "string", "minLength": 1, "maxLength": 100 }
//         },
//         "required": ["id", "name"],
//         "unknownKeys": "reject"
//     },
//     "definitions": {},
//     "extensions": {}
// }
```

### Importing Schemas

Schemas can be imported from a JSON document produced by any AnyVali SDK:

```php
// Import from a portable JSON document (associative array or JSON string)
$imported = AnyVali::import($doc);

// The imported schema works exactly like a locally-built one
$user = $imported->parse(['id' => 1, 'name' => 'Bob']);
```

This enables cross-language schema sharing. A schema authored in TypeScript, Go, or any other supported language can be exported, stored, and imported back into PHP.

## Type Inference with PHPStan and Psalm

The PHP SDK uses `@template` phpDoc annotations so that static analysis tools can infer the output type of parsed values:

```php
/** @var Schema<string> */
$name = AnyVali::string();

/** @var Schema<int> */
$age = AnyVali::int();

/** @var Schema<float> */
$price = AnyVali::float64();

/** @var Schema<bool> */
$active = AnyVali::bool();

// parse() return type is inferred by PHPStan/Psalm
$parsedName = $name->parse('Alice');   // PHPStan infers: string
$parsedAge  = $age->parse(25);         // PHPStan infers: int
```

`ParseResult` also carries the template type:

```php
$result = $name->safeParse('hello');
if ($result->success) {
    // $result->value is inferred as string by PHPStan/Psalm
    echo strtoupper($result->value);
}
```

## Practical Examples

### Form Validation

```php
$contactForm = AnyVali::object([
    'name'    => AnyVali::string()->minLength(1)->maxLength(200),
    'email'   => AnyVali::string()->format('email'),
    'subject' => AnyVali::enum(['support', 'sales', 'feedback']),
    'message' => AnyVali::string()->minLength(10)->maxLength(5000),
], required: ['name', 'email', 'subject', 'message']);

$result = $contactForm->safeParse($_POST);
if (!$result->success) {
    http_response_code(422);
    echo json_encode(['errors' => array_map(
        fn($issue) => [
            'field'   => implode('.', $issue->path),
            'message' => $issue->message,
        ],
        $result->issues,
    )]);
    return;
}

$data = $result->value;
// proceed with validated data
```

### API Request Validation

```php
$createOrderSchema = AnyVali::object([
    'customer_id' => AnyVali::int64(),
    'items' => AnyVali::array(
        AnyVali::object([
            'product_id' => AnyVali::int64(),
            'quantity'    => AnyVali::uint16()->min(1),
            'unit_price'  => AnyVali::float64()->min(0)->multipleOf(0.01),
        ], required: ['product_id', 'quantity', 'unit_price']),
    )->minItems(1),
    'notes' => AnyVali::optional(AnyVali::nullable(AnyVali::string()->maxLength(1000))),
], required: ['customer_id', 'items']);

$body = json_decode(file_get_contents('php://input'), true);

try {
    $order = $createOrderSchema->parse($body);
    // $order is now validated and safe to use
} catch (ValidationError $e) {
    http_response_code(400);
    echo json_encode(['errors' => $e->issues]);
}
```

### Configuration Validation

```php
$configSchema = AnyVali::object([
    'database' => AnyVali::object([
        'host'     => AnyVali::string(),
        'port'     => AnyVali::uint16()->min(1)->max(65535),
        'name'     => AnyVali::string()->minLength(1),
        'user'     => AnyVali::string(),
        'password' => AnyVali::string(),
    ], required: ['host', 'port', 'name', 'user', 'password']),
    'cache' => AnyVali::object([
        'driver' => AnyVali::enum(['redis', 'memcached', 'file']),
        'ttl'    => AnyVali::int()->min(0),
    ], required: ['driver', 'ttl']),
    'debug' => AnyVali::optional(AnyVali::bool()),
], required: ['database', 'cache']);

$config = $configSchema->parse(json_decode(
    file_get_contents(__DIR__ . '/config.json'),
    true,
));
```

### Cross-Language Schema Sharing

```php
// Receive a schema from a TypeScript service
$response = file_get_contents('https://api.example.com/schemas/user');
$doc = json_decode($response, true);
$userSchema = AnyVali::import($doc);

// Validate incoming data against the shared schema
$result = $userSchema->safeParse($incomingData);

// Build a local schema and share it back
$auditEvent = AnyVali::object([
    'action'    => AnyVali::enum(['create', 'update', 'delete']),
    'actor_id'  => AnyVali::int64(),
    'timestamp' => AnyVali::string()->format('date-time'),
    'payload'   => AnyVali::record(AnyVali::any()),
], required: ['action', 'actor_id', 'timestamp']);

$exportedDoc = $auditEvent->export();
// Send $exportedDoc to a schema registry for other services to consume
```

## Common Patterns

### Validating Configuration Files

Use `UnknownKeyMode::Strip` when parsing arrays that contain many extra keys you don't care about, like config files with additional entries:

```php
$envSchema = AnyVali::object(
    ['DATABASE_URL' => AnyVali::string()],
    ['DATABASE_URL'],
    UnknownKeyMode::Strip
);
```

Without `Strip`, parse would fail with `unknown_key` issues for every extra key because the default mode is `Reject`.

| Mode | What happens with extra keys |
|---|---|
| `Reject` (default) | Parse fails with `unknown_key` issues |
| `Strip` | Extra keys silently removed from output |
| `Allow` | Extra keys passed through to output |

### Eagerly Evaluated vs Lazy Defaults

`->default()` accepts any value of the correct type. Expressions like `getcwd()` are evaluated immediately when the schema is created and stored as a static value -- this works fine. What AnyVali does not support is lazy callable defaults that re-evaluate on each parse call. If you need a fresh value on every parse, apply it after:

```php
$configSchema = AnyVali::object(
    [
        'profile' => AnyVali::optional(AnyVali::string()->default('default')),
        'appDir'  => AnyVali::optional(AnyVali::string()),
    ],
    ['profile'],
    UnknownKeyMode::Strip
);

$config = $configSchema->parse($data);
if (!isset($config['appDir'])) {
    $config['appDir'] = getcwd();
}
```

This keeps the schema fully portable -- the same JSON document can be imported in Go, Python, or any other SDK without relying on language-specific function calls.

## API Reference

### AnyVali (Static Factory)

| Method | Returns | Description |
|---|---|---|
| `AnyVali::string()` | `Schema<string>` | String schema |
| `AnyVali::number()` | `Schema<float>` | IEEE 754 float64 schema |
| `AnyVali::int()` | `Schema<int>` | Signed int64 schema |
| `AnyVali::int8()` | `Schema<int>` | Signed 8-bit integer |
| `AnyVali::int16()` | `Schema<int>` | Signed 16-bit integer |
| `AnyVali::int32()` | `Schema<int>` | Signed 32-bit integer |
| `AnyVali::int64()` | `Schema<int>` | Signed 64-bit integer |
| `AnyVali::uint8()` | `Schema<int>` | Unsigned 8-bit integer |
| `AnyVali::uint16()` | `Schema<int>` | Unsigned 16-bit integer |
| `AnyVali::uint32()` | `Schema<int>` | Unsigned 32-bit integer |
| `AnyVali::uint64()` | `Schema<int>` | Unsigned 64-bit integer |
| `AnyVali::float32()` | `Schema<float>` | 32-bit float |
| `AnyVali::float64()` | `Schema<float>` | 64-bit float (same as `number()`) |
| `AnyVali::bool()` | `Schema<bool>` | Boolean schema |
| `AnyVali::null()` | `Schema<null>` | Null schema |
| `AnyVali::any()` | `Schema<mixed>` | Accepts any value |
| `AnyVali::unknown()` | `Schema<mixed>` | Accepts any value (explicit handling) |
| `AnyVali::never()` | `Schema<never>` | Rejects all values |
| `AnyVali::literal($value)` | `Schema` | Matches a single exact value |
| `AnyVali::enum($values)` | `Schema` | Matches one of the given values |
| `AnyVali::array($items)` | `Schema<array>` | Homogeneous array |
| `AnyVali::tuple($elements)` | `Schema<array>` | Fixed-position typed array |
| `AnyVali::object($properties, $required, $unknownKeys)` | `Schema<array>` | Object/associative array |
| `AnyVali::record($valueSchema)` | `Schema<array>` | String-keyed map with uniform value type |
| `AnyVali::union($variants)` | `Schema` | Value must match at least one variant |
| `AnyVali::intersection($schemas)` | `Schema` | Value must match all schemas |
| `AnyVali::optional($schema)` | `Schema` | Value may be absent |
| `AnyVali::nullable($schema)` | `Schema` | Value may be null |
| `AnyVali::ref($ref)` | `Schema` | Reference to a definition |
| `AnyVali::import($source)` | `Schema` | Import schema from portable JSON |

### Schema Instance Methods

| Method | Returns | Description |
|---|---|---|
| `$schema->parse($input)` | `mixed` | Parse and validate; throws `ValidationError` on failure |
| `$schema->safeParse($input)` | `ParseResult` | Parse and validate; never throws |
| `$schema->export()` | `array` | Export to portable JSON document |

### StringSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->minLength($n)` | `int` | Minimum character count |
| `->maxLength($n)` | `int` | Maximum character count |
| `->pattern($p)` | `string` | Regex pattern the value must match |
| `->startsWith($s)` | `string` | Required prefix |
| `->endsWith($s)` | `string` | Required suffix |
| `->includes($s)` | `string` | Required substring |
| `->format($f)` | `string` | Named format (`email`, `uri`, `uuid`, `date`, `date-time`, etc.) |

### NumberSchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->min($n)` | `int\|float` | Minimum value (inclusive) |
| `->max($n)` | `int\|float` | Maximum value (inclusive) |
| `->exclusiveMin($n)` | `int\|float` | Minimum value (exclusive) |
| `->exclusiveMax($n)` | `int\|float` | Maximum value (exclusive) |
| `->multipleOf($n)` | `int\|float` | Value must be a multiple of this |

### ArraySchema Constraints

| Method | Parameter | Description |
|---|---|---|
| `->minItems($n)` | `int` | Minimum number of elements |
| `->maxItems($n)` | `int` | Maximum number of elements |

### UnknownKeyMode (Enum)

| Value | Description |
|---|---|
| `UnknownKeyMode::Reject` | Error on unexpected keys (default) |
| `UnknownKeyMode::Strip` | Silently remove unexpected keys |
| `UnknownKeyMode::Allow` | Keep unexpected keys as-is |

### ParseResult

| Property | Type | Description |
|---|---|---|
| `$result->success` | `bool` | Whether parsing succeeded |
| `$result->value` | `mixed` | The parsed value (only meaningful when `success` is `true`) |
| `$result->issues` | `array` | List of validation issues (empty when `success` is `true`) |

### ValidationError

Extends `\RuntimeException`. Thrown by `parse()` on validation failure.

| Property | Type | Description |
|---|---|---|
| `$e->issues` | `array` | List of validation issues |
| `$e->getMessage()` | `string` | Summary error message |

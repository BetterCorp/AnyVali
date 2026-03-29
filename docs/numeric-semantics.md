# AnyVali Numeric Semantics Guide

## Overview

AnyVali's numeric type system is designed for cross-language safety. This document explains the rationale behind every numeric design decision, the available types, their ranges, and best practices for choosing the right type.

## Why `number` Means `float64`

When you write `v.number()` in any AnyVali SDK, you get IEEE 754 double-precision floating point (`float64`). This is a deliberate choice.

### The Problem

Different languages have different default numeric types:

| Language | Default "number" type | Precision |
|---|---|---|
| JavaScript | `number` (float64) | 64-bit IEEE 754 |
| Python | `float` (float64) / `int` (arbitrary) | 64-bit IEEE 754 for floats |
| Go | no single default; `float64` is common | 64-bit IEEE 754 |
| Java | `double` (float64) | 64-bit IEEE 754 |
| C# | `double` (float64) | 64-bit IEEE 754 |
| Rust | `f64` | 64-bit IEEE 754 |
| PHP | `float` (float64) | 64-bit IEEE 754 |
| Ruby | `Float` (float64) | 64-bit IEEE 754 |
| Kotlin | `Double` (float64) | 64-bit IEEE 754 |
| C++ | `double` (usually float64) | 64-bit IEEE 754 |

IEEE 754 `float64` is the one floating-point type that every target language supports natively. Using it as the default ensures that a `number` schema means the same thing everywhere.

### What float64 Gives You

- 64 bits total: 1 sign, 11 exponent, 52 mantissa
- Approximately 15-17 significant decimal digits
- Range: approximately +/- 1.7976931348623157 x 10^308
- Exact integer representation up to 2^53 (9,007,199,254,740,992)

## Why `int` Means `int64`

When you write `v.int()`, you get a signed 64-bit integer (`int64`). This might seem wasteful on platforms where `int32` is the native default, but it solves a real portability problem.

### The Problem

| Language | Default "int" type | Width |
|---|---|---|
| JavaScript | No native integer type | float64 (53 bits of integer precision) |
| Python | `int` | Arbitrary precision |
| Go | `int` | Platform-dependent (32 or 64 bits) |
| Java | `int` | 32 bits |
| C# | `int` | 32 bits |
| Rust | `i32` | 32 bits |
| PHP | `int` | Platform-dependent (32 or 64 bits) |
| Ruby | `Integer` | Arbitrary precision |
| Kotlin | `Int` | 32 bits |
| C++ | `int` | Platform-dependent (usually 32 bits) |

If AnyVali defaulted `int` to `int32`, a schema authored in Python (which can hold arbitrarily large integers) might silently produce values that overflow when imported into Go on a 32-bit platform. By defaulting to `int64`, AnyVali provides a range that:

- Covers every practical use case (database IDs, timestamps, counters)
- Is representable in every target language (even if it requires a helper type)
- Avoids silent overflow on any platform

### The Memory vs Safety Tradeoff

Using `int64` as the default costs 8 bytes per integer value instead of 4 bytes for `int32`. In practice:

- For validation schemas describing API payloads, this cost is negligible
- For high-volume data processing where memory matters, use explicit narrow types like `int32` or `int16`
- The safety gain (no silent overflow across 10 languages) outweighs the memory cost for the vast majority of use cases

AnyVali makes the safe choice the default and the optimized choice explicit. If you know your values fit in 32 bits, use `v.int32()`.

## Cross-Language Numeric Differences

### JavaScript

JavaScript has only one numeric type: `number`, which is IEEE 754 `float64`. This means:

- There is no native integer type
- Integers are exact only up to 2^53 (Number.MAX_SAFE_INTEGER = 9,007,199,254,740,991)
- `BigInt` exists but is a separate type that cannot be mixed with `number` in arithmetic
- The JS SDK must use `BigInt` or reject values outside safe integer range for `int64` and `uint64` schemas

**Implication**: `int64` and `uint64` schemas in the JS SDK will have special handling. Values that exceed `Number.MAX_SAFE_INTEGER` must use `BigInt` or the SDK must reject them. Silent precision loss is not allowed.

### Python

Python's `int` type has arbitrary precision -- there is no overflow. Python's `float` is IEEE 754 `float64`.

- `int` values never overflow, but AnyVali `int64` schemas must still enforce the `int64` range
- `float` is always 64-bit
- Python does not have native `float32`; the SDK must validate range and precision constraints

**Implication**: The Python SDK must explicitly enforce range limits for all fixed-width integer types, even though the language would happily accept larger values.

### Go

Go has explicit numeric types: `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`. The plain `int` type is platform-dependent (32 or 64 bits).

- Go's type system maps directly to AnyVali's numeric kinds
- The SDK should use `int64` for `int` schemas, not the platform-dependent `int`
- No helper types are needed for any standard AnyVali numeric kind

**Implication**: Go is the most natural fit for AnyVali's numeric model. The SDK implementation is straightforward.

### Java and Kotlin

Java has `byte` (8), `short` (16), `int` (32), `long` (64), `float` (32), `double` (64). All signed. Kotlin mirrors these.

- No unsigned integer types in Java (Kotlin has `UInt`, `ULong`, etc.)
- The Java SDK must use helper types or range validation for `uint32` and `uint64`
- `uint8` through `uint16` can be represented with wider signed types

**Implication**: Unsigned types require careful handling in the Java SDK.

### C#

C# has `sbyte`/`byte`, `short`/`ushort`, `int`/`uint`, `long`/`ulong`, `float`, `double`. Both signed and unsigned types are native.

**Implication**: C# maps directly to AnyVali's full numeric type set with no gaps.

### Rust

Rust has `i8`-`i128`, `u8`-`u128`, `f32`, `f64`. Complete coverage.

**Implication**: Rust maps directly, with even more precision available than AnyVali requires.

### PHP

PHP has a single `int` type (platform-dependent width, usually 64-bit) and `float` (always float64).

- No native unsigned integer types
- No native fixed-width integer types
- The SDK must enforce range constraints in validation logic

**Implication**: PHP requires range-based validation for all narrow and unsigned integer types.

### Ruby

Ruby has arbitrary-precision `Integer` and float64 `Float`.

- Similar to Python: no native fixed-width integers
- Range constraints must be enforced by the SDK

## Available Numeric Types

### Integer Types

| Kind | Width | Signed | Minimum | Maximum |
|---|---|---|---|---|
| `int8` | 8 bits | Yes | -128 | 127 |
| `int16` | 16 bits | Yes | -32,768 | 32,767 |
| `int32` | 32 bits | Yes | -2,147,483,648 | 2,147,483,647 |
| `int64` | 64 bits | Yes | -9,223,372,036,854,775,808 | 9,223,372,036,854,775,807 |
| `uint8` | 8 bits | No | 0 | 255 |
| `uint16` | 16 bits | No | 0 | 65,535 |
| `uint32` | 32 bits | No | 0 | 4,294,967,295 |
| `uint64` | 64 bits | No | 0 | 18,446,744,073,709,551,615 |

The alias `int` is equivalent to `int64`.

### Floating-Point Types

| Kind | Width | Approximate Range | Significant Digits |
|---|---|---|---|
| `float32` | 32 bits | +/- 3.4028235 x 10^38 | ~7 |
| `float64` | 64 bits | +/- 1.7976931348623157 x 10^308 | ~15-17 |

The alias `number` is equivalent to `float64`.

### Range Reference Table

For convenience, here are the exact boundary values SDKs must enforce:

```
int8:    [-128, 127]
int16:   [-32768, 32767]
int32:   [-2147483648, 2147483647]
int64:   [-9223372036854775808, 9223372036854775807]

uint8:   [0, 255]
uint16:  [0, 65535]
uint32:  [0, 4294967295]
uint64:  [0, 18446744073709551615]

float32: [-3.4028235e+38, 3.4028235e+38]  (plus special values)
float64: [-1.7976931348623157e+308, 1.7976931348623157e+308]  (plus special values)
```

## How Narrowing Works

AnyVali uses **rejection, not silent approximation**. This is a core safety guarantee.

### What Narrowing Means

When a value is checked against a numeric schema, the SDK must verify that the value fits within the exact range and precision of that type. If it does not, the parse fails with a structured error.

### Examples

```typescript
const schema = v.int8(); // Range: -128 to 127

schema.parse(42);    // OK: 42 is within int8 range
schema.parse(200);   // FAILS: 200 exceeds int8 maximum (127)
schema.parse(-200);  // FAILS: -200 is below int8 minimum (-128)
schema.parse(3.14);  // FAILS: 3.14 is not an integer
```

```python
schema = v.uint16()  # Range: 0 to 65535

schema.parse(1000)   # OK
schema.parse(-1)     # FAILS: negative value for unsigned type
schema.parse(70000)  # FAILS: exceeds uint16 maximum
```

### What SDKs Must Not Do

- Must not silently truncate values (e.g., converting 200 to 127 for `int8`)
- Must not silently round floats to integers
- Must not silently lose precision when widening (this is generally safe, but the schema semantics must be preserved)
- Must not approximate `int64` values as `float64` without explicit opt-in

### Error Reporting

When narrowing fails, the SDK returns an issue with code `invalid_type` or `too_small` / `too_large` depending on the nature of the failure:

```json
{
  "code": "too_large",
  "message": "Value 200 exceeds maximum for int8 (127)",
  "path": ["age"],
  "expected": "int8",
  "received": "200"
}
```

## Best Practices for Choosing Numeric Types

### Use the defaults unless you have a reason not to

- `v.number()` (float64) for decimal values
- `v.int()` (int64) for integer values

These are safe across all 10 languages and cover the vast majority of use cases.

### Use narrow types when you know the domain

- HTTP status codes: `v.uint16()` (range 100-599, fits in uint16)
- Percentages: `v.uint8()` or `v.int().min(0).max(100)`
- Database IDs: `v.int64()` (explicitly, for clarity)
- Port numbers: `v.uint16()` (range 0-65535)
- Unix timestamps (seconds): `v.int64()`
- Byte values: `v.uint8()`
- RGB color channels: `v.uint8()`

### Prefer constraints over narrow types for business rules

If you want values between 0 and 100, you can use either approach:

```typescript
// Option A: narrow type
const pct = v.uint8().max(100);

// Option B: constraints on default int
const pct = v.int().min(0).max(100);
```

Option B is often clearer because it separates the machine-level representation from the business rule. Option A is useful when you also want to communicate the storage width.

### Be cautious with `int64` and `uint64` in JavaScript contexts

If your schema will be used in JavaScript:

- Values within `Number.MAX_SAFE_INTEGER` (2^53 - 1) work seamlessly
- Values outside that range require `BigInt` support in the JS SDK
- Consider whether `int32` or `uint32` is sufficient for your use case

### Use `float32` only when you need it

`float32` loses precision compared to `float64`. Use it only when:

- You are interfacing with systems that use 32-bit floats (GPU data, certain binary protocols)
- Storage efficiency matters more than precision

## Known Limitations

### JavaScript int64 / uint64 Precision

JavaScript's `number` type can only represent integers exactly up to 2^53. The AnyVali JS SDK must handle `int64` and `uint64` schemas using one of these strategies:

1. Use `BigInt` for values outside the safe integer range
2. Reject values outside `Number.MAX_SAFE_INTEGER` with an error

The exact strategy is an SDK implementation decision, but silent precision loss is never acceptable.

When exporting schemas that use `int64` or `uint64` from JavaScript, the JSON serialization must preserve full precision. This may require serializing large integers as strings in the JSON payload.

### Python Arbitrary-Precision Integers

Python integers have no upper bound, but AnyVali `int64` schemas must still enforce the `int64` range. A Python developer accustomed to writing `x = 2**100` must understand that this value will fail validation against an `int64` schema.

### Platform-Dependent Integer Widths

Go's `int` and C++'s `int` are platform-dependent. AnyVali SDKs in these languages must use explicitly-sized types (`int64` in Go, `int64_t` in C++) for the default `int` schema kind.

### NaN and Infinity

The AnyVali spec does not define portable behavior for `NaN`, `+Infinity`, or `-Infinity` in v1. SDKs should reject these values for portable schemas. Language-specific handling may be added via extensions in future versions.

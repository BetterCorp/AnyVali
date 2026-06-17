<?php

declare(strict_types=1);

namespace AnyVali\Parse;

use AnyVali\IssueCodes;
use AnyVali\ValidationIssue;

final class Coercion
{
    /**
     * Integer-family schema kinds. A bare/"string" coercion on any of these
     * resolves to string->int.
     */
    private const INT_KINDS = [
        'int', 'int8', 'int16', 'int32', 'int64',
        'uint8', 'uint16', 'uint32', 'uint64',
    ];

    /**
     * Number-family schema kinds. A bare/"string" coercion on any of these
     * resolves to string->number.
     */
    private const NUMBER_KINDS = ['number', 'float32', 'float64'];

    private function __construct()
    {
    }

    /**
     * Apply a single coercion step.
     * @return array{0: mixed, 1: ?ValidationIssue} [coerced value, issue or null]
     * @param array<int|string> $path
     */
    public static function apply(
        string $coercion,
        mixed $value,
        string $targetKind,
        array $path,
    ): array {
        // Resolve the generic/default source token "string" to a concrete
        // typed coercion based on the schema's own kind. This is the no-arg /
        // bare-source ergonomic: ->coerce() (which stores "string") or
        // ->coerce('string') coerces string input to the schema target.
        //   - numeric kinds  => string->number
        //   - integer kinds  => string->int
        //   - bool           => string->bool
        //   - string kind    => identity passthrough (trim/lower/upper handle
        //                       the real string transforms)
        if ($coercion === 'string') {
            if (in_array($targetKind, self::INT_KINDS, true)) {
                $coercion = 'string->int';
            } elseif (in_array($targetKind, self::NUMBER_KINDS, true)) {
                $coercion = 'string->number';
            } elseif ($targetKind === 'bool') {
                $coercion = 'string->bool';
            } else {
                // string (or any non-coercible) target: nothing to convert.
                return [$value, null];
            }
        }

        return match ($coercion) {
            'string->int' => self::stringToInt($value, $path),
            'string->number' => self::stringToNumber($value, $path),
            'string->bool' => self::stringToBool($value, $path),
            'trim' => self::trim($value, $path),
            'lower' => self::lower($value, $path),
            'upper' => self::upper($value, $path),
            default => [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: "Unknown coercion: {$coercion}",
                path: $path,
                expected: $targetKind,
                received: self::describeValue($value),
            )],
        };
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function stringToInt(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }

        $trimmed = trim($value);
        if (!preg_match('/^-?\d+$/', $trimmed)) {
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: "Cannot coerce \"{$value}\" to int",
                path: $path,
                expected: 'int',
                received: $value,
            )];
        }

        return [(int)$trimmed, null];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function stringToNumber(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }

        // Strict ASCII decimal float (optionally signed, optional exponent).
        // PHP's is_numeric()/(float) cast are more permissive than the ECMA-262
        // reference: is_numeric accepts hex-ish forms in older PHP, leading/
        // trailing whitespace, and "(float)" silently turns garbage into 0.0.
        // Gate on a strict grammar so coercion behaves identically across SDKs
        // (spec 5.1: parse as DECIMAL floating-point only). Mirrors the Python
        // canonical regex.
        $trimmed = trim($value);
        if (!preg_match('/^[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?$/', $trimmed)) {
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: "Cannot coerce \"{$value}\" to number",
                path: $path,
                expected: 'number',
                received: $value,
            )];
        }

        $num = (float)$trimmed;
        // Guard against magnitudes that overflow to +/-INF (e.g. "1e400"):
        // INF/NaN are not decimal-representable finite numbers (spec 5.1).
        if (is_infinite($num) || is_nan($num)) {
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: "Cannot coerce \"{$value}\" to number",
                path: $path,
                expected: 'number',
                received: $value,
            )];
        }

        return [$num, null];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function stringToBool(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }

        $lower = strtolower(trim($value));
        if ($lower === 'true' || $lower === '1') {
            return [true, null];
        }
        if ($lower === 'false' || $lower === '0') {
            return [false, null];
        }

        return [$value, new ValidationIssue(
            code: IssueCodes::COERCION_FAILED,
            message: "Cannot coerce \"{$value}\" to bool",
            path: $path,
            expected: 'bool',
            received: $value,
        )];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function trim(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }
        return [trim($value), null];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function lower(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }
        return [mb_strtolower($value), null];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function upper(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, null];
        }
        return [mb_strtoupper($value), null];
    }

    private static function describeValue(mixed $value): string
    {
        if ($value === null) {
            return 'null';
        }
        if (is_string($value)) {
            return $value;
        }
        if (is_bool($value)) {
            return $value ? 'true' : 'false';
        }
        if (is_int($value) || is_float($value)) {
            return (string)$value;
        }
        return gettype($value);
    }
}

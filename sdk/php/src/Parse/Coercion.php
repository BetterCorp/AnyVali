<?php

declare(strict_types=1);

namespace AnyVali\Parse;

use AnyVali\IssueCodes;
use AnyVali\ValidationIssue;

final class Coercion
{
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
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: 'Expected string input for string->int coercion',
                path: $path,
                expected: 'int',
                received: self::describeValue($value),
            )];
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
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: 'Expected string input for string->number coercion',
                path: $path,
                expected: 'number',
                received: self::describeValue($value),
            )];
        }

        $trimmed = trim($value);
        if (!is_numeric($trimmed)) {
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: "Cannot coerce \"{$value}\" to number",
                path: $path,
                expected: 'number',
                received: $value,
            )];
        }

        return [(float)$trimmed, null];
    }

    /**
     * @param array<int|string> $path
     * @return array{0: mixed, 1: ?ValidationIssue}
     */
    private static function stringToBool(mixed $value, array $path): array
    {
        if (!is_string($value)) {
            return [$value, new ValidationIssue(
                code: IssueCodes::COERCION_FAILED,
                message: 'Expected string input for string->bool coercion',
                path: $path,
                expected: 'bool',
                received: self::describeValue($value),
            )];
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

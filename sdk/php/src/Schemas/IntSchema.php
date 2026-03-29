<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class IntSchema extends Schema
{
    private ?float $min = null;
    private ?float $max = null;
    private ?float $exclusiveMin = null;
    private ?float $exclusiveMax = null;
    private ?float $multipleOf = null;
    private string $kind;

    // Width ranges for typed integers
    private const RANGES = [
        'int8'   => [-128, 127],
        'int16'  => [-32768, 32767],
        'int32'  => [-2147483648, 2147483647],
        'int64'  => [PHP_INT_MIN, PHP_INT_MAX],
        'int'    => [PHP_INT_MIN, PHP_INT_MAX],
        'uint8'  => [0, 255],
        'uint16' => [0, 65535],
        'uint32' => [0, 4294967295],
        'uint64' => [0, PHP_INT_MAX], // PHP max, but semantically uint64
    ];

    public function __construct(string $kind = 'int')
    {
        $this->kind = $kind;
    }

    public function getKind(): string
    {
        return $this->kind;
    }

    public function min(float|int $min): self
    {
        $clone = clone $this;
        $clone->min = (float)$min;
        return $clone;
    }

    public function max(float|int $max): self
    {
        $clone = clone $this;
        $clone->max = (float)$max;
        return $clone;
    }

    public function exclusiveMin(float|int $min): self
    {
        $clone = clone $this;
        $clone->exclusiveMin = (float)$min;
        return $clone;
    }

    public function exclusiveMax(float|int $max): self
    {
        $clone = clone $this;
        $clone->exclusiveMax = (float)$max;
        return $clone;
    }

    public function multipleOf(float|int $value): self
    {
        $clone = clone $this;
        $clone->multipleOf = (float)$value;
        return $clone;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        // Must be numeric (int or float with no fractional part)
        $isInteger = is_int($value) || (is_float($value) && floor($value) === $value && !is_nan($value) && !is_infinite($value));

        if (!$isInteger) {
            $received = self::getTypeName($value);
            // For non-integer floats
            if (is_float($value)) {
                $received = 'number';
            }
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: "Expected {$this->kind}",
                path: $ctx->path,
                expected: $this->kind,
                received: $received,
            )]);
        }

        $intValue = is_float($value) ? (int)$value : $value;
        $numericValue = (float)$intValue;
        $issues = [];

        // Check width range
        if (isset(self::RANGES[$this->kind]) && $this->kind !== 'int' && $this->kind !== 'int64') {
            [$rangeMin, $rangeMax] = self::RANGES[$this->kind];
            if ($intValue < $rangeMin) {
                $issues[] = new ValidationIssue(
                    code: IssueCodes::TOO_SMALL,
                    message: "Value out of range for {$this->kind}",
                    path: $ctx->path,
                    expected: $this->kind,
                    received: (string)$intValue,
                );
            }
            if ($intValue > $rangeMax) {
                $issues[] = new ValidationIssue(
                    code: IssueCodes::TOO_LARGE,
                    message: "Value out of range for {$this->kind}",
                    path: $ctx->path,
                    expected: $this->kind,
                    received: (string)$intValue,
                );
            }
        }

        // User-specified constraints
        if ($this->min !== null && $numericValue < $this->min) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Value must be >= {$this->min}",
                path: $ctx->path,
                expected: self::formatNum($this->min),
                received: (string)$intValue,
            );
        }

        if ($this->max !== null && $numericValue > $this->max) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Value must be <= {$this->max}",
                path: $ctx->path,
                expected: self::formatNum($this->max),
                received: (string)$intValue,
            );
        }

        if ($this->exclusiveMin !== null && $numericValue <= $this->exclusiveMin) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Value must be > {$this->exclusiveMin}",
                path: $ctx->path,
                expected: self::formatNum($this->exclusiveMin),
                received: (string)$intValue,
            );
        }

        if ($this->exclusiveMax !== null && $numericValue >= $this->exclusiveMax) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Value must be < {$this->exclusiveMax}",
                path: $ctx->path,
                expected: self::formatNum($this->exclusiveMax),
                received: (string)$intValue,
            );
        }

        if ($this->multipleOf !== null) {
            $remainder = fmod($numericValue, $this->multipleOf);
            if (abs($remainder) > 1e-10 && abs($remainder - $this->multipleOf) > 1e-10) {
                $issues[] = new ValidationIssue(
                    code: IssueCodes::INVALID_NUMBER,
                    message: "Value must be a multiple of {$this->multipleOf}",
                    path: $ctx->path,
                    expected: self::formatNum($this->multipleOf),
                    received: (string)$intValue,
                );
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($intValue);
    }

    private static function formatNum(float $v): string
    {
        if (floor($v) === $v && !is_infinite($v) && abs($v) < PHP_INT_MAX) {
            return (string)(int)$v;
        }
        return (string)$v;
    }

    public function exportNode(): array
    {
        $node = ['kind' => $this->kind];
        if ($this->min !== null) $node['min'] = $this->min;
        if ($this->max !== null) $node['max'] = $this->max;
        if ($this->exclusiveMin !== null) $node['exclusiveMin'] = $this->exclusiveMin;
        if ($this->exclusiveMax !== null) $node['exclusiveMax'] = $this->exclusiveMax;
        if ($this->multipleOf !== null) $node['multipleOf'] = $this->multipleOf;
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        if ($this->coerce !== null) $node['coerce'] = $this->coerce;
        return $node;
    }
}

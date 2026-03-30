<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

/**
 * @extends Schema<float>
 */
final class NumberSchema extends Schema
{
    private ?float $min = null;
    private ?float $max = null;
    private ?float $exclusiveMin = null;
    private ?float $exclusiveMax = null;
    private ?float $multipleOf = null;
    private string $kind;

    public function __construct(string $kind = 'number')
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
        if (!is_int($value) && !is_float($value)) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: "Expected {$this->kind}",
                path: $ctx->path,
                expected: $this->kind,
                received: self::getTypeName($value),
            )]);
        }

        $numericValue = (float)$value;
        $issues = [];

        if ($this->min !== null && $numericValue < $this->min) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Value must be >= {$this->min}",
                path: $ctx->path,
                expected: self::formatNum($this->min),
                received: self::formatNum($numericValue),
            );
        }

        if ($this->max !== null && $numericValue > $this->max) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Value must be <= {$this->max}",
                path: $ctx->path,
                expected: self::formatNum($this->max),
                received: self::formatNum($numericValue),
            );
        }

        if ($this->exclusiveMin !== null && $numericValue <= $this->exclusiveMin) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Value must be > {$this->exclusiveMin}",
                path: $ctx->path,
                expected: self::formatNum($this->exclusiveMin),
                received: self::formatNum($numericValue),
            );
        }

        if ($this->exclusiveMax !== null && $numericValue >= $this->exclusiveMax) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Value must be < {$this->exclusiveMax}",
                path: $ctx->path,
                expected: self::formatNum($this->exclusiveMax),
                received: self::formatNum($numericValue),
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
                    received: self::formatNum($numericValue),
                );
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($value);
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

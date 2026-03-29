<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\Format\FormatValidators;
use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class StringSchema extends Schema
{
    private ?int $minLength = null;
    private ?int $maxLength = null;
    private ?string $pattern = null;
    private ?string $startsWith = null;
    private ?string $endsWith = null;
    private ?string $includes = null;
    private ?string $format = null;

    public function getKind(): string
    {
        return 'string';
    }

    public function minLength(int $min): self
    {
        $clone = clone $this;
        $clone->minLength = $min;
        return $clone;
    }

    public function maxLength(int $max): self
    {
        $clone = clone $this;
        $clone->maxLength = $max;
        return $clone;
    }

    public function pattern(string $pattern): self
    {
        $clone = clone $this;
        $clone->pattern = $pattern;
        return $clone;
    }

    public function startsWith(string $prefix): self
    {
        $clone = clone $this;
        $clone->startsWith = $prefix;
        return $clone;
    }

    public function endsWith(string $suffix): self
    {
        $clone = clone $this;
        $clone->endsWith = $suffix;
        return $clone;
    }

    public function includes(string $substring): self
    {
        $clone = clone $this;
        $clone->includes = $substring;
        return $clone;
    }

    public function format(string $format): self
    {
        $clone = clone $this;
        $clone->format = $format;
        return $clone;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if (!is_string($value)) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected string',
                path: $ctx->path,
                expected: 'string',
                received: self::getTypeName($value),
            )]);
        }

        $issues = [];
        $len = mb_strlen($value);

        if ($this->minLength !== null && $len < $this->minLength) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "String must be at least {$this->minLength} characters",
                path: $ctx->path,
                expected: (string)$this->minLength,
                received: (string)$len,
            );
        }

        if ($this->maxLength !== null && $len > $this->maxLength) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "String must be at most {$this->maxLength} characters",
                path: $ctx->path,
                expected: (string)$this->maxLength,
                received: (string)$len,
            );
        }

        if ($this->pattern !== null && !preg_match('/' . $this->pattern . '/', $value)) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::INVALID_STRING,
                message: "String does not match pattern {$this->pattern}",
                path: $ctx->path,
                expected: $this->pattern,
                received: $value,
            );
        }

        if ($this->startsWith !== null && !str_starts_with($value, $this->startsWith)) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::INVALID_STRING,
                message: "String must start with \"{$this->startsWith}\"",
                path: $ctx->path,
                expected: $this->startsWith,
                received: $value,
            );
        }

        if ($this->endsWith !== null && !str_ends_with($value, $this->endsWith)) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::INVALID_STRING,
                message: "String must end with \"{$this->endsWith}\"",
                path: $ctx->path,
                expected: $this->endsWith,
                received: $value,
            );
        }

        if ($this->includes !== null && !str_contains($value, $this->includes)) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::INVALID_STRING,
                message: "String must include \"{$this->includes}\"",
                path: $ctx->path,
                expected: $this->includes,
                received: $value,
            );
        }

        if ($this->format !== null && !FormatValidators::validate($this->format, $value)) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::INVALID_STRING,
                message: "String does not match format \"{$this->format}\"",
                path: $ctx->path,
                expected: $this->format,
                received: $value,
            );
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($value);
    }

    public function exportNode(): array
    {
        $node = ['kind' => 'string'];
        if ($this->minLength !== null) $node['minLength'] = $this->minLength;
        if ($this->maxLength !== null) $node['maxLength'] = $this->maxLength;
        if ($this->pattern !== null) $node['pattern'] = $this->pattern;
        if ($this->startsWith !== null) $node['startsWith'] = $this->startsWith;
        if ($this->endsWith !== null) $node['endsWith'] = $this->endsWith;
        if ($this->includes !== null) $node['includes'] = $this->includes;
        if ($this->format !== null) $node['format'] = $this->format;
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        if ($this->coerce !== null) $node['coerce'] = $this->coerce;
        return $node;
    }
}

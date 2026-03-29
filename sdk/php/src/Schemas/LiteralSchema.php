<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class LiteralSchema extends Schema
{
    public function __construct(
        private readonly mixed $literalValue,
    ) {
    }

    public function getKind(): string
    {
        return 'literal';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if ($value !== $this->literalValue) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_LITERAL,
                message: 'Expected literal ' . self::formatValue($this->literalValue),
                path: $ctx->path,
                expected: self::formatValue($this->literalValue),
                received: self::formatValue($value),
            )]);
        }
        return ParseResult::ok($value);
    }

    public function getLiteralValue(): mixed
    {
        return $this->literalValue;
    }

    public function exportNode(): array
    {
        return ['kind' => 'literal', 'value' => $this->literalValue];
    }
}

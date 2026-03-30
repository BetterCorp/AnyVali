<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

/**
 * @extends Schema<never>
 */
final class NeverSchema extends Schema
{
    public function getKind(): string
    {
        return 'never';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        return ParseResult::fail([new ValidationIssue(
            code: IssueCodes::INVALID_TYPE,
            message: 'No value is valid for never type',
            path: $ctx->path,
            expected: 'never',
            received: self::getTypeName($value),
        )]);
    }

    public function exportNode(): array
    {
        return ['kind' => 'never'];
    }
}

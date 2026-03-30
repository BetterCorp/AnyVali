<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

/**
 * @extends Schema<null>
 */
final class NullSchema extends Schema
{
    public function getKind(): string
    {
        return 'null';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if ($value !== null) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected null',
                path: $ctx->path,
                expected: 'null',
                received: self::getTypeName($value),
            )]);
        }
        return ParseResult::ok(null);
    }

    public function exportNode(): array
    {
        return ['kind' => 'null'];
    }
}

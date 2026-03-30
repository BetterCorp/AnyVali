<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

/**
 * @extends Schema<bool>
 */
final class BoolSchema extends Schema
{
    public function getKind(): string
    {
        return 'bool';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if (!is_bool($value)) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected bool',
                path: $ctx->path,
                expected: 'bool',
                received: self::getTypeName($value),
            )]);
        }
        return ParseResult::ok($value);
    }

    public function exportNode(): array
    {
        $node = ['kind' => 'bool'];
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        if ($this->coerce !== null) $node['coerce'] = $this->coerce;
        return $node;
    }
}

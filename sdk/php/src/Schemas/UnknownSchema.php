<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;

final class UnknownSchema extends Schema
{
    public function getKind(): string
    {
        return 'unknown';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        return ParseResult::ok($value);
    }

    public function exportNode(): array
    {
        return ['kind' => 'unknown'];
    }
}

<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;

final class AnySchema extends Schema
{
    public function getKind(): string
    {
        return 'any';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        return ParseResult::ok($value);
    }

    public function exportNode(): array
    {
        return ['kind' => 'any'];
    }
}

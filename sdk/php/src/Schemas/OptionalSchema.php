<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;

final class OptionalSchema extends Schema
{
    public function __construct(
        private readonly Schema $innerSchema,
    ) {
    }

    public function getKind(): string
    {
        return 'optional';
    }

    public function getInnerSchema(): Schema
    {
        return $this->innerSchema;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        // When value is present, delegate to inner schema
        return $this->innerSchema->safeParse($value, $ctx);
    }

    public function exportNode(): array
    {
        $node = ['kind' => 'optional', 'schema' => $this->innerSchema->exportNode()];
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        return $node;
    }
}

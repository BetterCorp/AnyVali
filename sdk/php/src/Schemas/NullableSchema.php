<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;

/**
 * @extends Schema<mixed>
 */
final class NullableSchema extends Schema
{
    public function __construct(
        private readonly Schema $innerSchema,
    ) {
    }

    public function getKind(): string
    {
        return 'nullable';
    }

    public function getInnerSchema(): Schema
    {
        return $this->innerSchema;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if ($value === null) {
            return ParseResult::ok(null);
        }
        return $this->innerSchema->safeParse($value, $ctx);
    }

    public function exportNode(): array
    {
        $node = ['kind' => 'nullable', 'schema' => $this->innerSchema->exportNode()];
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        return $node;
    }
}

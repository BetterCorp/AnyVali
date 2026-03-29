<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class ArraySchema extends Schema
{
    private ?int $minItems = null;
    private ?int $maxItems = null;

    public function __construct(
        private readonly Schema $items,
    ) {
    }

    public function getKind(): string
    {
        return 'array';
    }

    public function minItems(int $min): self
    {
        $clone = clone $this;
        $clone->minItems = $min;
        return $clone;
    }

    public function maxItems(int $max): self
    {
        $clone = clone $this;
        $clone->maxItems = $max;
        return $clone;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if (!is_array($value) || ($value !== [] && !array_is_list($value))) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected array',
                path: $ctx->path,
                expected: 'array',
                received: self::getTypeName($value),
            )]);
        }

        $count = count($value);
        $issues = [];

        if ($this->minItems !== null && $count < $this->minItems) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Array must have at least {$this->minItems} items",
                path: $ctx->path,
                expected: (string)$this->minItems,
                received: (string)$count,
            );
        }

        if ($this->maxItems !== null && $count > $this->maxItems) {
            $issues[] = new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Array must have at most {$this->maxItems} items",
                path: $ctx->path,
                expected: (string)$this->maxItems,
                received: (string)$count,
            );
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        $parsedItems = [];
        foreach ($value as $i => $item) {
            $result = $this->items->safeParse($item, $ctx->child($i));
            if (!$result->success) {
                $issues = array_merge($issues, $result->issues);
            } else {
                $parsedItems[] = $result->value;
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($parsedItems);
    }

    public function getItems(): Schema
    {
        return $this->items;
    }

    public function exportNode(): array
    {
        $node = ['kind' => 'array', 'items' => $this->items->exportNode()];
        if ($this->minItems !== null) $node['minItems'] = $this->minItems;
        if ($this->maxItems !== null) $node['maxItems'] = $this->maxItems;
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        return $node;
    }
}

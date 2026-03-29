<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class TupleSchema extends Schema
{
    /** @var Schema[] */
    private readonly array $elements;

    /**
     * @param Schema[] $elements
     */
    public function __construct(array $elements)
    {
        $this->elements = $elements;
    }

    public function getKind(): string
    {
        return 'tuple';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if (!is_array($value) || ($value !== [] && !array_is_list($value))) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected tuple',
                path: $ctx->path,
                expected: 'tuple',
                received: self::getTypeName($value),
            )]);
        }

        $count = count($value);
        $expected = count($this->elements);

        if ($count < $expected) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::TOO_SMALL,
                message: "Tuple requires exactly {$expected} elements",
                path: $ctx->path,
                expected: (string)$expected,
                received: (string)$count,
            )]);
        }

        if ($count > $expected) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::TOO_LARGE,
                message: "Tuple requires exactly {$expected} elements",
                path: $ctx->path,
                expected: (string)$expected,
                received: (string)$count,
            )]);
        }

        $parsedItems = [];
        $issues = [];

        foreach ($this->elements as $i => $schema) {
            $result = $schema->safeParse($value[$i], $ctx->child($i));
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

    /** @return Schema[] */
    public function getElements(): array
    {
        return $this->elements;
    }

    public function exportNode(): array
    {
        return [
            'kind' => 'tuple',
            'elements' => array_map(fn(Schema $s) => $s->exportNode(), $this->elements),
        ];
    }
}

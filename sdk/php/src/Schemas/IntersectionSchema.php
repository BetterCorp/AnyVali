<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;

final class IntersectionSchema extends Schema
{
    /** @var Schema[] */
    private readonly array $allOf;

    /**
     * @param Schema[] $schemas
     */
    public function __construct(array $schemas)
    {
        $this->allOf = $schemas;
    }

    public function getKind(): string
    {
        return 'intersection';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        $allIssues = [];
        $mergedResult = $value;

        foreach ($this->allOf as $schema) {
            $result = $schema->safeParse($value, $ctx);
            if (!$result->success) {
                $allIssues = array_merge($allIssues, $result->issues);
            } else {
                // For objects, merge parsed results
                if (is_array($result->value) && is_array($mergedResult) && !array_is_list($result->value)) {
                    $mergedResult = array_merge(
                        is_array($mergedResult) ? $mergedResult : [],
                        $result->value,
                    );
                } else {
                    $mergedResult = $result->value;
                }
            }
        }

        if (!empty($allIssues)) {
            return ParseResult::fail($allIssues);
        }

        return ParseResult::ok($mergedResult);
    }

    /** @return Schema[] */
    public function getAllOf(): array
    {
        return $this->allOf;
    }

    public function exportNode(): array
    {
        return [
            'kind' => 'intersection',
            'allOf' => array_map(fn(Schema $s) => $s->exportNode(), $this->allOf),
        ];
    }
}

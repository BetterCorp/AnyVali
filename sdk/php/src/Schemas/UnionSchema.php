<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class UnionSchema extends Schema
{
    /** @var Schema[] */
    private readonly array $variants;

    /**
     * @param Schema[] $variants
     */
    public function __construct(array $variants)
    {
        $this->variants = $variants;
    }

    public function getKind(): string
    {
        return 'union';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        foreach ($this->variants as $variant) {
            $result = $variant->safeParse($value, $ctx);
            if ($result->success) {
                return $result;
            }
        }

        $kinds = array_map(fn(Schema $s) => $s->getKind(), $this->variants);
        return ParseResult::fail([new ValidationIssue(
            code: IssueCodes::INVALID_UNION,
            message: 'Value does not match any variant',
            path: $ctx->path,
            expected: implode(' | ', $kinds),
            received: self::getTypeName($value),
        )]);
    }

    /** @return Schema[] */
    public function getVariants(): array
    {
        return $this->variants;
    }

    public function exportNode(): array
    {
        return [
            'kind' => 'union',
            'variants' => array_map(fn(Schema $s) => $s->exportNode(), $this->variants),
        ];
    }
}

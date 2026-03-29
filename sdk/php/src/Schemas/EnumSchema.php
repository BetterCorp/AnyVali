<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class EnumSchema extends Schema
{
    /** @var array<mixed> */
    private readonly array $values;

    /**
     * @param array<mixed> $values
     */
    public function __construct(array $values)
    {
        $this->values = $values;
    }

    public function getKind(): string
    {
        return 'enum';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        // Strict comparison
        foreach ($this->values as $allowed) {
            if ($value === $allowed) {
                return ParseResult::ok($value);
            }
        }

        $enumStr = implode(',', array_map(fn($v) => self::formatValue($v), $this->values));
        return ParseResult::fail([new ValidationIssue(
            code: IssueCodes::INVALID_TYPE,
            message: "Value is not in enum({$enumStr})",
            path: $ctx->path,
            expected: "enum({$enumStr})",
            received: self::formatValue($value),
        )]);
    }

    /** @return array<mixed> */
    public function getValues(): array
    {
        return $this->values;
    }

    public function exportNode(): array
    {
        return ['kind' => 'enum', 'values' => $this->values];
    }
}

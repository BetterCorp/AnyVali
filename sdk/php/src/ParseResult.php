<?php

declare(strict_types=1);

namespace AnyVali;

/**
 * @template T
 */
final class ParseResult
{
    /**
     * @param ValidationIssue[] $issues
     * @param T|null $value
     */
    public function __construct(
        public readonly bool $success,
        public readonly mixed $value = null,
        public readonly array $issues = [],
    ) {
    }

    /**
     * @template TValue
     * @param TValue $value
     * @return self<TValue>
     */
    public static function ok(mixed $value): self
    {
        return new self(success: true, value: $value);
    }

    /**
     * @param ValidationIssue[] $issues
     * @return self<null>
     */
    public static function fail(array $issues): self
    {
        return new self(success: false, issues: $issues);
    }
}

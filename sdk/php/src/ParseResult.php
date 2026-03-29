<?php

declare(strict_types=1);

namespace AnyVali;

final class ParseResult
{
    /**
     * @param ValidationIssue[] $issues
     */
    public function __construct(
        public readonly bool $success,
        public readonly mixed $value = null,
        public readonly array $issues = [],
    ) {
    }

    public static function ok(mixed $value): self
    {
        return new self(success: true, value: $value);
    }

    /**
     * @param ValidationIssue[] $issues
     */
    public static function fail(array $issues): self
    {
        return new self(success: false, issues: $issues);
    }
}

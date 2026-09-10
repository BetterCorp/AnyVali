<?php

declare(strict_types=1);

namespace AnyVali;

/**
 * @template-covariant T
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
     * Failure contains no successful value and is valid for every output type.
     * @param ValidationIssue[] $issues
     * @return self<never>
     */
    public static function fail(array $issues): self
    {
        // The failure value is always null; there is no successful T to infer.
        /** @var self<never> $failure */
        $failure = new self(success: false, issues: $issues);
        return $failure;
    }
}

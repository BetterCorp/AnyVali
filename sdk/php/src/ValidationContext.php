<?php

declare(strict_types=1);

namespace AnyVali;

final class ValidationContext
{
    public const MAX_DEPTH = 64;
    public readonly ValidationBudget $budget;
    /** @var array<string, array<string, mixed>> */
    public readonly array $definitions;

    /**
     * @param array<int|string> $path
     * @param array<string, array<string, mixed>> $definitions
     */
    public function __construct(
        public readonly array $path = [],
        array $definitions = [],
        public readonly ?UnknownKeyMode $inheritedUnknownKeys = null,
        public readonly ?string $sensitiveMode = null,
        public readonly mixed $sensitiveTransform = null,
        public readonly ?\ArrayObject $sensitiveCache = null,
        public readonly int $depth = 0,
        public readonly bool $skipCoercion = false,
        ?ValidationBudget $budget = null,
    ) {
        $this->definitions = $definitions;
        $this->budget = $budget ?? new ValidationBudget();
    }

    /**
     * Start a child schema pipeline, optionally appending a property or index to the path.
     * A composite member keeps the path while re-enabling its own coercion stage.
     * @param int|string|null $segment
     */
    public function child(int|string|null $segment = null): self
    {
        return new self(
            path: $segment === null ? $this->path : array_merge($this->path, [$segment]),
            definitions: $this->definitions,
            inheritedUnknownKeys: $this->inheritedUnknownKeys,
            sensitiveMode: $this->sensitiveMode,
            sensitiveTransform: $this->sensitiveTransform,
            sensitiveCache: $this->sensitiveCache,
            depth: $this->depth,
            skipCoercion: false,
            budget: $this->budget,
        );
    }

    public function descend(): self
    {
        return new self(
            path: $this->path,
            definitions: $this->definitions,
            inheritedUnknownKeys: $this->inheritedUnknownKeys,
            sensitiveMode: $this->sensitiveMode,
            sensitiveTransform: $this->sensitiveTransform,
            sensitiveCache: $this->sensitiveCache,
            depth: $this->depth + 1,
            skipCoercion: $this->skipCoercion,
            budget: $this->depth === 0 ? new ValidationBudget() : $this->budget,
        );
    }

    public function forDefault(): self
    {
        return new self(
            path: $this->path,
            definitions: $this->definitions,
            inheritedUnknownKeys: $this->inheritedUnknownKeys,
            sensitiveMode: $this->sensitiveMode,
            sensitiveTransform: $this->sensitiveTransform,
            sensitiveCache: $this->sensitiveCache,
            depth: $this->depth,
            skipCoercion: true,
            budget: $this->budget,
        );
    }

}

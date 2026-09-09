<?php

declare(strict_types=1);

namespace AnyVali;

final class ValidationContext
{
    public const MAX_DEPTH = 64;
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
    ) {
        $this->definitions = $definitions;
    }

    /**
     * @param int|string $segment
     */
    public function child(int|string $segment): self
    {
        return new self(
            path: array_merge($this->path, [$segment]),
            definitions: $this->definitions,
            inheritedUnknownKeys: $this->inheritedUnknownKeys,
            sensitiveMode: $this->sensitiveMode,
            sensitiveTransform: $this->sensitiveTransform,
            sensitiveCache: $this->sensitiveCache,
            depth: $this->depth,
            skipCoercion: $this->skipCoercion,
        );
    }

    public function descend(bool $skipCoercion = false): self
    {
        return new self(
            path: $this->path,
            definitions: $this->definitions,
            inheritedUnknownKeys: $this->inheritedUnknownKeys,
            sensitiveMode: $this->sensitiveMode,
            sensitiveTransform: $this->sensitiveTransform,
            sensitiveCache: $this->sensitiveCache,
            depth: $this->depth + 1,
            skipCoercion: $this->skipCoercion || $skipCoercion,
        );
    }

}

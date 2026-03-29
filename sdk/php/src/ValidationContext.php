<?php

declare(strict_types=1);

namespace AnyVali;

final class ValidationContext
{
    /** @var array<string, array<string, mixed>> */
    public readonly array $definitions;

    /**
     * @param array<int|string> $path
     * @param array<string, array<string, mixed>> $definitions
     */
    public function __construct(
        public readonly array $path = [],
        array $definitions = [],
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
        );
    }
}

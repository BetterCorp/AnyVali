<?php

declare(strict_types=1);

namespace AnyVali;

final class ValidationIssue
{
    /**
     * @param array<int|string> $path
     * @param array<string, mixed> $meta
     */
    public function __construct(
        public readonly string $code,
        public readonly string $message,
        public readonly array $path = [],
        public readonly ?string $expected = null,
        public readonly ?string $received = null,
        public readonly array $meta = [],
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function toArray(): array
    {
        $result = ['code' => $this->code, 'path' => $this->path];
        if ($this->expected !== null) {
            $result['expected'] = $this->expected;
        }
        if ($this->received !== null) {
            $result['received'] = $this->received;
        }
        return $result;
    }

    /**
     * @param array<int|string> $prefix
     */
    public function withPathPrefix(array $prefix): self
    {
        return new self(
            code: $this->code,
            message: $this->message,
            path: array_merge($prefix, $this->path),
            expected: $this->expected,
            received: $this->received,
            meta: $this->meta,
        );
    }
}

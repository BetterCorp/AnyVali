<?php

declare(strict_types=1);

namespace AnyVali;

final class AnyValiDocument
{
    /**
     * @param array<string, mixed> $root
     * @param array<string, array<string, mixed>> $definitions
     * @param array<string, mixed> $extensions
     */
    public function __construct(
        public readonly string $anyvaliVersion = '1.0',
        public readonly string $schemaVersion = '1',
        public readonly array $root = [],
        public readonly array $definitions = [],
        public readonly array $extensions = [],
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function toArray(): array
    {
        return [
            'anyvaliVersion' => $this->anyvaliVersion,
            'schemaVersion' => $this->schemaVersion,
            'root' => $this->root,
            'definitions' => (object)$this->definitions,
            'extensions' => (object)$this->extensions,
        ];
    }

    public function toJson(): string
    {
        return json_encode($this->toArray(), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
    }

    /**
     * @param array<string, mixed> $data
     */
    public static function fromArray(array $data): self
    {
        return new self(
            anyvaliVersion: $data['anyvaliVersion'] ?? '1.0',
            schemaVersion: $data['schemaVersion'] ?? '1',
            root: $data['root'] ?? [],
            definitions: (array)($data['definitions'] ?? []),
            extensions: (array)($data['extensions'] ?? []),
        );
    }

    public static function fromJson(string $json): self
    {
        $data = json_decode($json, true, 512, JSON_THROW_ON_ERROR);
        return self::fromArray($data);
    }
}

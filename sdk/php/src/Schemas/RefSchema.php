<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class RefSchema extends Schema
{
    public function __construct(
        private readonly string $ref,
        private ?Schema $resolvedSchema = null,
    ) {
    }

    public function getKind(): string
    {
        return 'ref';
    }

    public function getRef(): string
    {
        return $this->ref;
    }

    public function resolve(Schema $schema): void
    {
        $this->resolvedSchema = $schema;
    }

    public function getResolvedSchema(): ?Schema
    {
        return $this->resolvedSchema;
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        if ($this->resolvedSchema !== null) {
            return $this->resolvedSchema->safeParse($value, $ctx);
        }

        // Try to resolve from context definitions
        $refName = $this->extractRefName();
        if ($refName !== null && isset($ctx->definitions[$refName])) {
            // Import the definition and validate
            $defSchema = \AnyVali\Interchange\Importer::importNode(
                $ctx->definitions[$refName],
                $ctx->definitions,
            );
            return $defSchema->safeParse($value, $ctx);
        }

        return ParseResult::fail([new ValidationIssue(
            code: IssueCodes::INVALID_TYPE,
            message: "Cannot resolve ref: {$this->ref}",
            path: $ctx->path,
            expected: 'ref',
            received: self::getTypeName($value),
        )]);
    }

    private function extractRefName(): ?string
    {
        if (str_starts_with($this->ref, '#/definitions/')) {
            return substr($this->ref, strlen('#/definitions/'));
        }
        return null;
    }

    public function exportNode(): array
    {
        return ['kind' => 'ref', 'ref' => $this->ref];
    }
}

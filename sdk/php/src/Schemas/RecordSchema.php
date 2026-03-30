<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

final class RecordSchema extends Schema
{
    public function __construct(
        private readonly Schema $valueSchema,
    ) {
    }

    public function getKind(): string
    {
        return 'record';
    }

    protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult
    {
        // Accept stdClass as an object (e.g. empty JSON {} decoded with json_decode(…, false))
        if ($value instanceof \stdClass) {
            $value = (array)$value;
        }
        if (!is_array($value) || ($value !== [] && array_is_list($value))) {
            return ParseResult::fail([new ValidationIssue(
                code: IssueCodes::INVALID_TYPE,
                message: 'Expected record',
                path: $ctx->path,
                expected: 'record',
                received: self::getTypeName($value),
            )]);
        }

        $parsed = [];
        $issues = [];

        foreach ($value as $key => $v) {
            $result = $this->valueSchema->safeParse($v, $ctx->child($key));
            if (!$result->success) {
                $issues = array_merge($issues, $result->issues);
            } else {
                $parsed[$key] = $result->value;
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($parsed);
    }

    public function getValueSchema(): Schema
    {
        return $this->valueSchema;
    }

    public function exportNode(): array
    {
        return ['kind' => 'record', 'values' => $this->valueSchema->exportNode()];
    }
}

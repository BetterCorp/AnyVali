<?php

declare(strict_types=1);

namespace AnyVali\Schemas;

use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\UnknownKeyMode;
use AnyVali\ValidationContext;
use AnyVali\ValidationIssue;

/**
 * @extends Schema<array<string, mixed>>
 */
final class ObjectSchema extends Schema
{
    /** @var array<string, Schema> */
    private readonly array $properties;
    /** @var string[] */
    private readonly array $required;
    private UnknownKeyMode $unknownKeys;

    /**
     * @param array<string, Schema> $properties
     * @param string[] $required
     */
    public function __construct(
        array $properties,
        array $required = [],
        UnknownKeyMode $unknownKeys = UnknownKeyMode::Reject,
    ) {
        $this->properties = $properties;
        $this->required = $required;
        $this->unknownKeys = $unknownKeys;
    }

    public function getKind(): string
    {
        return 'object';
    }

    public function strict(): self
    {
        return $this->unknownKeys(UnknownKeyMode::Reject);
    }

    public function strip(): self
    {
        return $this->unknownKeys(UnknownKeyMode::Strip);
    }

    public function passthrough(): self
    {
        return $this->unknownKeys(UnknownKeyMode::Allow);
    }

    public function unknownKeys(UnknownKeyMode $mode): self
    {
        $clone = clone $this;
        $clone->unknownKeys = $mode;
        return $clone;
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
                message: 'Expected object',
                path: $ctx->path,
                expected: 'object',
                received: self::getTypeName($value),
            )]);
        }

        $issues = [];
        $parsed = [];

        // Check required fields
        foreach ($this->required as $key) {
            if (!array_key_exists($key, $value)) {
                $expectedKind = isset($this->properties[$key])
                    ? $this->properties[$key]->getKind()
                    : 'unknown';
                $issues[] = new ValidationIssue(
                    code: IssueCodes::REQUIRED,
                    message: "Required field \"{$key}\" is missing",
                    path: array_merge($ctx->path, [$key]),
                    expected: $expectedKind,
                    received: 'undefined',
                );
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        // Validate known properties
        foreach ($this->properties as $key => $schema) {
            if (array_key_exists($key, $value)) {
                // Apply defaults for nested schemas that have defaults
                $fieldValue = $value[$key];

                $result = $schema->safeParse($fieldValue, $ctx->child($key));
                if (!$result->success) {
                    $issues = array_merge($issues, $result->issues);
                } else {
                    $parsed[$key] = $result->value;
                }
            } elseif ($schema->hasDefaultValue()) {
                // Apply default
                $defaultVal = $schema->getDefaultValue();
                // Validate the default value
                $defResult = $schema->safeParse($defaultVal, $ctx->child($key));
                if (!$defResult->success) {
                    // Default is invalid
                    $issues[] = new ValidationIssue(
                        code: IssueCodes::DEFAULT_INVALID,
                        message: "Default value for \"{$key}\" is invalid",
                        path: array_merge($ctx->path, [$key]),
                        expected: $defResult->issues[0]->expected ?? (string)$schema->getKind(),
                        received: self::formatValue($defaultVal),
                    );
                } else {
                    $parsed[$key] = $defResult->value;
                }
            } elseif ($schema instanceof OptionalSchema) {
                // Optional field that is absent -- skip
            }
        }

        // Handle unknown keys
        $knownKeys = array_keys($this->properties);
        foreach ($value as $key => $v) {
            if (!in_array($key, $knownKeys, true)) {
                switch ($this->unknownKeys) {
                    case UnknownKeyMode::Reject:
                        $issues[] = new ValidationIssue(
                            code: IssueCodes::UNKNOWN_KEY,
                            message: "Unknown key \"{$key}\"",
                            path: array_merge($ctx->path, [$key]),
                            expected: 'undefined',
                            received: (string)$key,
                        );
                        break;
                    case UnknownKeyMode::Allow:
                        $parsed[$key] = $v;
                        break;
                    case UnknownKeyMode::Strip:
                        // Silently drop
                        break;
                }
            }
        }

        if (!empty($issues)) {
            return ParseResult::fail($issues);
        }

        return ParseResult::ok($parsed);
    }

    /** @return array<string, Schema> */
    public function getProperties(): array
    {
        return $this->properties;
    }

    /** @return string[] */
    public function getRequired(): array
    {
        return $this->required;
    }

    public function getUnknownKeys(): UnknownKeyMode
    {
        return $this->unknownKeys;
    }

    public function exportNode(): array
    {
        $props = [];
        foreach ($this->properties as $key => $schema) {
            $props[$key] = $schema->exportNode();
        }
        $node = [
            'kind' => 'object',
            'properties' => $props,
            'required' => $this->required,
            'unknownKeys' => $this->unknownKeys->value,
        ];
        if ($this->hasDefault) $node['default'] = $this->defaultValue;
        return $node;
    }
}

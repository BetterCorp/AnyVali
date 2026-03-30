<?php

declare(strict_types=1);

namespace AnyVali;

use AnyVali\Interchange\Exporter;
use AnyVali\Parse\Coercion;

/**
 * @template TOutput
 */
abstract class Schema
{
    protected mixed $defaultValue = null;
    protected bool $hasDefault = false;
    /** @var string|string[]|null */
    protected string|array|null $coerce = null;
    /** @var callable[] */
    protected array $customValidators = [];

    abstract public function getKind(): string;

    /**
     * Validate and return parsed value or issues.
     * @param ValidationContext $ctx
     * @return ParseResult
     */
    abstract protected function validateValue(mixed $value, ValidationContext $ctx): ParseResult;

    /**
     * Export this schema node to array form.
     * @return array<string, mixed>
     */
    abstract public function exportNode(): array;

    /**
     * Throwing parse.
     *
     * @return TOutput
     */
    public function parse(mixed $input): mixed
    {
        $result = $this->safeParse($input);
        if (!$result->success) {
            throw new ValidationError($result->issues);
        }
        return $result->value;
    }

    /**
     * Non-throwing parse.
     *
     * @return ParseResult<TOutput>
     */
    public function safeParse(mixed $input, ?ValidationContext $ctx = null): ParseResult
    {
        $ctx ??= new ValidationContext();
        $value = $input;

        // Step 1: Coercion (only if value is present)
        if ($this->coerce !== null) {
            $coercions = is_array($this->coerce) ? $this->coerce : [$this->coerce];
            foreach ($coercions as $c) {
                [$value, $issue] = Coercion::apply($c, $value, $this->getKind(), $ctx->path);
                if ($issue !== null) {
                    return ParseResult::fail([$issue]);
                }
            }
        }

        // Step 2: Validate
        $result = $this->validateValue($value, $ctx);

        // Step 3: Custom validators (local-only)
        if ($result->success && !empty($this->customValidators)) {
            $issues = [];
            foreach ($this->customValidators as $validator) {
                $validatorResult = $validator($result->value, $ctx);
                if ($validatorResult instanceof ValidationIssue) {
                    $issues[] = $validatorResult;
                }
            }
            if (!empty($issues)) {
                return ParseResult::fail($issues);
            }
        }

        return $result;
    }

    /**
     * Export as AnyValiDocument.
     */
    public function export(ExportMode $mode = ExportMode::Portable): AnyValiDocument
    {
        if ($mode === ExportMode::Portable && !empty($this->customValidators)) {
            throw new \RuntimeException(
                'Cannot export schema with custom validators in portable mode'
            );
        }
        return Exporter::export($this, $mode);
    }

    /**
     * Set default value. Returns cloned instance.
     */
    public function default(mixed $value): static
    {
        $clone = clone $this;
        $clone->defaultValue = $value;
        $clone->hasDefault = true;
        return $clone;
    }

    public function hasDefaultValue(): bool
    {
        return $this->hasDefault;
    }

    public function getDefaultValue(): mixed
    {
        return $this->defaultValue;
    }

    /**
     * Set coercion. Returns cloned instance.
     * @param string|string[] $coerce
     */
    public function coerce(string|array $coerce): static
    {
        $clone = clone $this;
        $clone->coerce = $coerce;
        return $clone;
    }

    public function getCoerce(): string|array|null
    {
        return $this->coerce;
    }

    /**
     * Add a custom validator (local-only, non-portable).
     */
    public function refine(callable $validator): static
    {
        $clone = clone $this;
        $clone->customValidators[] = $validator;
        return $clone;
    }

    public function hasCustomValidators(): bool
    {
        return !empty($this->customValidators);
    }

    /**
     * Get the PHP type name for a value, using AnyVali naming conventions.
     */
    protected static function getTypeName(mixed $value): string
    {
        if ($value === null) {
            return 'null';
        }
        if (is_bool($value)) {
            return 'boolean';
        }
        if (is_int($value)) {
            return 'number';
        }
        if (is_float($value)) {
            return 'number';
        }
        if (is_string($value)) {
            return 'string';
        }
        if (is_array($value)) {
            // Check if sequential (array) or associative (object)
            if (array_is_list($value)) {
                return 'array';
            }
            return 'object';
        }
        if (is_object($value)) {
            return 'object';
        }
        return gettype($value);
    }

    /**
     * Format a value for display in issue messages.
     */
    protected static function formatValue(mixed $value): string
    {
        if ($value === null) {
            return 'null';
        }
        if (is_bool($value)) {
            return $value ? 'true' : 'false';
        }
        if (is_int($value) || is_float($value)) {
            // For integers that are float due to JSON decode, show without decimal
            if (is_float($value) && floor($value) === $value && !is_infinite($value)) {
                return (string)(int)$value;
            }
            return (string)$value;
        }
        if (is_string($value)) {
            return $value;
        }
        return json_encode($value) ?: 'unknown';
    }
}

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
    private const RESERVED_METADATA_KEYS = [
        'title', 'description', 'deprecated', 'deprecatedMessage',
        'notStable', 'since', 'sensitive', 'readonly', 'writeonly', 'examples',
    ];

    protected mixed $defaultValue = null;
    protected bool $hasDefault = false;
    /** @var string|string[]|null */
    protected string|array|null $coerce = null;
    /** @var callable[] */
    protected array $customValidators = [];
    /** @var array<string, mixed> */
    protected array $metadata = [];

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
     * Add documentation metadata. Returns cloned instance.
     * @param array{title?: string, deprecated?: bool, deprecatedMessage?: string, notStable?: bool, since?: string, sensitive?: bool, readonly?: bool, writeonly?: bool, examples?: array} $opts
     */
    public function describe(string $description, array $opts = []): static
    {
        $clone = clone $this;
        $clone->metadata = array_merge($clone->metadata, ['description' => $description]);

        if (isset($opts['title'])) {
            if (!is_string($opts['title'])) {
                throw new \InvalidArgumentException('describe(): title must be a string');
            }
            $clone->metadata['title'] = $opts['title'];
        }
        if (isset($opts['deprecated'])) {
            if (!is_bool($opts['deprecated'])) {
                throw new \InvalidArgumentException('describe(): deprecated must be a boolean');
            }
            $clone->metadata['deprecated'] = $opts['deprecated'];
        }
        if (isset($opts['deprecatedMessage'])) {
            if (!is_string($opts['deprecatedMessage'])) {
                throw new \InvalidArgumentException('describe(): deprecatedMessage must be a string');
            }
            if (empty($opts['deprecated'])) {
                throw new \InvalidArgumentException('describe(): deprecatedMessage requires deprecated to be true');
            }
            $clone->metadata['deprecatedMessage'] = $opts['deprecatedMessage'];
        }
        if (isset($opts['notStable'])) {
            if (!is_bool($opts['notStable'])) {
                throw new \InvalidArgumentException('describe(): notStable must be a boolean');
            }
            $clone->metadata['notStable'] = $opts['notStable'];
        }
        if (isset($opts['since'])) {
            if (!is_string($opts['since'])) {
                throw new \InvalidArgumentException('describe(): since must be a string');
            }
            $clone->metadata['since'] = $opts['since'];
        }
        if (isset($opts['sensitive'])) {
            if (!is_bool($opts['sensitive'])) {
                throw new \InvalidArgumentException('describe(): sensitive must be a boolean');
            }
            $clone->metadata['sensitive'] = $opts['sensitive'];
        }
        if (isset($opts['readonly'])) {
            if (!is_bool($opts['readonly'])) {
                throw new \InvalidArgumentException('describe(): readonly must be a boolean');
            }
            $clone->metadata['readonly'] = $opts['readonly'];
        }
        if (isset($opts['writeonly'])) {
            if (!is_bool($opts['writeonly'])) {
                throw new \InvalidArgumentException('describe(): writeonly must be a boolean');
            }
            $clone->metadata['writeonly'] = $opts['writeonly'];
        }
        if (!empty($opts['readonly']) && !empty($opts['writeonly'])) {
            throw new \InvalidArgumentException('describe(): readonly and writeonly cannot both be true');
        }
        if (isset($opts['examples'])) {
            if (!is_array($opts['examples'])) {
                throw new \InvalidArgumentException('describe(): examples must be an array');
            }
            $clone->metadata['examples'] = $opts['examples'];
        }

        return $clone;
    }

    /**
     * Attach arbitrary metadata. Reserved keys must use describe().
     * @param array<string, mixed> $meta
     */
    public function metadata(array $meta, bool $replace = false): static
    {
        foreach (array_keys($meta) as $key) {
            if (in_array($key, self::RESERVED_METADATA_KEYS, true)) {
                throw new \InvalidArgumentException(
                    "metadata(): \"{$key}\" is a reserved key. Use describe() instead."
                );
            }
        }

        $clone = clone $this;
        if ($replace) {
            $preserved = [];
            foreach ($clone->metadata as $k => $v) {
                if (in_array($k, self::RESERVED_METADATA_KEYS, true)) {
                    $preserved[$k] = $v;
                }
            }
            $clone->metadata = array_merge($preserved, $meta);
        } else {
            $clone->metadata = array_merge($clone->metadata, $meta);
        }
        return $clone;
    }

    /**
     * @return array<string, mixed>
     */
    public function getMetadata(): array
    {
        return $this->metadata;
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
     * Merge metadata into an export node array.
     */
    protected function addMetadataToNode(array &$node): void
    {
        if (!empty($this->metadata)) {
            $node['metadata'] = $this->metadata;
        }
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

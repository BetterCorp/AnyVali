<?php

declare(strict_types=1);

namespace AnyVali\Interchange;

use AnyVali\AnyValiDocument;
use AnyVali\Schema;
use AnyVali\UnknownKeyMode;
use AnyVali\Schemas\{
    AnySchema,
    ArraySchema,
    BoolSchema,
    EnumSchema,
    IntSchema,
    IntersectionSchema,
    LiteralSchema,
    NeverSchema,
    NullSchema,
    NullableSchema,
    NumberSchema,
    ObjectSchema,
    OptionalSchema,
    RecordSchema,
    RefSchema,
    StringSchema,
    TupleSchema,
    UnionSchema,
    UnknownSchema,
};

final class Importer
{
    /** @var array<string, Schema> */
    private array $resolved = [];
    /** @var array<string, bool> */
    private array $active = [];
    /** @var array<string, list<RefSchema>> */
    private array $pending = [];

    private function __construct()
    {
    }

    /**
     * Import a schema from an AnyValiDocument.
     */
    public static function import(AnyValiDocument $doc): Schema
    {
        return self::importNode($doc->root, $doc->definitions);
    }

    /**
     * Import from a JSON string.
     */
    public static function importJson(string $json): Schema
    {
        return self::import(AnyValiDocument::fromJson($json));
    }

    /**
     * Import from array.
     * @param array<string, mixed> $data
     */
    public static function importArray(array $data): Schema
    {
        return self::import(AnyValiDocument::fromArray($data));
    }

    /**
     * Import a single schema node.
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    public static function importNode(array $node, array $definitions = []): Schema
    {
        return (new self())->node($node, $definitions);
    }

    private function node(array $node, array $definitions): Schema
    {
        $kind = $node['kind'] ?? null;

        if ($kind === null) {
            throw new \RuntimeException('Schema node missing "kind" field');
        }

        $schema = match ($kind) {
            'string' => $this->importString($node),
            'number', 'float64', 'float32' => $this->importNumber($node, $kind),
            'int', 'int8', 'int16', 'int32', 'int64',
            'uint8', 'uint16', 'uint32', 'uint64' => $this->importInt($node, $kind),
            'bool' => $this->importBool($node),
            'null' => new NullSchema(),
            'any' => new AnySchema(),
            'unknown' => new UnknownSchema(),
            'never' => new NeverSchema(),
            'literal' => new LiteralSchema($node['value'] ?? null),
            'enum' => new EnumSchema($node['values'] ?? []),
            'array' => $this->importArray2($node, $definitions),
            'tuple' => $this->importTuple($node, $definitions),
            'object' => $this->importObject($node, $definitions),
            'record' => $this->importRecord($node, $definitions),
            'union' => $this->importUnion($node, $definitions),
            'intersection' => $this->importIntersection($node, $definitions),
            'optional' => $this->importOptional($node, $definitions),
            'nullable' => $this->importNullable($node, $definitions),
            'ref' => new RefSchema($node['ref'] ?? ''),
            default => throw new \RuntimeException("Unsupported schema kind: {$kind}"),
        };

        // Apply coerce
        if (isset($node['coerce'])) {
            $schema = $schema->coerce($node['coerce']);
        }

        // Apply default
        if (array_key_exists('default', $node)) {
            $schema = $schema->default($node['default']);
        }

        if ($schema instanceof RefSchema) {
            $this->resolveRef($schema, $definitions);
        }
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     */
    private function importString(array $node): StringSchema
    {
        $schema = new StringSchema();
        if (isset($node['minLength'])) $schema = $schema->minLength((int)$node['minLength']);
        if (isset($node['maxLength'])) $schema = $schema->maxLength((int)$node['maxLength']);
        if (isset($node['pattern'])) $schema = $schema->pattern($node['pattern']);
        if (isset($node['startsWith'])) $schema = $schema->startsWith($node['startsWith']);
        if (isset($node['endsWith'])) $schema = $schema->endsWith($node['endsWith']);
        if (isset($node['includes'])) $schema = $schema->includes($node['includes']);
        if (isset($node['format'])) $schema = $schema->format($node['format']);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     */
    private function importNumber(array $node, string $kind): NumberSchema
    {
        $schema = new NumberSchema($kind);
        if (isset($node['min'])) $schema = $schema->min($node['min']);
        if (isset($node['max'])) $schema = $schema->max($node['max']);
        if (isset($node['exclusiveMin'])) $schema = $schema->exclusiveMin($node['exclusiveMin']);
        if (isset($node['exclusiveMax'])) $schema = $schema->exclusiveMax($node['exclusiveMax']);
        if (isset($node['multipleOf'])) $schema = $schema->multipleOf($node['multipleOf']);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     */
    private function importInt(array $node, string $kind): IntSchema
    {
        $schema = new IntSchema($kind);
        if (isset($node['min'])) $schema = $schema->min($node['min']);
        if (isset($node['max'])) $schema = $schema->max($node['max']);
        if (isset($node['exclusiveMin'])) $schema = $schema->exclusiveMin($node['exclusiveMin']);
        if (isset($node['exclusiveMax'])) $schema = $schema->exclusiveMax($node['exclusiveMax']);
        if (isset($node['multipleOf'])) $schema = $schema->multipleOf($node['multipleOf']);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     */
    private function importBool(array $node): BoolSchema
    {
        return new BoolSchema();
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importArray2(array $node, array $definitions): ArraySchema
    {
        // Accept "items", "item", and "array.items" keys for compatibility
        $itemsNode = $node['items'] ?? $node['item'] ?? $node['array.items'] ?? ['kind' => 'any'];
        $items = $this->node($itemsNode, $definitions);
        $schema = new ArraySchema($items);
        if (isset($node['minItems'])) $schema = $schema->minItems((int)$node['minItems']);
        if (isset($node['maxItems'])) $schema = $schema->maxItems((int)$node['maxItems']);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importTuple(array $node, array $definitions): TupleSchema
    {
        $elements = array_map(
            fn(array $el) => $this->node($el, $definitions),
            $node['elements'] ?? [],
        );
        return new TupleSchema($elements);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importObject(array $node, array $definitions): ObjectSchema
    {
        $properties = [];
        foreach (($node['properties'] ?? []) as $key => $propNode) {
            $properties[$key] = $this->node($propNode, $definitions);
        }

        $required = $node['required'] ?? [];
        $unknownKeys = UnknownKeyMode::tryFrom($node['unknownKeys'] ?? 'strip')
            ?? UnknownKeyMode::Strip;

        return new ObjectSchema($properties, $required, $unknownKeys, true);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importRecord(array $node, array $definitions): RecordSchema
    {
        $valueNode = array_key_exists('values', $node) ? $node['values'] : ($node['valueSchema'] ?? null);
        if (!is_array($valueNode)) throw new \RuntimeException('Record schema missing or invalid "values"');
        $values = $this->node($valueNode, $definitions);
        return new RecordSchema($values);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importUnion(array $node, array $definitions): UnionSchema
    {
        // Accept "variants", "schemas", and "union.variants" keys for compatibility
        $variantsData = $node['variants'] ?? $node['schemas'] ?? $node['union.variants'] ?? [];
        $variants = array_map(
            fn(array $v) => $this->node($v, $definitions),
            $variantsData,
        );
        return new UnionSchema($variants);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importIntersection(array $node, array $definitions): IntersectionSchema
    {
        $allOf = array_map(
            fn(array $s) => $this->node($s, $definitions),
            $node['allOf'] ?? [],
        );
        return new IntersectionSchema($allOf);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importOptional(array $node, array $definitions): OptionalSchema
    {
        $inner = $this->node($node['schema'] ?? ['kind' => 'any'], $definitions);
        $schema = new OptionalSchema($inner);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private function importNullable(array $node, array $definitions): NullableSchema
    {
        $inner = $this->node($node['schema'] ?? ['kind' => 'any'], $definitions);
        $schema = new NullableSchema($inner);
        return $schema;
    }

    /**
     * @param array<string, array<string, mixed>> $definitions
     */
    private function resolveRef(RefSchema $schema, array $definitions): void
    {
        $ref = $schema->getRef();
        $prefix = '#/definitions/';
        if (!str_starts_with($ref, $prefix)) return;
        $name = substr($ref, strlen($prefix));
        if (!isset($definitions[$name])) return;
        if (isset($this->resolved[$name])) {
            $schema->resolve($this->resolved[$name]);
            return;
        }
        if (isset($this->active[$name])) {
            $this->pending[$name][] = $schema;
            return;
        }
        $this->active[$name] = true;
        $this->pending[$name] = [$schema];
        try {
            $target = $this->node($definitions[$name], $definitions);
            $this->resolved[$name] = $target;
            foreach ($this->pending[$name] as $pending) {
                $pending->resolve($target);
            }
        } finally {
            unset($this->active[$name], $this->pending[$name]);
        }
    }
}

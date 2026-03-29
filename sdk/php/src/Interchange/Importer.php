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
        $kind = $node['kind'] ?? null;

        if ($kind === null) {
            throw new \RuntimeException('Schema node missing "kind" field');
        }

        $schema = match ($kind) {
            'string' => self::importString($node),
            'number', 'float64', 'float32' => self::importNumber($node, $kind),
            'int', 'int8', 'int16', 'int32', 'int64',
            'uint8', 'uint16', 'uint32', 'uint64' => self::importInt($node, $kind),
            'bool' => self::importBool($node),
            'null' => new NullSchema(),
            'any' => new AnySchema(),
            'unknown' => new UnknownSchema(),
            'never' => new NeverSchema(),
            'literal' => new LiteralSchema($node['value'] ?? null),
            'enum' => new EnumSchema($node['values'] ?? []),
            'array' => self::importArray2($node, $definitions),
            'tuple' => self::importTuple($node, $definitions),
            'object' => self::importObject($node, $definitions),
            'record' => self::importRecord($node, $definitions),
            'union' => self::importUnion($node, $definitions),
            'intersection' => self::importIntersection($node, $definitions),
            'optional' => self::importOptional($node, $definitions),
            'nullable' => self::importNullable($node, $definitions),
            'ref' => self::importRef($node, $definitions),
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

        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     */
    private static function importString(array $node): StringSchema
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
    private static function importNumber(array $node, string $kind): NumberSchema
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
    private static function importInt(array $node, string $kind): IntSchema
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
    private static function importBool(array $node): BoolSchema
    {
        return new BoolSchema();
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importArray2(array $node, array $definitions): ArraySchema
    {
        $items = self::importNode($node['items'] ?? ['kind' => 'any'], $definitions);
        $schema = new ArraySchema($items);
        if (isset($node['minItems'])) $schema = $schema->minItems((int)$node['minItems']);
        if (isset($node['maxItems'])) $schema = $schema->maxItems((int)$node['maxItems']);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importTuple(array $node, array $definitions): TupleSchema
    {
        $elements = array_map(
            fn(array $el) => self::importNode($el, $definitions),
            $node['elements'] ?? [],
        );
        return new TupleSchema($elements);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importObject(array $node, array $definitions): ObjectSchema
    {
        $properties = [];
        foreach (($node['properties'] ?? []) as $key => $propNode) {
            $properties[$key] = self::importNode($propNode, $definitions);
        }

        $required = $node['required'] ?? [];
        $unknownKeys = UnknownKeyMode::tryFrom($node['unknownKeys'] ?? 'reject')
            ?? UnknownKeyMode::Reject;

        return new ObjectSchema($properties, $required, $unknownKeys);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importRecord(array $node, array $definitions): RecordSchema
    {
        $values = self::importNode($node['values'] ?? ['kind' => 'any'], $definitions);
        return new RecordSchema($values);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importUnion(array $node, array $definitions): UnionSchema
    {
        $variants = array_map(
            fn(array $v) => self::importNode($v, $definitions),
            $node['variants'] ?? [],
        );
        return new UnionSchema($variants);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importIntersection(array $node, array $definitions): IntersectionSchema
    {
        $allOf = array_map(
            fn(array $s) => self::importNode($s, $definitions),
            $node['allOf'] ?? [],
        );
        return new IntersectionSchema($allOf);
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importOptional(array $node, array $definitions): OptionalSchema
    {
        $inner = self::importNode($node['schema'] ?? ['kind' => 'any'], $definitions);
        $schema = new OptionalSchema($inner);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importNullable(array $node, array $definitions): NullableSchema
    {
        $inner = self::importNode($node['schema'] ?? ['kind' => 'any'], $definitions);
        $schema = new NullableSchema($inner);
        return $schema;
    }

    /**
     * @param array<string, mixed> $node
     * @param array<string, array<string, mixed>> $definitions
     */
    private static function importRef(array $node, array $definitions): RefSchema
    {
        $ref = $node['ref'] ?? '';
        $schema = new RefSchema($ref);

        // Try to resolve from definitions (but keep as ref for recursion support)
        return $schema;
    }
}

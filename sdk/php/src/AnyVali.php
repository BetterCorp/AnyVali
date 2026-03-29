<?php

declare(strict_types=1);

namespace AnyVali;

use AnyVali\Interchange\Importer;
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

final class AnyVali
{
    private function __construct()
    {
    }

    public static function string(): StringSchema
    {
        return new StringSchema();
    }

    public static function number(): NumberSchema
    {
        return new NumberSchema('number');
    }

    public static function int(): IntSchema
    {
        return new IntSchema('int');
    }

    public static function int8(): IntSchema
    {
        return new IntSchema('int8');
    }

    public static function int16(): IntSchema
    {
        return new IntSchema('int16');
    }

    public static function int32(): IntSchema
    {
        return new IntSchema('int32');
    }

    public static function int64(): IntSchema
    {
        return new IntSchema('int64');
    }

    public static function uint8(): IntSchema
    {
        return new IntSchema('uint8');
    }

    public static function uint16(): IntSchema
    {
        return new IntSchema('uint16');
    }

    public static function uint32(): IntSchema
    {
        return new IntSchema('uint32');
    }

    public static function uint64(): IntSchema
    {
        return new IntSchema('uint64');
    }

    public static function float32(): NumberSchema
    {
        return new NumberSchema('float32');
    }

    public static function float64(): NumberSchema
    {
        return new NumberSchema('float64');
    }

    public static function bool(): BoolSchema
    {
        return new BoolSchema();
    }

    public static function null(): NullSchema
    {
        return new NullSchema();
    }

    public static function any(): AnySchema
    {
        return new AnySchema();
    }

    public static function unknown(): UnknownSchema
    {
        return new UnknownSchema();
    }

    public static function never(): NeverSchema
    {
        return new NeverSchema();
    }

    public static function literal(mixed $value): LiteralSchema
    {
        return new LiteralSchema($value);
    }

    /**
     * @param array<mixed> $values
     */
    public static function enum(array $values): EnumSchema
    {
        return new EnumSchema($values);
    }

    public static function array(Schema $items): ArraySchema
    {
        return new ArraySchema($items);
    }

    /**
     * @param Schema[] $elements
     */
    public static function tuple(array $elements): TupleSchema
    {
        return new TupleSchema($elements);
    }

    /**
     * @param array<string, Schema> $properties
     * @param string[] $required
     */
    public static function object(
        array $properties,
        array $required = [],
        UnknownKeyMode $unknownKeys = UnknownKeyMode::Reject,
    ): ObjectSchema {
        return new ObjectSchema($properties, $required, $unknownKeys);
    }

    public static function record(Schema $valueSchema): RecordSchema
    {
        return new RecordSchema($valueSchema);
    }

    /**
     * @param Schema[] $variants
     */
    public static function union(array $variants): UnionSchema
    {
        return new UnionSchema($variants);
    }

    /**
     * @param Schema[] $schemas
     */
    public static function intersection(array $schemas): IntersectionSchema
    {
        return new IntersectionSchema($schemas);
    }

    public static function optional(Schema $schema): OptionalSchema
    {
        return new OptionalSchema($schema);
    }

    public static function nullable(Schema $schema): NullableSchema
    {
        return new NullableSchema($schema);
    }

    public static function ref(string $ref): RefSchema
    {
        return new RefSchema($ref);
    }

    /**
     * Import a schema from a portable document.
     */
    public static function import(AnyValiDocument|string|array $source): Schema
    {
        if ($source instanceof AnyValiDocument) {
            return Importer::import($source);
        }
        if (is_string($source)) {
            return Importer::importJson($source);
        }
        return Importer::importArray($source);
    }
}

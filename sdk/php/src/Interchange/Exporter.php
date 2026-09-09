<?php

declare(strict_types=1);

namespace AnyVali\Interchange;

use AnyVali\AnyValiDocument;
use AnyVali\ExportMode;
use AnyVali\Schema;
use AnyVali\Schemas\{ArraySchema, ObjectSchema, RecordSchema, TupleSchema, UnionSchema,
    IntersectionSchema, OptionalSchema, NullableSchema, RefSchema};

final class Exporter
{
    private function __construct()
    {
    }

    public static function export(Schema $schema, ExportMode $mode = ExportMode::Portable): AnyValiDocument
    {
        if ($mode === ExportMode::Portable && $schema->hasCustomValidators()) {
            throw new \RuntimeException(
                'Cannot export schema with custom validators in portable mode'
            );
        }

        $definitions = [];
        self::collectDefinitions($schema, $definitions, new \SplObjectStorage(), $mode);
        return new AnyValiDocument(
            root: $schema->exportNode(),
            definitions: $definitions,
        );
    }

    private static function collectDefinitions(Schema $schema, array &$definitions, \SplObjectStorage $seen, ExportMode $mode): void
    {
        if ($seen->contains($schema)) return;
        $seen->attach($schema);
        if ($mode === ExportMode::Portable && $schema->hasCustomValidators()) {
            throw new \RuntimeException('Cannot export schema with custom validators in portable mode');
        }
        $children = match (true) {
            $schema instanceof ObjectSchema => array_values($schema->getProperties()),
            $schema instanceof ArraySchema => [$schema->getItems()],
            $schema instanceof RecordSchema => [$schema->getValueSchema()],
            $schema instanceof TupleSchema => $schema->getElements(),
            $schema instanceof UnionSchema => $schema->getVariants(),
            $schema instanceof IntersectionSchema => $schema->getAllOf(),
            $schema instanceof OptionalSchema, $schema instanceof NullableSchema => [$schema->getInnerSchema()],
            default => [],
        };
        if ($schema instanceof RefSchema && $schema->getResolvedSchema() !== null) {
            $target = $schema->getResolvedSchema();
            $prefix = '#/definitions/';
            if (str_starts_with($schema->getRef(), $prefix)) {
                $name = substr($schema->getRef(), strlen($prefix));
                $node = $target->exportNode();
                if (isset($definitions[$name]) && self::definitionKey($definitions[$name]) !== self::definitionKey($node)) {
                    throw new \RuntimeException("Conflicting definition: {$name}");
                }
                $definitions[$name] = $node;
            }
            $children[] = $target;
        }
        foreach ($children as $child) self::collectDefinitions($child, $definitions, $seen, $mode);
    }


    private static function definitionKey(array $node): string
    {
        return json_encode(self::canonicalize($node), JSON_THROW_ON_ERROR | JSON_PRESERVE_ZERO_FRACTION);
    }

    private static function canonicalize(mixed $value): mixed
    {
        // Normalize integral floats only where an exact native integer exists.
        // Avoid PHP's loose int/float comparison, which rounds large integers.
        if (is_float($value) && is_finite($value) && floor($value) === $value
            && $value >= PHP_INT_MIN && $value < -(float)PHP_INT_MIN) {
            return (int)$value;
        }
        if ($value instanceof \stdClass) {
            $fields = get_object_vars($value);
            ksort($fields);
            return (object)array_map(self::canonicalize(...), $fields);
        }
        if (!is_array($value)) return $value;
        if (!array_is_list($value)) ksort($value);
        return array_map(self::canonicalize(...), $value);
    }

}

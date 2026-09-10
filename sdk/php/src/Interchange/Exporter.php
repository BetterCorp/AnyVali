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
    private const NUMERIC_KINDS = ['number', 'float32', 'float64', 'int', 'int8', 'int16', 'int32', 'int64', 'uint8', 'uint16', 'uint32', 'uint64'];
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

    private static function canonicalize(mixed $value, bool $normalizeNumber = false): mixed
    {
        // Normalize integral floats only where an exact native integer exists.
        // Avoid PHP's loose int/float comparison, which rounds large integers.
        if ($normalizeNumber && is_float($value) && is_finite($value) && floor($value) === $value
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
        $kind = $value['kind'] ?? null;
        $numeric = in_array($kind, self::NUMERIC_KINDS, true);
        foreach ($value as $key => $child) {
            // Literal/enum payloads and arbitrary defaults use strict PHP types.
            // Do not interpret schema-looking data inside them as schema nodes.
            if (($key === 'value' && $kind === 'literal') || ($key === 'values' && $kind === 'enum') || ($key === 'default' && is_string($kind))) {
                if ($key === 'default' && $numeric && is_scalar($child)) {
                    $value[$key] = self::canonicalize($child, true);
                } elseif ($key === 'default') {
                    $value[$key] = self::canonicalizeDefault($child, $value);
                }
                continue;
            }
            $value[$key] = self::canonicalize($child, $numeric && in_array($key, ['min', 'max', 'exclusiveMin', 'exclusiveMax', 'multipleOf'], true));
        }
        return $value;
    }

    /** Sort unordered object defaults without changing type-sensitive child payloads. */
    private static function canonicalizeDefault(mixed $value, array $node): mixed
    {
        $kind = $node['kind'] ?? null;
        if ($kind === 'optional' || $kind === 'nullable') {
            return self::canonicalizeDefault($value, $node['schema']);
        }
        if (is_scalar($value) && in_array($kind, self::NUMERIC_KINDS, true)) {
            return self::canonicalize($value, true);
        }
        if (!is_array($value) && !$value instanceof \stdClass) return $value;
        $object = $value instanceof \stdClass;
        $fields = $object ? get_object_vars($value) : $value;
        // Literal/enum matching is strict; preserve their key order and scalar types.
        // References and alternatives need their resolved runtime type to compare safely.
        if (!in_array($kind, ['object', 'record', 'array', 'tuple', 'any', 'unknown'], true)) return $value;
        if (!array_is_list($fields)) ksort($fields);
        foreach ($fields as $key => $child) {
            $childNode = match ($kind) {
                'object' => $node['properties'][$key] ?? ['kind' => 'any'],
                'record' => $node['values'],
                'array' => $node['items'],
                'tuple' => $node['elements'][$key] ?? ['kind' => 'any'],
                default => ['kind' => 'any'],
            };
            $fields[$key] = self::canonicalizeDefault($child, $childNode);
        }
        return $object ? (object)$fields : $fields;
    }

}

<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\AnyValiDocument;
use AnyVali\ExportMode;
use AnyVali\IssueCodes;
use AnyVali\Interchange\Exporter;
use AnyVali\Interchange\Importer;
use AnyVali\UnknownKeyMode;
use AnyVali\ValidationIssue;
use PHPUnit\Framework\TestCase;

final class InterchangeTest extends TestCase
{
    // ── Export ──────────────────────────────────────

    public function testExportStringSchema(): void
    {
        $s = AnyVali::string()->minLength(1)->maxLength(100);
        $doc = $s->export();

        $this->assertSame('1.0', $doc->anyvaliVersion);
        $this->assertSame('1', $doc->schemaVersion);
        $this->assertSame('string', $doc->root['kind']);
        $this->assertSame(1, $doc->root['minLength']);
        $this->assertSame(100, $doc->root['maxLength']);
    }

    public function testExportNumberSchema(): void
    {
        $s = AnyVali::number()->min(0)->max(100);
        $doc = $s->export();

        $this->assertSame('number', $doc->root['kind']);
        $this->assertSame(0.0, $doc->root['min']);
        $this->assertSame(100.0, $doc->root['max']);
    }

    public function testExportIntSchema(): void
    {
        $s = AnyVali::int32();
        $doc = $s->export();
        $this->assertSame('int32', $doc->root['kind']);
    }

    public function testExportBoolSchema(): void
    {
        $doc = AnyVali::bool()->export();
        $this->assertSame('bool', $doc->root['kind']);
    }

    public function testExportNullSchema(): void
    {
        $doc = AnyVali::null()->export();
        $this->assertSame('null', $doc->root['kind']);
    }

    public function testExportAnySchema(): void
    {
        $doc = AnyVali::any()->export();
        $this->assertSame('any', $doc->root['kind']);
    }

    public function testExportUnknownSchema(): void
    {
        $doc = AnyVali::unknown()->export();
        $this->assertSame('unknown', $doc->root['kind']);
    }

    public function testExportNeverSchema(): void
    {
        $doc = AnyVali::never()->export();
        $this->assertSame('never', $doc->root['kind']);
    }

    public function testExportLiteralSchema(): void
    {
        $doc = AnyVali::literal('hello')->export();
        $this->assertSame('literal', $doc->root['kind']);
        $this->assertSame('hello', $doc->root['value']);
    }

    public function testExportEnumSchema(): void
    {
        $doc = AnyVali::enum(['a', 'b'])->export();
        $this->assertSame('enum', $doc->root['kind']);
        $this->assertSame(['a', 'b'], $doc->root['values']);
    }

    public function testExportArraySchema(): void
    {
        $doc = AnyVali::array(AnyVali::string())->minItems(1)->export();
        $this->assertSame('array', $doc->root['kind']);
        $this->assertSame('string', $doc->root['items']['kind']);
        $this->assertSame(1, $doc->root['minItems']);
    }

    public function testExportTupleSchema(): void
    {
        $doc = AnyVali::tuple([AnyVali::string(), AnyVali::int()])->export();
        $this->assertSame('tuple', $doc->root['kind']);
        $this->assertCount(2, $doc->root['elements']);
    }

    public function testExportObjectSchema(): void
    {
        $doc = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
        )->export();
        $this->assertSame('object', $doc->root['kind']);
        $this->assertSame('string', $doc->root['properties']['name']['kind']);
        $this->assertSame(['name'], $doc->root['required']);
        $this->assertSame('reject', $doc->root['unknownKeys']);
    }

    public function testExportRecordSchema(): void
    {
        $doc = AnyVali::record(AnyVali::int())->export();
        $this->assertSame('record', $doc->root['kind']);
        $this->assertSame('int', $doc->root['values']['kind']);
    }

    public function testExportUnionSchema(): void
    {
        $doc = AnyVali::union([AnyVali::string(), AnyVali::int()])->export();
        $this->assertSame('union', $doc->root['kind']);
        $this->assertCount(2, $doc->root['variants']);
    }

    public function testExportIntersectionSchema(): void
    {
        $doc = AnyVali::intersection([
            AnyVali::number()->min(0),
            AnyVali::number()->max(100),
        ])->export();
        $this->assertSame('intersection', $doc->root['kind']);
        $this->assertCount(2, $doc->root['allOf']);
    }

    public function testExportOptionalSchema(): void
    {
        $doc = AnyVali::optional(AnyVali::string())->export();
        $this->assertSame('optional', $doc->root['kind']);
        $this->assertSame('string', $doc->root['schema']['kind']);
    }

    public function testExportNullableSchema(): void
    {
        $doc = AnyVali::nullable(AnyVali::string())->export();
        $this->assertSame('nullable', $doc->root['kind']);
        $this->assertSame('string', $doc->root['schema']['kind']);
    }

    public function testExportRefSchema(): void
    {
        $doc = AnyVali::ref('#/definitions/User')->export();
        $this->assertSame('ref', $doc->root['kind']);
        $this->assertSame('#/definitions/User', $doc->root['ref']);
    }

    public function testExportWithDefault(): void
    {
        $doc = AnyVali::string()->default('hello')->export();
        $this->assertSame('hello', $doc->root['default']);
    }

    public function testExportWithCoerce(): void
    {
        $doc = AnyVali::int()->coerce('string->int')->export();
        $this->assertSame('string->int', $doc->root['coerce']);
    }

    public function testExportPortableFailsWithCustomValidators(): void
    {
        $s = AnyVali::string()->refine(fn($v) => null);
        $this->expectException(\RuntimeException::class);
        $s->export(ExportMode::Portable);
    }

    public function testExportExtendedSucceedsWithCustomValidators(): void
    {
        $s = AnyVali::string()->refine(fn($v) => null);
        $doc = $s->export(ExportMode::Extended);
        $this->assertSame('string', $doc->root['kind']);
    }

    // ── Import ──────────────────────────────────────

    public function testImportStringSchema(): void
    {
        $doc = new AnyValiDocument(root: ['kind' => 'string', 'minLength' => 3]);
        $s = AnyVali::import($doc);

        $this->assertSame('abc', $s->parse('abc'));

        $result = $s->safeParse('ab');
        $this->assertFalse($result->success);
    }

    public function testImportNumberSchema(): void
    {
        $s = AnyVali::import(['anyvaliVersion' => '1.0', 'schemaVersion' => '1',
            'root' => ['kind' => 'number', 'min' => 10], 'definitions' => [], 'extensions' => []]);
        $this->assertSame(10, $s->parse(10));
    }

    public function testImportFromJson(): void
    {
        $json = json_encode([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1',
            'root' => ['kind' => 'string'],
            'definitions' => new \stdClass(),
            'extensions' => new \stdClass(),
        ]);
        $s = AnyVali::import($json);
        $this->assertSame('hello', $s->parse('hello'));
    }

    public function testImportObjectSchema(): void
    {
        $doc = new AnyValiDocument(root: [
            'kind' => 'object',
            'properties' => [
                'name' => ['kind' => 'string'],
                'age' => ['kind' => 'int'],
            ],
            'required' => ['name', 'age'],
            'unknownKeys' => 'reject',
        ]);
        $s = AnyVali::import($doc);
        $result = $s->parse(['name' => 'Alice', 'age' => 30]);
        $this->assertSame(['name' => 'Alice', 'age' => 30], $result);
    }

    public function testImportUnionSchema(): void
    {
        $doc = new AnyValiDocument(root: [
            'kind' => 'union',
            'variants' => [
                ['kind' => 'string'],
                ['kind' => 'int'],
            ],
        ]);
        $s = AnyVali::import($doc);
        $this->assertSame('hello', $s->parse('hello'));
        $this->assertSame(42, $s->parse(42));
    }

    public function testImportWithDefault(): void
    {
        $doc = new AnyValiDocument(root: [
            'kind' => 'object',
            'properties' => [
                'name' => ['kind' => 'string'],
                'role' => ['kind' => 'string', 'default' => 'user'],
            ],
            'required' => ['name'],
            'unknownKeys' => 'reject',
        ]);
        $s = AnyVali::import($doc);
        $result = $s->parse(['name' => 'Alice']);
        $this->assertSame(['name' => 'Alice', 'role' => 'user'], $result);
    }

    public function testImportWithCoerce(): void
    {
        $doc = new AnyValiDocument(root: [
            'kind' => 'int',
            'coerce' => 'string->int',
        ]);
        $s = AnyVali::import($doc);
        $this->assertSame(42, $s->parse('42'));
    }

    public function testImportUnsupportedKindThrows(): void
    {
        $doc = new AnyValiDocument(root: ['kind' => 'custom_unsupported']);
        $this->expectException(\RuntimeException::class);
        AnyVali::import($doc);
    }

    public function testImportMissingKindThrows(): void
    {
        $doc = new AnyValiDocument(root: []);
        $this->expectException(\RuntimeException::class);
        AnyVali::import($doc);
    }

    // ── Round-trip ──────────────────────────────────

    public function testRoundTripStringSchema(): void
    {
        $original = AnyVali::string()->minLength(1)->maxLength(50)->pattern('^[a-z]+$');
        $doc = $original->export();
        $imported = AnyVali::import($doc);

        $this->assertSame('abc', $imported->parse('abc'));

        $result = $imported->safeParse('ABC');
        $this->assertFalse($result->success);

        $result = $imported->safeParse('');
        $this->assertFalse($result->success);
    }

    public function testRoundTripObjectSchema(): void
    {
        $original = AnyVali::object(
            [
                'name' => AnyVali::string(),
                'age' => AnyVali::int()->min(0),
            ],
            ['name', 'age'],
        );
        $doc = $original->export();
        $imported = AnyVali::import($doc);

        $this->assertSame(
            ['name' => 'Alice', 'age' => 30],
            $imported->parse(['name' => 'Alice', 'age' => 30])
        );

        $result = $imported->safeParse(['name' => 'Alice']);
        $this->assertFalse($result->success);
    }

    // ── AnyValiDocument ──────────────────────────────

    public function testDocumentToJson(): void
    {
        $doc = new AnyValiDocument(root: ['kind' => 'string']);
        $json = $doc->toJson();
        $decoded = json_decode($json, true);
        $this->assertSame('1.0', $decoded['anyvaliVersion']);
        $this->assertSame('string', $decoded['root']['kind']);
    }

    public function testDocumentFromJson(): void
    {
        $json = '{"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"string"},"definitions":{},"extensions":{}}';
        $doc = AnyValiDocument::fromJson($json);
        $this->assertSame('string', $doc->root['kind']);
    }

    public function testDocumentFromArray(): void
    {
        $doc = AnyValiDocument::fromArray([
            'root' => ['kind' => 'int'],
        ]);
        $this->assertSame('int', $doc->root['kind']);
        $this->assertSame('1.0', $doc->anyvaliVersion);
    }

    public function testDocumentToArray(): void
    {
        $doc = new AnyValiDocument(root: ['kind' => 'bool']);
        $arr = $doc->toArray();
        $this->assertSame('bool', $arr['root']['kind']);
        $this->assertSame('1.0', $arr['anyvaliVersion']);
    }

    // ── Import Int/Float Widths ──────────────────────

    public function testImportInt8(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: ['kind' => 'int8']));
        $this->assertSame(127, $s->parse(127));
        $result = $s->safeParse(128);
        $this->assertFalse($result->success);
    }

    public function testImportFloat32(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: ['kind' => 'float32']));
        $this->assertSame(1.5, $s->parse(1.5));
    }

    public function testImportFloat64(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: ['kind' => 'float64']));
        $this->assertSame(3.14, $s->parse(3.14));
    }

    // ── Import Array/Tuple/Record ──────────────────

    public function testImportArraySchema(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'array',
            'items' => ['kind' => 'int'],
            'minItems' => 1,
        ]));
        $this->assertSame([1, 2], $s->parse([1, 2]));
        $result = $s->safeParse([]);
        $this->assertFalse($result->success);
    }

    public function testImportTupleSchema(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'tuple',
            'elements' => [['kind' => 'string'], ['kind' => 'int']],
        ]));
        $this->assertSame(['hello', 42], $s->parse(['hello', 42]));
    }

    public function testImportRecordSchema(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'record',
            'values' => ['kind' => 'int'],
        ]));
        $this->assertSame(['a' => 1], $s->parse(['a' => 1]));
    }

    // ── Import Optional/Nullable ──────────────────

    public function testImportOptional(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'object',
            'properties' => [
                'name' => ['kind' => 'optional', 'schema' => ['kind' => 'string']],
            ],
            'required' => [],
            'unknownKeys' => 'reject',
        ]));
        $this->assertSame([], $s->parse([]));
        $this->assertSame(['name' => 'hello'], $s->parse(['name' => 'hello']));
    }

    public function testImportNullable(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'nullable',
            'schema' => ['kind' => 'string'],
        ]));
        $this->assertNull($s->parse(null));
        $this->assertSame('hello', $s->parse('hello'));
    }

    // ── Import Intersection ──────────────────────

    public function testImportIntersection(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'intersection',
            'allOf' => [
                ['kind' => 'number', 'min' => 0],
                ['kind' => 'number', 'max' => 100],
            ],
        ]));
        $this->assertSame(50, $s->parse(50));
        $result = $s->safeParse(-5);
        $this->assertFalse($result->success);
    }

    // ── Import Literal/Enum ──────────────────────

    public function testImportLiteral(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: ['kind' => 'literal', 'value' => 'hello']));
        $this->assertSame('hello', $s->parse('hello'));
    }

    public function testImportEnum(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: ['kind' => 'enum', 'values' => ['a', 'b']]));
        $this->assertSame('a', $s->parse('a'));
    }

    // ── Import with strip unknownKeys ──────────────

    public function testImportObjectStrip(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'object',
            'properties' => ['name' => ['kind' => 'string']],
            'required' => ['name'],
            'unknownKeys' => 'strip',
        ]));
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertSame(['name' => 'Alice'], $result);
    }

    public function testImportObjectAllow(): void
    {
        $s = AnyVali::import(new AnyValiDocument(root: [
            'kind' => 'object',
            'properties' => ['name' => ['kind' => 'string']],
            'required' => ['name'],
            'unknownKeys' => 'allow',
        ]));
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertSame(['name' => 'Alice', 'extra' => 'value'], $result);
    }

    // ── Import with chained coercion ──────────────

    public function testImportChainedCoerce(): void
    {
        $doc = new AnyValiDocument(root: [
            'kind' => 'string',
            'coerce' => ['trim', 'lower'],
        ]);
        $s = AnyVali::import($doc);
        $this->assertSame('hello', $s->parse('  HELLO  '));
    }

    // ── Exporter direct ──────────────────────────

    public function testExporterDirect(): void
    {
        $doc = Exporter::export(AnyVali::string());
        $this->assertSame('string', $doc->root['kind']);
    }

    // ── Number export with constraints ──────────

    public function testExportNumberAllConstraints(): void
    {
        $s = AnyVali::number()->min(1)->max(100)->exclusiveMin(0)->exclusiveMax(101)->multipleOf(5);
        $doc = $s->export();
        $this->assertSame(1.0, $doc->root['min']);
        $this->assertSame(100.0, $doc->root['max']);
        $this->assertSame(0.0, $doc->root['exclusiveMin']);
        $this->assertSame(101.0, $doc->root['exclusiveMax']);
        $this->assertSame(5.0, $doc->root['multipleOf']);
    }

    // ── String export with all constraints ──────

    public function testExportStringAllConstraints(): void
    {
        $s = AnyVali::string()
            ->minLength(1)->maxLength(10)
            ->pattern('^[a-z]+$')
            ->startsWith('a')->endsWith('z')
            ->includes('b')->format('email');
        $doc = $s->export();
        $this->assertSame(1, $doc->root['minLength']);
        $this->assertSame(10, $doc->root['maxLength']);
        $this->assertSame('^[a-z]+$', $doc->root['pattern']);
        $this->assertSame('a', $doc->root['startsWith']);
        $this->assertSame('z', $doc->root['endsWith']);
        $this->assertSame('b', $doc->root['includes']);
        $this->assertSame('email', $doc->root['format']);
    }
}

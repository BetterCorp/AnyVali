<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\Interchange\Importer;
use PHPUnit\Framework\TestCase;

final class ImportRegressionTest extends TestCase
{
    private function document(string $kind = 'string'): array
    {
        $ref = ['kind' => 'ref', 'ref' => '#/definitions/JsonValue'];
        return [
            'anyvaliVersion' => '1.0', 'schemaVersion' => '1.1', 'root' => $ref,
            'definitions' => ['JsonValue' => ['kind' => 'union', 'variants' => [
                ['kind' => $kind], ['kind' => 'array', 'items' => $ref],
                ['kind' => 'record', 'valueSchema' => $ref],
            ]]],
        ];
    }

    public function testRecursiveImportsAndNativeParentsRoundTrip(): void
    {
        $schema = AnyVali::import($this->document());
        for ($round = 0; $round < 2; $round++) {
            foreach (['leaf', ['leaf', ['nested']], ['key' => ['leaf']]] as $input) {
                $this->assertTrue($schema->safeParse($input)->success);
            }
            $this->assertFalse($schema->safeParse(['key' => [true]])->success);
            $schema = AnyVali::import($schema->export()->toJson());
        }
        $parent = AnyVali::object(['value' => $schema]);
        $restored = AnyVali::import($parent->export()->toJson());
        $this->assertTrue($restored->safeParse(['value' => ['nested']])->success);
        $this->assertFalse($restored->safeParse(['value' => [false]])->success);
    }

    public function testDocumentsRemainIsolated(): void
    {
        $first = AnyVali::import($this->document());
        $second = AnyVali::import($this->document('bool'));
        $this->assertTrue($first->safeParse('leaf')->success);
        $this->assertFalse($second->safeParse('leaf')->success);
        $this->assertTrue($second->safeParse(true)->success);
        $this->assertFalse($first->safeParse(true)->success);
        $this->expectException(\RuntimeException::class);
        AnyVali::object(['first' => $first, 'second' => $second])->export();
    }

    public function testUnguardedReferenceCyclesFailValidation(): void
    {
        $doc = $this->document();
        $doc['definitions']['JsonValue'] = $doc['root'];
        $this->assertFalse(AnyVali::import($doc)->safeParse('leaf')->success);
    }

    public function testRequiredDefaultsAndNullRoundTrip(): void
    {
        foreach ([['string', 'fallback'], ['bool', true], ['number', 12.0], ['int32', 12]] as [$kind, $default]) {
            $doc = ['root' => ['kind' => 'object', 'required' => ['value'], 'properties' => [
                'value' => ['kind' => $kind, 'default' => $default],
            ]]];
            $schema = AnyVali::import($doc);
            for ($round = 0; $round < 2; $round++) {
                $result = $schema->safeParse(new \stdClass());
                $this->assertTrue($result->success);
                $this->assertEquals(['value' => $default], $result->value);
                $this->assertFalse($schema->safeParse(['value' => null])->success);
                $schema = AnyVali::import($schema->export()->toJson());
            }
        }
    }

    public function testNullableAndReferenceDefaults(): void
    {
        foreach ([null, 'fallback'] as $default) {
            $doc = ['root' => ['kind' => 'object', 'required' => ['value'], 'properties' => [
                'value' => ['kind' => 'nullable', 'schema' => ['kind' => 'string'], 'default' => $default],
            ]]];
            $schema = AnyVali::import($doc);
            for ($round = 0; $round < 2; $round++) {
                $this->assertSame(['value' => $default], $schema->parse(new \stdClass()));
                $this->assertSame(['value' => null], $schema->parse(['value' => null]));
                $schema = AnyVali::import($schema->export()->toJson());
            }
        }
        $doc = $this->document();
        $doc['root'] = ['kind' => 'object', 'required' => ['value'], 'properties' => [
            'value' => $doc['root'] + ['default' => ['leaf']],
        ]];
        $schema = AnyVali::import($doc);
        $this->assertSame(['value' => ['leaf']], $schema->parse(new \stdClass()));
        $this->assertSame(['value' => ['leaf']], AnyVali::import($schema->export()->toJson())->parse(new \stdClass()));
    }

    public function testCanonicalRecordAndLegacyAlias(): void
    {
        foreach (['valueSchema', 'values'] as $key) {
            $schema = Importer::importNode(['kind' => 'record', $key => ['kind' => 'string']]);
            $this->assertTrue($schema->safeParse(['key' => 'leaf'])->success);
            $this->assertFalse($schema->safeParse(['key' => true])->success);
            $this->assertArrayHasKey('values', $schema->exportNode());
            $this->assertArrayNotHasKey('valueSchema', $schema->exportNode());
        }
    }

    public function testEquivalentDefinitionsIgnoreObjectKeyOrderButKeepListOrder(): void
    {
        $ref = ['kind' => 'ref', 'ref' => '#/definitions/Value'];
        $first = ['kind' => 'object', 'properties' => ['a' => ['kind' => 'string'], 'b' => ['kind' => 'bool']], 'required' => ['a', 'b']];
        $second = ['required' => ['a', 'b'], 'properties' => ['b' => ['kind' => 'bool'], 'a' => ['kind' => 'string']], 'kind' => 'object'];
        $schemas = [];
        foreach ([$first, $second] as $node) {
            $schemas[] = AnyVali::import(['root' => $ref, 'definitions' => ['Value' => $node]]);
        }
        $this->assertCount(1, AnyVali::object(['first' => $schemas[0], 'second' => $schemas[1]])->export()->definitions);
        $schemas = [];
        foreach ([['a', 'b'], ['b', 'a']] as $values) {
            $schemas[] = AnyVali::import(['root' => $ref, 'definitions' => ['Value' => ['kind' => 'literal', 'value' => $values]]]);
        }
        $this->expectException(\RuntimeException::class);
        AnyVali::object(['first' => $schemas[0], 'second' => $schemas[1]])->export();
    }

    public function testDefaultsInheritedThroughReferencesAndNotCoerced(): void
    {
        foreach ([['string', 'fallback', null, true], ['int', '12', 'string->int', false], ['int', 12, 'string->int', true]] as [$kind, $default, $coerce, $valid]) {
            $target = ['kind' => $kind, 'default' => $default];
            if ($coerce !== null) $target['coerce'] = $coerce;
            $schema = AnyVali::import([
                'root' => ['kind' => 'object', 'required' => ['value'], 'properties' => [
                    'value' => ['kind' => 'ref', 'ref' => '#/definitions/Alias'],
                ]],
                'definitions' => ['Alias' => ['kind' => 'ref', 'ref' => '#/definitions/Value'], 'Value' => $target],
            ]);
            for ($round = 0; $round < 2; $round++) {
                $result = $schema->safeParse([]);
                $this->assertSame($valid, $result->success);
                if ($valid) $this->assertSame(['value' => $default], $result->value);
                else $this->assertSame('default_invalid', $result->issues[0]->code);
                $this->assertFalse($schema->safeParse(['value' => null])->success);
                $schema = AnyVali::import($schema->export()->toJson());
            }
        }
        $schema = AnyVali::object(['value' => AnyVali::int()->coerce('string->int')->default('12')], required: ['value']);
        $this->assertSame('default_invalid', $schema->safeParse([])->issues[0]->code);
        $this->assertSame(['value' => 12], $schema->parse(['value' => '12']));
    }

    public function testRecursiveInputValidationIsBounded(): void
    {
        $schema = AnyVali::import([
            'root' => ['kind' => 'ref', 'ref' => '#/definitions/Node'],
            'definitions' => ['Node' => ['kind' => 'object', 'required' => [], 'properties' => [
                'child' => ['kind' => 'ref', 'ref' => '#/definitions/Node'],
            ]]],
        ]);
        $this->assertTrue($schema->safeParse(['child' => []])->success);
        $deep = [];
        for ($i = 0; $i < 1000; $i++) $deep = ['child' => $deep];
        $result = $schema->safeParse($deep);
        $this->assertFalse($result->success);
        $this->assertSame('Maximum validation depth exceeded', $result->issues[0]->message);
        $cycle = [];
        $cycle['child'] = &$cycle;
        $this->assertFalse($schema->safeParse($cycle)->success);
        $this->assertTrue($schema->safeParse([])->success);
    }

    public function testPortableDefinitionsRejectCustomValidators(): void
    {
        $ref = new \AnyVali\Schemas\RefSchema('#/definitions/Value');
        $ref->resolve(AnyVali::object(['value' => AnyVali::string()->refine(fn() => null)]));
        $this->assertCount(1, $ref->export(\AnyVali\ExportMode::Extended)->definitions);
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('custom validators');
        $ref->export();
    }


    public function testSameDepthCycleWithTerminatingVariant(): void
    {
        $ref = ['kind' => 'ref', 'ref' => '#/definitions/A'];
        $schema = AnyVali::import(['root' => $ref, 'definitions' => ['A' => [
            'kind' => 'union', 'variants' => [['kind' => 'string'], $ref],
        ]]]);
        $this->assertTrue($schema->safeParse('leaf')->success);
        $this->assertFalse($schema->safeParse(true)->success);
        $this->assertTrue(AnyVali::import($schema->export()->toJson())->safeParse('leaf')->success);
    }

    public function testRecursiveWorkBudgetAcrossFailingBranches(): void
    {
        $ref = ['kind' => 'ref', 'ref' => '#/definitions/A'];
        $variant = ['kind' => 'array', 'items' => $ref];
        $schema = AnyVali::import(['root' => $ref, 'definitions' => ['A' => [
            'kind' => 'union', 'variants' => [$variant, $variant],
        ]]]);
        $value = true;
        for ($i = 0; $i < 24; $i++) $value = [$value];
        $start = microtime(true);
        $ctx = new \AnyVali\ValidationContext();
        $result = $schema->safeParse($value, $ctx);
        $this->assertFalse($result->success);
        $this->assertSame('Maximum validation work exceeded', $result->issues[0]->message);
        $this->assertLessThan(5.0, microtime(true) - $start);
        $this->assertTrue($schema->safeParse([], $ctx)->success);
    }

    public function testDefaultAndPresentValuesUseEqualDepth(): void
    {
        $schema = AnyVali::object(['value' => AnyVali::int()->default(12)], required: ['value']);
        $present = ['value' => 12];
        $missing = [];
        for ($i = 0; $i < 62; $i++) {
            $schema = AnyVali::object(['child' => $schema], required: ['child']);
            $present = ['child' => $present];
            $missing = ['child' => $missing];
        }
        $this->assertTrue($schema->safeParse($present)->success);
        $this->assertTrue($schema->safeParse($missing)->success);
    }

    public function testChildrenOfDefaultObjectsStillCoerce(): void
    {
        $schema = AnyVali::object(['config' => AnyVali::object([
            'count' => AnyVali::int()->coerce('string->int'),
        ], required: ['count'])->default(['count' => '12'])], required: ['config']);
        $this->assertSame(['config' => ['count' => 12]], $schema->parse([]));
        $this->assertSame($schema->parse([]), $schema->parse(['config' => ['count' => '12']]));
    }


    public function testEquivalentNumericDefaultsPreserveIntegerPrecision(): void
    {
        $make = fn($default) => AnyVali::import([
            'root' => ['kind' => 'ref', 'ref' => '#/definitions/Number'],
            'definitions' => ['Number' => ['kind' => 'number', 'default' => $default]],
        ]);
        $this->assertCount(1, AnyVali::object(['a' => $make(1), 'b' => $make(1.0)])->export()->definitions);
        $this->expectException(\RuntimeException::class);
        AnyVali::object(['a' => $make(9007199254740993), 'b' => $make(9007199254740992.0)])->export();
    }

    public function testImportDepthBoundsLongReferencesAndInlineSchemas(): void
    {
        $definitions = ['A1000' => ['kind' => 'string']];
        for ($i = 999; $i >= 0; $i--) {
            $definitions['A' . $i] = ['kind' => 'ref', 'ref' => '#/definitions/A' . ($i + 1)];
        }
        $documents = [['root' => ['kind' => 'ref', 'ref' => '#/definitions/A0'], 'definitions' => $definitions]];
        foreach (['items', 'elements', 'variants', 'allOf'] as $key) {
            $kind = ['items' => 'array', 'elements' => 'tuple', 'variants' => 'union', 'allOf' => 'intersection'][$key];
            $node = ['kind' => 'string'];
            for ($i = 0; $i < 63; $i++) {
                $node = ['kind' => $kind, $key => $key === 'items' ? $node : [$node]];
            }
            $this->assertInstanceOf(\AnyVali\Schema::class, Importer::importNode($node));
            for ($i = 0; $i < 1000; $i++) {
                $node = ['kind' => $kind, $key => $key === 'items' ? $node : [$node]];
            }
            $documents[] = ['root' => $node];
        }
        foreach ($documents as $document) {
            try {
                AnyVali::import($document);
                $this->fail('Expected excessive import depth to be rejected');
            } catch (\RuntimeException $error) {
                $this->assertSame('Maximum schema import depth exceeded', $error->getMessage());
            }
            $this->assertSame('leaf', AnyVali::import(['root' => ['kind' => 'string']])->parse('leaf'));
        }
    }

    public function testTypeSensitiveDefinitionsKeepNumericPayloadTypes(): void
    {
        foreach (['literal', 'enum'] as $kind) {
            foreach ([[1, 1.0], [['kind' => 'number', 'default' => 1], ['kind' => 'number', 'default' => 1.0]]] as [$integer, $floating]) {
                $schemas = [];
                foreach ([$integer, $floating] as $value) {
                    $payload = match ($kind) {
                        'literal' => ['value' => $value],
                        'enum' => ['values' => [$value]],
                        default => ['default' => $value],
                    };
                    $schemas[] = AnyVali::import([
                        'root' => ['kind' => 'ref', 'ref' => '#/definitions/Value'],
                        'definitions' => ['Value' => ['kind' => $kind] + $payload],
                    ]);
                }
                if ($kind !== 'any') {
                    $this->assertTrue($schemas[0]->safeParse($integer)->success);
                    $this->assertFalse($schemas[0]->safeParse($floating)->success);
                    $this->assertTrue($schemas[1]->safeParse($floating)->success);
                    $this->assertFalse($schemas[1]->safeParse($integer)->success);
                }
                try {
                    AnyVali::object(['a' => $schemas[0], 'b' => $schemas[1]])->export();
                    $this->fail('Expected type-sensitive definitions to conflict');
                } catch (\RuntimeException $error) {
                    $this->assertSame('Conflicting definition: Value', $error->getMessage());
                }
            }
        }
    }

    public function testDefaultsSurviveExportForAllReferencedKinds(): void
    {
        $string = ['kind' => 'string'];
        $cases = [
            [$string, 'fallback'],
            [['kind' => 'number'], 12],
            [['kind' => 'int'], 12],
            [['kind' => 'bool'], false],
            [['kind' => 'null'], null],
            [['kind' => 'any'], ['value' => null]],
            [['kind' => 'unknown'], null],
            [['kind' => 'literal', 'value' => 'fallback'], 'fallback'],
            [['kind' => 'enum', 'values' => ['fallback']], 'fallback'],
            [['kind' => 'array', 'items' => $string], ['fallback']],
            [['kind' => 'tuple', 'elements' => [$string]], ['fallback']],
            [['kind' => 'object', 'properties' => ['value' => $string], 'required' => ['value']], ['value' => 'fallback']],
            [['kind' => 'record', 'values' => $string], ['value' => 'fallback']],
            [['kind' => 'union', 'variants' => [$string, ['kind' => 'bool']]], 'fallback'],
            [['kind' => 'intersection', 'allOf' => [$string, ['kind' => 'string', 'minLength' => 2]]], 'fallback'],
            [['kind' => 'optional', 'schema' => $string], 'fallback'],
            [['kind' => 'nullable', 'schema' => $string], null],
            [['kind' => 'ref', 'ref' => '#/definitions/Leaf'], 'fallback'],
        ];
        foreach ($cases as [$target, $default]) {
            foreach ([true, false] as $referenced) {
                $target['default'] = $default;
                $schema = AnyVali::import([
                    'root' => ['kind' => 'object', 'required' => ['value'], 'properties' => [
                        'value' => $referenced ? ['kind' => 'ref', 'ref' => '#/definitions/Value'] : $target,
                    ]],
                    'definitions' => ['Value' => $target, 'Leaf' => $string],
                ]);
                for ($round = 0; $round < 3; $round++) {
                    $result = $schema->safeParse([]);
                    $this->assertTrue($result->success, $target['kind'] . ' round ' . $round);
                    $this->assertEquals(['value' => $default], $result->value);
                    $schema = AnyVali::import($schema->export()->toJson());
                }
            }
        }
        $schema = AnyVali::import([
            'root' => ['kind' => 'object', 'required' => ['value'], 'properties' => [
                'value' => ['kind' => 'ref', 'ref' => '#/definitions/Never'],
            ]],
            'definitions' => ['Never' => ['kind' => 'never', 'default' => null]],
        ]);
        for ($round = 0; $round < 3; $round++) {
            $this->assertSame('default_invalid', $schema->safeParse([])->issues[0]->code);
            $schema = AnyVali::import($schema->export()->toJson());
        }
    }

    public function testCompositeCoercionsSurviveReferencedDefinitionExport(): void
    {
        $int = ['kind' => 'int'];
        foreach ([
            ['kind' => 'union', 'variants' => [$int]],
            ['kind' => 'intersection', 'allOf' => [$int]],
            ['kind' => 'optional', 'schema' => $int],
            ['kind' => 'nullable', 'schema' => $int],
            ['kind' => 'literal', 'value' => 12],
            ['kind' => 'enum', 'values' => [12]],
            ['kind' => 'any'],
            ['kind' => 'unknown'],
        ] as $target) {
            $target['coerce'] = ['trim', 'string->int'];
            $schema = AnyVali::import([
                'root' => ['kind' => 'ref', 'ref' => '#/definitions/Value'],
                'definitions' => ['Value' => $target],
            ]);
            for ($round = 0; $round < 3; $round++) {
                $this->assertSame(12, $schema->parse(' 12 '));
                $this->assertFalse($schema->safeParse('invalid')->success);
                $schema = AnyVali::import($schema->export()->toJson());
            }
        }
    }

}

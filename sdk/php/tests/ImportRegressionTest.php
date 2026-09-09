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

    public function testUnguardedReferenceCyclesAreRejected(): void
    {
        $doc = $this->document();
        $doc['definitions']['JsonValue'] = $doc['root'];
        $this->expectException(\RuntimeException::class);
        AnyVali::import($doc);
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
            $this->assertArrayHasKey('valueSchema', $schema->exportNode());
            $this->assertArrayNotHasKey('values', $schema->exportNode());
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

}

<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\IssueCodes;
use PHPUnit\Framework\TestCase;

final class DefaultsTest extends TestCase
{
    public function testMissingFieldGetsDefault(): void
    {
        $schema = AnyVali::object([
            'name' => AnyVali::string(),
            'role' => AnyVali::string()->default('user'),
        ], required: ['name']);

        $result = $schema->safeParse(['name' => 'Alice']);

        $this->assertTrue($result->success, json_encode(array_map(fn($i) => $i->toArray(), $result->issues)));
        $this->assertSame('user', $result->value['role']);
    }

    public function testPresentFieldIsNotOverwritten(): void
    {
        $schema = AnyVali::object([
            'role' => AnyVali::string()->default('user'),
        ]);

        $result = $schema->safeParse(['role' => 'admin']);

        $this->assertTrue($result->success);
        $this->assertSame('admin', $result->value['role']);
    }

    public function testInvalidDefaultProducesDefaultInvalid(): void
    {
        $schema = AnyVali::object([
            'count' => AnyVali::int()->min(10)->default(5),
        ]);

        $result = $schema->safeParse([]);

        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::DEFAULT_INVALID, $result->issues[0]->code);
        $this->assertSame(['count'], $result->issues[0]->path);
    }

    public function testNullIsNotAbsentForNullableDefault(): void
    {
        $schema = AnyVali::object([
            'value' => AnyVali::nullable(AnyVali::string())->default('fallback'),
        ]);

        $result = $schema->safeParse(['value' => null]);

        $this->assertTrue($result->success);
        $this->assertNull($result->value['value']);
    }

    public function testFalsyDefaultsAreApplied(): void
    {
        $schema = AnyVali::object([
            'count' => AnyVali::int()->default(0),
            'name' => AnyVali::string()->default(''),
            'active' => AnyVali::bool()->default(false),
        ]);

        $result = $schema->safeParse([]);

        $this->assertTrue($result->success, json_encode(array_map(fn($i) => $i->toArray(), $result->issues)));
        $this->assertSame(['count' => 0, 'name' => '', 'active' => false], $result->value);
    }

    public function testNestedObjectFieldGetsDefault(): void
    {
        $schema = AnyVali::object([
            'user' => AnyVali::object([
                'name' => AnyVali::string(),
                'role' => AnyVali::string()->default('guest'),
            ], required: ['name']),
        ], required: ['user']);

        $result = $schema->safeParse(['user' => ['name' => 'Bob']]);

        $this->assertTrue($result->success, json_encode(array_map(fn($i) => $i->toArray(), $result->issues)));
        $this->assertSame('guest', $result->value['user']['role']);
    }
}

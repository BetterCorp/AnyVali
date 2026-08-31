<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\ValidationError;
use PHPUnit\Framework\TestCase;

final class SensitiveDataTest extends TestCase
{
    public function testEncryptedRoundTripAndValidation(): void
    {
        $schema = AnyVali::object([
            'name' => AnyVali::string(),
            'secret' => AnyVali::string()->describe('secret', ['sensitive' => true]),
            'profile' => AnyVali::object(['token' => AnyVali::string()])
                ->describe('profile', ['sensitive' => true]),
            'aliases' => AnyVali::array(
                AnyVali::string()->describe('alias', ['sensitive' => true]),
            ),
        ], ['name', 'secret', 'profile', 'aliases']);
        $plain = [
            'name' => 'Ada',
            'secret' => 'abc',
            'profile' => ['token' => 'xyz'],
            'aliases' => ['one', 'two'],
        ];

        $encrypted = AnyVali::encrypt(
            $schema,
            $plain,
            fn(array $path, mixed $value) => 'encrypted:' . json_encode($value, JSON_THROW_ON_ERROR),
        );
        $this->assertTrue(AnyVali::safeParseEncrypted($schema, $encrypted)->success);
        $this->assertSame($plain, AnyVali::decrypt(
            $schema,
            $encrypted,
            fn(array $path, string $value) => json_decode(substr($value, 10), true, flags: JSON_THROW_ON_ERROR),
        ));

        $this->expectException(ValidationError::class);
        AnyVali::encrypt($schema, $plain, fn() => 'broken');
    }
}

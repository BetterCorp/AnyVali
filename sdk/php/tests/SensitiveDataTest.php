<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\ValidationError;
use AnyVali\Interchange\Importer;
use PHPUnit\Framework\TestCase;

final class SensitiveDataTest extends TestCase
{
    public function testImportedSensitiveMetadataEncryptsInsteadOfReturningPlaintext(): void
    {
        $schema = Importer::importNode([
            'kind' => 'string',
            'metadata' => ['sensitive' => true, 'vendor' => ['scope' => 'test']],
        ]);
        $encrypted = AnyVali::encrypt($schema, 'secret', fn(array $path, mixed $value) => 'encrypted:' . $value);
        $this->assertSame('encrypted:secret', $encrypted);
        $this->assertFalse(AnyVali::safeParseEncrypted($schema, 'secret')->success);
        $this->assertTrue($schema->getMetadata()['sensitive']);
    }

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

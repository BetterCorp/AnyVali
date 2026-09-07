<?php

declare(strict_types=1);

namespace AnyVali;

final class SensitiveData
{
    private function __construct()
    {
    }

    public static function safeParseEncrypted(Schema $schema, mixed $data): ParseResult
    {
        return $schema->safeParse($data, new ValidationContext(sensitiveMode: 'encrypted'));
    }

    public static function encrypt(Schema $schema, mixed $data, callable $transform): mixed
    {
        $encrypted = self::transform($schema, $schema->parse($data), 'encrypt', $transform);
        $result = self::safeParseEncrypted($schema, $encrypted);
        if (!$result->success) throw new ValidationError($result->issues);
        return $result->value;
    }

    public static function decrypt(Schema $schema, mixed $data, callable $transform): mixed
    {
        $encrypted = self::safeParseEncrypted($schema, $data);
        if (!$encrypted->success) throw new ValidationError($encrypted->issues);
        return $schema->parse(self::transform($schema, $encrypted->value, 'decrypt', $transform));
    }

    private static function transform(Schema $schema, mixed $data, string $mode, callable $transform): mixed
    {
        $result = $schema->safeParse($data, new ValidationContext(
            sensitiveMode: $mode,
            sensitiveTransform: $transform,
            sensitiveCache: new \ArrayObject(),
        ));
        if (!$result->success) throw new ValidationError($result->issues);
        return $result->value;
    }
}

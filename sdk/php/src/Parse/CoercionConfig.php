<?php

declare(strict_types=1);

namespace AnyVali\Parse;

final class CoercionConfig
{
    /**
     * @param string[] $coercions
     */
    public function __construct(
        public readonly array $coercions = [],
    ) {
    }

    /**
     * @param string|string[] $coerce
     */
    public static function from(string|array $coerce): self
    {
        if (is_string($coerce)) {
            return new self([$coerce]);
        }
        return new self($coerce);
    }
}

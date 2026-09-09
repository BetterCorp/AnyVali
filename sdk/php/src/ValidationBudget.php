<?php

declare(strict_types=1);

namespace AnyVali;

/** Shared work accounting for one parse, including all union branches. */
final class ValidationBudget
{
    public const MAX_CALLS = 100000;
    private int $calls = 0;
    private bool $exhausted = false;

    public function consume(): bool
    {
        if ($this->calls >= self::MAX_CALLS) {
            $this->exhausted = true;
            return false;
        }
        $this->calls++;
        return true;
    }

    public function exhausted(): bool
    {
        return $this->exhausted;
    }
}

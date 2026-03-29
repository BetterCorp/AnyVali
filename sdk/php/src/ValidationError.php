<?php

declare(strict_types=1);

namespace AnyVali;

final class ValidationError extends \RuntimeException
{
    /**
     * @param ValidationIssue[] $issues
     */
    public function __construct(
        public readonly array $issues,
    ) {
        $messages = array_map(
            fn(ValidationIssue $issue) => sprintf(
                '[%s] %s at path [%s]',
                $issue->code,
                $issue->message,
                implode('.', array_map('strval', $issue->path)),
            ),
            $issues,
        );
        parent::__construct(implode('; ', $messages));
    }
}

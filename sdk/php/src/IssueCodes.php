<?php

declare(strict_types=1);

namespace AnyVali;

final class IssueCodes
{
    public const INVALID_TYPE = 'invalid_type';
    public const REQUIRED = 'required';
    public const UNKNOWN_KEY = 'unknown_key';
    public const TOO_SMALL = 'too_small';
    public const TOO_LARGE = 'too_large';
    public const INVALID_STRING = 'invalid_string';
    public const INVALID_NUMBER = 'invalid_number';
    public const INVALID_LITERAL = 'invalid_literal';
    public const INVALID_UNION = 'invalid_union';
    public const CUSTOM_VALIDATION_NOT_PORTABLE = 'custom_validation_not_portable';
    public const UNSUPPORTED_EXTENSION = 'unsupported_extension';
    public const UNSUPPORTED_SCHEMA_KIND = 'unsupported_schema_kind';
    public const COERCION_FAILED = 'coercion_failed';
    public const DEFAULT_INVALID = 'default_invalid';

    private function __construct()
    {
    }
}

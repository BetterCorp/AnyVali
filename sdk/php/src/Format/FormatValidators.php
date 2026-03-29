<?php

declare(strict_types=1);

namespace AnyVali\Format;

final class FormatValidators
{
    private function __construct()
    {
    }

    public static function validate(string $format, string $value): bool
    {
        return match ($format) {
            'email' => self::isEmail($value),
            'url' => self::isUrl($value),
            'uuid' => self::isUuid($value),
            'ipv4' => self::isIpv4($value),
            'ipv6' => self::isIpv6($value),
            'date' => self::isDate($value),
            'date-time' => self::isDateTime($value),
            default => false,
        };
    }

    public static function isEmail(string $value): bool
    {
        // Must contain @ and a dot after @
        if (filter_var($value, FILTER_VALIDATE_EMAIL) === false) {
            return false;
        }
        // Additional check: must have a dot in domain part
        $parts = explode('@', $value);
        if (count($parts) !== 2) {
            return false;
        }
        return str_contains($parts[1], '.');
    }

    public static function isUrl(string $value): bool
    {
        if (filter_var($value, FILTER_VALIDATE_URL) === false) {
            return false;
        }
        $scheme = parse_url($value, PHP_URL_SCHEME);
        return $scheme === 'http' || $scheme === 'https';
    }

    public static function isUuid(string $value): bool
    {
        return (bool)preg_match(
            '/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i',
            $value
        );
    }

    public static function isIpv4(string $value): bool
    {
        // Reject leading zeros
        $parts = explode('.', $value);
        if (count($parts) !== 4) {
            return false;
        }
        foreach ($parts as $part) {
            if ($part === '') {
                return false;
            }
            if (strlen($part) > 1 && $part[0] === '0') {
                return false;
            }
            if (!ctype_digit($part)) {
                return false;
            }
            $num = (int)$part;
            if ($num < 0 || $num > 255) {
                return false;
            }
        }
        return true;
    }

    public static function isIpv6(string $value): bool
    {
        return filter_var($value, FILTER_VALIDATE_IP, FILTER_FLAG_IPV6) !== false;
    }

    public static function isDate(string $value): bool
    {
        if (!preg_match('/^\d{4}-\d{2}-\d{2}$/', $value)) {
            return false;
        }
        $parts = explode('-', $value);
        return checkdate((int)$parts[1], (int)$parts[2], (int)$parts[0]);
    }

    public static function isDateTime(string $value): bool
    {
        // ISO 8601 with required timezone (Z or +/-HH:MM)
        if (!preg_match(
            '/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/',
            $value
        )) {
            return false;
        }
        // Validate the date portion
        $datePart = substr($value, 0, 10);
        return self::isDate($datePart);
    }
}

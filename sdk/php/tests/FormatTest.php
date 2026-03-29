<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\Format\FormatValidators;
use AnyVali\IssueCodes;
use PHPUnit\Framework\TestCase;

final class FormatTest extends TestCase
{
    // ── Email ──────────────────────────────────────

    public function testEmailAcceptsValid(): void
    {
        $s = AnyVali::string()->format('email');
        $this->assertSame('user@example.com', $s->parse('user@example.com'));
    }

    public function testEmailRejectsWithoutAt(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('not-an-email');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
        $this->assertSame('email', $result->issues[0]->expected);
    }

    public function testEmailRejectsWithoutDot(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user@localhost');
        $this->assertFalse($result->success);
    }

    // ── URL ──────────────────────────────────────

    public function testUrlAcceptsHttps(): void
    {
        $s = AnyVali::string()->format('url');
        $this->assertSame('https://example.com', $s->parse('https://example.com'));
    }

    public function testUrlAcceptsHttp(): void
    {
        $s = AnyVali::string()->format('url');
        $this->assertSame('http://example.com/path?q=1', $s->parse('http://example.com/path?q=1'));
    }

    public function testUrlRejectsFtp(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('ftp://files.example.com');
        $this->assertFalse($result->success);
    }

    // ── UUID ──────────────────────────────────────

    public function testUuidAcceptsValid(): void
    {
        $s = AnyVali::string()->format('uuid');
        $this->assertSame(
            '550e8400-e29b-41d4-a716-446655440000',
            $s->parse('550e8400-e29b-41d4-a716-446655440000')
        );
    }

    public function testUuidRejectsInvalid(): void
    {
        $s = AnyVali::string()->format('uuid');
        $result = $s->safeParse('not-a-uuid');
        $this->assertFalse($result->success);
    }

    // ── IPv4 ──────────────────────────────────────

    public function testIpv4AcceptsValid(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $this->assertSame('192.168.1.1', $s->parse('192.168.1.1'));
    }

    public function testIpv4RejectsLeadingZeros(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.01.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsOutOfRange(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('256.1.1.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsTooFewOctets(): void
    {
        $this->assertFalse(FormatValidators::isIpv4('192.168.1'));
    }

    public function testIpv4RejectsEmptyOctet(): void
    {
        $this->assertFalse(FormatValidators::isIpv4('192..1.1'));
    }

    public function testIpv4RejectsNonDigit(): void
    {
        $this->assertFalse(FormatValidators::isIpv4('192.168.a.1'));
    }

    // ── IPv6 ──────────────────────────────────────

    public function testIpv6AcceptsValid(): void
    {
        $s = AnyVali::string()->format('ipv6');
        $this->assertSame(
            '2001:0db8:85a3:0000:0000:8a2e:0370:7334',
            $s->parse('2001:0db8:85a3:0000:0000:8a2e:0370:7334')
        );
    }

    public function testIpv6AcceptsCompressed(): void
    {
        $s = AnyVali::string()->format('ipv6');
        $this->assertSame('::1', $s->parse('::1'));
    }

    public function testIpv6RejectsInvalid(): void
    {
        $s = AnyVali::string()->format('ipv6');
        $result = $s->safeParse('not:an:ipv6');
        $this->assertFalse($result->success);
    }

    // ── Date ──────────────────────────────────────

    public function testDateAcceptsValid(): void
    {
        $s = AnyVali::string()->format('date');
        $this->assertSame('2024-02-29', $s->parse('2024-02-29'));
    }

    public function testDateRejectsInvalidLeapDay(): void
    {
        $s = AnyVali::string()->format('date');
        $result = $s->safeParse('2023-02-29');
        $this->assertFalse($result->success);
    }

    public function testDateRejectsBadFormat(): void
    {
        $this->assertFalse(FormatValidators::isDate('2024/02/29'));
    }

    // ── DateTime ──────────────────────────────────

    public function testDateTimeAcceptsWithZ(): void
    {
        $s = AnyVali::string()->format('date-time');
        $this->assertSame('2024-01-15T10:30:00Z', $s->parse('2024-01-15T10:30:00Z'));
    }

    public function testDateTimeAcceptsWithOffset(): void
    {
        $s = AnyVali::string()->format('date-time');
        $this->assertSame(
            '2024-01-15T10:30:00+05:30',
            $s->parse('2024-01-15T10:30:00+05:30')
        );
    }

    public function testDateTimeRejectsWithoutTimezone(): void
    {
        $s = AnyVali::string()->format('date-time');
        $result = $s->safeParse('2024-01-15T10:30:00');
        $this->assertFalse($result->success);
    }

    // ── Unknown Format ──────────────────────────────

    public function testUnknownFormatFails(): void
    {
        $this->assertFalse(FormatValidators::validate('unknown-format', 'test'));
    }

    // ── Direct FormatValidators calls ──────────────

    public function testFormatValidatorsDirect(): void
    {
        $this->assertTrue(FormatValidators::isEmail('a@b.com'));
        $this->assertTrue(FormatValidators::isUrl('https://x.com'));
        $this->assertTrue(FormatValidators::isUuid('550e8400-e29b-41d4-a716-446655440000'));
        $this->assertTrue(FormatValidators::isIpv4('127.0.0.1'));
        $this->assertTrue(FormatValidators::isIpv6('::1'));
        $this->assertTrue(FormatValidators::isDate('2024-01-01'));
        $this->assertTrue(FormatValidators::isDateTime('2024-01-01T00:00:00Z'));
    }
}

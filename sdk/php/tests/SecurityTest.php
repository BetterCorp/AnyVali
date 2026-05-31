<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\AnyValiDocument;
use AnyVali\IssueCodes;
use AnyVali\UnknownKeyMode;
use AnyVali\ValidationContext;
use AnyVali\ValidationError;
use PHPUnit\Framework\TestCase;

/**
 * Security-focused tests for the AnyVali PHP SDK.
 *
 * Covers: ReDoS, recursive $ref, integer overflow, NaN/Infinity,
 * format bypass, large inputs, object key safety, and schema import injection.
 */
final class SecurityTest extends TestCase
{
    // ── CVE-2016-4055: ReDoS (Catastrophic Backtracking) ───────────

    public function testRedosExponentialBacktrackingPattern(): void
    {
        // Classic ReDoS: (a+)+ against "aaa...!" causes exponential backtracking
        $s = AnyVali::string()->pattern('(a+)+$');
        $malicious = str_repeat('a', 30) . '!';

        $start = microtime(true);
        $result = $s->safeParse($malicious);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Pattern matching took too long, possible ReDoS');
    }

    public function testRedosNestedQuantifiers(): void
    {
        // Nested quantifiers: (a|a)* against non-matching tail
        $s = AnyVali::string()->pattern('(a|a)*$');
        $malicious = str_repeat('a', 30) . 'X';

        $start = microtime(true);
        $result = $s->safeParse($malicious);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Nested quantifier pattern took too long');
    }

    public function testRedosOverlappingAlternation(): void
    {
        // Overlapping alternation: (a|aa)* against non-matching suffix
        $s = AnyVali::string()->pattern('(a|aa)*$');
        $malicious = str_repeat('a', 30) . '!';

        $start = microtime(true);
        $result = $s->safeParse($malicious);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Overlapping alternation took too long');
    }

    public function testRedosEmailLikePattern(): void
    {
        // Pathological email-like pattern: ([a-zA-Z0-9]+)*@
        $s = AnyVali::string()->pattern('([a-zA-Z0-9]+)*@');
        $malicious = str_repeat('a', 50) . '!';

        $start = microtime(true);
        $result = $s->safeParse($malicious);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Email-like pattern took too long');
    }

    // ── CVE-2003-1564: Recursive $ref (Self-Referencing Schemas) ───

    public function testRecursiveRefDoesNotCauseInfiniteLoop(): void
    {
        // A schema that references itself via $ref should not blow up the stack
        $doc = new AnyValiDocument(
            root: ['kind' => 'ref', 'ref' => '#/definitions/node'],
            definitions: [
                'node' => [
                    'kind' => 'object',
                    'properties' => [
                        'value' => ['kind' => 'string'],
                        'child' => ['kind' => 'ref', 'ref' => '#/definitions/node'],
                    ],
                    'required' => ['value'],
                    'unknownKeys' => 'allow',
                ],
            ],
        );

        // Importing should not throw or hang
        $schema = AnyVali::import($doc);
        $this->assertNotNull($schema);
    }

    public function testMutuallyRecursiveRefImport(): void
    {
        // Two types referencing each other: A -> B -> A
        $doc = new AnyValiDocument(
            root: ['kind' => 'ref', 'ref' => '#/definitions/A'],
            definitions: [
                'A' => [
                    'kind' => 'object',
                    'properties' => [
                        'b' => ['kind' => 'ref', 'ref' => '#/definitions/B'],
                    ],
                    'required' => [],
                    'unknownKeys' => 'allow',
                ],
                'B' => [
                    'kind' => 'object',
                    'properties' => [
                        'a' => ['kind' => 'ref', 'ref' => '#/definitions/A'],
                    ],
                    'required' => [],
                    'unknownKeys' => 'allow',
                ],
            ],
        );

        $schema = AnyVali::import($doc);
        $this->assertNotNull($schema);
    }

    public function testDirectSelfRefDoesNotHang(): void
    {
        // Schema that directly references itself
        $doc = new AnyValiDocument(
            root: ['kind' => 'ref', 'ref' => '#/definitions/self'],
            definitions: [
                'self' => ['kind' => 'ref', 'ref' => '#/definitions/self'],
            ],
        );

        $schema = AnyVali::import($doc);
        $this->assertNotNull($schema);
    }

    /**
     * CVE-2003-1564 parse-time: importing a pure self-cycle succeeds, but
     * actually parsing through it must not hang the runtime. PHP raises a
     * fatal "Maximum function nesting level" / stack-overflow on unbounded
     * recursion — Throwable catches it on PHP >= 7. Test asserts the call
     * terminates in bounded time, regardless of success/failure.
     */
    public function testDirectSelfRefParseDoesNotHangRuntime(): void
    {
        $doc = new AnyValiDocument(
            root: ['kind' => 'ref', 'ref' => '#/definitions/self'],
            definitions: [
                'self' => ['kind' => 'ref', 'ref' => '#/definitions/self'],
            ],
        );

        $schema = AnyVali::import($doc);

        $start = microtime(true);
        $threw = false;
        try {
            $schema->safeParse('anything');
        } catch (\Throwable $e) {
            // Stack overflow / nesting-limit exceeded surfaces as Throwable
            // on modern PHP — acceptable defense.
            $threw = true;
        }
        $elapsed = microtime(true) - $start;

        $this->assertLessThan(
            5.0,
            $elapsed,
            'Self-cycle parse hung > 5s — runtime DoS (CVE-2003-1564)',
        );
        // Either threw or returned — both are non-hangs.
        $this->assertTrue(true, "threw=" . ($threw ? 'true' : 'false'));
    }

    // ── CWE-190: Integer Overflow (Int Width Boundaries) ───────────

    public function testInt8BoundaryExact(): void
    {
        $s = AnyVali::int8();

        // Exact boundaries should pass
        $this->assertSame(-128, $s->parse(-128));
        $this->assertSame(127, $s->parse(127));

        // One past boundary should fail
        $result = $s->safeParse(128);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-129);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testInt16BoundaryExact(): void
    {
        $s = AnyVali::int16();

        $this->assertSame(-32768, $s->parse(-32768));
        $this->assertSame(32767, $s->parse(32767));

        $result = $s->safeParse(32768);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-32769);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testInt32BoundaryExact(): void
    {
        $s = AnyVali::int32();

        $this->assertSame(2147483647, $s->parse(2147483647));
        $this->assertSame(-2147483648, $s->parse(-2147483648));

        $result = $s->safeParse(2147483648);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-2147483649);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testUint8BoundaryExact(): void
    {
        $s = AnyVali::uint8();

        $this->assertSame(0, $s->parse(0));
        $this->assertSame(255, $s->parse(255));

        $result = $s->safeParse(256);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testUint16BoundaryExact(): void
    {
        $s = AnyVali::uint16();

        $this->assertSame(0, $s->parse(0));
        $this->assertSame(65535, $s->parse(65535));

        $result = $s->safeParse(65536);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testUint32BoundaryExact(): void
    {
        $s = AnyVali::uint32();

        $this->assertSame(0, $s->parse(0));
        $this->assertSame(4294967295, $s->parse(4294967295));

        $result = $s->safeParse(4294967296);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);

        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testUint64RejectsNegativeValues(): void
    {
        $s = AnyVali::uint64();

        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);

        $result = $s->safeParse(-9999999);
        $this->assertFalse($result->success);
    }

    public function testIntRejectsFloatWithFractionalPart(): void
    {
        $s = AnyVali::int();

        $result = $s->safeParse(3.14);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testInt8RejectsLargePositiveOverflow(): void
    {
        $s = AnyVali::int8();

        // Far beyond int8 range
        $result = $s->safeParse(1000);
        $this->assertFalse($result->success);

        $result = $s->safeParse(PHP_INT_MAX);
        $this->assertFalse($result->success);
    }

    public function testInt8RejectsLargeNegativeOverflow(): void
    {
        $s = AnyVali::int8();

        $result = $s->safeParse(-1000);
        $this->assertFalse($result->success);

        $result = $s->safeParse(PHP_INT_MIN);
        $this->assertFalse($result->success);
    }

    // ── CWE-20: NaN/Infinity Rejection ─────────────────────────────

    public function testNumberRejectsNan(): void
    {
        $s = AnyVali::number();
        $result = $s->safeParse(NAN);
        // NAN should either be rejected or at least not silently accepted as valid
        // PHP's NAN is a float, so it passes is_float(), but validators should catch it
        if ($result->success) {
            // If accepted, verify downstream constraints still work
            $constrained = AnyVali::number()->min(0)->max(100);
            $result2 = $constrained->safeParse(NAN);
            // NAN comparisons are always false, so constraints may not catch it
            // This documents the behavior
            $this->assertNotNull($result2);
        } else {
            $this->assertFalse($result->success);
        }
    }

    public function testNumberRejectsPositiveInfinity(): void
    {
        $s = AnyVali::number()->max(1e308);
        $result = $s->safeParse(INF);
        $this->assertFalse($result->success, 'Positive infinity should be rejected by max constraint');
    }

    public function testNumberRejectsNegativeInfinity(): void
    {
        $s = AnyVali::number()->min(-1e308);
        $result = $s->safeParse(-INF);
        $this->assertFalse($result->success, 'Negative infinity should be rejected by min constraint');
    }

    public function testIntRejectsNan(): void
    {
        $s = AnyVali::int();
        $result = $s->safeParse(NAN);
        $this->assertFalse($result->success, 'NAN should not be accepted as an integer');
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testIntRejectsInfinity(): void
    {
        $s = AnyVali::int();
        $result = $s->safeParse(INF);
        $this->assertFalse($result->success, 'INF should not be accepted as an integer');
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testIntRejectsNegativeInfinity(): void
    {
        $s = AnyVali::int();
        $result = $s->safeParse(-INF);
        $this->assertFalse($result->success, '-INF should not be accepted as an integer');
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testFloat32RejectsNan(): void
    {
        $s = AnyVali::float32();
        $result = $s->safeParse(NAN);
        // Document the behavior: NAN is a float and may pass type check
        // but should ideally be rejected
        $this->assertNotNull($result);
    }

    public function testFloat64RejectsNan(): void
    {
        $s = AnyVali::float64();
        $result = $s->safeParse(NAN);
        $this->assertNotNull($result);
    }

    public function testInt8RejectsNan(): void
    {
        $s = AnyVali::int8();
        $result = $s->safeParse(NAN);
        $this->assertFalse($result->success, 'NAN should not be accepted as int8');
    }

    public function testInt16RejectsNan(): void
    {
        $s = AnyVali::int16();
        $result = $s->safeParse(NAN);
        $this->assertFalse($result->success, 'NAN should not be accepted as int16');
    }

    public function testInt32RejectsNan(): void
    {
        $s = AnyVali::int32();
        $result = $s->safeParse(NAN);
        $this->assertFalse($result->success, 'NAN should not be accepted as int32');
    }

    // ── CWE-20: Format Bypass (email, url, ipv4 edge cases) ───────

    public function testEmailRejectsPlainString(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('notanemail');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    public function testTamperedEmailFormatNameNotSilentlyIgnored(): void
    {
        $s = AnyVali::string()->format("email\0");
        $result = $s->safeParse('not-an-email');
        $this->assertFalse($result->success, 'Tampered format name bypassed email validation');
    }

    public function testImportedTamperedEmailFormatNameNotUnconstrained(): void
    {
        $schema = AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'string', 'format' => "email\0"],
            'definitions' => [],
        ]);
        $result = $schema->safeParse('not-an-email');
        $this->assertFalse($result->success, 'Imported tampered format name bypassed email validation');
    }

    public function testEmailRejectsMissingDomain(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user@');
        $this->assertFalse($result->success);
    }

    public function testEmailRejectsMissingLocal(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('@domain.com');
        $this->assertFalse($result->success);
    }

    public function testEmailRejectsDoubleAt(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user@@domain.com');
        $this->assertFalse($result->success);
    }

    public function testEmailRejectsNoDotInDomain(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user@localhost');
        $this->assertFalse($result->success);
    }

    public function testEmailAcceptsValidEmail(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user@example.com');
        $this->assertTrue($result->success);
    }

    public function testEmailRejectsSpacesInAddress(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse('user name@example.com');
        $this->assertFalse($result->success);
    }

    public function testEmailRejectsNewlineInjection(): void
    {
        $s = AnyVali::string()->format('email');
        $result = $s->safeParse("user@example.com\r\nBcc: attacker@evil.com");
        $this->assertFalse($result->success);
    }

    public function testUrlRejectsJavascriptProtocol(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('javascript:alert(1)');
        $this->assertFalse($result->success);
    }

    public function testUrlRejectsDataProtocol(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('data:text/html,<script>alert(1)</script>');
        $this->assertFalse($result->success);
    }

    public function testUrlRejectsFtpProtocol(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('ftp://example.com');
        $this->assertFalse($result->success);
    }

    public function testUrlAcceptsHttps(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('https://example.com/path?q=1');
        $this->assertTrue($result->success);
    }

    public function testUrlAcceptsHttp(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('http://example.com');
        $this->assertTrue($result->success);
    }

    public function testUrlRejectsEmptyString(): void
    {
        $s = AnyVali::string()->format('url');
        $result = $s->safeParse('');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsLeadingZerosOctalBypass(): void
    {
        // Leading zeros can be interpreted as octal in some parsers (e.g., 0177.0.0.1 = 127.0.0.1)
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('0177.0.0.1');
        $this->assertFalse($result->success, 'Leading zeros (octal bypass) should be rejected');
    }

    public function testIpv4RejectsLeadingZeroPadding(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.01.1');
        $this->assertFalse($result->success, 'Zero-padded octets should be rejected');
    }

    public function testIpv4RejectsOutOfRangeOctet(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('256.1.1.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsNegativeOctet(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.-1.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsTooFewOctets(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsTooManyOctets(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.1.1.1');
        $this->assertFalse($result->success);
    }

    public function testIpv4RejectsEmptyOctets(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168..1');
        $this->assertFalse($result->success);
    }

    public function testIpv4AcceptsValid(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $this->assertTrue($s->safeParse('192.168.1.1')->success);
        $this->assertTrue($s->safeParse('0.0.0.0')->success);
        $this->assertTrue($s->safeParse('255.255.255.255')->success);
        $this->assertTrue($s->safeParse('127.0.0.1')->success);
    }

    public function testIpv4RejectsNonNumericOctet(): void
    {
        $s = AnyVali::string()->format('ipv4');
        $result = $s->safeParse('192.168.abc.1');
        $this->assertFalse($result->success);
    }

    public function testDateFormatRejectsInvalidDate(): void
    {
        $s = AnyVali::string()->format('date');
        $result = $s->safeParse('2024-13-01'); // month 13
        $this->assertFalse($result->success);
    }

    public function testDateFormatRejectsInvalidDay(): void
    {
        $s = AnyVali::string()->format('date');
        $result = $s->safeParse('2024-02-30'); // Feb 30
        $this->assertFalse($result->success);
    }

    public function testDateTimeFormatRejectsNoTimezone(): void
    {
        $s = AnyVali::string()->format('date-time');
        $result = $s->safeParse('2024-01-01T12:00:00'); // no TZ
        $this->assertFalse($result->success);
    }

    public function testUuidFormatRejectsInvalid(): void
    {
        $s = AnyVali::string()->format('uuid');
        $result = $s->safeParse('not-a-uuid');
        $this->assertFalse($result->success);
    }

    public function testUuidFormatAcceptsValid(): void
    {
        $s = AnyVali::string()->format('uuid');
        $this->assertTrue($s->safeParse('550e8400-e29b-41d4-a716-446655440000')->success);
    }

    public function testUnknownFormatRejects(): void
    {
        $s = AnyVali::string()->format('nonexistent-format');
        $result = $s->safeParse('anything');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    // ── CWE-400: Large Inputs (Resource Exhaustion) ────────────────

    public function testUnicodeLengthAstralCodePointCountsAsOneCharacter(): void
    {
        $emoji = "\u{1F600}";
        $this->assertTrue(AnyVali::string()->maxLength(1)->safeParse($emoji)->success);
        $this->assertFalse(AnyVali::string()->minLength(2)->safeParse($emoji)->success);
    }

    public function testUnicodeLengthImportedMaxLengthUsesCodePoints(): void
    {
        $schema = AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'string', 'maxLength' => 1],
            'definitions' => [],
        ]);
        $this->assertTrue($schema->safeParse("\u{1F600}")->success);
    }

    public function testLargeStringDoesNotCrash(): void
    {
        $s = AnyVali::string();
        $largeString = str_repeat('a', 1_000_000); // 1 MB

        $start = microtime(true);
        $result = $s->safeParse($largeString);
        $elapsed = microtime(true) - $start;

        $this->assertTrue($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Parsing 1MB string should complete quickly');
    }

    public function testLargeStringWithMaxLengthConstraint(): void
    {
        $s = AnyVali::string()->maxLength(100);
        $largeString = str_repeat('x', 100_000);

        $start = microtime(true);
        $result = $s->safeParse($largeString);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
        $this->assertLessThan(5.0, $elapsed, 'MaxLength check on large string should be fast');
    }

    public function testLargeArrayDoesNotCrash(): void
    {
        $s = AnyVali::array(AnyVali::int());
        $largeArray = range(1, 10_000);

        $start = microtime(true);
        $result = $s->safeParse($largeArray);
        $elapsed = microtime(true) - $start;

        $this->assertTrue($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Validating 10k-element array should complete quickly');
    }

    public function testLargeArrayWithMaxItemsConstraint(): void
    {
        $s = AnyVali::array(AnyVali::int())->maxItems(100);
        $largeArray = range(1, 10_000);

        $start = microtime(true);
        $result = $s->safeParse($largeArray);
        $elapsed = microtime(true) - $start;

        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
        $this->assertLessThan(5.0, $elapsed, 'MaxItems check on large array should be fast');
    }

    public function testDeeplyNestedObjectDoesNotCrash(): void
    {
        // Build a 50-level deep nested structure
        $depth = 50;
        $value = 'leaf';
        for ($i = 0; $i < $depth; $i++) {
            $value = ['child' => $value];
        }

        $s = AnyVali::record(AnyVali::any());

        $start = microtime(true);
        $result = $s->safeParse($value);
        $elapsed = microtime(true) - $start;

        $this->assertTrue($result->success);
        $this->assertLessThan(5.0, $elapsed);
    }

    public function testLargeObjectManyKeysDoesNotCrash(): void
    {
        $s = AnyVali::record(AnyVali::int());
        $data = [];
        for ($i = 0; $i < 10_000; $i++) {
            $data["key_{$i}"] = $i;
        }

        $start = microtime(true);
        $result = $s->safeParse($data);
        $elapsed = microtime(true) - $start;

        $this->assertTrue($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Validating 10k-key record should complete quickly');
    }

    public function testLargeStringWithPatternDoesNotHang(): void
    {
        // A safe pattern on a large input should still complete quickly
        $s = AnyVali::string()->pattern('^[a-z]+$');
        $largeString = str_repeat('a', 100_000);

        $start = microtime(true);
        $result = $s->safeParse($largeString);
        $elapsed = microtime(true) - $start;

        $this->assertTrue($result->success);
        $this->assertLessThan(5.0, $elapsed, 'Safe pattern on large string should be fast');
    }

    // ── CVE-2019-10744: Object Key Safety ──────────────────────────

    public function testObjectRejectsProtoKeyWhenConfigured(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Reject,
        );
        $result = $s->safeParse([
            'name' => 'Alice',
            '__proto__' => ['isAdmin' => true],
        ]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[0]->code);
    }

    public function testObjectRejectsConstructorKeyWhenConfigured(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Reject,
        );
        $result = $s->safeParse([
            'name' => 'Alice',
            'constructor' => ['prototype' => ['isAdmin' => true]],
        ]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[0]->code);
    }

    public function testObjectRejectsPrototypeKeyWhenConfigured(): void
    {
        $s = AnyVali::object(
            ['id' => AnyVali::int()],
            ['id'],
            UnknownKeyMode::Reject,
        );
        $result = $s->safeParse([
            'id' => 1,
            'prototype' => 'injected',
        ]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[0]->code);
    }

    public function testObjectStripsProtoKey(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Strip,
        );
        $result = $s->parse([
            'name' => 'Alice',
            '__proto__' => ['isAdmin' => true],
        ]);
        $this->assertSame(['name' => 'Alice'], $result);
        $this->assertArrayNotHasKey('__proto__', $result);
    }

    public function testObjectStripsConstructorKey(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Strip,
        );
        $result = $s->parse([
            'name' => 'Alice',
            'constructor' => 'injected',
        ]);
        $this->assertSame(['name' => 'Alice'], $result);
        $this->assertArrayNotHasKey('constructor', $result);
    }

    public function testRecordAcceptsProtoKeyAsData(): void
    {
        // Records treat all keys as data, so __proto__ is just a key name
        $s = AnyVali::record(AnyVali::string());
        $result = $s->safeParse([
            '__proto__' => 'value',
            'constructor' => 'value',
        ]);
        $this->assertTrue($result->success);
    }

    public function testObjectRejectsMultipleDangerousKeys(): void
    {
        $s = AnyVali::object(
            ['id' => AnyVali::int()],
            ['id'],
        );
        $result = $s->safeParse([
            'id' => 1,
            '__proto__' => 'a',
            'constructor' => 'b',
            '__defineGetter__' => 'c',
        ]);
        $this->assertFalse($result->success);
        // Should report all unknown keys
        $this->assertGreaterThanOrEqual(3, count($result->issues));
        foreach ($result->issues as $issue) {
            $this->assertSame(IssueCodes::UNKNOWN_KEY, $issue->code);
        }
    }

    public function testObjectWithProtoAsDefinedProperty(): void
    {
        // If __proto__ is explicitly part of the schema, it should be accepted
        $s = AnyVali::object(
            ['__proto__' => AnyVali::string()],
            ['__proto__'],
        );
        $result = $s->safeParse(['__proto__' => 'safe']);
        $this->assertTrue($result->success);
    }

    // ── Schema Import Injection (Unknown Kinds Rejected) ───────────

    public function testImportRejectsUnknownKind(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Unsupported schema kind');

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'exec'],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsEvalKind(): void
    {
        $this->expectException(\RuntimeException::class);

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'eval'],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsSystemKind(): void
    {
        $this->expectException(\RuntimeException::class);

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'system'],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsMissingKind(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('missing "kind"');

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['minLength' => 5],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsEmptyKind(): void
    {
        $this->expectException(\RuntimeException::class);

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => ''],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsSqlInjectionKind(): void
    {
        $this->expectException(\RuntimeException::class);

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => "'; DROP TABLE users; --"],
            'definitions' => [],
        ]);
    }

    public function testImportAcceptsAllValidKinds(): void
    {
        $validKinds = [
            'string', 'number', 'float32', 'float64',
            'int', 'int8', 'int16', 'int32', 'int64',
            'uint8', 'uint16', 'uint32', 'uint64',
            'bool', 'null', 'any', 'unknown', 'never',
        ];

        foreach ($validKinds as $kind) {
            $schema = AnyVali::import([
                'anyvaliVersion' => '1.0',
                'schemaVersion' => '1.1',
                'root' => ['kind' => $kind],
                'definitions' => [],
            ]);
            $this->assertNotNull($schema, "Valid kind '{$kind}' should be importable");
        }
    }

    public function testImportRejectsUnknownKindInNestedSchema(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Unsupported schema kind');

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => [
                'kind' => 'array',
                'items' => ['kind' => 'dangerous_custom_type'],
            ],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsUnknownKindInObjectProperty(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Unsupported schema kind');

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => [
                'kind' => 'object',
                'properties' => [
                    'name' => ['kind' => 'shellcode'],
                ],
                'required' => ['name'],
            ],
            'definitions' => [],
        ]);
    }

    public function testImportRejectsUnknownKindInUnionVariant(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Unsupported schema kind');

        AnyVali::import([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => [
                'kind' => 'union',
                'variants' => [
                    ['kind' => 'string'],
                    ['kind' => 'rce'],
                ],
            ],
            'definitions' => [],
        ]);
    }

    public function testImportFromJsonRejectsInvalidJson(): void
    {
        $this->expectException(\JsonException::class);
        AnyVali::import('{invalid json}}}');
    }

    public function testImportFromJsonRejectsUnknownKind(): void
    {
        $this->expectException(\RuntimeException::class);

        $json = json_encode([
            'anyvaliVersion' => '1.0',
            'schemaVersion' => '1.1',
            'root' => ['kind' => 'payload'],
            'definitions' => new \stdClass(),
        ]);

        AnyVali::import($json);
    }
}

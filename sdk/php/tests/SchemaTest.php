<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\IssueCodes;
use AnyVali\ParseResult;
use AnyVali\Schema;
use AnyVali\UnknownKeyMode;
use AnyVali\ValidationError;
use AnyVali\ValidationIssue;
use PHPUnit\Framework\TestCase;

final class SchemaTest extends TestCase
{
    // ── String ──────────────────────────────────────────

    public function testStringAcceptsValidStrings(): void
    {
        $s = AnyVali::string();
        $this->assertSame('hello', $s->parse('hello'));
        $this->assertSame('', $s->parse(''));
        $this->assertSame("\u{00e9}\u{00e0}\u{00fc}\u{00f1}\u{00f6}", $s->parse("\u{00e9}\u{00e0}\u{00fc}\u{00f1}\u{00f6}"));
        $this->assertSame("line1\nline2\ttab", $s->parse("line1\nline2\ttab"));
    }

    public function testStringRejectsNonStrings(): void
    {
        $s = AnyVali::string();

        $result = $s->safeParse(42);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('string', $result->issues[0]->expected);
        $this->assertSame('number', $result->issues[0]->received);

        $result = $s->safeParse(true);
        $this->assertSame('boolean', $result->issues[0]->received);

        $result = $s->safeParse(null);
        $this->assertSame('null', $result->issues[0]->received);

        $result = $s->safeParse(['a', 'b']);
        $this->assertSame('array', $result->issues[0]->received);

        $result = $s->safeParse(['key' => 'value']);
        $this->assertSame('object', $result->issues[0]->received);
    }

    public function testStringParseThrowsOnInvalid(): void
    {
        $this->expectException(ValidationError::class);
        AnyVali::string()->parse(42);
    }

    public function testStringMinLength(): void
    {
        $s = AnyVali::string()->minLength(3);
        $this->assertSame('abc', $s->parse('abc'));
        $this->assertSame('abcd', $s->parse('abcd'));

        $result = $s->safeParse('ab');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
        $this->assertSame('3', $result->issues[0]->expected);
        $this->assertSame('2', $result->issues[0]->received);
    }

    public function testStringMaxLength(): void
    {
        $s = AnyVali::string()->maxLength(5);
        $this->assertSame('hello', $s->parse('hello'));

        $result = $s->safeParse('hello!');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
        $this->assertSame('5', $result->issues[0]->expected);
        $this->assertSame('6', $result->issues[0]->received);
    }

    public function testStringPattern(): void
    {
        $s = AnyVali::string()->pattern('^[a-z]+$');
        $this->assertSame('abc', $s->parse('abc'));

        $result = $s->safeParse('ABC');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    public function testStringStartsWith(): void
    {
        $s = AnyVali::string()->startsWith('hello');
        $this->assertSame('hello world', $s->parse('hello world'));

        $result = $s->safeParse('world hello');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    public function testStringEndsWith(): void
    {
        $s = AnyVali::string()->endsWith('.json');
        $this->assertSame('file.json', $s->parse('file.json'));

        $result = $s->safeParse('file.xml');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    public function testStringIncludes(): void
    {
        $s = AnyVali::string()->includes('world');
        $this->assertSame('hello world!', $s->parse('hello world!'));

        $result = $s->safeParse('hello there');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_STRING, $result->issues[0]->code);
    }

    // ── Number ──────────────────────────────────────────

    public function testNumberAcceptsValidNumbers(): void
    {
        $s = AnyVali::number();
        $this->assertSame(42, $s->parse(42));
        $this->assertSame(0, $s->parse(0));
        $this->assertSame(-3.14, $s->parse(-3.14));
        $this->assertSame(1.7976931348623157e+308, $s->parse(1.7976931348623157e+308));
        $this->assertSame(5e-324, $s->parse(5e-324));
    }

    public function testNumberRejectsNonNumbers(): void
    {
        $s = AnyVali::number();

        $result = $s->safeParse('42');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('string', $result->issues[0]->received);

        $result = $s->safeParse(true);
        $this->assertSame('boolean', $result->issues[0]->received);

        $result = $s->safeParse(null);
        $this->assertSame('null', $result->issues[0]->received);

        $result = $s->safeParse([]);
        $this->assertSame('array', $result->issues[0]->received);

        $result = $s->safeParse(['key' => 'val']);
        $this->assertSame('object', $result->issues[0]->received);
    }

    public function testNumberMin(): void
    {
        $s = AnyVali::number()->min(10);
        $this->assertSame(10, $s->parse(10));
        $this->assertSame(11, $s->parse(11));

        $result = $s->safeParse(9);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testNumberMax(): void
    {
        $s = AnyVali::number()->max(100);
        $this->assertSame(100, $s->parse(100));

        $result = $s->safeParse(101);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testNumberExclusiveMin(): void
    {
        $s = AnyVali::number()->exclusiveMin(0);
        $this->assertSame(0.001, $s->parse(0.001));

        $result = $s->safeParse(0);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testNumberExclusiveMax(): void
    {
        $s = AnyVali::number()->exclusiveMax(100);
        $this->assertSame(99.999, $s->parse(99.999));

        $result = $s->safeParse(100);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testNumberMultipleOf(): void
    {
        $s = AnyVali::number()->multipleOf(3);
        $this->assertSame(9, $s->parse(9));

        $result = $s->safeParse(10);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_NUMBER, $result->issues[0]->code);
    }

    public function testNumberMultipleOfFloat(): void
    {
        $s = AnyVali::number()->multipleOf(0.5);
        $this->assertSame(2.5, $s->parse(2.5));
    }

    // ── Int ──────────────────────────────────────────

    public function testIntAcceptsValidIntegers(): void
    {
        $s = AnyVali::int();
        $this->assertSame(42, $s->parse(42));
        $this->assertSame(0, $s->parse(0));
        $this->assertSame(-100, $s->parse(-100));
    }

    public function testIntRejectsFloat(): void
    {
        $s = AnyVali::int();
        $result = $s->safeParse(3.14);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('int', $result->issues[0]->expected);
        $this->assertSame('number', $result->issues[0]->received);
    }

    public function testIntRejectsString(): void
    {
        $s = AnyVali::int();
        $result = $s->safeParse('42');
        $this->assertFalse($result->success);
        $this->assertSame('string', $result->issues[0]->received);
    }

    public function testIntWithConstraints(): void
    {
        $s = AnyVali::int()->min(1)->max(10);
        $this->assertSame(5, $s->parse(5));

        $result = $s->safeParse(0);
        $this->assertFalse($result->success);
    }

    // ── Int Widths ──────────────────────────────────────

    public function testInt8Range(): void
    {
        $s = AnyVali::int8();
        $this->assertSame(127, $s->parse(127));
        $this->assertSame(-128, $s->parse(-128));

        $result = $s->safeParse(128);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
        $this->assertSame('int8', $result->issues[0]->expected);

        $result = $s->safeParse(-129);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testInt16Range(): void
    {
        $s = AnyVali::int16();
        $this->assertSame(32767, $s->parse(32767));

        $result = $s->safeParse(32768);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testInt32Range(): void
    {
        $s = AnyVali::int32();
        $this->assertSame(2147483647, $s->parse(2147483647));

        $result = $s->safeParse(2147483648);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testUint8Range(): void
    {
        $s = AnyVali::uint8();
        $this->assertSame(0, $s->parse(0));
        $this->assertSame(255, $s->parse(255));

        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);

        $result = $s->safeParse(256);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testUint16Range(): void
    {
        $s = AnyVali::uint16();
        $this->assertSame(65535, $s->parse(65535));
    }

    public function testUint32Range(): void
    {
        $s = AnyVali::uint32();
        $this->assertSame(4294967295, $s->parse(4294967295));
    }

    public function testUint64RejectsNegative(): void
    {
        $s = AnyVali::uint64();
        $result = $s->safeParse(-1);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    // ── Float Widths ──────────────────────────────────────

    public function testFloat64AcceptsFloats(): void
    {
        $s = AnyVali::float64();
        $this->assertSame(3.141592653589793, $s->parse(3.141592653589793));
        $this->assertSame(42, $s->parse(42));
    }

    public function testFloat64RejectsString(): void
    {
        $s = AnyVali::float64();
        $result = $s->safeParse('3.14');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testFloat32AcceptsValue(): void
    {
        $s = AnyVali::float32();
        $this->assertSame(1.5, $s->parse(1.5));

        $result = $s->safeParse(true);
        $this->assertFalse($result->success);
    }

    // ── Bool ──────────────────────────────────────────

    public function testBoolAcceptsBooleans(): void
    {
        $s = AnyVali::bool();
        $this->assertTrue($s->parse(true));
        $this->assertFalse($s->parse(false));
    }

    public function testBoolRejectsNonBooleans(): void
    {
        $s = AnyVali::bool();

        $result = $s->safeParse(1);
        $this->assertFalse($result->success);
        $this->assertSame('number', $result->issues[0]->received);

        $result = $s->safeParse('true');
        $this->assertSame('string', $result->issues[0]->received);

        $result = $s->safeParse(null);
        $this->assertSame('null', $result->issues[0]->received);
    }

    // ── Null ──────────────────────────────────────────

    public function testNullAcceptsNull(): void
    {
        $s = AnyVali::null();
        $this->assertNull($s->parse(null));
    }

    public function testNullRejectsNonNull(): void
    {
        $s = AnyVali::null();

        $result = $s->safeParse('null');
        $this->assertFalse($result->success);

        $result = $s->safeParse(0);
        $this->assertFalse($result->success);

        $result = $s->safeParse(false);
        $this->assertFalse($result->success);

        $result = $s->safeParse('');
        $this->assertFalse($result->success);
    }

    // ── Any ──────────────────────────────────────────

    public function testAnyAcceptsAnything(): void
    {
        $s = AnyVali::any();
        $this->assertSame('hello', $s->parse('hello'));
        $this->assertSame(42, $s->parse(42));
        $this->assertNull($s->parse(null));
        $this->assertSame(['key' => 'value'], $s->parse(['key' => 'value']));
        $this->assertSame([1, 'two', true], $s->parse([1, 'two', true]));
    }

    // ── Unknown ──────────────────────────────────────

    public function testUnknownAcceptsAnything(): void
    {
        $s = AnyVali::unknown();
        $this->assertSame('hello', $s->parse('hello'));
        $this->assertSame(99, $s->parse(99));
        $this->assertNull($s->parse(null));
        $this->assertFalse($s->parse(false));
    }

    // ── Never ──────────────────────────────────────

    public function testNeverRejectsEverything(): void
    {
        $s = AnyVali::never();

        $result = $s->safeParse('hello');
        $this->assertFalse($result->success);
        $this->assertSame('never', $result->issues[0]->expected);

        $result = $s->safeParse(0);
        $this->assertFalse($result->success);

        $result = $s->safeParse(null);
        $this->assertFalse($result->success);

        $result = $s->safeParse(true);
        $this->assertFalse($result->success);

        $result = $s->safeParse([]);
        $this->assertFalse($result->success);
    }

    // ── Literal ──────────────────────────────────────

    public function testLiteralMatchesExactValue(): void
    {
        $this->assertSame('hello', AnyVali::literal('hello')->parse('hello'));
        $this->assertSame(42, AnyVali::literal(42)->parse(42));
        $this->assertTrue(AnyVali::literal(true)->parse(true));
        $this->assertNull(AnyVali::literal(null)->parse(null));
    }

    public function testLiteralRejectsNonMatch(): void
    {
        $result = AnyVali::literal('hello')->safeParse('world');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_LITERAL, $result->issues[0]->code);
    }

    public function testLiteralRejectsWrongType(): void
    {
        $result = AnyVali::literal(42)->safeParse('42');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_LITERAL, $result->issues[0]->code);
    }

    // ── Enum ──────────────────────────────────────

    public function testEnumAcceptsValidValues(): void
    {
        $s = AnyVali::enum(['red', 'green', 'blue']);
        $this->assertSame('red', $s->parse('red'));
        $this->assertSame('blue', $s->parse('blue'));
    }

    public function testEnumRejectsInvalidValue(): void
    {
        $s = AnyVali::enum(['red', 'green', 'blue']);
        $result = $s->safeParse('yellow');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('enum(red,green,blue)', $result->issues[0]->expected);
    }

    public function testEnumWithNumericValues(): void
    {
        $s = AnyVali::enum([1, 2, 3]);
        $this->assertSame(2, $s->parse(2));

        // Strict type check
        $result = $s->safeParse('1');
        $this->assertFalse($result->success);
    }

    // ── Array ──────────────────────────────────────

    public function testArrayAcceptsValidElements(): void
    {
        $s = AnyVali::array(AnyVali::string());
        $this->assertSame(['a', 'b', 'c'], $s->parse(['a', 'b', 'c']));
    }

    public function testArrayAcceptsEmpty(): void
    {
        $s = AnyVali::array(AnyVali::int());
        $this->assertSame([], $s->parse([]));
    }

    public function testArrayRejectsNonArray(): void
    {
        $s = AnyVali::array(AnyVali::string());
        $result = $s->safeParse('not an array');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('array', $result->issues[0]->expected);
    }

    public function testArrayRejectsInvalidElement(): void
    {
        $s = AnyVali::array(AnyVali::int());
        $result = $s->safeParse([1, 2, 'three']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame([2], $result->issues[0]->path);
    }

    public function testArrayReportsMultipleInvalidElements(): void
    {
        $s = AnyVali::array(AnyVali::bool());
        $result = $s->safeParse([true, 'yes', false, 1]);
        $this->assertFalse($result->success);
        $this->assertCount(2, $result->issues);
        $this->assertSame([1], $result->issues[0]->path);
        $this->assertSame([3], $result->issues[1]->path);
    }

    public function testArrayMinItems(): void
    {
        $s = AnyVali::array(AnyVali::int())->minItems(2);
        $this->assertSame([1, 2], $s->parse([1, 2]));

        $result = $s->safeParse([1]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testArrayMaxItems(): void
    {
        $s = AnyVali::array(AnyVali::string())->maxItems(3);
        $this->assertSame(['a', 'b', 'c'], $s->parse(['a', 'b', 'c']));

        $result = $s->safeParse(['a', 'b', 'c', 'd']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    // ── Tuple ──────────────────────────────────────

    public function testTupleAcceptsValid(): void
    {
        $s = AnyVali::tuple([AnyVali::string(), AnyVali::int()]);
        $this->assertSame(['hello', 42], $s->parse(['hello', 42]));
    }

    public function testTupleRejectsTooFew(): void
    {
        $s = AnyVali::tuple([AnyVali::string(), AnyVali::int()]);
        $result = $s->safeParse(['hello']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    public function testTupleRejectsTooMany(): void
    {
        $s = AnyVali::tuple([AnyVali::string(), AnyVali::int()]);
        $result = $s->safeParse(['hello', 42, true]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    public function testTupleRejectsWrongType(): void
    {
        $s = AnyVali::tuple([AnyVali::string(), AnyVali::int()]);
        $result = $s->safeParse([42, 'hello']);
        $this->assertFalse($result->success);
        $this->assertCount(2, $result->issues);
        $this->assertSame([0], $result->issues[0]->path);
        $this->assertSame([1], $result->issues[1]->path);
    }

    public function testTupleRejectsNonArray(): void
    {
        $s = AnyVali::tuple([AnyVali::string()]);
        $result = $s->safeParse('not a tuple');
        $this->assertFalse($result->success);
        $this->assertSame('tuple', $result->issues[0]->expected);
    }

    // ── Object ──────────────────────────────────────

    public function testObjectAcceptsAllRequired(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string(), 'age' => AnyVali::int()],
            ['name', 'age'],
        );
        $this->assertSame(['name' => 'Alice', 'age' => 30], $s->parse(['name' => 'Alice', 'age' => 30]));
    }

    public function testObjectRejectsMissingRequired(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string(), 'age' => AnyVali::int()],
            ['name', 'age'],
        );
        $result = $s->safeParse(['name' => 'Alice']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::REQUIRED, $result->issues[0]->code);
        $this->assertSame(['age'], $result->issues[0]->path);
    }

    public function testObjectRejectsMissingAllRequired(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string(), 'age' => AnyVali::int()],
            ['name', 'age'],
        );
        $result = $s->safeParse([]);
        $this->assertFalse($result->success);
        $this->assertCount(2, $result->issues);
    }

    public function testObjectAcceptsOptionalAbsent(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string(), 'nickname' => AnyVali::string()],
            ['name'],
        );
        $this->assertSame(['name' => 'Alice'], $s->parse(['name' => 'Alice']));
    }

    public function testObjectRejectsNonObject(): void
    {
        $s = AnyVali::object(['name' => AnyVali::string()], ['name']);
        $result = $s->safeParse('not an object');
        $this->assertFalse($result->success);
        $this->assertSame('object', $result->issues[0]->expected);
    }

    // ── Object Unknown Keys ──────────────────────────

    public function testObjectRejectsUnknownKeysByDefault(): void
    {
        $s = AnyVali::object(['name' => AnyVali::string()], ['name']);
        $result = $s->safeParse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[0]->code);
        $this->assertSame(['extra'], $result->issues[0]->path);
    }

    public function testObjectStripsUnknownKeys(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Strip,
        );
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value', 'another' => 42]);
        $this->assertSame(['name' => 'Alice'], $result);
    }

    public function testObjectAllowsUnknownKeys(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::string()],
            ['name'],
            UnknownKeyMode::Allow,
        );
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertSame(['name' => 'Alice', 'extra' => 'value'], $result);
    }

    public function testObjectReportsMultipleUnknownKeys(): void
    {
        $s = AnyVali::object(['id' => AnyVali::int()], ['id']);
        $result = $s->safeParse(['id' => 1, 'foo' => 'bar', 'baz' => true]);
        $this->assertFalse($result->success);
        $this->assertCount(2, $result->issues);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[0]->code);
        $this->assertSame(IssueCodes::UNKNOWN_KEY, $result->issues[1]->code);
    }

    // ── Record ──────────────────────────────────────

    public function testRecordAcceptsValid(): void
    {
        $s = AnyVali::record(AnyVali::int());
        $this->assertSame(['a' => 1, 'b' => 2, 'c' => 3], $s->parse(['a' => 1, 'b' => 2, 'c' => 3]));
    }

    public function testRecordAcceptsEmpty(): void
    {
        $s = AnyVali::record(AnyVali::string());
        $this->assertSame([], $s->parse([]));
    }

    public function testRecordRejectsInvalidValue(): void
    {
        $s = AnyVali::record(AnyVali::int());
        $result = $s->safeParse(['a' => 1, 'b' => 'two']);
        $this->assertFalse($result->success);
        $this->assertSame(['b'], $result->issues[0]->path);
    }

    public function testRecordRejectsNonObject(): void
    {
        $s = AnyVali::record(AnyVali::string());
        $result = $s->safeParse([1, 2, 3]);
        $this->assertFalse($result->success);
        $this->assertSame('record', $result->issues[0]->expected);
    }

    // ── Union ──────────────────────────────────────

    public function testUnionAcceptsFirstVariant(): void
    {
        $s = AnyVali::union([AnyVali::string(), AnyVali::int()]);
        $this->assertSame('hello', $s->parse('hello'));
    }

    public function testUnionAcceptsSecondVariant(): void
    {
        $s = AnyVali::union([AnyVali::string(), AnyVali::int()]);
        $this->assertSame(42, $s->parse(42));
    }

    public function testUnionRejectsNoVariant(): void
    {
        $s = AnyVali::union([AnyVali::string(), AnyVali::int()]);
        $result = $s->safeParse(true);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_UNION, $result->issues[0]->code);
        $this->assertSame('string | int', $result->issues[0]->expected);
    }

    public function testUnionWithNull(): void
    {
        $s = AnyVali::union([AnyVali::string(), AnyVali::null()]);
        $this->assertNull($s->parse(null));
    }

    // ── Intersection ──────────────────────────────────

    public function testIntersectionAcceptsValid(): void
    {
        $s = AnyVali::intersection([
            AnyVali::object(
                ['name' => AnyVali::string()],
                ['name'],
                UnknownKeyMode::Allow,
            ),
            AnyVali::object(
                ['age' => AnyVali::int()],
                ['age'],
                UnknownKeyMode::Allow,
            ),
        ]);
        $result = $s->parse(['name' => 'Alice', 'age' => 30]);
        $this->assertSame('Alice', $result['name']);
        $this->assertSame(30, $result['age']);
    }

    public function testIntersectionRejectsPartialFailure(): void
    {
        $s = AnyVali::intersection([
            AnyVali::object(
                ['name' => AnyVali::string()],
                ['name'],
                UnknownKeyMode::Allow,
            ),
            AnyVali::object(
                ['age' => AnyVali::int()],
                ['age'],
                UnknownKeyMode::Allow,
            ),
        ]);
        $result = $s->safeParse(['name' => 'Alice']);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::REQUIRED, $result->issues[0]->code);
    }

    public function testIntersectionNumericRanges(): void
    {
        $s = AnyVali::intersection([
            AnyVali::number()->min(0),
            AnyVali::number()->max(100),
        ]);
        $this->assertSame(50, $s->parse(50));

        $result = $s->safeParse(-5);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
    }

    // ── Optional ──────────────────────────────────────

    public function testOptionalAcceptsPresent(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::optional(AnyVali::string())],
            [],
        );
        $this->assertSame(['name' => 'Alice'], $s->parse(['name' => 'Alice']));
    }

    public function testOptionalAcceptsAbsent(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::optional(AnyVali::string())],
            [],
        );
        $this->assertSame([], $s->parse([]));
    }

    public function testOptionalRejectsInvalidPresent(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::optional(AnyVali::string())],
            [],
        );
        $result = $s->safeParse(['name' => 123]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame(['name'], $result->issues[0]->path);
    }

    public function testOptionalNullIsNotAbsent(): void
    {
        $s = AnyVali::object(
            ['name' => AnyVali::optional(AnyVali::string())],
            [],
        );
        $result = $s->safeParse(['name' => null]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
        $this->assertSame('null', $result->issues[0]->received);
    }

    // ── Nullable ──────────────────────────────────────

    public function testNullableAcceptsNull(): void
    {
        $s = AnyVali::nullable(AnyVali::string());
        $this->assertNull($s->parse(null));
    }

    public function testNullableAcceptsValid(): void
    {
        $s = AnyVali::nullable(AnyVali::string());
        $this->assertSame('hello', $s->parse('hello'));
    }

    public function testNullableRejectsInvalid(): void
    {
        $s = AnyVali::nullable(AnyVali::string());
        $result = $s->safeParse(42);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testNullableIntAcceptsNull(): void
    {
        $s = AnyVali::nullable(AnyVali::int());
        $this->assertNull($s->parse(null));
        $this->assertSame(99, $s->parse(99));
    }

    // ── Defaults ──────────────────────────────────────

    public function testDefaultAppliedWhenAbsent(): void
    {
        $s = AnyVali::object(
            [
                'name' => AnyVali::string(),
                'role' => AnyVali::string()->default('user'),
            ],
            ['name'],
        );
        $result = $s->parse(['name' => 'Alice']);
        $this->assertSame(['name' => 'Alice', 'role' => 'user'], $result);
    }

    public function testDefaultNotOverwritePresent(): void
    {
        $s = AnyVali::object(
            [
                'name' => AnyVali::string(),
                'role' => AnyVali::string()->default('user'),
            ],
            ['name'],
        );
        $result = $s->parse(['name' => 'Alice', 'role' => 'admin']);
        $this->assertSame(['name' => 'Alice', 'role' => 'admin'], $result);
    }

    public function testDefaultValueIsValidated(): void
    {
        $s = AnyVali::object(
            ['count' => AnyVali::int()->min(1)->default(5)],
            [],
        );
        $result = $s->parse([]);
        $this->assertSame(['count' => 5], $result);
    }

    public function testInvalidDefaultProducesIssue(): void
    {
        $s = AnyVali::object(
            ['count' => AnyVali::int()->min(10)->default(5)],
            [],
        );
        $result = $s->safeParse([]);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::DEFAULT_INVALID, $result->issues[0]->code);
        $this->assertSame(['count'], $result->issues[0]->path);
    }

    public function testDefaultWithBooleanValue(): void
    {
        $s = AnyVali::object(
            ['active' => AnyVali::bool()->default(true)],
            [],
        );
        $result = $s->parse([]);
        $this->assertSame(['active' => true], $result);
    }

    public function testNullNotTreatedAsAbsentForDefaults(): void
    {
        $s = AnyVali::object(
            ['value' => AnyVali::nullable(AnyVali::string())->default('fallback')],
            [],
        );
        $result = $s->parse(['value' => null]);
        $this->assertSame(['value' => null], $result);
    }

    // ── ParseResult ──────────────────────────────────

    public function testParseResultOk(): void
    {
        $r = ParseResult::ok(42);
        $this->assertTrue($r->success);
        $this->assertSame(42, $r->value);
        $this->assertEmpty($r->issues);
    }

    public function testParseResultFail(): void
    {
        $issues = [new ValidationIssue(code: 'test', message: 'test')];
        $r = ParseResult::fail($issues);
        $this->assertFalse($r->success);
        $this->assertNull($r->value);
        $this->assertCount(1, $r->issues);
    }

    // ── ValidationIssue ──────────────────────────────

    public function testValidationIssueToArray(): void
    {
        $issue = new ValidationIssue(
            code: IssueCodes::INVALID_TYPE,
            message: 'test',
            path: ['foo'],
            expected: 'string',
            received: 'number',
        );
        $arr = $issue->toArray();
        $this->assertSame('invalid_type', $arr['code']);
        $this->assertSame(['foo'], $arr['path']);
        $this->assertSame('string', $arr['expected']);
        $this->assertSame('number', $arr['received']);
    }

    public function testValidationIssueWithPathPrefix(): void
    {
        $issue = new ValidationIssue(
            code: IssueCodes::INVALID_TYPE,
            message: 'test',
            path: ['bar'],
        );
        $prefixed = $issue->withPathPrefix(['foo']);
        $this->assertSame(['foo', 'bar'], $prefixed->path);
    }

    // ── ValidationError ──────────────────────────────

    public function testValidationErrorMessage(): void
    {
        $error = new ValidationError([
            new ValidationIssue(code: 'test', message: 'bad value', path: ['a', 'b']),
        ]);
        $this->assertStringContainsString('test', $error->getMessage());
        $this->assertStringContainsString('a.b', $error->getMessage());
    }

    // ── Method Chaining Immutability ──────────────────

    public function testMethodChainingReturnsNewInstance(): void
    {
        $s1 = AnyVali::string();
        $s2 = $s1->minLength(3);
        $s3 = $s2->maxLength(10);

        $this->assertNotSame($s1, $s2);
        $this->assertNotSame($s2, $s3);

        // s1 should still accept short strings
        $this->assertSame('ab', $s1->parse('ab'));
        // s2 should reject
        $result = $s2->safeParse('ab');
        $this->assertFalse($result->success);
    }

    // ── Numeric Safety ──────────────────────────────

    public function testNumberRoundtripsAsFloat64(): void
    {
        $s = AnyVali::number();
        $this->assertSame(1.7976931348623157e+308, $s->parse(1.7976931348623157e+308));
    }

    public function testIntRoundtripsAsInt64(): void
    {
        $s = AnyVali::int();
        $this->assertSame(9007199254740991, $s->parse(9007199254740991));
    }

    public function testFloat64AndNumberIdentical(): void
    {
        $s = AnyVali::float64();
        $this->assertSame(42.5, $s->parse(42.5));
    }

    public function testInt64AndIntIdentical(): void
    {
        $s = AnyVali::int64();
        $this->assertSame(42, $s->parse(42));
    }

    public function testNarrowingRejected(): void
    {
        $s = AnyVali::int8();
        $result = $s->safeParse(200);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_LARGE, $result->issues[0]->code);
    }

    // ── Object Helper Methods ──────────────────────────

    public function testObjectStrip(): void
    {
        $s = AnyVali::object(['name' => AnyVali::string()], ['name'])->strip();
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertSame(['name' => 'Alice'], $result);
    }

    public function testObjectPassthrough(): void
    {
        $s = AnyVali::object(['name' => AnyVali::string()], ['name'])->passthrough();
        $result = $s->parse(['name' => 'Alice', 'extra' => 'value']);
        $this->assertSame(['name' => 'Alice', 'extra' => 'value'], $result);
    }

    // ── Custom Validators ──────────────────────────

    public function testRefineAddsCustomValidation(): void
    {
        $s = AnyVali::string()->refine(function (mixed $value) {
            if ($value === 'bad') {
                return new ValidationIssue(
                    code: 'custom',
                    message: 'Value cannot be "bad"',
                );
            }
            return null;
        });

        $this->assertSame('good', $s->parse('good'));

        $result = $s->safeParse('bad');
        $this->assertFalse($result->success);
        $this->assertSame('custom', $result->issues[0]->code);
    }

    public function testHasCustomValidators(): void
    {
        $s = AnyVali::string();
        $this->assertFalse($s->hasCustomValidators());

        $s2 = $s->refine(fn($v) => null);
        $this->assertTrue($s2->hasCustomValidators());
    }
}

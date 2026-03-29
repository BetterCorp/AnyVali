<?php

declare(strict_types=1);

namespace AnyVali\Tests;

use AnyVali\AnyVali;
use AnyVali\IssueCodes;
use PHPUnit\Framework\TestCase;

final class CoercionTest extends TestCase
{
    public function testStringToIntValid(): void
    {
        $s = AnyVali::int()->coerce('string->int');
        $this->assertSame(42, $s->parse('42'));
    }

    public function testStringToIntTrimsWhitespace(): void
    {
        $s = AnyVali::int()->coerce('string->int');
        $this->assertSame(42, $s->parse('  42  '));
    }

    public function testStringToNumberValid(): void
    {
        $s = AnyVali::number()->coerce('string->number');
        $this->assertSame(3.14, $s->parse('3.14'));
    }

    public function testStringToBoolTrue(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $this->assertTrue($s->parse('true'));
    }

    public function testStringToBoolFalse(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $this->assertFalse($s->parse('false'));
    }

    public function testStringToBool1(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $this->assertTrue($s->parse('1'));
    }

    public function testStringToBool0(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $this->assertFalse($s->parse('0'));
    }

    public function testStringToBoolCaseInsensitive(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $this->assertTrue($s->parse('TRUE'));
    }

    public function testTrimCoercion(): void
    {
        $s = AnyVali::string()->coerce('trim');
        $this->assertSame('hello', $s->parse('  hello  '));
    }

    public function testLowerCoercion(): void
    {
        $s = AnyVali::string()->coerce('lower');
        $this->assertSame('hello world', $s->parse('HELLO World'));
    }

    public function testUpperCoercion(): void
    {
        $s = AnyVali::string()->coerce('upper');
        $this->assertSame('HELLO WORLD', $s->parse('hello world'));
    }

    public function testCoercionFailureProducesIssue(): void
    {
        $s = AnyVali::int()->coerce('string->int');
        $result = $s->safeParse('not-a-number');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
        $this->assertSame('int', $result->issues[0]->expected);
        $this->assertSame('not-a-number', $result->issues[0]->received);
    }

    public function testCoercionBeforeValidation(): void
    {
        $s = AnyVali::int()->min(10)->coerce('string->int');
        $result = $s->safeParse('5');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::TOO_SMALL, $result->issues[0]->code);
        $this->assertSame('10', $result->issues[0]->expected);
        $this->assertSame('5', $result->issues[0]->received);
    }

    public function testCoercionThenValidationSuccess(): void
    {
        $s = AnyVali::int()->min(1)->max(100)->coerce('string->int');
        $this->assertSame(50, $s->parse('50'));
    }

    public function testChainedCoercions(): void
    {
        $s = AnyVali::string()->coerce(['trim', 'lower']);
        $this->assertSame('hello', $s->parse('  HELLO  '));
    }

    public function testCoercionImmutability(): void
    {
        $s1 = AnyVali::int();
        $s2 = $s1->coerce('string->int');
        $this->assertNotSame($s1, $s2);

        // s1 should NOT coerce
        $result = $s1->safeParse('42');
        $this->assertFalse($result->success);

        // s2 should coerce
        $this->assertSame(42, $s2->parse('42'));
    }

    public function testStringToBoolInvalidString(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $result = $s->safeParse('maybe');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
    }

    public function testStringToIntNonStringInput(): void
    {
        $s = AnyVali::int()->coerce('string->int');
        $result = $s->safeParse(42);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
    }

    public function testStringToNumberNonStringInput(): void
    {
        $s = AnyVali::number()->coerce('string->number');
        $result = $s->safeParse(3.14);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
    }

    public function testStringToBoolNonStringInput(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $result = $s->safeParse(true);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
    }

    public function testTrimNonStringPassesThrough(): void
    {
        $s = AnyVali::string()->coerce('trim');
        // trim on non-string passes through, then validation fails
        $result = $s->safeParse(42);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::INVALID_TYPE, $result->issues[0]->code);
    }

    public function testUnknownCoercion(): void
    {
        $s = AnyVali::string()->coerce('unknown_coerce');
        $result = $s->safeParse('hello');
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
    }

    public function testGetCoerce(): void
    {
        $s = AnyVali::string()->coerce('trim');
        $this->assertSame('trim', $s->getCoerce());

        $s2 = AnyVali::string()->coerce(['trim', 'lower']);
        $this->assertSame(['trim', 'lower'], $s2->getCoerce());

        $s3 = AnyVali::string();
        $this->assertNull($s3->getCoerce());
    }
}

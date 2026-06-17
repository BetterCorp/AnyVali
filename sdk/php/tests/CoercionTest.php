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
        // SPEC §Coercions: string->int only coerces strings; non-string input
        // bypasses coercion and is validated as-is.
        $this->assertTrue($result->success);
        $this->assertSame(42, $result->value);
    }

    public function testStringToNumberNonStringInput(): void
    {
        $s = AnyVali::number()->coerce('string->number');
        $result = $s->safeParse(3.14);
        // SPEC §Coercions: string->number only coerces strings; non-string
        // input bypasses coercion and is validated as-is.
        $this->assertTrue($result->success);
        $this->assertSame(3.14, $result->value);
    }

    public function testStringToBoolNonStringInput(): void
    {
        $s = AnyVali::bool()->coerce('string->bool');
        $result = $s->safeParse(true);
        // SPEC §Coercions: string->bool only coerces strings; non-string input
        // bypasses coercion and is validated as-is.
        $this->assertTrue($result->success);
        $this->assertTrue($result->value);
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

    // ------------------------------------------------------------------
    // Default / no-explicit-source coercion.
    //
    // Enabling coercion the SIMPLEST way — naming only the source ('string')
    // with no explicit target arrow ('string->int' etc.) — must still coerce
    // string input to the schema's own target kind. "string" is the only
    // portable coercion source, so a bare source is the default form.
    // (Regression for the JS bug where the default/no-source form silently
    // no-ops and string input fails with invalid_type.)
    // ------------------------------------------------------------------

    public function testNumberDefaultCoerce(): void
    {
        // No-arg ergonomic: ->coerce() infers string->number from kind.
        $s = AnyVali::number()->coerce();
        $this->assertSame(3.14, $s->parse('3.14'));
    }

    public function testIntDefaultCoerce(): void
    {
        $s = AnyVali::int()->coerce();
        $this->assertSame(42, $s->parse('42'));
    }

    public function testBoolDefaultCoerceTrue(): void
    {
        $s = AnyVali::bool()->coerce();
        $this->assertTrue($s->parse('true'));
    }

    public function testBoolDefaultCoerceFalse(): void
    {
        $s = AnyVali::bool()->coerce();
        $this->assertFalse($s->parse('false'));
    }

    public function testDefaultCoerceStoresStringToken(): void
    {
        // The no-arg form is sugar for the bare "string" source token, which
        // is what round-trips through interchange export/import.
        $s = AnyVali::number()->coerce();
        $this->assertSame('string', $s->getCoerce());
    }

    public function testExplicitStringTokenEqualsNoArg(): void
    {
        // ->coerce('string') and ->coerce() are equivalent.
        $a = AnyVali::int()->coerce('string');
        $b = AnyVali::int()->coerce();
        $this->assertSame(42, $a->parse('42'));
        $this->assertSame(42, $b->parse('42'));
        $this->assertSame($a->getCoerce(), $b->getCoerce());
    }

    public function testObjectNumericFieldsDefaultCoerce(): void
    {
        $s = AnyVali::object([
            'lumpSum' => AnyVali::number()->coerce(),
            'monthlyContributions' => AnyVali::number()->coerce(),
            'investmentTerm' => AnyVali::int()->coerce(),
        ], ['lumpSum', 'monthlyContributions', 'investmentTerm']);

        $result = $s->safeParse([
            'lumpSum' => '1000000',
            'monthlyContributions' => '1000',
            'investmentTerm' => '20',
        ]);

        $this->assertTrue($result->success);
        $this->assertSame(1000000.0, $result->value['lumpSum']);
        $this->assertSame(1000.0, $result->value['monthlyContributions']);
        $this->assertSame(20, $result->value['investmentTerm']);
    }

    // ------------------------------------------------------------------
    // CANONICAL COERCION MATRIX (all FROM STRING) via the no-arg ergonomic.
    //
    // Each row coerces string input to the schema's target kind using the
    // bare ->coerce() form. ACCEPT rows must parse to the expected typed
    // value; REJECT rows must fail with coercion_failed.
    // ------------------------------------------------------------------

    /**
     * string -> int: ASCII ^-?\d+$ trimmed.
     * @dataProvider provideStringToIntAccept
     */
    public function testMatrixStringToIntAccept(string $input, int $expected): void
    {
        $s = AnyVali::int()->coerce();
        $this->assertSame($expected, $s->parse($input));
    }

    /** @return array<string, array{string, int}> */
    public static function provideStringToIntAccept(): array
    {
        return [
            '"42"'      => ['42', 42],
            '"  42  "'  => ['  42  ', 42],
            '"-7"'      => ['-7', -7],
        ];
    }

    /**
     * string -> int rejects non-ASCII-integer forms.
     * @dataProvider provideStringToIntReject
     */
    public function testMatrixStringToIntReject(string $input): void
    {
        $s = AnyVali::int()->coerce();
        $result = $s->safeParse($input);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
        $this->assertSame('int', $result->issues[0]->expected);
    }

    /** @return array<string, array{string}> */
    public static function provideStringToIntReject(): array
    {
        return [
            '"3.14"'      => ['3.14'],
            '"0x10"'      => ['0x10'],
            '"1_000"'     => ['1_000'],
            '"+5"'        => ['+5'],
            '"Infinity"'  => ['Infinity'],
            'empty'       => [''],
            '"abc"'       => ['abc'],
        ];
    }

    /**
     * string -> number: ASCII decimal float incl exponent, trimmed.
     * @dataProvider provideStringToNumberAccept
     */
    public function testMatrixStringToNumberAccept(string $input, float $expected): void
    {
        $s = AnyVali::number()->coerce();
        $this->assertSame($expected, $s->parse($input));
    }

    /** @return array<string, array{string, float}> */
    public static function provideStringToNumberAccept(): array
    {
        return [
            '"3.14"'    => ['3.14', 3.14],
            '"-1.5e3"'  => ['-1.5e3', -1500.0],
            '"  2  "'   => ['  2  ', 2.0],
            '"0"'       => ['0', 0.0],
        ];
    }

    /**
     * string -> number rejects non-decimal-float forms.
     * @dataProvider provideStringToNumberReject
     */
    public function testMatrixStringToNumberReject(string $input): void
    {
        $s = AnyVali::number()->coerce();
        $result = $s->safeParse($input);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
        $this->assertSame('number', $result->issues[0]->expected);
    }

    /** @return array<string, array{string}> */
    public static function provideStringToNumberReject(): array
    {
        return [
            '"0x10"'      => ['0x10'],
            '"Infinity"'  => ['Infinity'],
            '"NaN"'       => ['NaN'],
            'empty'       => [''],
            '"1_000"'     => ['1_000'],
            '"abc"'       => ['abc'],
        ];
    }

    /**
     * string -> bool: trim + case-insensitive; only true/1 and false/0.
     * @dataProvider provideStringToBoolAccept
     */
    public function testMatrixStringToBoolAccept(string $input, bool $expected): void
    {
        $s = AnyVali::bool()->coerce();
        $this->assertSame($expected, $s->parse($input));
    }

    /** @return array<string, array{string, bool}> */
    public static function provideStringToBoolAccept(): array
    {
        return [
            '"true"'  => ['true', true],
            '"TRUE"'  => ['TRUE', true],
            '"1"'     => ['1', true],
            '"false"' => ['false', false],
            '"0"'     => ['0', false],
        ];
    }

    /**
     * string -> bool rejects everything that is not true/1/false/0.
     * @dataProvider provideStringToBoolReject
     */
    public function testMatrixStringToBoolReject(string $input): void
    {
        $s = AnyVali::bool()->coerce();
        $result = $s->safeParse($input);
        $this->assertFalse($result->success);
        $this->assertSame(IssueCodes::COERCION_FAILED, $result->issues[0]->code);
        $this->assertSame('bool', $result->issues[0]->expected);
    }

    /** @return array<string, array{string}> */
    public static function provideStringToBoolReject(): array
    {
        return [
            '"yes"' => ['yes'],
            '"no"'  => ['no'],
            '"on"'  => ['on'],
            '"off"' => ['off'],
            '"t"'   => ['t'],
            '"f"'   => ['f'],
            '"2"'   => ['2'],
            'empty' => [''],
        ];
    }

    // ------------------------------------------------------------------
    // String transforms (string kind): trim, lower, upper; chainable.
    // The no-arg ->coerce() on a string schema is an identity passthrough
    // (no numeric/bool target to infer); real transforms use the tokens.
    // ------------------------------------------------------------------

    public function testStringKindNoArgCoerceIsIdentity(): void
    {
        $s = AnyVali::string()->coerce();
        $this->assertSame('  Hello  ', $s->parse('  Hello  '));
    }

    public function testStringTransformsChainable(): void
    {
        $s = AnyVali::string()->coerce(['trim', 'upper']);
        $this->assertSame('HELLO', $s->parse('  hello  '));
    }

    public function testTypedTokenStillWorksAlongsideNoArg(): void
    {
        // Explicit typed token must keep working unchanged.
        $s = AnyVali::int()->coerce('string->int');
        $this->assertSame(42, $s->parse('42'));
    }
}

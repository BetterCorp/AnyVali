using AnyVali;
using AnyVali.Parse;
using Xunit;

namespace AnyVali.Tests;

public class CoercionTests
{
    [Fact]
    public void StringToIntCoercion()
    {
        var s = V.Int().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(42L, s.Parse("42"));
    }

    [Fact]
    public void StringToIntTrimsWhitespace()
    {
        var s = V.Int().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(42L, s.Parse("  42  "));
    }

    [Fact]
    public void StringToNumberCoercion()
    {
        var s = V.Number().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(3.14, s.Parse("3.14"));
    }

    [Fact]
    public void StringToBoolTrue()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.True(s.Parse("true"));
    }

    [Fact]
    public void StringToBoolFalse()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.False(s.Parse("false"));
    }

    [Fact]
    public void StringToBoolOne()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.True(s.Parse("1"));
    }

    [Fact]
    public void StringToBoolZero()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.False(s.Parse("0"));
    }

    [Fact]
    public void StringToBoolCaseInsensitive()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.True(s.Parse("TRUE"));
    }

    [Fact]
    public void TrimCoercion()
    {
        var s = V.String().Coerce(new CoercionConfig { Trim = true });
        Assert.Equal("hello", s.Parse("  hello  "));
    }

    [Fact]
    public void LowerCoercion()
    {
        var s = V.String().Coerce(new CoercionConfig { Lower = true });
        Assert.Equal("hello world", s.Parse("HELLO World"));
    }

    [Fact]
    public void UpperCoercion()
    {
        var s = V.String().Coerce(new CoercionConfig { Upper = true });
        Assert.Equal("HELLO WORLD", s.Parse("hello world"));
    }

    [Fact]
    public void CoercionFailureProducesIssue()
    {
        var s = V.Int().Coerce(new CoercionConfig { From = "string" });
        var result = s.SafeParse("not-a-number");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.CoercionFailed, result.Issues[0].Code);
    }

    [Fact]
    public void CoercionHappensBeforeValidation()
    {
        var s = V.Int().Min(10).Coerce(new CoercionConfig { From = "string" });
        var result = s.SafeParse("5");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.TooSmall, result.Issues[0].Code);
    }

    [Fact]
    public void CoercionThenValidationSuccess()
    {
        var s = V.Int().Min(1).Max(100).Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(50L, s.Parse("50"));
    }

    [Fact]
    public void NormalizeConfigFromString()
    {
        var config = Coercion.NormalizeConfig("string->int");
        Assert.Equal("string", config.From);
    }

    [Fact]
    public void NormalizeConfigFromArray()
    {
        var config = Coercion.NormalizeConfig(new List<object?> { "trim", "lower" });
        Assert.True(config.Trim);
        Assert.True(config.Lower);
    }

    [Fact]
    public void NormalizeConfigFromDict()
    {
        var config = Coercion.NormalizeConfig(new Dictionary<string, object?>
        {
            ["from"] = "string",
            ["trim"] = true,
        });
        Assert.Equal("string", config.From);
        Assert.True(config.Trim);
    }

    [Fact]
    public void ChainedCoercions()
    {
        var config = Coercion.NormalizeConfig(new List<object?> { "trim", "lower" });
        var s = V.String().Coerce(config);
        Assert.Equal("hello", s.Parse("  HELLO  "));
    }

    // Default coercion (no explicit source). Enabling coercion with no args should
    // still coerce string input on numeric/bool schemas, since "string" is the only
    // portable coercion source.

    [Fact]
    public void DefaultCoerceNumberFromString()
    {
        var s = V.Number().Coerce();
        Assert.Equal(3.14, s.Parse("3.14"));
    }

    [Fact]
    public void DefaultCoerceIntFromString()
    {
        var s = V.Int().Coerce();
        Assert.Equal(42L, s.Parse("42"));
    }

    [Fact]
    public void DefaultCoerceBoolTrueFromString()
    {
        var s = V.Bool().Coerce();
        Assert.True(s.Parse("true"));
    }

    [Fact]
    public void DefaultCoerceBoolFalseFromString()
    {
        var s = V.Bool().Coerce();
        Assert.False(s.Parse("false"));
    }

    [Fact]
    public void DefaultCoerceObjectWithNumericFields()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["age"] = V.Int().Coerce(),
            ["height"] = V.Number().Coerce(),
            ["active"] = V.Bool().Coerce(),
        });

        var result = schema.SafeParse(new Dictionary<string, object?>
        {
            ["age"] = "30",
            ["height"] = "1.75",
            ["active"] = "true",
        });

        Assert.True(result.Success, string.Join(", ", result.Issues.Select(i => i.Code)));
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        Assert.Equal(30L, data["age"]);
        Assert.Equal(1.75, data["height"]);
        Assert.Equal(true, data["active"]);
    }

    // ---------------------------------------------------------------------
    // Canonical coercion matrix (all FROM STRING), using the bare .Coerce()
    // form. Every ACCEPT/REJECT row from the spec is covered. Reject cases
    // assert failure + issue code "coercion_failed".
    // ---------------------------------------------------------------------

    // string -> int: ASCII ^-?\d+$ trimmed.
    [Theory]
    [InlineData("42", 42L)]
    [InlineData("  42  ", 42L)]
    [InlineData("-7", -7L)]
    public void MatrixIntAccept(string input, long expected)
    {
        var s = V.Int().Coerce();
        Assert.Equal(expected, s.Parse(input));
    }

    [Theory]
    [InlineData("3.14")]
    [InlineData("0x10")]
    [InlineData("1_000")]
    [InlineData("+5")]
    [InlineData("Infinity")]
    [InlineData("")]
    [InlineData("abc")]
    public void MatrixIntReject(string input)
    {
        var s = V.Int().Coerce();
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.CoercionFailed, result.Issues[0].Code);
    }

    // string -> number: ASCII decimal float incl exponent, trimmed.
    [Theory]
    [InlineData("3.14", 3.14)]
    [InlineData("-1.5e3", -1500.0)]
    [InlineData("  2  ", 2.0)]
    [InlineData("0", 0.0)]
    public void MatrixNumberAccept(string input, double expected)
    {
        var s = V.Number().Coerce();
        Assert.Equal(expected, s.Parse(input));
    }

    [Theory]
    [InlineData("0x10")]
    [InlineData("Infinity")]
    [InlineData("NaN")]
    [InlineData("")]
    [InlineData("1_000")]
    [InlineData("abc")]
    public void MatrixNumberReject(string input)
    {
        var s = V.Number().Coerce();
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.CoercionFailed, result.Issues[0].Code);
    }

    // string -> bool: trim + case-insensitive.
    [Theory]
    [InlineData("true", true)]
    [InlineData("TRUE", true)]
    [InlineData("1", true)]
    [InlineData("false", false)]
    [InlineData("0", false)]
    public void MatrixBoolAccept(string input, bool expected)
    {
        var s = V.Bool().Coerce();
        Assert.Equal(expected, s.Parse(input));
    }

    [Theory]
    [InlineData("yes")]
    [InlineData("no")]
    [InlineData("on")]
    [InlineData("off")]
    [InlineData("t")]
    [InlineData("f")]
    [InlineData("2")]
    [InlineData("")]
    public void MatrixBoolReject(string input)
    {
        var s = V.Bool().Coerce();
        var result = s.SafeParse(input);
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.CoercionFailed, result.Issues[0].Code);
    }

    // string transforms (string kind): trim, lower, upper; chainable.
    [Fact]
    public void MatrixStringTrim()
    {
        var s = V.String().Coerce(new CoercionConfig { Trim = true });
        Assert.Equal("hello", s.Parse("  hello  "));
    }

    [Fact]
    public void MatrixStringLower()
    {
        var s = V.String().Coerce(new CoercionConfig { Lower = true });
        Assert.Equal("hello", s.Parse("HELLO"));
    }

    [Fact]
    public void MatrixStringUpper()
    {
        var s = V.String().Coerce(new CoercionConfig { Upper = true });
        Assert.Equal("HELLO", s.Parse("hello"));
    }

    [Fact]
    public void MatrixStringChained()
    {
        var s = V.String().Coerce(new CoercionConfig { Trim = true, Lower = true });
        Assert.Equal("hello", s.Parse("  HELLO  "));
    }

    // Bare .Coerce() on a string schema does NOT interchange-coerce; it just
    // passes the string through (no transforms enabled).
    [Fact]
    public void MatrixBareCoerceOnStringPassesThrough()
    {
        var s = V.String().Coerce();
        Assert.Equal("42", s.Parse("42"));
    }
}

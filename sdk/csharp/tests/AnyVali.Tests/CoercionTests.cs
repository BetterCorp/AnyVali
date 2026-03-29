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
        Assert.Equal(true, s.Parse("true"));
    }

    [Fact]
    public void StringToBoolFalse()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(false, s.Parse("false"));
    }

    [Fact]
    public void StringToBoolOne()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(true, s.Parse("1"));
    }

    [Fact]
    public void StringToBoolZero()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(false, s.Parse("0"));
    }

    [Fact]
    public void StringToBoolCaseInsensitive()
    {
        var s = V.Bool().Coerce(new CoercionConfig { From = "string" });
        Assert.Equal(true, s.Parse("TRUE"));
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
}

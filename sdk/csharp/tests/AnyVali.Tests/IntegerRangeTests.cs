using AnyVali;
using AnyVali.Parse;
using System.Text.Json;
using Xunit;

namespace AnyVali.Tests;

public class IntegerRangeTests
{
    private static Schema Roundtrip(Schema schema) => V.Import(AnyValiDocument.FromJson(V.Export(schema).ToJson()));

    [Fact]
    public void SignedOverflowIsRejectedBeforeConversion()
    {
        foreach (var native in new Schema[] { V.Int(), V.Int64() })
        foreach (var schema in new[] { native, Roundtrip(native) })
        {
            foreach (var value in new object[] { 1e20, 1e21, double.MaxValue, 1e21f, (double)long.MaxValue,
                (float)long.MaxValue, (decimal)long.MaxValue + 1, decimal.MaxValue, (ulong)long.MaxValue + 1, ulong.MaxValue })
            {
                var result = schema.SafeParse(value);
                Assert.False(result.Success);
                Assert.Null(result.Data);
                Assert.Equal(IssueCodes.TooLarge, Assert.Single(result.Issues).Code);
            }
            foreach (var value in new object[] { -1e20, -1e21, -double.MaxValue, -1e21f,
                Math.BitDecrement((double)long.MinValue), (decimal)long.MinValue - 1, decimal.MinValue })
                Assert.Equal(IssueCodes.TooSmall, Assert.Single(schema.SafeParse(value).Issues).Code);

            foreach (var value in new[] { long.MinValue, long.MaxValue, 0L })
                Assert.Equal(value, Assert.IsType<long>(schema.SafeParse(value).Data));
            Assert.Equal(long.MinValue, Assert.IsType<long>(schema.Parse((double)long.MinValue)));
            var lastDouble = Math.BitDecrement((double)long.MaxValue);
            Assert.Equal(checked((long)lastDouble), Assert.IsType<long>(schema.Parse(lastDouble)));
            Assert.Equal(long.MaxValue, Assert.IsType<long>(schema.Parse((decimal)long.MaxValue)));
            Assert.Equal(long.MaxValue, Assert.IsType<long>(schema.Parse((ulong)long.MaxValue)));
        }
    }

    [Fact]
    public void IntegerNumberUnionFallsBackWithoutSaturating()
    {
        var native = V.Union(V.Int(), V.Number());
        foreach (var schema in new Schema[] { native, Roundtrip(native) })
        foreach (var value in new[] { 1e20, 1e21, -1e20, -1e21 })
            Assert.Equal(value, Assert.IsType<double>(schema.Parse(value)));
    }

    [Fact]
    public void Uint64PreservesTheFullRangeAndDefaults()
    {
        foreach (var value in new[] { 0UL, (ulong)long.MaxValue, (ulong)long.MaxValue + 1,
            9223372036854775809UL, ulong.MaxValue })
        {
            var native = V.Uint64().Default(value);
            foreach (var schema in new Schema[] { native, Roundtrip(native) })
            {
                Assert.Equal(value, Assert.IsType<ulong>(schema.SafeParse(value).Data));
                Assert.Equal(value, Assert.IsType<ulong>(schema.Parse(Absent.Value)));
                Assert.Equal(value, Assert.IsType<ulong>(schema.Parse((decimal)value)));
            }
            var exported = V.Export(native).ToJson();
            Assert.Equal(exported, V.Export(Roundtrip(native)).ToJson());
        }
        var parent = Roundtrip(V.Object(new() { ["value"] = V.Uint64().Default(ulong.MaxValue) }));
        var parsed = Assert.IsType<Dictionary<string, object?>>(parent.Parse(new Dictionary<string, object?>()));
        Assert.Equal(ulong.MaxValue, Assert.IsType<ulong>(parsed["value"]));
    }

    [Fact]
    public void Uint64RejectsOutsideItsRange()
    {
        foreach (var schema in new Schema[] { V.Uint64(), Roundtrip(V.Uint64()) })
        {
            foreach (var value in new object[] { -1L, long.MinValue, -1e21 })
                Assert.Equal(IssueCodes.TooSmall, Assert.Single(schema.SafeParse(value).Issues).Code);
            foreach (var value in new object[] { (double)ulong.MaxValue, (float)ulong.MaxValue,
                (decimal)ulong.MaxValue + 1, decimal.MaxValue, 1e21, double.MaxValue })
                Assert.Equal(IssueCodes.TooLarge, Assert.Single(schema.SafeParse(value).Issues).Code);
            var lastDouble = Math.BitDecrement((double)ulong.MaxValue);
            Assert.Equal(checked((ulong)lastDouble), Assert.IsType<ulong>(schema.Parse(lastDouble)));
        }
        Assert.Equal(IssueCodes.DefaultInvalid, Assert.Single(V.Int().Default(1e21).SafeParse(Absent.Value).Issues).Code);
        Assert.Equal(IssueCodes.DefaultInvalid, Assert.Single(V.Uint64().Default(-1L).SafeParse(Absent.Value).Issues).Code);
    }

    [Fact]
    public void IntegerConstraintsKeepPrecisionThroughClonesAndRoundtrips()
    {
        var cases = new (Schema schema, object valid, object invalid, string code)[]
        {
            (V.Uint64().MultipleOf(3), ulong.MaxValue, ulong.MaxValue - 1, IssueCodes.InvalidNumber),
            (V.Int64().MultipleOf(2), long.MaxValue - 1, long.MaxValue, IssueCodes.InvalidNumber),
            (V.Uint64().Max(9007199254740992d), 9007199254740992UL, 9007199254740993UL, IssueCodes.TooLarge),
            (V.Uint64().Min(9007199254740992d), 9007199254740992UL, 9007199254740991UL, IssueCodes.TooSmall),
            (V.Uint64().ExclusiveMin(9007199254740992d), 9007199254740993UL, 9007199254740992UL, IssueCodes.TooSmall),
            (V.Uint64().ExclusiveMax(9007199254740992d), 9007199254740991UL, 9007199254740992UL, IssueCodes.TooLarge),
            (V.Int().Min(-1.5).Max(1.5), 1L, 2L, IssueCodes.TooLarge),
            (V.Int().MultipleOf(1.5), 3L, 2L, IssueCodes.InvalidNumber),
        };
        foreach (var (native, valid, invalid, code) in cases)
        foreach (var schema in new[] { native, native.Describe("clone"), Roundtrip(native) })
        {
            Assert.True(schema.SafeParse(valid).Success);
            Assert.Equal(code, Assert.Single(schema.SafeParse(invalid).Issues).Code);
        }
        Assert.True(V.Uint64().MultipleOf(0.1).SafeParse(ulong.MaxValue).Success);
        Assert.True(V.Uint64().MultipleOf(1.5e-20).SafeParse(ulong.MaxValue).Success);
        Assert.False(V.Uint64().MultipleOf(1.5e-20).SafeParse(ulong.MaxValue - 1).Success);
    }

    [Fact]
    public void Native64BitConstraintArgumentsRemainExact()
    {
        var cases = new (Schema schema, object valid, object invalid)[]
        {
            (V.Uint64().Max(ulong.MaxValue - 1), ulong.MaxValue - 1, ulong.MaxValue),
            (V.Uint64().Min(ulong.MaxValue), ulong.MaxValue, ulong.MaxValue - 1),
            (V.Uint64().ExclusiveMin(ulong.MaxValue - 1), ulong.MaxValue, ulong.MaxValue - 1),
            (V.Uint64().ExclusiveMax(ulong.MaxValue), ulong.MaxValue - 1, ulong.MaxValue),
            (V.Uint64().MultipleOf(ulong.MaxValue), ulong.MaxValue, ulong.MaxValue - 1),
            (V.Int64().Max(long.MaxValue - 1), long.MaxValue - 1, long.MaxValue),
            (V.Int64().Min(long.MinValue + 1), long.MinValue + 1, long.MinValue),
            (V.Int64().ExclusiveMin(long.MinValue), long.MinValue + 1, long.MinValue),
            (V.Int64().ExclusiveMax(long.MaxValue), long.MaxValue - 1, long.MaxValue),
            (V.Int64().MultipleOf(long.MaxValue), long.MaxValue, long.MaxValue - 1),
        };
        foreach (var (native, valid, invalid) in cases)
        foreach (var schema in new[] { native, native.Describe("clone"), Roundtrip(native) })
        {
            Assert.True(schema.SafeParse(valid).Success);
            Assert.False(schema.SafeParse(invalid).Success);
        }
    }

    [Fact]
    public void InvalidNumericConstraintsFailAtConfiguration()
    {
        foreach (var bound in new[] { double.NaN, double.NegativeInfinity, double.PositiveInfinity })
        {
            Assert.Throws<InvalidOperationException>(() => V.Int().Max(bound));
            Assert.Throws<InvalidOperationException>(() => V.Number().Min(bound));
        }
        Assert.Throws<InvalidOperationException>(() => V.Int().MultipleOf(0));
        Assert.Throws<InvalidOperationException>(() => V.Int().MultipleOf(-1));
    }

    [Theory]
    [InlineData("min", "18446744073709551614", "18446744073709551613", "too_small")]
    [InlineData("max", "18446744073709551614", "18446744073709551615", "too_large")]
    [InlineData("exclusiveMin", "18446744073709551614", "18446744073709551614", "too_small")]
    [InlineData("exclusiveMax", "18446744073709551614", "18446744073709551614", "too_large")]
    [InlineData("multipleOf", "18446744073709551614", "18446744073709551615", "invalid_number")]
    public void ImportedIntegerConstraintsRemainExact(string name, string bound, string input, string code)
    {
        var doc = AnyValiDocument.FromJson($$$"""
            {"anyvaliVersion":"1.0","schemaVersion":"1.1",
             "root":{"kind":"uint64","{{{name}}}":{{{bound}}}},"definitions":{},"extensions":{}}
            """);
        var native = V.Import(doc);
        foreach (var schema in new[] { native, native.Default(0UL).Describe("clone"), Roundtrip(native) })
        {
            Assert.Equal(code, Assert.Single(schema.SafeParse(ulong.Parse(input)).Issues).Code);
            Assert.Equal(ulong.Parse(bound), Assert.IsType<ulong>(V.Export(schema).Root[name]));
        }
    }

    [Theory]
    [InlineData("string", "minLength")]
    [InlineData("string", "maxLength")]
    [InlineData("array", "minItems")]
    [InlineData("array", "maxItems")]
    public void OversizedLengthConstraintsAreRejected(string kind, string constraint)
    {
        foreach (var size in new[] { "9223372036854775808", "18446744073709551615", "2147483648", "-1", "1.5" })
        {
            var doc = AnyValiDocument.FromJson($$$"""
                {"anyvaliVersion":"1.0","schemaVersion":"1.1",
                 "root":{"kind":"{{{kind}}}","{{{constraint}}}":{{{size}}}},"definitions":{},"extensions":{}}
                """);
            Assert.Throws<InvalidOperationException>(() => V.Import(doc));
        }
    }

    [Fact]
    public void ImportedUnsignedConstraintsAreNotConvertedToZero()
    {
        var schema = V.Import(AnyValiDocument.FromJson("""
            {"anyvaliVersion":"1.0","schemaVersion":"1.1",
             "root":{"kind":"uint64","min":9223372036854775808,"max":18446744073709551615},
             "definitions":{},"extensions":{}}
            """));
        Assert.Equal(ulong.MaxValue, Assert.IsType<ulong>(schema.Parse(ulong.MaxValue)));
        Assert.Equal(IssueCodes.TooSmall, Assert.Single(schema.SafeParse(0UL).Issues).Code);
    }

    [Fact]
    public void Uint64CoercionAndJsonDecodingPreserveUnsignedValues()
    {
        foreach (var schema in new Schema[] { V.Uint64().Coerce(), Roundtrip(V.Uint64().Coerce()) })
        {
            Assert.Equal(ulong.MaxValue, Assert.IsType<ulong>(schema.Parse("18446744073709551615")));
            Assert.Equal(9223372036854775809UL, Assert.IsType<ulong>(schema.Parse("9223372036854775809")));
            Assert.False(schema.SafeParse("18446744073709551616").Success);
            Assert.False(schema.SafeParse("-1").Success);
            Assert.False(schema.SafeParse("1.5").Success);
        }
        using var json = JsonDocument.Parse("18446744073709551615");
        Assert.Equal(ulong.MaxValue, Assert.IsType<ulong>(JsonHelper.ElementToObject(json.RootElement)));
    }
}

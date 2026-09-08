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

using AnyVali;
using Xunit;

namespace AnyVali.Tests;

public class DefaultsTests
{
    [Fact]
    public void MissingFieldGetsDefault()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["name"] = V.String(),
            ["role"] = V.String().Default("user"),
        });

        var result = schema.SafeParse(new Dictionary<string, object?> { ["name"] = "Alice" });

        Assert.True(result.Success, string.Join(", ", result.Issues.Select(i => i.Code)));
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        Assert.Equal("user", data["role"]);
    }

    [Fact]
    public void PresentFieldIsNotOverwritten()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["role"] = V.String().Default("user"),
        });

        var result = schema.SafeParse(new Dictionary<string, object?> { ["role"] = "admin" });

        Assert.True(result.Success);
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        Assert.Equal("admin", data["role"]);
    }

    [Fact]
    public void InvalidDefaultProducesDefaultInvalid()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["count"] = V.Int().Min(10).Default(5),
        });

        var result = schema.SafeParse(new Dictionary<string, object?>());

        Assert.False(result.Success);
        Assert.Equal(IssueCodes.DefaultInvalid, result.Issues[0].Code);
        Assert.Equal(new object[] { "count" }, result.Issues[0].Path);
    }

    [Fact]
    public void NullIsNotAbsentForNullableDefault()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["value"] = V.Nullable(V.String()).Default("fallback"),
        });

        var result = schema.SafeParse(new Dictionary<string, object?> { ["value"] = null });

        Assert.True(result.Success);
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        Assert.Null(data["value"]);
    }

    [Fact]
    public void FalsyDefaultsAreApplied()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["count"] = V.Int().Default(0),
            ["name"] = V.String().Default(""),
            ["active"] = V.Bool().Default(false),
        });

        var result = schema.SafeParse(new Dictionary<string, object?>());

        Assert.True(result.Success, string.Join(", ", result.Issues.Select(i => i.Code)));
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        Assert.Equal(0L, data["count"]);
        Assert.Equal("", data["name"]);
        Assert.Equal(false, data["active"]);
    }

    [Fact]
    public void NestedObjectFieldGetsDefault()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["user"] = V.Object(new Dictionary<string, Schema>
            {
                ["name"] = V.String(),
                ["role"] = V.String().Default("guest"),
            }),
        });

        var result = schema.SafeParse(new Dictionary<string, object?>
        {
            ["user"] = new Dictionary<string, object?> { ["name"] = "Bob" },
        });

        Assert.True(result.Success, string.Join(", ", result.Issues.Select(i => i.Code)));
        var data = Assert.IsType<Dictionary<string, object?>>(result.Data);
        var user = Assert.IsType<Dictionary<string, object?>>(data["user"]);
        Assert.Equal("guest", user["role"]);
    }

    [Fact]
    public void MutableDefaultIsNotSharedBetweenParses()
    {
        var schema = V.Object(new Dictionary<string, Schema>
        {
            ["meta"] = V.Any().Default(new Dictionary<string, object?>
            {
                ["items"] = new List<object?>(),
            }),
        });

        var first = Assert.IsType<Dictionary<string, object?>>(
            schema.Parse(new Dictionary<string, object?>()));
        var firstMeta = Assert.IsType<Dictionary<string, object?>>(first["meta"]);
        var firstItems = Assert.IsType<List<object?>>(firstMeta["items"]);
        firstItems.Add("mutated");

        var second = Assert.IsType<Dictionary<string, object?>>(
            schema.Parse(new Dictionary<string, object?>()));
        var secondMeta = Assert.IsType<Dictionary<string, object?>>(second["meta"]);
        Assert.Empty(Assert.IsType<List<object?>>(secondMeta["items"]));
    }
}

using AnyVali;
using Xunit;

namespace AnyVali.Tests;

public class SensitiveDataTests
{
    [Fact]
    public void EncryptedRoundTripAndValidation()
    {
        var sensitive = new DescribeOptions { Sensitive = true };
        var schema = V.Object(new()
        {
            ["name"] = V.String(),
            ["secret"] = V.String().Describe("secret", sensitive),
            ["profile"] = V.Object(new() { ["token"] = V.String() }).Describe("profile", sensitive),
            ["aliases"] = V.Array(V.String().Describe("alias", sensitive)),
        });
        var plain = new Dictionary<string, object?>
        {
            ["name"] = "Ada", ["secret"] = "abc",
            ["profile"] = new Dictionary<string, object?> { ["token"] = "xyz" },
            ["aliases"] = new List<object?> { "one", "two" },
        };

        var encrypted = V.Encrypt(schema, plain, (_, value) => $"encrypted:{value}");
        Assert.True(V.SafeParseEncrypted(schema, encrypted).Success);
        Assert.Throws<ValidationError>(() => V.Encrypt(schema, plain, (_, _) => "broken"));
        Assert.NotNull(V.Decrypt(schema, encrypted, (path, _) => string.Join(".", path) switch
        {
            "profile" => new Dictionary<string, object?> { ["token"] = "xyz" },
            "aliases.0" => "one",
            "aliases.1" => "two",
            _ => "abc",
        }));
    }
}

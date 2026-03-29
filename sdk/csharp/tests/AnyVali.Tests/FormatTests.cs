using AnyVali;
using AnyVali.Format;
using Xunit;

namespace AnyVali.Tests;

public class FormatTests
{
    [Theory]
    [InlineData("test@example.com", true)]
    [InlineData("user.name+tag@domain.co.uk", true)]
    [InlineData("not-email", false)]
    [InlineData("@no-local.com", false)]
    [InlineData("no-domain@", false)]
    public void EmailFormat(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "email"));
    }

    [Theory]
    [InlineData("https://example.com", true)]
    [InlineData("http://example.com/path?q=1", true)]
    [InlineData("ftp://example.com", false)]
    [InlineData("not a url", false)]
    public void UrlFormat(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "url"));
    }

    [Theory]
    [InlineData("550e8400-e29b-41d4-a716-446655440000", true)]
    [InlineData("not-a-uuid", false)]
    [InlineData("550e8400-e29b-41d4-a716-44665544000", false)] // too short
    public void UuidFormat(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "uuid"));
    }

    [Theory]
    [InlineData("192.168.1.1", true)]
    [InlineData("0.0.0.0", true)]
    [InlineData("255.255.255.255", true)]
    [InlineData("256.1.1.1", false)]
    [InlineData("1.2.3", false)]
    public void Ipv4Format(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "ipv4"));
    }

    [Theory]
    [InlineData("::1", true)]
    [InlineData("::", true)]
    [InlineData("2001:0db8:85a3:0000:0000:8a2e:0370:7334", true)]
    [InlineData("not-ipv6", false)]
    public void Ipv6Format(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "ipv6"));
    }

    [Theory]
    [InlineData("2024-01-15", true)]
    [InlineData("2024-02-29", true)] // leap year
    [InlineData("2023-02-29", false)] // not a leap year
    [InlineData("2024-13-01", false)]
    [InlineData("not-a-date", false)]
    public void DateFormat(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "date"));
    }

    [Theory]
    [InlineData("2024-01-15T10:30:00Z", true)]
    [InlineData("2024-01-15T10:30:00+05:30", true)]
    [InlineData("2024-01-15T10:30:00.123Z", true)]
    [InlineData("2024-01-15T25:00:00Z", false)]
    [InlineData("not-datetime", false)]
    public void DateTimeFormat(string value, bool expected)
    {
        Assert.Equal(expected, FormatValidators.Validate(value, "date-time"));
    }

    [Fact]
    public void StringSchemaWithFormat()
    {
        var s = V.String().Format("email");
        Assert.Equal("test@example.com", s.Parse("test@example.com"));

        var result = s.SafeParse("not-email");
        Assert.False(result.Success);
        Assert.Equal(IssueCodes.InvalidString, result.Issues[0].Code);
    }

    [Fact]
    public void UnknownFormatPasses()
    {
        Assert.True(FormatValidators.Validate("anything", "unknown-format"));
    }
}

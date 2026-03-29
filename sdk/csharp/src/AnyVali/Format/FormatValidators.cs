using System.Text.RegularExpressions;

namespace AnyVali.Format;

/// <summary>
/// Portable string format validators.
/// </summary>
public static class FormatValidators
{
    private static readonly Regex EmailRe = new(
        @"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$",
        RegexOptions.Compiled);

    private static readonly Regex UuidRe = new(
        @"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$",
        RegexOptions.Compiled);

    private static readonly Regex Ipv4Re = new(
        @"^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)$",
        RegexOptions.Compiled);

    private static readonly Regex Ipv6Re = new(
        @"^(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?::[0-9a-fA-F]{1,4}){1,6}|:(?::[0-9a-fA-F]{1,4}){1,7}|::)$",
        RegexOptions.Compiled);

    private static readonly Regex DateRe = new(
        @"^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$",
        RegexOptions.Compiled);

    private static readonly Regex DateTimeRe = new(
        @"^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$",
        RegexOptions.Compiled);

    public static bool Validate(string value, string format)
    {
        return format switch
        {
            "email" => EmailRe.IsMatch(value),
            "url" => IsValidUrl(value),
            "uuid" => UuidRe.IsMatch(value),
            "ipv4" => Ipv4Re.IsMatch(value),
            "ipv6" => Ipv6Re.IsMatch(value),
            "date" => IsValidDate(value),
            "date-time" => IsValidDateTime(value),
            _ => true // unknown formats pass
        };
    }

    private static bool IsValidUrl(string value)
    {
        if (!Uri.TryCreate(value, UriKind.Absolute, out var uri))
            return false;
        return uri.Scheme is "http" or "https";
    }

    private static bool IsValidDate(string value)
    {
        if (!DateRe.IsMatch(value)) return false;
        var parts = value.Split('-');
        var y = int.Parse(parts[0]);
        var m = int.Parse(parts[1]);
        var d = int.Parse(parts[2]);
        try
        {
            var dt = new DateTime(y, m, d);
            return dt.Year == y && dt.Month == m && dt.Day == d;
        }
        catch
        {
            return false;
        }
    }

    private static bool IsValidDateTime(string value)
    {
        if (!DateTimeRe.IsMatch(value)) return false;
        return IsValidDate(value[..10]);
    }
}

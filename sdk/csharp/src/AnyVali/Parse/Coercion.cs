using System.Globalization;
using System.Text.RegularExpressions;

namespace AnyVali.Parse;

/// <summary>
/// Result of a coercion attempt.
/// </summary>
public readonly record struct CoercionResult(bool Success, object? Value, string? Message);

/// <summary>
/// Portable coercion logic.
/// </summary>
public static class Coercion
{
    private static readonly Regex IntPattern = new(@"^-?\d+$", RegexOptions.Compiled);

    /// <summary>
    /// Normalize raw coercion config from interchange format.
    /// Handles string, array of strings, or dictionary forms.
    /// </summary>
    public static CoercionConfig NormalizeConfig(object? raw)
    {
        if (raw is CoercionConfig cc) return cc;

        if (raw is Dictionary<string, object?> dict)
        {
            return new CoercionConfig
            {
                From = dict.TryGetValue("from", out var f) ? f?.ToString() : null,
                Trim = dict.TryGetValue("trim", out var t) && t is true,
                Lower = dict.TryGetValue("lower", out var l) && l is true,
                Upper = dict.TryGetValue("upper", out var u) && u is true,
            };
        }

        var items = new List<string>();
        if (raw is string s)
        {
            items.Add(s);
        }
        else if (raw is List<object?> list)
        {
            foreach (var item in list)
                if (item is string str) items.Add(str);
        }

        var config = new CoercionConfig();
        foreach (var item in items)
        {
            config = item switch
            {
                "string->int" or "string->number" or "string->bool" =>
                    config with { From = "string" },
                "trim" => config with { Trim = true },
                "lower" => config with { Lower = true },
                "upper" => config with { Upper = true },
                _ => config
            };
        }

        return config;
    }

    /// <summary>
    /// Apply coercion to an input value.
    /// </summary>
    public static CoercionResult Apply(object? input, CoercionConfig config, string targetType)
    {
        var value = input;

        // String transformations
        if (value is string str)
        {
            if (config.Trim) str = str.Trim();
            if (config.Lower) str = str.ToLowerInvariant();
            if (config.Upper) str = str.ToUpperInvariant();
            value = str;
        }

        // Type coercion from string to target
        if (config.From == "string" && value is string sv)
        {
            switch (targetType)
            {
                case "int":
                case "int8":
                case "int16":
                case "int32":
                case "int64":
                case "uint8":
                case "uint16":
                case "uint32":
                case "uint64":
                {
                    var trimmed = sv.Trim();
                    if (trimmed.Length == 0 || !IntPattern.IsMatch(trimmed))
                        return new CoercionResult(false, null, $"Cannot coerce \"{sv}\" to {targetType}");

                    if (!long.TryParse(trimmed, NumberStyles.Integer, CultureInfo.InvariantCulture, out var num))
                        return new CoercionResult(false, null, $"Cannot coerce \"{sv}\" to {targetType}");

                    value = num;
                    break;
                }
                case "number":
                case "float32":
                case "float64":
                {
                    var trimmed = sv.Trim();
                    if (trimmed.Length == 0)
                        return new CoercionResult(false, null, $"Cannot coerce empty string to {targetType}");

                    if (!double.TryParse(trimmed, NumberStyles.Float | NumberStyles.AllowLeadingSign,
                            CultureInfo.InvariantCulture, out var num) || double.IsInfinity(num) || double.IsNaN(num))
                        return new CoercionResult(false, null, $"Cannot coerce \"{sv}\" to {targetType}");

                    value = num;
                    break;
                }
                case "bool":
                {
                    var lower = sv.Trim().ToLowerInvariant();
                    if (lower is "true" or "1")
                        value = true;
                    else if (lower is "false" or "0")
                        value = false;
                    else
                        return new CoercionResult(false, null, $"Cannot coerce \"{sv}\" to bool");
                    break;
                }
            }
        }

        return new CoercionResult(true, value, null);
    }
}

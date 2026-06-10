using System.Text.RegularExpressions;
using AnyVali.Format;

namespace AnyVali.Schemas;

public sealed class StringSchema : Schema<string>
{
    private int? _minLength;
    private int? _maxLength;
    private string? _pattern;
    private string? _startsWith;
    private string? _endsWith;
    private string? _includes;
    private string? _format;

    internal override string GetCoercionTarget() => "string";

    public StringSchema MinLength(int n) { var c = (StringSchema)Clone(); c._minLength = n; return c; }
    public StringSchema MaxLength(int n) { var c = (StringSchema)Clone(); c._maxLength = n; return c; }
    public StringSchema Pattern(string p) { var c = (StringSchema)Clone(); c._pattern = p; return c; }
    public StringSchema StartsWith(string s) { var c = (StringSchema)Clone(); c._startsWith = s; return c; }
    public StringSchema EndsWith(string s) { var c = (StringSchema)Clone(); c._endsWith = s; return c; }
    public StringSchema Includes(string s) { var c = (StringSchema)Clone(); c._includes = s; return c; }
    public StringSchema Format(string f) { var c = (StringSchema)Clone(); c._format = f; return c; }

    public new StringSchema Default(object? value) => (StringSchema)base.Default(value);
    public new StringSchema Coerce(Parse.CoercionConfig? config = null) => (StringSchema)base.Coerce(config);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not string val)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected string, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "string",
                Received = DescribeType(input),
            });
            return null;
        }

        var length = CodePointLength(val);

        if (_minLength.HasValue && length < _minLength.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"String must have at least {_minLength.Value} character(s)",
                Path = ctx.ClonePath(),
                Expected = _minLength.Value.ToString(),
                Received = length.ToString(),
            });
        }

        if (_maxLength.HasValue && length > _maxLength.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"String must have at most {_maxLength.Value} character(s)",
                Path = ctx.ClonePath(),
                Expected = _maxLength.Value.ToString(),
                Received = length.ToString(),
            });
        }

        if (_pattern is not null)
        {
            try
            {
                var re = new Regex(ToEcmaAnchors(_pattern), RegexOptions.None, TimeSpan.FromSeconds(1));
                if (!re.IsMatch(val))
                {
                    ctx.Issues.Add(new ValidationIssue
                    {
                        Code = IssueCodes.InvalidString,
                        Message = $"String does not match pattern: {_pattern}",
                        Path = ctx.ClonePath(),
                        Expected = _pattern,
                        Received = val,
                    });
                }
            }
            catch (Exception)
            {
                // Invalid regex pattern or timeout - treat as validation failure
                ctx.Issues.Add(new ValidationIssue
                {
                    Code = IssueCodes.InvalidString,
                    Message = $"Invalid regex pattern: {_pattern}",
                    Path = ctx.ClonePath(),
                    Expected = _pattern,
                    Received = val,
                });
            }
        }

        if (_startsWith is not null && !val.StartsWith(_startsWith, StringComparison.Ordinal))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidString,
                Message = $"String must start with \"{_startsWith}\"",
                Path = ctx.ClonePath(),
                Expected = _startsWith,
                Received = val,
            });
        }

        if (_endsWith is not null && !val.EndsWith(_endsWith, StringComparison.Ordinal))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidString,
                Message = $"String must end with \"{_endsWith}\"",
                Path = ctx.ClonePath(),
                Expected = _endsWith,
                Received = val,
            });
        }

        if (_includes is not null && !val.Contains(_includes, StringComparison.Ordinal))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidString,
                Message = $"String must include \"{_includes}\"",
                Path = ctx.ClonePath(),
                Expected = _includes,
                Received = val,
            });
        }

        if (_format is not null && !FormatValidators.Validate(val, _format))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidString,
                Message = $"Invalid {_format} format",
                Path = ctx.ClonePath(),
                Expected = _format,
                Received = val,
            });
        }

        return val;
    }

    /// <summary>
    /// Rewrite ECMA-262 anchors to .NET's absolute anchors. In ECMA without the
    /// multiline flag, "^"/"$" match only the start/end of the whole string.
    /// .NET's "$" also matches just before a trailing "\n", so an anchored
    /// whitelist like ^[a-z]+$ would accept "abc\n" -- a newline-injection
    /// bypass that diverges from the JS reference. Translate unescaped, top-level
    /// "^" -> "\A" and "$" -> "\z" (absolute end). Anchors inside character
    /// classes and escaped "\^"/"\$" are left untouched.
    /// </summary>
    private static string ToEcmaAnchors(string pattern)
    {
        var sb = new System.Text.StringBuilder(pattern.Length + 4);
        var escaped = false;
        var inClass = false;
        foreach (var ch in pattern)
        {
            if (escaped) { sb.Append(ch); escaped = false; }
            else if (ch == '\\') { sb.Append(ch); escaped = true; }
            else if (ch == '[') { inClass = true; sb.Append(ch); }
            else if (ch == ']' && inClass) { inClass = false; sb.Append(ch); }
            else if (ch == '^' && !inClass) { sb.Append(@"\A"); }
            else if (ch == '$' && !inClass) { sb.Append(@"\z"); }
            else { sb.Append(ch); }
        }
        return sb.ToString();
    }

    private static int CodePointLength(string value)
    {
        var count = 0;
        for (var i = 0; i < value.Length; i++)
        {
            if (char.IsHighSurrogate(value[i]) &&
                i + 1 < value.Length &&
                char.IsLowSurrogate(value[i + 1]))
            {
                i++;
            }
            count++;
        }
        return count;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "string" };
        if (_minLength.HasValue) node["minLength"] = (long)_minLength.Value;
        if (_maxLength.HasValue) node["maxLength"] = (long)_maxLength.Value;
        if (_pattern is not null) node["pattern"] = _pattern;
        if (_startsWith is not null) node["startsWith"] = _startsWith;
        if (_endsWith is not null) node["endsWith"] = _endsWith;
        if (_includes is not null) node["includes"] = _includes;
        if (_format is not null) node["format"] = _format;
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone()
    {
        return new StringSchema
        {
            _minLength = _minLength, _maxLength = _maxLength,
            _pattern = _pattern, _startsWith = _startsWith,
            _endsWith = _endsWith, _includes = _includes,
            _format = _format,
            DefaultValue = DefaultValue, CoercionCfg = CoercionCfg,
            IsPortable = IsPortable, MetadataMap = MetadataMap,
        };
    }
}

namespace AnyVali.Schemas;

public sealed class EnumSchema : Schema<object?>
{
    private readonly List<object> _values;

    public EnumSchema(IEnumerable<object> values)
    {
        _values = values.ToList();
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        foreach (var v in _values)
        {
            if (v is null && input is null) return input;
            if (v is not null && input is not null)
            {
                // Numeric comparison
                if (IsFiniteNumber(v) && IsFiniteNumber(input) && ToDouble(v) == ToDouble(input))
                    return input;
                if (v.Equals(input))
                    return input;
            }
        }

        var valuesStr = string.Join(",", _values.Select(v => v?.ToString() ?? "null"));
        ctx.Issues.Add(new ValidationIssue
        {
            Code = IssueCodes.InvalidType,
            Message = $"Expected one of enum({valuesStr}), received {input}",
            Path = ctx.ClonePath(),
            Expected = $"enum({valuesStr})",
            Received = input?.ToString() ?? "null",
        });
        return null;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "enum",
            ["values"] = _values.Select(v => (object?)v).ToList(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new EnumSchema(_values)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

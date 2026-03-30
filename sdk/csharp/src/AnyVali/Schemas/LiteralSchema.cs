namespace AnyVali.Schemas;

public sealed class LiteralSchema : Schema<object?>
{
    private readonly object? _value;

    public LiteralSchema(object? value)
    {
        _value = value;
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (!ValuesEqual(input, _value))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidLiteral,
                Message = $"Expected literal {FormatLiteral(_value)}, received {FormatLiteral(input)}",
                Path = ctx.ClonePath(),
                Expected = FormatLiteral(_value),
                Received = FormatLiteral(input),
            });
            return null;
        }
        return input;
    }

    private static bool ValuesEqual(object? a, object? b)
    {
        if (a is null && b is null) return true;
        if (a is null || b is null) return false;
        // Numeric comparison: normalize to double
        if (IsFiniteNumber(a) && IsFiniteNumber(b))
            return ToDouble(a) == ToDouble(b);
        return a.Equals(b);
    }

    private static string FormatLiteral(object? value) =>
        value?.ToString() ?? "null";

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "literal", ["value"] = _value };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new LiteralSchema(_value)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

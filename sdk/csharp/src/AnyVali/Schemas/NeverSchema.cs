namespace AnyVali.Schemas;

public sealed class NeverSchema : Schema<object?>
{
    internal override object? Validate(object? input, ValidationContext ctx)
    {
        ctx.Issues.Add(new ValidationIssue
        {
            Code = IssueCodes.InvalidType,
            Message = "Expected never (no value is valid)",
            Path = ctx.ClonePath(),
            Expected = "never",
            Received = DescribeType(input),
        });
        return null;
    }

    internal override Dictionary<string, object?> ToNode() =>
        new() { ["kind"] = "never" };

    internal override Schema Clone() => new NeverSchema
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

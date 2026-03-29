namespace AnyVali.Schemas;

public sealed class NullSchema : Schema
{
    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not null)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected null, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "null",
                Received = DescribeType(input),
            });
            return null;
        }
        return null;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "null" };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new NullSchema
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

namespace AnyVali.Schemas;

public sealed class BoolSchema : Schema
{
    internal override string GetCoercionTarget() => "bool";

    public new BoolSchema Default(object? value) => (BoolSchema)base.Default(value);
    public new BoolSchema Coerce(Parse.CoercionConfig? config = null) => (BoolSchema)base.Coerce(config);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not bool)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected boolean, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "bool",
                Received = DescribeType(input),
            });
            return null;
        }
        return input;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "bool" };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new BoolSchema
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

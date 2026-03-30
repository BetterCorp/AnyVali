namespace AnyVali.Schemas;

public sealed class NullableSchema : Schema<object?>
{
    internal Schema Inner { get; }

    public NullableSchema(Schema inner)
    {
        Inner = inner;
    }

    public new NullableSchema Default(object? value) => (NullableSchema)base.Default(value);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is null) return null;
        return Inner.Validate(input, ctx);
    }

    internal override object? RunPipeline(object? input, ValidationContext ctx)
    {
        if (input is null) return null;
        return Inner.RunPipeline(input, ctx);
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "nullable",
            ["inner"] = Inner.ToNode(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new NullableSchema(Inner)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

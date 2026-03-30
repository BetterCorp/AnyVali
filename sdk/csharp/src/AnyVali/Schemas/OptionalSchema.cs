using AnyVali.Parse;

namespace AnyVali.Schemas;

public sealed class OptionalSchema : Schema<object?>
{
    internal Schema Inner { get; }

    public OptionalSchema(Schema inner)
    {
        Inner = inner;
        DefaultValue = inner.DefaultValue;
        CoercionCfg = inner.CoercionCfg;
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is null && IsAbsentNull(input))
            return null;
        if (Absent.IsAbsent(input))
            return null;
        return Inner.Validate(input, ctx);
    }

    internal override object? RunPipeline(object? input, ValidationContext ctx)
    {
        var isAbsent = Absent.IsAbsent(input);

        // If absent and inner has default, delegate to inner
        if (isAbsent && !Absent.IsAbsent(Inner.DefaultValue))
            return Inner.RunPipeline(input, ctx);

        if (isAbsent)
            return null;

        // Delegate to inner's pipeline for coercion etc.
        return Inner.RunPipeline(input, ctx);
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "optional",
            ["inner"] = Inner.ToNode(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new OptionalSchema(Inner)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

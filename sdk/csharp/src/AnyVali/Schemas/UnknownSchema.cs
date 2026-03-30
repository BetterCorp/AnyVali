namespace AnyVali.Schemas;

public sealed class UnknownSchema : Schema<object?>
{
    internal override object? Validate(object? input, ValidationContext ctx) => input;

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "unknown" };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new UnknownSchema
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

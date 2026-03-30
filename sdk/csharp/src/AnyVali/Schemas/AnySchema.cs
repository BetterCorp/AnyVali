namespace AnyVali.Schemas;

public sealed class AnySchema : Schema<object?>
{
    internal override object? Validate(object? input, ValidationContext ctx) => input;

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = "any" };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new AnySchema
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

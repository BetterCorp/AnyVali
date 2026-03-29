namespace AnyVali.Schemas;

public sealed class RefSchema : Schema
{
    private readonly string _ref;
    private readonly Func<Schema>? _resolver;

    public RefSchema(string refPath, Func<Schema>? resolver = null)
    {
        _ref = refPath;
        _resolver = resolver;
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (_resolver is not null)
        {
            var resolved = _resolver();
            return resolved.Validate(input, ctx);
        }

        ctx.Issues.Add(new ValidationIssue
        {
            Code = IssueCodes.UnsupportedSchemaKind,
            Message = $"Unresolved ref: {_ref}",
            Path = ctx.ClonePath(),
        });
        return null;
    }

    internal override object? RunPipeline(object? input, ValidationContext ctx)
    {
        if (_resolver is not null)
        {
            var resolved = _resolver();
            return resolved.RunPipeline(input, ctx);
        }

        return Validate(input, ctx);
    }

    internal override Dictionary<string, object?> ToNode() =>
        new() { ["kind"] = "ref", ["ref"] = _ref };

    internal override Schema Clone() => new RefSchema(_ref, _resolver)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

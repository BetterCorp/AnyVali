namespace AnyVali.Schemas;

public sealed class UnionSchema : Schema
{
    private readonly List<Schema> _variants;

    public UnionSchema(IEnumerable<Schema> variants)
    {
        _variants = variants.ToList();
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        foreach (var variant in _variants)
        {
            var innerCtx = new ValidationContext();
            foreach (var p in ctx.Path)
                innerCtx.PushPath(p);

            var result = variant.RunPipeline(input, innerCtx);
            if (innerCtx.Issues.Count == 0)
                return result;
        }

        var variantKinds = _variants.Select(v => v.ToNode()["kind"]?.ToString() ?? "unknown");
        ctx.Issues.Add(new ValidationIssue
        {
            Code = IssueCodes.InvalidUnion,
            Message = "Input did not match any variant of the union",
            Path = ctx.ClonePath(),
            Expected = string.Join(" | ", variantKinds),
            Received = DescribeType(input),
        });
        return null;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "union",
            ["variants"] = _variants.Select(v => (object?)v.ToNode()).ToList(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new UnionSchema(_variants)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

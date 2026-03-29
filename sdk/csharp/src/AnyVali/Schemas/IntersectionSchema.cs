namespace AnyVali.Schemas;

public sealed class IntersectionSchema : Schema
{
    private readonly List<Schema> _schemas;

    public IntersectionSchema(IEnumerable<Schema> schemas)
    {
        _schemas = schemas.ToList();
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        object? result = input;
        var anyFailed = false;

        foreach (var schema in _schemas)
        {
            var innerCtx = new ValidationContext();
            foreach (var p in ctx.Path)
                innerCtx.PushPath(p);

            var validated = schema.RunPipeline(input, innerCtx);

            if (innerCtx.Issues.Count > 0)
            {
                ctx.Issues.AddRange(innerCtx.Issues);
                anyFailed = true;
            }
            else
            {
                // Merge object results
                if (result is Dictionary<string, object?> rd &&
                    validated is Dictionary<string, object?> vd)
                {
                    var merged = new Dictionary<string, object?>(rd);
                    foreach (var (k, v) in vd)
                        merged[k] = v;
                    result = merged;
                }
                else
                {
                    result = validated;
                }
            }
        }

        return anyFailed ? null : result;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "intersection",
            ["allOf"] = _schemas.Select(s => (object?)s.ToNode()).ToList(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new IntersectionSchema(_schemas)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

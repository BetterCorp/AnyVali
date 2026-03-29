namespace AnyVali.Schemas;

public sealed class TupleSchema : Schema
{
    private readonly List<Schema> _items;

    public TupleSchema(IEnumerable<Schema> items)
    {
        _items = items.ToList();
    }

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not List<object?> arr)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected tuple, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "tuple",
                Received = DescribeType(input),
            });
            return null;
        }

        if (arr.Count < _items.Count)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Tuple must have exactly {_items.Count} element(s), received {arr.Count}",
                Path = ctx.ClonePath(),
                Expected = _items.Count.ToString(),
                Received = arr.Count.ToString(),
            });
            return null;
        }

        if (arr.Count > _items.Count)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Tuple must have exactly {_items.Count} element(s), received {arr.Count}",
                Path = ctx.ClonePath(),
                Expected = _items.Count.ToString(),
                Received = arr.Count.ToString(),
            });
            return null;
        }

        var result = new List<object?>();
        for (var i = 0; i < _items.Count; i++)
        {
            ctx.PushPath(i);
            var val = _items[i].RunPipeline(arr[i], ctx);
            result.Add(val);
            ctx.PopPath();
        }
        return result;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "tuple",
            ["elements"] = _items.Select(s => (object?)s.ToNode()).ToList(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new TupleSchema(_items)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

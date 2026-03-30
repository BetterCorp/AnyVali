namespace AnyVali.Schemas;

public sealed class ArraySchema : Schema<List<object?>>
{
    private readonly Schema _items;
    private int? _minItems;
    private int? _maxItems;

    public ArraySchema(Schema items)
    {
        _items = items;
    }

    public ArraySchema MinItems(int n) { var c = (ArraySchema)Clone(); c._minItems = n; return c; }
    public ArraySchema MaxItems(int n) { var c = (ArraySchema)Clone(); c._maxItems = n; return c; }
    public new ArraySchema Default(object? value) => (ArraySchema)base.Default(value);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not List<object?> arr)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected array, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "array",
                Received = DescribeType(input),
            });
            return null;
        }

        if (_minItems.HasValue && arr.Count < _minItems.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Array must have at least {_minItems.Value} item(s)",
                Path = ctx.ClonePath(),
                Expected = _minItems.Value.ToString(),
                Received = arr.Count.ToString(),
            });
        }

        if (_maxItems.HasValue && arr.Count > _maxItems.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Array must have at most {_maxItems.Value} item(s)",
                Path = ctx.ClonePath(),
                Expected = _maxItems.Value.ToString(),
                Received = arr.Count.ToString(),
            });
        }

        var result = new List<object?>();
        for (var i = 0; i < arr.Count; i++)
        {
            ctx.PushPath(i);
            var val = _items.RunPipeline(arr[i], ctx);
            result.Add(val);
            ctx.PopPath();
        }
        return result;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "array",
            ["items"] = _items.ToNode(),
        };
        if (_minItems.HasValue) node["minItems"] = (long)_minItems.Value;
        if (_maxItems.HasValue) node["maxItems"] = (long)_maxItems.Value;
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new ArraySchema(_items)
    {
        _minItems = _minItems, _maxItems = _maxItems,
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

namespace AnyVali.Schemas;

public sealed class RecordSchema : Schema<Dictionary<string, object?>>
{
    private readonly Schema _valueSchema;

    public RecordSchema(Schema valueSchema)
    {
        _valueSchema = valueSchema;
    }

    public new RecordSchema Default(object? value) => (RecordSchema)base.Default(value);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not Dictionary<string, object?> obj)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected record, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "record",
                Received = DescribeType(input),
            });
            return null;
        }

        var result = new Dictionary<string, object?>();
        foreach (var (key, value) in obj)
        {
            ctx.PushPath(key);
            result[key] = _valueSchema.RunPipeline(value, ctx);
            ctx.PopPath();
        }
        return result;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?>
        {
            ["kind"] = "record",
            ["valueSchema"] = _valueSchema.ToNode(),
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new RecordSchema(_valueSchema)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

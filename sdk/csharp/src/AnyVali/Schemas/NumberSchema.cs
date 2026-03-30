namespace AnyVali.Schemas;

public class NumberSchema : Schema<double>
{
    protected string Kind;
    protected double? _min;
    protected double? _max;
    protected double? _exclusiveMin;
    protected double? _exclusiveMax;
    protected double? _multipleOf;

    public NumberSchema(string kind = "number")
    {
        Kind = kind;
    }

    internal override string GetCoercionTarget() => Kind;

    public NumberSchema Min(double n) { var c = (NumberSchema)Clone(); c._min = n; return c; }
    public NumberSchema Max(double n) { var c = (NumberSchema)Clone(); c._max = n; return c; }
    public NumberSchema ExclusiveMin(double n) { var c = (NumberSchema)Clone(); c._exclusiveMin = n; return c; }
    public NumberSchema ExclusiveMax(double n) { var c = (NumberSchema)Clone(); c._exclusiveMax = n; return c; }
    public NumberSchema MultipleOf(double n) { var c = (NumberSchema)Clone(); c._multipleOf = n; return c; }

    public new NumberSchema Default(object? value) => (NumberSchema)base.Default(value);
    public new NumberSchema Coerce(Parse.CoercionConfig? config = null) => (NumberSchema)base.Coerce(config);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (!IsFiniteNumber(input))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected {Kind}, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = Kind,
                Received = DescribeType(input),
            });
            return null;
        }

        var val = ToDouble(input);
        ValidateConstraints(val, ctx);
        return val;
    }

    protected void ValidateConstraints(double val, ValidationContext ctx)
    {
        if (_min.HasValue && val < _min.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Number must be >= {FormatNum(_min.Value)}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(_min.Value),
                Received = FormatNum(val),
            });
        }

        if (_max.HasValue && val > _max.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Number must be <= {FormatNum(_max.Value)}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(_max.Value),
                Received = FormatNum(val),
            });
        }

        if (_exclusiveMin.HasValue && val <= _exclusiveMin.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Number must be > {FormatNum(_exclusiveMin.Value)}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(_exclusiveMin.Value),
                Received = FormatNum(val),
            });
        }

        if (_exclusiveMax.HasValue && val >= _exclusiveMax.Value)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Number must be < {FormatNum(_exclusiveMax.Value)}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(_exclusiveMax.Value),
                Received = FormatNum(val),
            });
        }

        if (_multipleOf.HasValue)
        {
            var remainder = val % _multipleOf.Value;
            if (Math.Abs(remainder) > 1e-10 && Math.Abs(remainder - _multipleOf.Value) > 1e-10)
            {
                ctx.Issues.Add(new ValidationIssue
                {
                    Code = IssueCodes.InvalidNumber,
                    Message = $"Number must be a multiple of {FormatNum(_multipleOf.Value)}",
                    Path = ctx.ClonePath(),
                    Expected = FormatNum(_multipleOf.Value),
                    Received = FormatNum(val),
                });
            }
        }
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = Kind };
        if (_min.HasValue) node["min"] = _min.Value;
        if (_max.HasValue) node["max"] = _max.Value;
        if (_exclusiveMin.HasValue) node["exclusiveMin"] = _exclusiveMin.Value;
        if (_exclusiveMax.HasValue) node["exclusiveMax"] = _exclusiveMax.Value;
        if (_multipleOf.HasValue) node["multipleOf"] = _multipleOf.Value;
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone()
    {
        return new NumberSchema(Kind)
        {
            _min = _min, _max = _max,
            _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax,
            _multipleOf = _multipleOf,
            DefaultValue = DefaultValue, CoercionCfg = CoercionCfg,
            IsPortable = IsPortable,
        };
    }

    internal static string FormatNum(double d)
    {
        if (d == Math.Floor(d) && !double.IsInfinity(d))
            return ((long)d).ToString();
        return d.ToString("G");
    }
}

public sealed class Float32Schema : NumberSchema
{
    public Float32Schema() : base("float32") { }
    internal override Schema Clone() => new Float32Schema
    {
        _min = _min, _max = _max, _exclusiveMin = _exclusiveMin,
        _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf,
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

public sealed class Float64Schema : NumberSchema
{
    public Float64Schema() : base("float64") { }
    internal override Schema Clone() => new Float64Schema
    {
        _min = _min, _max = _max, _exclusiveMin = _exclusiveMin,
        _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf,
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

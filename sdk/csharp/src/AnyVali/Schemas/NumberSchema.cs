namespace AnyVali.Schemas;

public class NumberSchema : Schema<double>
{
    protected string Kind;
    protected object? _min;
    protected object? _max;
    protected object? _exclusiveMin;
    protected object? _exclusiveMax;
    protected object? _multipleOf;

    public NumberSchema(string kind = "number")
    {
        Kind = kind;
    }

    internal override string GetCoercionTarget() => Kind;

    public NumberSchema Min(double n) => WithConstraint("min", n);
    public NumberSchema Max(double n) => WithConstraint("max", n);
    public NumberSchema ExclusiveMin(double n) => WithConstraint("exclusiveMin", n);
    public NumberSchema ExclusiveMax(double n) => WithConstraint("exclusiveMax", n);
    public NumberSchema MultipleOf(double n) => WithConstraint("multipleOf", n);

    internal NumberSchema WithConstraint(string name, object value)
    {
        if (!IsFiniteNumber(value) || name == "multipleOf" && ToDouble(value) <= 0)
            throw new InvalidOperationException($"Invalid numeric constraint: {name}");
        var clone = (NumberSchema)Clone();
        switch (name)
        {
            case "min": clone._min = value; break;
            case "max": clone._max = value; break;
            case "exclusiveMin": clone._exclusiveMin = value; break;
            case "exclusiveMax": clone._exclusiveMax = value; break;
            case "multipleOf": clone._multipleOf = value; break;
            default: throw new InvalidOperationException($"Unknown numeric constraint: {name}");
        }
        return clone;
    }

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

        // float32 MUST reject values outside the binary32 representable range
        // (spec 1.4). Without this, float32 silently accepts any double value,
        // defeating the narrowing guarantee.
        if (Kind == "float32" && val != 0.0 && Math.Abs(val) > Float32Max)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Value {FormatNum(val)} is outside the float32 range",
                Path = ctx.ClonePath(),
                Expected = "float32",
                Received = FormatNum(val),
            });
            return null;
        }

        ValidateConstraints(val, ctx);
        return val;
    }

    private const double Float32Max = 3.4028234663852886e38;

    protected void ValidateConstraints(double val, ValidationContext ctx)
    {
        if (_min is not null && val < ToDouble(_min))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Number must be >= {FormatNum(ToDouble(_min))}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(ToDouble(_min)),
                Received = FormatNum(val),
            });
        }

        if (_max is not null && val > ToDouble(_max))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Number must be <= {FormatNum(ToDouble(_max))}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(ToDouble(_max)),
                Received = FormatNum(val),
            });
        }

        if (_exclusiveMin is not null && val <= ToDouble(_exclusiveMin))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Number must be > {FormatNum(ToDouble(_exclusiveMin))}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(ToDouble(_exclusiveMin)),
                Received = FormatNum(val),
            });
        }

        if (_exclusiveMax is not null && val >= ToDouble(_exclusiveMax))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Number must be < {FormatNum(ToDouble(_exclusiveMax))}",
                Path = ctx.ClonePath(),
                Expected = FormatNum(ToDouble(_exclusiveMax)),
                Received = FormatNum(val),
            });
        }

        if (_multipleOf is not null)
        {
            var remainder = val % ToDouble(_multipleOf);
            if (Math.Abs(remainder) > 1e-10 && Math.Abs(remainder - ToDouble(_multipleOf)) > 1e-10)
            {
                ctx.Issues.Add(new ValidationIssue
                {
                    Code = IssueCodes.InvalidNumber,
                    Message = $"Number must be a multiple of {FormatNum(ToDouble(_multipleOf))}",
                    Path = ctx.ClonePath(),
                    Expected = FormatNum(ToDouble(_multipleOf)),
                    Received = FormatNum(val),
                });
            }
        }
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var node = new Dictionary<string, object?> { ["kind"] = Kind };
        if (_min is not null) node["min"] = _min;
        if (_max is not null) node["max"] = _max;
        if (_exclusiveMin is not null) node["exclusiveMin"] = _exclusiveMin;
        if (_exclusiveMax is not null) node["exclusiveMax"] = _exclusiveMax;
        if (_multipleOf is not null) node["multipleOf"] = _multipleOf;
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
            IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions,
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
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions,
    };
}

public sealed class Float64Schema : NumberSchema
{
    public Float64Schema() : base("float64") { }
    internal override Schema Clone() => new Float64Schema
    {
        _min = _min, _max = _max, _exclusiveMin = _exclusiveMin,
        _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf,
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions,
    };
}

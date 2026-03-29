namespace AnyVali.Schemas;

public class IntSchema : NumberSchema
{
    private readonly long _rangeMin;
    private readonly long _rangeMax;

    private static readonly Dictionary<string, (long min, long max)> IntRanges = new()
    {
        ["int8"] = (-128, 127),
        ["int16"] = (-32768, 32767),
        ["int32"] = (-2147483648, 2147483647),
        ["int64"] = (long.MinValue, long.MaxValue),
        ["uint8"] = (0, 255),
        ["uint16"] = (0, 65535),
        ["uint32"] = (0, 4294967295),
        ["uint64"] = (0, long.MaxValue), // capped at int64 max for safety
        ["int"] = (long.MinValue, long.MaxValue),
    };

    public IntSchema(string kind = "int") : base(kind)
    {
        var range = IntRanges.GetValueOrDefault(kind, IntRanges["int"]);
        _rangeMin = range.min;
        _rangeMax = range.max;
    }

    public new IntSchema Min(double n) => (IntSchema)base.Min(n);
    public new IntSchema Max(double n) => (IntSchema)base.Max(n);
    public new IntSchema ExclusiveMin(double n) => (IntSchema)base.ExclusiveMin(n);
    public new IntSchema ExclusiveMax(double n) => (IntSchema)base.ExclusiveMax(n);
    public new IntSchema MultipleOf(double n) => (IntSchema)base.MultipleOf(n);
    public new IntSchema Default(object? value) => (IntSchema)base.Default(value);
    public new IntSchema Coerce(Parse.CoercionConfig? config = null) => (IntSchema)base.Coerce(config);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (!IsFiniteNumber(input))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected integer, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = Kind,
                Received = DescribeType(input),
            });
            return null;
        }

        if (!IsInteger(input))
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = "Expected integer, received float",
                Path = ctx.ClonePath(),
                Expected = Kind,
                Received = "number",
            });
            return null;
        }

        var val = ToLong(input);

        if (val > _rangeMax)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooLarge,
                Message = $"Value {val} is above the maximum for {Kind}",
                Path = ctx.ClonePath(),
                Expected = Kind,
                Received = val.ToString(),
            });
            return null;
        }

        if (val < _rangeMin)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.TooSmall,
                Message = $"Value {val} is below the minimum for {Kind}",
                Path = ctx.ClonePath(),
                Expected = Kind,
                Received = val.ToString(),
            });
            return null;
        }

        ValidateConstraints(val, ctx);
        // Return as long for int types
        return val;
    }

    internal override Schema Clone()
    {
        return new IntSchema(Kind)
        {
            _min = _min, _max = _max, _exclusiveMin = _exclusiveMin,
            _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf,
            DefaultValue = DefaultValue, CoercionCfg = CoercionCfg,
            IsPortable = IsPortable,
        };
    }
}

public sealed class Int8Schema : IntSchema { public Int8Schema() : base("int8") { } internal override Schema Clone() => new Int8Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Int16Schema : IntSchema { public Int16Schema() : base("int16") { } internal override Schema Clone() => new Int16Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Int32Schema : IntSchema { public Int32Schema() : base("int32") { } internal override Schema Clone() => new Int32Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Int64Schema : IntSchema { public Int64Schema() : base("int64") { } internal override Schema Clone() => new Int64Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Uint8Schema : IntSchema { public Uint8Schema() : base("uint8") { } internal override Schema Clone() => new Uint8Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Uint16Schema : IntSchema { public Uint16Schema() : base("uint16") { } internal override Schema Clone() => new Uint16Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Uint32Schema : IntSchema { public Uint32Schema() : base("uint32") { } internal override Schema Clone() => new Uint32Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }
public sealed class Uint64Schema : IntSchema { public Uint64Schema() : base("uint64") { } internal override Schema Clone() => new Uint64Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable }; }

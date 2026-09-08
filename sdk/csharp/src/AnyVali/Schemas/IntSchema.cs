using System.Numerics;
using System.Globalization;

namespace AnyVali.Schemas;

public class IntSchema : NumberSchema
{
    private readonly BigInteger _rangeMin;
    private readonly BigInteger _rangeMax;

    private static readonly Dictionary<string, (BigInteger min, BigInteger max)> IntRanges = new()
    {
        ["int8"] = (-128, 127),
        ["int16"] = (-32768, 32767),
        ["int32"] = (-2147483648, 2147483647),
        ["int64"] = (long.MinValue, long.MaxValue),
        ["uint8"] = (0, 255),
        ["uint16"] = (0, 65535),
        ["uint32"] = (0, 4294967295),
        ["uint64"] = (0, ulong.MaxValue),
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
    public IntSchema Min(long n) => (IntSchema)WithConstraint("min", n);
    public IntSchema Min(ulong n) => (IntSchema)WithConstraint("min", n);
    public IntSchema Max(long n) => (IntSchema)WithConstraint("max", n);
    public IntSchema Max(ulong n) => (IntSchema)WithConstraint("max", n);
    public IntSchema ExclusiveMin(long n) => (IntSchema)WithConstraint("exclusiveMin", n);
    public IntSchema ExclusiveMin(ulong n) => (IntSchema)WithConstraint("exclusiveMin", n);
    public IntSchema ExclusiveMax(long n) => (IntSchema)WithConstraint("exclusiveMax", n);
    public IntSchema ExclusiveMax(ulong n) => (IntSchema)WithConstraint("exclusiveMax", n);
    public IntSchema MultipleOf(long n) => (IntSchema)WithConstraint("multipleOf", n);
    public IntSchema MultipleOf(ulong n) => (IntSchema)WithConstraint("multipleOf", n);
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

        // Compare exact integer values before narrowing; double bounds would round
        // Int64.MaxValue/UInt64.MaxValue up to the first out-of-range integer.
        var val = ToInteger(input!);

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

        ValidateIntegerConstraints(val, ctx);
        return Kind == "uint64" ? (object)(ulong)val : (long)val;
    }

    private static BigInteger ToInteger(object value) => value switch
    {
        ulong u => new BigInteger(u),
        double d => new BigInteger(d),
        float f => new BigInteger(f),
        decimal m => new BigInteger(m),
        _ => new BigInteger(Convert.ToInt64(value)),
    };

    private void ValidateIntegerConstraints(BigInteger value, ValidationContext ctx)
    {
        foreach (var (name, bound) in new[] { ("min", _min), ("max", _max),
            ("exclusiveMin", _exclusiveMin), ("exclusiveMax", _exclusiveMax), ("multipleOf", _multipleOf) })
        {
            if (bound is null) continue;
            var (numerator, denominator) = ConstraintFraction(bound);
            var scaled = value * denominator;
            var invalid = name switch
            {
                "min" => scaled < numerator,
                "max" => scaled > numerator,
                "exclusiveMin" => scaled <= numerator,
                "exclusiveMax" => scaled >= numerator,
                _ => scaled % numerator != 0,
            };
            if (invalid)
                ctx.Issues.Add(new ValidationIssue
                {
                    Code = name is "min" or "exclusiveMin" ? IssueCodes.TooSmall
                        : name is "max" or "exclusiveMax" ? IssueCodes.TooLarge : IssueCodes.InvalidNumber,
                    Message = $"Integer violates {name}: {Convert.ToString(bound, CultureInfo.InvariantCulture)}",
                    Path = ctx.ClonePath(),
                    Expected = Convert.ToString(bound, CultureInfo.InvariantCulture),
                    Received = value.ToString(CultureInfo.InvariantCulture),
                });
        }
    }

    private static (BigInteger numerator, BigInteger denominator) ConstraintFraction(object bound)
    {
        if (IsInteger(bound)) return (ToInteger(bound), BigInteger.One);

        // Fractional constraints use their decimal representation, including
        // scientific notation, so neither the integer nor the remainder is rounded.
        var parts = Convert.ToString(bound, CultureInfo.InvariantCulture)!.Split('E', 'e');
        var mantissa = parts[0];
        var exponent = parts.Length == 2 ? int.Parse(parts[1], CultureInfo.InvariantCulture) : 0;
        var point = mantissa.IndexOf('.');
        var scale = (point < 0 ? 0 : mantissa.Length - point - 1) - exponent;
        var numerator = BigInteger.Parse(mantissa.Replace(".", ""), CultureInfo.InvariantCulture);
        return (numerator, BigInteger.Pow(10, scale));
    }

    internal override Schema Clone()
    {
        return new IntSchema(Kind)
        {
            _min = _min, _max = _max, _exclusiveMin = _exclusiveMin,
            _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf,
            DefaultValue = DefaultValue, CoercionCfg = CoercionCfg,
            IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions,
        };
    }
}

public sealed class Int8Schema : IntSchema { public Int8Schema() : base("int8") { } internal override Schema Clone() => new Int8Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Int16Schema : IntSchema { public Int16Schema() : base("int16") { } internal override Schema Clone() => new Int16Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Int32Schema : IntSchema { public Int32Schema() : base("int32") { } internal override Schema Clone() => new Int32Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Int64Schema : IntSchema { public Int64Schema() : base("int64") { } internal override Schema Clone() => new Int64Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Uint8Schema : IntSchema { public Uint8Schema() : base("uint8") { } internal override Schema Clone() => new Uint8Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Uint16Schema : IntSchema { public Uint16Schema() : base("uint16") { } internal override Schema Clone() => new Uint16Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Uint32Schema : IntSchema { public Uint32Schema() : base("uint32") { } internal override Schema Clone() => new Uint32Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }
public sealed class Uint64Schema : IntSchema { public Uint64Schema() : base("uint64") { } internal override Schema Clone() => new Uint64Schema { _min = _min, _max = _max, _exclusiveMin = _exclusiveMin, _exclusiveMax = _exclusiveMax, _multipleOf = _multipleOf, DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable, MetadataMap = MetadataMap, ImportedDefinitions = ImportedDefinitions, ImportedExtensions = ImportedExtensions }; }

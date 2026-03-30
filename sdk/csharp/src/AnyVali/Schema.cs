using AnyVali.Parse;

namespace AnyVali;

/// <summary>
/// Abstract base for all AnyVali schemas.
/// </summary>
public abstract class Schema
{
    internal const string AnyvaliVersionValue = "1.0";
    internal const string SchemaVersionValue = "1";

    internal object? DefaultValue { get; set; } = Absent.Value;
    internal CoercionConfig? CoercionCfg { get; set; }
    internal bool IsPortable { get; set; } = true;

    // ---- Public API ----

    /// <summary>
    /// Parse input. Throws ValidationError on failure.
    /// </summary>
    public object? Parse(object? input)
    {
        var result = SafeParse(input);
        if (result.Success) return result.Data;
        throw new ValidationError(result.Issues);
    }

    /// <summary>
    /// Parse input. Returns ParseResult with success/failure.
    /// </summary>
    public ParseResult SafeParse(object? input)
    {
        var ctx = new ValidationContext();
        var output = RunPipeline(input, ctx);
        if (ctx.Issues.Count > 0)
            return ParseResult.Fail(ctx.Issues);
        return ParseResult.Ok(output);
    }

    /// <summary>
    /// Run the 5-step parse pipeline.
    /// </summary>
    internal virtual object? RunPipeline(object? input, ValidationContext ctx)
    {
        // Step 1: detect presence
        var isAbsent = Absent.IsAbsent(input) || input is null && IsAbsentNull(input);

        var value = input;

        // Step 2: coercion (only for present values)
        if (!isAbsent && CoercionCfg is not null)
        {
            var coerced = Coercion.Apply(value, CoercionCfg, GetCoercionTarget());
            if (coerced.Success)
            {
                value = coerced.Value;
            }
            else
            {
                ctx.Issues.Add(new ValidationIssue
                {
                    Code = IssueCodes.CoercionFailed,
                    Message = coerced.Message ?? "Coercion failed",
                    Path = ctx.ClonePath(),
                    Expected = GetCoercionTarget(),
                    Received = DescribeValue(input),
                });
                return null;
            }
        }

        // Step 3: default materialization (only for absent values)
        var usedDefault = false;
        if (isAbsent && !Absent.IsAbsent(DefaultValue))
        {
            value = DefaultValue;
            usedDefault = true;
        }

        // Step 4: validate
        var issuesBefore = ctx.Issues.Count;
        var result = Validate(value, ctx);

        // If default was materialized and validation failed, remap to default_invalid
        if (usedDefault && ctx.Issues.Count > issuesBefore)
        {
            for (var i = issuesBefore; i < ctx.Issues.Count; i++)
            {
                var issue = ctx.Issues[i];
                ctx.Issues[i] = new ValidationIssue
                {
                    Code = IssueCodes.DefaultInvalid,
                    Message = issue.Message,
                    Path = issue.Path,
                    Expected = issue.Expected,
                    Received = issue.Received,
                    Meta = issue.Meta,
                };
            }
        }

        return result;
    }

    /// <summary>
    /// Whether a null input should be treated as absent.
    /// By default, null is NOT absent (it's a concrete null value).
    /// Only overridden in contexts where null means "missing".
    /// </summary>
    protected virtual bool IsAbsentNull(object? input) => false;

    /// <summary>
    /// Override to provide the coercion target type name.
    /// </summary>
    internal virtual string GetCoercionTarget() => "unknown";

    /// <summary>
    /// Core validation logic. Override in subclasses.
    /// </summary>
    internal abstract object? Validate(object? input, ValidationContext ctx);

    /// <summary>
    /// Convert to interchange schema node.
    /// </summary>
    internal abstract Dictionary<string, object?> ToNode();

    /// <summary>
    /// Set a default value. Returns a clone.
    /// </summary>
    public Schema Default(object? value)
    {
        var clone = Clone();
        clone.DefaultValue = value;
        return clone;
    }

    /// <summary>
    /// Enable coercion. Returns a clone.
    /// </summary>
    public Schema Coerce(CoercionConfig? config = null)
    {
        var clone = Clone();
        clone.CoercionCfg = config ?? new CoercionConfig();
        return clone;
    }

    /// <summary>
    /// Export to AnyValiDocument.
    /// </summary>
    public AnyValiDocument Export(ExportMode mode = ExportMode.Portable)
    {
        if (mode == ExportMode.Portable && !IsPortable)
            throw new InvalidOperationException(
                "Cannot export in portable mode: schema contains non-portable features");

        var node = ToNode();
        return new AnyValiDocument
        {
            AnyvaliVersion = AnyvaliVersionValue,
            SchemaVersion = SchemaVersionValue,
            Root = node,
            Definitions = new Dictionary<string, object?>(),
            Extensions = new Dictionary<string, object?>(),
        };
    }

    // ---- Internal helpers ----

    internal abstract Schema Clone();

    internal void AddDefaultAndCoercion(Dictionary<string, object?> node)
    {
        if (!Absent.IsAbsent(DefaultValue))
            node["default"] = DefaultValue;
        if (CoercionCfg is not null)
        {
            var coerce = new Dictionary<string, object?>();
            if (CoercionCfg.From is not null) coerce["from"] = CoercionCfg.From;
            if (CoercionCfg.Trim) coerce["trim"] = true;
            if (CoercionCfg.Lower) coerce["lower"] = true;
            if (CoercionCfg.Upper) coerce["upper"] = true;
            node["coerce"] = coerce;
        }
    }

    internal static string DescribeType(object? value)
    {
        if (value is null) return "null";
        if (Absent.IsAbsent(value)) return "undefined";
        if (value is bool) return "boolean";
        if (value is string) return "string";
        if (value is long or int or short or byte or sbyte or ushort or uint or ulong) return "number";
        if (value is double or float or decimal) return "number";
        if (value is List<object?> or object[]) return "array";
        if (value is Dictionary<string, object?>) return "object";
        return value.GetType().Name.ToLowerInvariant();
    }

    internal static string DescribeValue(object? value)
    {
        if (value is null) return "null";
        if (value is string s) return s;
        return value.ToString() ?? "null";
    }

    /// <summary>
    /// Convert a numeric value to double for comparison.
    /// </summary>
    internal static double ToDouble(object? value)
    {
        return value switch
        {
            double d => d,
            long l => l,
            int i => i,
            float f => f,
            decimal m => (double)m,
            short s => s,
            byte b => b,
            sbyte sb => sb,
            ushort us => us,
            uint ui => ui,
            ulong ul => ul,
            _ => double.NaN
        };
    }

    /// <summary>
    /// Check if a value is a finite number.
    /// </summary>
    internal static bool IsFiniteNumber(object? value)
    {
        if (value is double d) return double.IsFinite(d);
        if (value is float f) return float.IsFinite(f);
        return value is long or int or short or byte or sbyte or ushort or uint or ulong or decimal;
    }

    /// <summary>
    /// Check if a value is an integer (no fractional part).
    /// </summary>
    internal static bool IsInteger(object? value)
    {
        if (value is long or int or short or byte or sbyte or ushort or uint or ulong) return true;
        if (value is double d) return double.IsFinite(d) && d == Math.Floor(d);
        if (value is float f) return float.IsFinite(f) && f == MathF.Floor(f);
        if (value is decimal m) return m == Math.Floor(m);
        return false;
    }

    /// <summary>
    /// Convert a numeric value to long.
    /// </summary>
    internal static long ToLong(object? value)
    {
        return value switch
        {
            long l => l,
            int i => i,
            double d => (long)d,
            float f => (long)f,
            decimal m => (long)m,
            short s => s,
            byte b => b,
            sbyte sb => sb,
            ushort us => us,
            uint ui => ui,
            ulong ul => (long)ul,
            _ => 0
        };
    }
}

/// <summary>
/// Generic base for schemas that provides typed Parse and SafeParse methods.
/// </summary>
public abstract class Schema<T> : Schema
{
    /// <summary>
    /// Parse input with strong typing. Throws ValidationError on failure.
    /// </summary>
    public new T Parse(object? input)
    {
        var result = base.SafeParse(input);
        if (result.Success) return CastData(result.Data)!;
        throw new ValidationError(result.Issues);
    }

    /// <summary>
    /// Parse input with strong typing. Returns ParseResult&lt;T&gt; with success/failure.
    /// </summary>
    public ParseResult<T> SafeParseTyped(object? input)
    {
        var result = base.SafeParse(input);
        if (result.Success)
            return ParseResult<T>.Ok(CastData(result.Data));
        return ParseResult<T>.Fail(result.Issues);
    }

    private static T? CastData(object? data)
    {
        if (data is null) return default;
        if (data is T typed) return typed;
        // Handle numeric type conversions (e.g., long -> double for IntSchema via NumberSchema)
        return (T)Convert.ChangeType(data, Nullable.GetUnderlyingType(typeof(T)) ?? typeof(T));
    }
}

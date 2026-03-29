using AnyVali.Parse;

namespace AnyVali.Schemas;

public sealed class ObjectSchema : Schema
{
    internal sealed class PropertyDef
    {
        public required Schema Schema { get; init; }
        public required bool Required { get; init; }
    }

    private readonly Dictionary<string, PropertyDef> _properties;
    private UnknownKeyMode _unknownKeys;

    public ObjectSchema(Dictionary<string, Schema> shape, UnknownKeyMode unknownKeys = UnknownKeyMode.Reject)
    {
        _unknownKeys = unknownKeys;
        _properties = new Dictionary<string, PropertyDef>();

        foreach (var (key, schema) in shape)
        {
            var isOptional = schema is OptionalSchema;
            _properties[key] = new PropertyDef
            {
                Schema = schema,
                Required = !isOptional,
            };
        }
    }

    // Internal constructor for cloning
    private ObjectSchema(Dictionary<string, PropertyDef> properties, UnknownKeyMode unknownKeys)
    {
        _properties = properties;
        _unknownKeys = unknownKeys;
    }

    public ObjectSchema UnknownKeys(UnknownKeyMode mode)
    {
        var c = (ObjectSchema)Clone();
        c._unknownKeys = mode;
        return c;
    }

    public new ObjectSchema Default(object? value) => (ObjectSchema)base.Default(value);

    internal override object? Validate(object? input, ValidationContext ctx)
    {
        if (input is not Dictionary<string, object?> obj)
        {
            ctx.Issues.Add(new ValidationIssue
            {
                Code = IssueCodes.InvalidType,
                Message = $"Expected object, received {DescribeType(input)}",
                Path = ctx.ClonePath(),
                Expected = "object",
                Received = DescribeType(input),
            });
            return null;
        }

        var result = new Dictionary<string, object?>();
        var inputKeys = new HashSet<string>(obj.Keys);

        // Validate declared properties
        foreach (var (key, prop) in _properties)
        {
            ctx.PushPath(key);
            var hasKey = obj.ContainsKey(key);
            inputKeys.Remove(key);

            if (!hasKey)
            {
                if (prop.Required && Absent.IsAbsent(prop.Schema.DefaultValue))
                {
                    var expectedKind = prop.Schema.ToNode()["kind"]?.ToString() ?? "unknown";
                    ctx.Issues.Add(new ValidationIssue
                    {
                        Code = IssueCodes.Required,
                        Message = $"Required property \"{key}\" is missing",
                        Path = ctx.ClonePath(),
                        Expected = expectedKind,
                        Received = "undefined",
                    });
                    ctx.PopPath();
                    continue;
                }
            }

            var rawValue = hasKey ? obj[key] : Absent.Value;
            var val = prop.Schema.RunPipeline(rawValue, ctx);

            if (val is not null || hasKey || !Absent.IsAbsent(prop.Schema.DefaultValue))
            {
                result[key] = val;
            }

            ctx.PopPath();
        }

        // Handle unknown keys
        foreach (var key in inputKeys)
        {
            switch (_unknownKeys)
            {
                case UnknownKeyMode.Reject:
                    ctx.Issues.Add(new ValidationIssue
                    {
                        Code = IssueCodes.UnknownKey,
                        Message = $"Unknown key \"{key}\"",
                        Path = [..ctx.Path, key],
                        Expected = "undefined",
                        Received = key,
                    });
                    break;
                case UnknownKeyMode.Allow:
                    result[key] = obj[key];
                    break;
                case UnknownKeyMode.Strip:
                    break;
            }
        }

        return result;
    }

    internal override Dictionary<string, object?> ToNode()
    {
        var properties = new Dictionary<string, object?>();
        var required = new List<object?>();

        foreach (var (key, prop) in _properties)
        {
            properties[key] = prop.Schema.ToNode();
            if (prop.Required)
                required.Add(key);
        }

        var node = new Dictionary<string, object?>
        {
            ["kind"] = "object",
            ["properties"] = properties,
            ["required"] = required,
            ["unknownKeys"] = _unknownKeys switch
            {
                UnknownKeyMode.Reject => "reject",
                UnknownKeyMode.Strip => "strip",
                UnknownKeyMode.Allow => "allow",
                _ => "reject"
            },
        };
        AddDefaultAndCoercion(node);
        return node;
    }

    internal override Schema Clone() => new ObjectSchema(_properties, _unknownKeys)
    {
        DefaultValue = DefaultValue, CoercionCfg = CoercionCfg, IsPortable = IsPortable,
    };
}

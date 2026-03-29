using AnyVali.Parse;
using AnyVali.Schemas;

namespace AnyVali.Interchange;

/// <summary>
/// Import an AnyValiDocument back into a live Schema.
/// </summary>
public static class Importer
{
    public static Schema ImportSchema(AnyValiDocument doc)
    {
        var definitions = doc.Definitions ?? new Dictionary<string, object?>();
        var resolvedDefs = new Dictionary<string, Schema>();

        Schema ImportNode(object? nodeObj)
        {
            var node = ToDict(nodeObj);
            var kind = node.TryGetValue("kind", out var k) ? k?.ToString() : null;

            Schema schema;

            switch (kind)
            {
                case "string":
                {
                    var s = new StringSchema();
                    if (node.TryGetValue("minLength", out var ml) && ml is not null)
                        s = s.MinLength(ToInt(ml));
                    if (node.TryGetValue("maxLength", out var mxl) && mxl is not null)
                        s = s.MaxLength(ToInt(mxl));
                    if (node.TryGetValue("pattern", out var p) && p is string ps)
                        s = s.Pattern(ps);
                    if (node.TryGetValue("startsWith", out var sw) && sw is string sws)
                        s = s.StartsWith(sws);
                    if (node.TryGetValue("endsWith", out var ew) && ew is string ews)
                        s = s.EndsWith(ews);
                    if (node.TryGetValue("includes", out var inc) && inc is string incs)
                        s = s.Includes(incs);
                    if (node.TryGetValue("format", out var f) && f is string fs)
                        s = s.Format(fs);
                    schema = s;
                    break;
                }
                case "number":
                case "float64":
                {
                    var s = kind == "float64" ? (NumberSchema)new Float64Schema() : new NumberSchema();
                    schema = ApplyNumericConstraints(s, node);
                    break;
                }
                case "float32":
                    schema = ApplyNumericConstraints(new Float32Schema(), node);
                    break;
                case "int":
                case "int64":
                {
                    var s = kind == "int64" ? (IntSchema)new Int64Schema() : new IntSchema();
                    schema = ApplyNumericConstraints(s, node);
                    break;
                }
                case "int8":
                    schema = ApplyNumericConstraints(new Int8Schema(), node);
                    break;
                case "int16":
                    schema = ApplyNumericConstraints(new Int16Schema(), node);
                    break;
                case "int32":
                    schema = ApplyNumericConstraints(new Int32Schema(), node);
                    break;
                case "uint8":
                    schema = ApplyNumericConstraints(new Uint8Schema(), node);
                    break;
                case "uint16":
                    schema = ApplyNumericConstraints(new Uint16Schema(), node);
                    break;
                case "uint32":
                    schema = ApplyNumericConstraints(new Uint32Schema(), node);
                    break;
                case "uint64":
                    schema = ApplyNumericConstraints(new Uint64Schema(), node);
                    break;
                case "bool":
                    schema = new BoolSchema();
                    break;
                case "null":
                    schema = new NullSchema();
                    break;
                case "any":
                    schema = new AnySchema();
                    break;
                case "unknown":
                    schema = new UnknownSchema();
                    break;
                case "never":
                    schema = new NeverSchema();
                    break;
                case "literal":
                {
                    var value = node.TryGetValue("value", out var v) ? v : null;
                    schema = new LiteralSchema(value);
                    break;
                }
                case "enum":
                {
                    var values = node.TryGetValue("values", out var v) && v is List<object?> list
                        ? list.Where(x => x is not null).Cast<object>().ToArray()
                        : Array.Empty<object>();
                    schema = new EnumSchema(values);
                    break;
                }
                case "array":
                {
                    var items = node.TryGetValue("items", out var it) ? ImportNode(it) : new AnySchema();
                    var s = new ArraySchema(items);
                    if (node.TryGetValue("minItems", out var mi) && mi is not null)
                        s = s.MinItems(ToInt(mi));
                    if (node.TryGetValue("maxItems", out var mxi) && mxi is not null)
                        s = s.MaxItems(ToInt(mxi));
                    schema = s;
                    break;
                }
                case "tuple":
                {
                    // Corpus uses "elements", export uses "items"
                    var elements = node.TryGetValue("elements", out var el) ? el
                        : node.TryGetValue("items", out var it) ? it : null;
                    var items = elements is List<object?> list
                        ? list.Select(ImportNode).ToArray()
                        : Array.Empty<Schema>();
                    schema = new TupleSchema(items);
                    break;
                }
                case "object":
                {
                    var shape = new Dictionary<string, Schema>();
                    var requiredSet = new HashSet<string>();
                    if (node.TryGetValue("required", out var req) && req is List<object?> reqList)
                    {
                        foreach (var r in reqList)
                            if (r is string rs) requiredSet.Add(rs);
                    }
                    if (node.TryGetValue("properties", out var props) && props is Dictionary<string, object?> propsDict)
                    {
                        foreach (var (propKey, propNode) in propsDict)
                        {
                            var propSchema = ImportNode(propNode);
                            if (!requiredSet.Contains(propKey))
                                propSchema = new OptionalSchema(propSchema);
                            shape[propKey] = propSchema;
                        }
                    }
                    var ukMode = UnknownKeyMode.Reject;
                    if (node.TryGetValue("unknownKeys", out var uk) && uk is string uks)
                    {
                        ukMode = uks switch
                        {
                            "strip" => UnknownKeyMode.Strip,
                            "allow" => UnknownKeyMode.Allow,
                            _ => UnknownKeyMode.Reject,
                        };
                    }
                    schema = new ObjectSchema(shape, ukMode);
                    break;
                }
                case "record":
                {
                    // Corpus uses "values", export uses "valueSchema"
                    var valueNode = node.TryGetValue("values", out var v) ? v
                        : node.TryGetValue("valueSchema", out var vs) ? vs : null;
                    schema = new RecordSchema(ImportNode(valueNode));
                    break;
                }
                case "union":
                {
                    var variants = node.TryGetValue("variants", out var v) && v is List<object?> list
                        ? list.Select(ImportNode).ToArray()
                        : Array.Empty<Schema>();
                    schema = new UnionSchema(variants);
                    break;
                }
                case "intersection":
                {
                    var allOf = node.TryGetValue("allOf", out var a) && a is List<object?> list
                        ? list.Select(ImportNode).ToArray()
                        : Array.Empty<Schema>();
                    schema = new IntersectionSchema(allOf);
                    break;
                }
                case "optional":
                {
                    var innerNode = node.TryGetValue("schema", out var s) ? s
                        : node.TryGetValue("inner", out var i) ? i : null;
                    schema = new OptionalSchema(ImportNode(innerNode));
                    break;
                }
                case "nullable":
                {
                    var innerNode = node.TryGetValue("schema", out var s) ? s
                        : node.TryGetValue("inner", out var i) ? i : null;
                    schema = new NullableSchema(ImportNode(innerNode));
                    break;
                }
                case "ref":
                {
                    var refPath = node.TryGetValue("ref", out var r) ? r?.ToString() ?? "" : "";
                    var defName = refPath.Replace("#/definitions/", "");
                    schema = new RefSchema(refPath, () =>
                    {
                        if (resolvedDefs.TryGetValue(defName, out var cached))
                            return cached;
                        if (!definitions.TryGetValue(defName, out var defNode))
                            throw new InvalidOperationException($"Unresolved definition: {defName}");
                        var resolved = ImportNode(defNode);
                        resolvedDefs[defName] = resolved;
                        return resolved;
                    });
                    break;
                }
                default:
                    throw new InvalidOperationException($"Unsupported schema kind: {kind}");
            }

            // Apply default
            if (node.TryGetValue("default", out var defaultVal) && defaultVal is not null)
            {
                schema = schema.Default(defaultVal);
            }

            // Apply coercion
            if (node.TryGetValue("coerce", out var coerceVal) && coerceVal is not null)
            {
                var config = Coercion.NormalizeConfig(coerceVal);
                schema = schema.Coerce(config);
            }

            return schema;
        }

        return ImportNode(doc.Root);
    }

    private static NumberSchema ApplyNumericConstraints(NumberSchema schema, Dictionary<string, object?> node)
    {
        var s = schema;
        if (node.TryGetValue("min", out var min) && min is not null)
            s = s.Min(ToDouble(min));
        if (node.TryGetValue("max", out var max) && max is not null)
            s = s.Max(ToDouble(max));
        if (node.TryGetValue("exclusiveMin", out var emin) && emin is not null)
            s = s.ExclusiveMin(ToDouble(emin));
        if (node.TryGetValue("exclusiveMax", out var emax) && emax is not null)
            s = s.ExclusiveMax(ToDouble(emax));
        if (node.TryGetValue("multipleOf", out var mo) && mo is not null)
            s = s.MultipleOf(ToDouble(mo));
        return s;
    }

    private static Dictionary<string, object?> ToDict(object? obj)
    {
        return obj as Dictionary<string, object?> ?? new Dictionary<string, object?>();
    }

    private static int ToInt(object? value)
    {
        return value switch
        {
            int i => i,
            long l => (int)l,
            double d => (int)d,
            string s => int.Parse(s),
            _ => 0
        };
    }

    private static double ToDouble(object? value)
    {
        return value switch
        {
            double d => d,
            long l => l,
            int i => i,
            float f => f,
            string s => double.Parse(s),
            _ => 0
        };
    }
}

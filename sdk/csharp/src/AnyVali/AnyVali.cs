using AnyVali.Schemas;

namespace AnyVali;

/// <summary>
/// Static factory for creating AnyVali schemas.
/// </summary>
public static class V
{
    public static StringSchema String() => new();
    public static NumberSchema Number() => new();
    public static Float32Schema Float32() => new();
    public static Float64Schema Float64() => new();
    public static IntSchema Int() => new();
    public static Int8Schema Int8() => new();
    public static Int16Schema Int16() => new();
    public static Int32Schema Int32() => new();
    public static Int64Schema Int64() => new();
    public static Uint8Schema Uint8() => new();
    public static Uint16Schema Uint16() => new();
    public static Uint32Schema Uint32() => new();
    public static Uint64Schema Uint64() => new();
    public static BoolSchema Bool() => new();
    public static NullSchema Null() => new();
    public static AnySchema Any() => new();
    public static UnknownSchema Unknown() => new();
    public static NeverSchema Never() => new();
    public static LiteralSchema Literal(object? value) => new(value);
    public static EnumSchema Enum(params object[] values) => new(values);

    public static ArraySchema Array(Schema items) => new(items);
    public static TupleSchema Tuple(params Schema[] items) => new(items);

    public static ObjectSchema Object(
        Dictionary<string, Schema> shape,
        UnknownKeyMode unknownKeys = UnknownKeyMode.Reject) => new(shape, unknownKeys);

    public static RecordSchema Record(Schema valueSchema) => new(valueSchema);
    public static UnionSchema Union(params Schema[] variants) => new(variants);
    public static IntersectionSchema Intersection(params Schema[] schemas) => new(schemas);
    public static OptionalSchema Optional(Schema inner) => new(inner);
    public static NullableSchema Nullable(Schema inner) => new(inner);

    /// <summary>Parse input using the given schema. Throws ValidationError on failure.</summary>
    public static object? Parse(Schema schema, object? input) => schema.Parse(input);

    /// <summary>Parse input using the given schema. Returns ParseResult.</summary>
    public static ParseResult SafeParse(Schema schema, object? input) => schema.SafeParse(input);

    /// <summary>Export a schema to an AnyValiDocument.</summary>
    public static AnyValiDocument Export(Schema schema, ExportMode mode = ExportMode.Portable) =>
        Interchange.Exporter.ExportSchema(schema, mode);

    /// <summary>Import an AnyValiDocument to a live schema.</summary>
    public static Schema Import(AnyValiDocument doc) =>
        Interchange.Importer.ImportSchema(doc);
}

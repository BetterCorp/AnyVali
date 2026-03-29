namespace AnyVali.Interchange;

/// <summary>
/// Export a schema to an AnyValiDocument.
/// </summary>
public static class Exporter
{
    public static AnyValiDocument ExportSchema(Schema schema, ExportMode mode = ExportMode.Portable)
    {
        return schema.Export(mode);
    }
}

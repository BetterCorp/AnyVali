namespace AnyVali;

/// <summary>
/// Export mode for schema interchange.
/// </summary>
public enum ExportMode
{
    /// <summary>Fails if the schema contains non-portable features.</summary>
    Portable,
    /// <summary>Emits core schema plus extension namespaces.</summary>
    Extended
}

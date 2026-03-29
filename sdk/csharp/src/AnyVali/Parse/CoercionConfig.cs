namespace AnyVali.Parse;

/// <summary>
/// Configuration for value coercion.
/// </summary>
public sealed record CoercionConfig
{
    public string? From { get; init; }
    public bool Trim { get; init; }
    public bool Lower { get; init; }
    public bool Upper { get; init; }
}

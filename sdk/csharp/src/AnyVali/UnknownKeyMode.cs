namespace AnyVali;

/// <summary>
/// How to handle unknown keys in object schemas.
/// </summary>
public enum UnknownKeyMode
{
    /// <summary>Reject unknown keys.</summary>
    Reject,
    /// <summary>Remove unknown keys from output (default).</summary>
    Strip,
    /// <summary>Pass unknown keys through.</summary>
    Allow
}

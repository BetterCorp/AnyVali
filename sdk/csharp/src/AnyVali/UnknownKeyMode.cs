namespace AnyVali;

/// <summary>
/// How to handle unknown keys in object schemas.
/// </summary>
public enum UnknownKeyMode
{
    /// <summary>Reject unknown keys (default).</summary>
    Reject,
    /// <summary>Remove unknown keys from output.</summary>
    Strip,
    /// <summary>Pass unknown keys through.</summary>
    Allow
}

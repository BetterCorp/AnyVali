namespace AnyVali;

/// <summary>
/// Internal context for tracking path and issues during validation.
/// </summary>
public sealed class ValidationContext
{
    public List<object> Path { get; } = new(); // string or int
    public List<ValidationIssue> Issues { get; } = new();
    public UnknownKeyMode? InheritedUnknownKeys { get; set; }
    internal string? SensitiveMode { get; set; }
    internal Func<IReadOnlyList<object>, object?, object?>? SensitiveTransform { get; set; }
    internal Dictionary<string, object?>? SensitiveCache { get; set; }

    public List<object> ClonePath() => new(Path);

    public void PushPath(object segment) => Path.Add(segment);
    public void PopPath() => Path.RemoveAt(Path.Count - 1);

    internal void InheritSensitive(ValidationContext parent)
    {
        SensitiveMode = parent.SensitiveMode;
        SensitiveTransform = parent.SensitiveTransform;
        SensitiveCache = parent.SensitiveCache;
    }
}

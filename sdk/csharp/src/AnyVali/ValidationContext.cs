namespace AnyVali;

/// <summary>
/// Internal context for tracking path and issues during validation.
/// </summary>
public sealed class ValidationContext
{
    public List<object> Path { get; } = new(); // string or int
    public List<ValidationIssue> Issues { get; } = new();

    public List<object> ClonePath() => new(Path);

    public void PushPath(object segment) => Path.Add(segment);
    public void PopPath() => Path.RemoveAt(Path.Count - 1);
}

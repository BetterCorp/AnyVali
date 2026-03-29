namespace AnyVali;

/// <summary>
/// Result of a SafeParse operation.
/// </summary>
public sealed class ParseResult
{
    public bool Success { get; }
    public object? Data { get; }
    public IReadOnlyList<ValidationIssue> Issues { get; }

    private ParseResult(bool success, object? data, IReadOnlyList<ValidationIssue> issues)
    {
        Success = success;
        Data = data;
        Issues = issues;
    }

    public static ParseResult Ok(object? data) =>
        new(true, data, Array.Empty<ValidationIssue>());

    public static ParseResult Fail(IReadOnlyList<ValidationIssue> issues) =>
        new(false, null, issues);
}

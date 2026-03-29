namespace AnyVali;

/// <summary>
/// Exception thrown by Parse() when validation fails.
/// </summary>
public sealed class ValidationError : Exception
{
    public IReadOnlyList<ValidationIssue> Issues { get; }

    public ValidationError(IReadOnlyList<ValidationIssue> issues)
        : base(FormatMessage(issues))
    {
        Issues = issues;
    }

    private static string FormatMessage(IReadOnlyList<ValidationIssue> issues)
    {
        return string.Join("\n", issues.Select(i =>
        {
            var pathStr = i.Path.Count > 0
                ? string.Join(".", i.Path) + ": "
                : "";
            return $"[{i.Code}] {pathStr}{i.Message}";
        }));
    }
}

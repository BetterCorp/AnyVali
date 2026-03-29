namespace AnyVali.Parse;

/// <summary>
/// Sentinel for "value not present".
/// </summary>
public static class Absent
{
    public static readonly object Value = new AbsentMarker();

    public static bool IsAbsent(object? value) => value is AbsentMarker;

    private sealed class AbsentMarker
    {
        public override string ToString() => "<absent>";
    }
}

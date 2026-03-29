using System.Text.Json.Serialization;

namespace AnyVali;

/// <summary>
/// Represents a single validation issue.
/// </summary>
public sealed class ValidationIssue
{
    [JsonPropertyName("code")]
    public required string Code { get; init; }

    [JsonPropertyName("message")]
    public required string Message { get; init; }

    [JsonPropertyName("path")]
    public required List<object> Path { get; init; } // string or int

    [JsonPropertyName("expected")]
    public string? Expected { get; init; }

    [JsonPropertyName("received")]
    public string? Received { get; init; }

    [JsonPropertyName("meta")]
    public Dictionary<string, object>? Meta { get; init; }
}

using System.Text.Json;
using System.Text.Json.Serialization;

namespace AnyVali;

/// <summary>
/// Represents a portable AnyVali interchange document.
/// </summary>
public sealed class AnyValiDocument
{
    [JsonPropertyName("anyvaliVersion")]
    public required string AnyvaliVersion { get; init; }

    [JsonPropertyName("schemaVersion")]
    public required string SchemaVersion { get; init; }

    [JsonPropertyName("root")]
    public required Dictionary<string, object?> Root { get; init; }

    [JsonPropertyName("definitions")]
    public required Dictionary<string, object?> Definitions { get; init; }

    [JsonPropertyName("extensions")]
    public required Dictionary<string, object?> Extensions { get; init; }

    public string ToJson()
    {
        return JsonSerializer.Serialize(this, JsonHelper.SerializerOptions);
    }

    public static AnyValiDocument FromJson(string json)
    {
        return JsonSerializer.Deserialize<AnyValiDocument>(json, JsonHelper.SerializerOptions)
            ?? throw new InvalidOperationException("Failed to deserialize AnyValiDocument");
    }
}

internal static class JsonHelper
{
    public static readonly JsonSerializerOptions SerializerOptions = new()
    {
        WriteIndented = false,
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        Converters = { new ObjectJsonConverter() }
    };

    /// <summary>
    /// Converts a JsonElement to a plain .NET object (used during deserialization).
    /// </summary>
    public static object? ElementToObject(JsonElement element)
    {
        return element.ValueKind switch
        {
            JsonValueKind.String => element.GetString(),
            JsonValueKind.Number => element.TryGetInt64(out var l) ? (object)l : element.GetDouble(),
            JsonValueKind.True => true,
            JsonValueKind.False => false,
            JsonValueKind.Null => null,
            JsonValueKind.Array => element.EnumerateArray().Select(ElementToObject).ToList(),
            JsonValueKind.Object => element.EnumerateObject()
                .ToDictionary(p => p.Name, p => ElementToObject(p.Value)),
            _ => null
        };
    }
}

/// <summary>
/// Custom converter that deserializes arbitrary JSON into plain .NET objects
/// (Dictionary, List, string, long, double, bool, null).
/// </summary>
internal sealed class ObjectJsonConverter : JsonConverter<object?>
{
    public override object? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
    {
        return reader.TokenType switch
        {
            JsonTokenType.String => reader.GetString(),
            JsonTokenType.Number => reader.TryGetInt64(out var l) ? (object)l : reader.GetDouble(),
            JsonTokenType.True => true,
            JsonTokenType.False => false,
            JsonTokenType.Null => null,
            JsonTokenType.StartArray => ReadArray(ref reader, options),
            JsonTokenType.StartObject => ReadObject(ref reader, options),
            _ => throw new JsonException($"Unexpected token {reader.TokenType}")
        };
    }

    private static List<object?> ReadArray(ref Utf8JsonReader reader, JsonSerializerOptions options)
    {
        var list = new List<object?>();
        while (reader.Read())
        {
            if (reader.TokenType == JsonTokenType.EndArray) return list;
            list.Add(JsonSerializer.Deserialize<object?>(ref reader, options));
        }
        throw new JsonException("Unexpected end of array");
    }

    private static Dictionary<string, object?> ReadObject(ref Utf8JsonReader reader, JsonSerializerOptions options)
    {
        var dict = new Dictionary<string, object?>();
        while (reader.Read())
        {
            if (reader.TokenType == JsonTokenType.EndObject) return dict;
            if (reader.TokenType != JsonTokenType.PropertyName)
                throw new JsonException("Expected property name");
            var key = reader.GetString()!;
            reader.Read();
            dict[key] = JsonSerializer.Deserialize<object?>(ref reader, options);
        }
        throw new JsonException("Unexpected end of object");
    }

    public override void Write(Utf8JsonWriter writer, object? value, JsonSerializerOptions options)
    {
        WriteValue(writer, value, options);
    }

    private static void WriteValue(Utf8JsonWriter writer, object? value, JsonSerializerOptions options)
    {
        switch (value)
        {
            case null:
                writer.WriteNullValue();
                break;
            case string s:
                writer.WriteStringValue(s);
                break;
            case bool b:
                writer.WriteBooleanValue(b);
                break;
            case int i:
                writer.WriteNumberValue(i);
                break;
            case long l:
                writer.WriteNumberValue(l);
                break;
            case double d:
                writer.WriteNumberValue(d);
                break;
            case float f:
                writer.WriteNumberValue(f);
                break;
            case Dictionary<string, object?> dict:
                writer.WriteStartObject();
                foreach (var (k, v) in dict)
                {
                    writer.WritePropertyName(k);
                    WriteValue(writer, v, options);
                }
                writer.WriteEndObject();
                break;
            case List<object?> list:
                writer.WriteStartArray();
                foreach (var item in list)
                    WriteValue(writer, item, options);
                writer.WriteEndArray();
                break;
            case IList<object?> list:
                writer.WriteStartArray();
                foreach (var item in list)
                    WriteValue(writer, item, options);
                writer.WriteEndArray();
                break;
            default:
                // Fallback: use default serializer
                JsonSerializer.Serialize(writer, value, value.GetType(), options);
                break;
        }
    }
}

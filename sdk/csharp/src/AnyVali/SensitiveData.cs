namespace AnyVali;

/// <summary>Explicit encrypted-storage helpers for sensitive schema nodes.</summary>
public static class SensitiveData
{
    public static ParseResult SafeParseEncrypted(Schema schema, object? data)
    {
        var ctx = new ValidationContext { SensitiveMode = "encrypted" };
        var output = schema.RunPipeline(data, ctx);
        return ctx.Issues.Count > 0 ? ParseResult.Fail(ctx.Issues) : ParseResult.Ok(output);
    }

    public static object? Encrypt(
        Schema schema,
        object? data,
        Func<IReadOnlyList<object>, object?, object?> transform)
    {
        var encrypted = Transform(schema, schema.Parse(data), "encrypt", transform);
        var checkedData = SafeParseEncrypted(schema, encrypted);
        if (!checkedData.Success) throw new ValidationError(checkedData.Issues);
        return checkedData.Data;
    }

    public static object? Decrypt(
        Schema schema,
        object? data,
        Func<IReadOnlyList<object>, object?, object?> transform)
    {
        var checkedData = SafeParseEncrypted(schema, data);
        if (!checkedData.Success) throw new ValidationError(checkedData.Issues);
        return schema.Parse(Transform(schema, checkedData.Data, "decrypt", transform));
    }

    private static object? Transform(
        Schema schema,
        object? data,
        string mode,
        Func<IReadOnlyList<object>, object?, object?> transform)
    {
        var ctx = new ValidationContext
        {
            SensitiveMode = mode,
            SensitiveTransform = transform,
            SensitiveCache = new Dictionary<string, object?>(),
        };
        var output = schema.RunPipeline(data, ctx);
        if (ctx.Issues.Count > 0) throw new ValidationError(ctx.Issues);
        return output;
    }
}

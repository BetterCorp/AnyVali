package com.anyvali;

import java.util.HashMap;
import java.util.List;
import java.util.function.BiFunction;

/** Explicit encrypted-storage helpers for sensitive schema nodes. */
public final class SensitiveData {
    private SensitiveData() {}

    public static ParseResult<Object> safeParseEncrypted(Schema<?> schema, Object data) {
        var ctx = new ValidationContext();
        ctx.setSensitiveMode("encrypted");
        Object output = schema.runPipeline(data, ctx);
        return ctx.hasIssues() ? ParseResult.failure(ctx.getIssues()) : ParseResult.success(output);
    }

    public static Object encrypt(Schema<?> schema, Object data,
                                 BiFunction<List<Object>, Object, Object> transform) {
        Object encrypted = transform(schema, schema.parse(data), "encrypt", transform);
        var checked = safeParseEncrypted(schema, encrypted);
        if (!checked.success()) throw new ValidationError(checked.issues());
        return checked.data();
    }

    public static <T> T decrypt(Schema<T> schema, Object data,
                                BiFunction<List<Object>, Object, Object> transform) {
        var checked = safeParseEncrypted(schema, data);
        if (!checked.success()) throw new ValidationError(checked.issues());
        return schema.parse(transform(schema, checked.data(), "decrypt", transform));
    }

    private static Object transform(Schema<?> schema, Object data, String mode,
                                    BiFunction<List<Object>, Object, Object> transform) {
        var ctx = new ValidationContext();
        ctx.setSensitiveMode(mode);
        ctx.setSensitiveTransform(transform);
        ctx.setSensitiveCache(new HashMap<>());
        Object output = schema.runPipeline(data, ctx);
        if (ctx.hasIssues()) throw new ValidationError(ctx.getIssues());
        return output;
    }
}

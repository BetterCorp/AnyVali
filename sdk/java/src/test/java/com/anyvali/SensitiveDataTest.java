package com.anyvali;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

class SensitiveDataTest {
    @Test
    void encryptedRoundTripAndValidation() {
        var sensitive = new Schema.DescribeOptions().sensitive(true);
        Schema<?> schema = object_(Map.of(
                "name", string(),
                "secret", string().describe("secret", sensitive),
                "profile", object_(Map.of("token", string())).describe("profile", sensitive),
                "aliases", array(string().describe("alias", sensitive))
        ));
        var plain = Map.of(
                "name", "Ada", "secret", "abc",
                "profile", Map.of("token", "xyz"), "aliases", List.of("one", "two")
        );

        Object encrypted = encrypt(schema, plain, (path, value) -> "encrypted:" + value);
        assertTrue(safeParseEncrypted(schema, encrypted).success());
        assertThrows(ValidationError.class,
                () -> encrypt(schema, plain, (path, value) -> "broken"));
        assertNotNull(decrypt(schema, encrypted,
                (path, value) -> switch (String.join(".", path.stream().map(Object::toString).toList())) {
                    case "profile" -> Map.of("token", "xyz");
                    case "aliases.0" -> "one";
                    case "aliases.1" -> "two";
                    default -> "abc";
                }));
    }
}

package com.anyvali;

import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

class DefaultsTest {

    @Test
    void missingFieldGetsDefault() {
        var schema = object_(Map.of(
                "name", string(),
                "role", string().withDefault("user")
        ), Set.of("name"));

        var result = schema.safeParse(Map.of("name", "Alice"));

        assertTrue(result.success(), () -> result.issues().toString());
        assertEquals("user", result.data().get("role"));
    }

    @Test
    void presentFieldIsNotOverwritten() {
        var schema = object_(Map.of(
                "role", string().withDefault("user")
        ));

        var result = schema.safeParse(Map.of("role", "admin"));

        assertTrue(result.success(), () -> result.issues().toString());
        assertEquals("admin", result.data().get("role"));
    }

    @Test
    void invalidDefaultProducesDefaultInvalid() {
        var schema = object_(Map.of(
                "count", int_().min(10).withDefault(5)
        ));

        var result = schema.safeParse(Map.of());

        assertFalse(result.success());
        assertEquals(IssueCodes.DEFAULT_INVALID, result.issues().get(0).code());
        assertEquals(List.of("count"), result.issues().get(0).path());
    }

    @Test
    void nullIsNotAbsentForNullableDefault() {
        var schema = object_(Map.of(
                "value", nullable(string()).withDefault("fallback")
        ));

        var input = new LinkedHashMap<String, Object>();
        input.put("value", null);
        var result = schema.safeParse(input);

        assertTrue(result.success(), () -> result.issues().toString());
        assertNull(result.data().get("value"));
    }

    @Test
    void falsyDefaultsAreApplied() {
        var schema = object_(Map.of(
                "count", int_().withDefault(0),
                "name", string().withDefault(""),
                "active", bool_().withDefault(false)
        ));

        var result = schema.safeParse(Map.of());

        assertTrue(result.success(), () -> result.issues().toString());
        assertEquals(0, result.data().get("count"));
        assertEquals("", result.data().get("name"));
        assertEquals(false, result.data().get("active"));
    }

    @Test
    void optionalWrapperFieldGetsDefault() {
        var schema = object_(Map.of(
                "host", optional(string()).withDefault("localhost")
        ));

        var result = schema.safeParse(Map.of());

        assertTrue(result.success(), () -> result.issues().toString());
        assertEquals("localhost", result.data().get("host"));
    }

    @Test
    void optionalWrapperDefaultDoesNotOverridePresentField() {
        var schema = object_(Map.of(
                "host", optional(string()).withDefault("localhost")
        ));

        var result = schema.safeParse(Map.of("host", "example.com"));

        assertTrue(result.success(), () -> result.issues().toString());
        assertEquals("example.com", result.data().get("host"));
    }

    @Test
    void optionalWrapperDefaultIsValidated() {
        var schema = object_(Map.of(
                "host", optional(string().minLength(5)).withDefault("hi")
        ));

        var result = schema.safeParse(Map.of());

        assertFalse(result.success());
        assertEquals(IssueCodes.DEFAULT_INVALID, result.issues().get(0).code());
        assertEquals(List.of("host"), result.issues().get(0).path());
    }

    @Test
    void optionalWrapperDefaultIsExported() {
        var doc = optional(string()).withDefault("localhost").export();
        @SuppressWarnings("unchecked")
        var root = (Map<String, Object>) doc.get("root");

        assertEquals("optional", root.get("kind"));
        assertEquals("localhost", root.get("default"));
    }

    @Test
    void nestedObjectFieldGetsDefault() {
        var schema = object_(Map.of(
                "user", object_(Map.of(
                        "name", string(),
                        "role", string().withDefault("guest")
                ), Set.of("name"))
        ), Set.of("user"));

        var result = schema.safeParse(Map.of("user", Map.of("name", "Bob")));

        assertTrue(result.success(), () -> result.issues().toString());
        @SuppressWarnings("unchecked")
        var user = (Map<String, Object>) result.data().get("user");
        assertEquals("guest", user.get("role"));
    }

    @Test
    void mutableDefaultIsNotSharedBetweenParses() {
        var schema = object_(Map.of(
                "meta", any_().withDefault(Map.of("items", new ArrayList<>()))
        ));

        var first = schema.parse(Map.of());
        @SuppressWarnings("unchecked")
        var firstMeta = (Map<String, Object>) first.get("meta");
        @SuppressWarnings("unchecked")
        var firstItems = (List<Object>) firstMeta.get("items");
        firstItems.add("mutated");

        var second = schema.parse(Map.of());
        @SuppressWarnings("unchecked")
        var secondMeta = (Map<String, Object>) second.get("meta");
        assertEquals(List.of(), secondMeta.get("items"));
    }

    @Test
    void mutableOptionalWrapperDefaultIsNotSharedBetweenParses() {
        var schema = object_(Map.of(
                "meta", optional(any_()).withDefault(Map.of("items", new ArrayList<>()))
        ));

        var first = schema.parse(Map.of());
        @SuppressWarnings("unchecked")
        var firstMeta = (Map<String, Object>) first.get("meta");
        @SuppressWarnings("unchecked")
        var firstItems = (List<Object>) firstMeta.get("items");
        firstItems.add("mutated");

        var second = schema.parse(Map.of());
        @SuppressWarnings("unchecked")
        var secondMeta = (Map<String, Object>) second.get("meta");
        assertEquals(List.of(), secondMeta.get("items"));
    }
}

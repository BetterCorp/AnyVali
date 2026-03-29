package com.anyvali;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Top-level AnyVali interchange document.
 */
public record AnyValiDocument(
        String anyvaliVersion,
        String schemaVersion,
        Map<String, Object> root,
        Map<String, Object> definitions,
        Map<String, Object> extensions
) {
    public AnyValiDocument {
        anyvaliVersion = anyvaliVersion != null ? anyvaliVersion : "1.0";
        schemaVersion = schemaVersion != null ? schemaVersion : "1";
        root = root != null ? root : Map.of();
        definitions = definitions != null ? definitions : Map.of();
        extensions = extensions != null ? extensions : Map.of();
    }

    public AnyValiDocument(Map<String, Object> root) {
        this("1.0", "1", root, Map.of(), Map.of());
    }

    public AnyValiDocument(Map<String, Object> root, Map<String, Object> definitions,
                           Map<String, Object> extensions) {
        this("1.0", "1", root, definitions, extensions);
    }

    /**
     * Convert to a Map suitable for JSON serialization.
     */
    public Map<String, Object> toMap() {
        var map = new LinkedHashMap<String, Object>();
        map.put("anyvaliVersion", anyvaliVersion);
        map.put("schemaVersion", schemaVersion);
        map.put("root", root);
        if (definitions != null && !definitions.isEmpty()) {
            map.put("definitions", definitions);
        }
        if (extensions != null && !extensions.isEmpty()) {
            map.put("extensions", extensions);
        }
        return map;
    }

    /**
     * Create from a Map (e.g. parsed from JSON).
     */
    @SuppressWarnings("unchecked")
    public static AnyValiDocument fromMap(Map<String, Object> map) {
        return new AnyValiDocument(
                (String) map.getOrDefault("anyvaliVersion", "1.0"),
                (String) map.getOrDefault("schemaVersion", "1"),
                (Map<String, Object>) map.getOrDefault("root", Map.of()),
                (Map<String, Object>) map.getOrDefault("definitions", Map.of()),
                (Map<String, Object>) map.getOrDefault("extensions", Map.of())
        );
    }
}

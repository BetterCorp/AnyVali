package com.anyvali.interchange;

import com.anyvali.AnyValiDocument;
import com.anyvali.ExportMode;
import com.anyvali.Schema;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Export schemas to AnyVali interchange format.
 */
public final class Exporter {
    private Exporter() {}

    /**
     * Export a schema to an AnyVali document map.
     */
    public static Map<String, Object> exportSchema(Schema<?> schema) {
        return exportSchema(schema, ExportMode.PORTABLE, null, null);
    }

    /**
     * Export a schema to an AnyVali document map.
     */
    public static Map<String, Object> exportSchema(Schema<?> schema, ExportMode mode,
                                                    Map<String, ? extends Schema<?>> definitions,
                                                    Map<String, Object> extensions) {
        Map<String, Object> rootNode = schema.toPortableNode();

        Map<String, Object> defs = new LinkedHashMap<>();
        if (definitions != null) {
            for (var entry : definitions.entrySet()) {
                defs.put(entry.getKey(), entry.getValue().toPortableNode());
            }
        }

        Map<String, Object> ext = new LinkedHashMap<>();
        if (mode == ExportMode.EXTENDED && extensions != null) {
            ext.putAll(extensions);
        }

        AnyValiDocument doc = new AnyValiDocument(rootNode, defs, ext);
        return doc.toMap();
    }

    /**
     * Export a schema to a JSON string.
     */
    public static String exportSchemaJson(Schema<?> schema) {
        return exportSchemaJson(schema, ExportMode.PORTABLE, null, null);
    }

    /**
     * Export a schema to a JSON string.
     */
    public static String exportSchemaJson(Schema<?> schema, ExportMode mode,
                                          Map<String, ? extends Schema<?>> definitions,
                                          Map<String, Object> extensions) {
        Map<String, Object> doc = exportSchema(schema, mode, definitions, extensions);
        return JsonHelper.toJson(doc);
    }
}

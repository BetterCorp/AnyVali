package com.anyvali;

import com.anyvali.interchange.Exporter;
import com.anyvali.interchange.Importer;
import com.anyvali.interchange.JsonHelper;
import com.anyvali.parse.CoercionConfig;
import com.anyvali.schemas.*;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.*;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Export/import round-trip tests.
 */
class InterchangeTest {

    // ---- JSON Helper ----
    @Nested
    class JsonHelperTests {
        @Test
        void serializeString() {
            assertEquals("\"hello\"", JsonHelper.toJson("hello"));
        }

        @Test
        void serializeNull() {
            assertEquals("null", JsonHelper.toJson(null));
        }

        @Test
        void serializeBoolean() {
            assertEquals("true", JsonHelper.toJson(true));
            assertEquals("false", JsonHelper.toJson(false));
        }

        @Test
        void serializeNumber() {
            assertEquals("42", JsonHelper.toJson(42));
            assertEquals("42", JsonHelper.toJson(42L));
            assertEquals("3.14", JsonHelper.toJson(3.14));
        }

        @Test
        void serializeMap() {
            var map = new LinkedHashMap<String, Object>();
            map.put("a", 1);
            map.put("b", "hello");
            String json = JsonHelper.toJson(map);
            assertTrue(json.contains("\"a\""));
            assertTrue(json.contains("\"b\""));
        }

        @Test
        void serializeList() {
            String json = JsonHelper.toJson(List.of(1, "two", true));
            assertTrue(json.contains("1"));
            assertTrue(json.contains("\"two\""));
        }

        @Test
        void serializeEscapedString() {
            String json = JsonHelper.toJson("hello\n\"world\"");
            assertTrue(json.contains("\\n"));
            assertTrue(json.contains("\\\""));
        }

        @Test
        void parseString() {
            assertEquals("hello", JsonHelper.parseJson("\"hello\""));
        }

        @Test
        void parseNumber() {
            assertEquals(42, JsonHelper.parseJson("42"));
            assertEquals(3.14, JsonHelper.parseJson("3.14"));
        }

        @Test
        void parseBoolean() {
            assertEquals(true, JsonHelper.parseJson("true"));
            assertEquals(false, JsonHelper.parseJson("false"));
        }

        @Test
        void parseNull() {
            assertNull(JsonHelper.parseJson("null"));
        }

        @Test
        @SuppressWarnings("unchecked")
        void parseObject() {
            var map = (Map<String, Object>) JsonHelper.parseJson("{\"a\": 1, \"b\": \"hello\"}");
            assertEquals(1, map.get("a"));
            assertEquals("hello", map.get("b"));
        }

        @Test
        @SuppressWarnings("unchecked")
        void parseArray() {
            var list = (List<Object>) JsonHelper.parseJson("[1, \"two\", true, null]");
            assertEquals(4, list.size());
            assertEquals(1, list.get(0));
            assertEquals("two", list.get(1));
            assertEquals(true, list.get(2));
            assertNull(list.get(3));
        }

        @Test
        void parseEmpty() {
            assertEquals(Map.of(), JsonHelper.parseJson("{}"));
            assertEquals(List.of(), JsonHelper.parseJson("[]"));
        }

        @Test
        void parseUnicode() {
            assertEquals("A", JsonHelper.parseJson("\"\\u0041\""));
        }

        @Test
        void parseNegativeNumber() {
            assertEquals(-42, JsonHelper.parseJson("-42"));
            assertEquals(-3.14, JsonHelper.parseJson("-3.14"));
        }

        @Test
        void parseScientificNotation() {
            assertEquals(1e10, JsonHelper.parseJson("1e10"));
            assertEquals(1.5e-3, JsonHelper.parseJson("1.5e-3"));
        }

        @Test
        void roundTrip() {
            var original = new LinkedHashMap<String, Object>();
            original.put("name", "test");
            original.put("count", 42);
            original.put("valid", true);
            original.put("items", List.of(1, 2, 3));
            original.put("nested", Map.of("key", "value"));

            String json = JsonHelper.toJson(original);
            @SuppressWarnings("unchecked")
            var parsed = (Map<String, Object>) JsonHelper.parseJson(json);

            assertEquals("test", parsed.get("name"));
            assertEquals(42, parsed.get("count"));
            assertEquals(true, parsed.get("valid"));
        }
    }

    // ---- Export/Import Round-Trip ----
    @Nested
    class RoundTripTests {
        @Test
        void stringSchema() {
            var schema = string().minLength(1).maxLength(100);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
            assertFalse(imported.safeParse("").success());
        }

        @Test
        void numberSchema() {
            var schema = number().min(0).max(100);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(50.0).success());
            assertFalse(imported.safeParse(-1.0).success());
        }

        @Test
        void intSchema() {
            var schema = int_().min(0).max(100);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(50).success());
            assertFalse(imported.safeParse(-1).success());
        }

        @Test
        void int8Schema() {
            var doc = Exporter.exportSchema(int8());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(127).success());
            assertFalse(imported.safeParse(128).success());
        }

        @Test
        void uint32Schema() {
            var doc = Exporter.exportSchema(uint32());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(0).success());
            assertFalse(imported.safeParse(-1).success());
        }

        @Test
        void float32Schema() {
            var doc = Exporter.exportSchema(float32());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(1.0).success());
        }

        @Test
        void float64Schema() {
            var doc = Exporter.exportSchema(float64());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(3.14).success());
        }

        @Test
        void boolSchema() {
            var doc = Exporter.exportSchema(bool_());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(true).success());
            assertFalse(imported.safeParse("true").success());
        }

        @Test
        void nullSchema() {
            var doc = Exporter.exportSchema(null_());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(null).success());
            assertFalse(imported.safeParse("null").success());
        }

        @Test
        void anySchema() {
            var doc = Exporter.exportSchema(any_());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("anything").success());
            assertTrue(imported.safeParse(null).success());
        }

        @Test
        void unknownSchema() {
            var doc = Exporter.exportSchema(unknown());
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("anything").success());
        }

        @Test
        void neverSchema() {
            var doc = Exporter.exportSchema(never());
            Schema imported = Importer.importSchema(doc);
            assertFalse(imported.safeParse("anything").success());
        }

        @Test
        void literalSchema() {
            var doc = Exporter.exportSchema(literal("hello"));
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
            assertFalse(imported.safeParse("world").success());
        }

        @Test
        void enumSchema() {
            var doc = Exporter.exportSchema(enum_(List.of("a", "b", "c")));
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("a").success());
            assertFalse(imported.safeParse("d").success());
        }

        @Test
        void arraySchema() {
            var schema = array(string()).minItems(1).maxItems(3);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(List.of("a")).success());
            assertFalse(imported.safeParse(List.of()).success());
        }

        @Test
        void tupleSchema() {
            var schema = tuple_(List.of(string(), int_()));
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(List.of("hello", 42)).success());
            assertFalse(imported.safeParse(List.of("hello")).success());
        }

        @Test
        void objectSchema() {
            var schema = object_(Map.of("name", string(), "age", int_()));
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(Map.of("name", "Alice", "age", 30)).success());
            assertFalse(imported.safeParse(Map.of("name", "Alice")).success());
        }

        @Test
        void recordSchema() {
            var schema = record(int_());
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(Map.of("a", 1, "b", 2)).success());
        }

        @Test
        void unionSchema() {
            var schema = union(List.of(string(), int_()));
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
            assertTrue(imported.safeParse(42).success());
            assertFalse(imported.safeParse(true).success());
        }

        @Test
        void intersectionSchema() {
            var s1 = object_(Map.of("a", string()), null, UnknownKeyMode.STRIP);
            var s2 = object_(Map.of("b", int_()), null, UnknownKeyMode.STRIP);
            var schema = intersection(List.of(s1, s2));
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(Map.of("a", "hello", "b", 42)).success());
        }

        @Test
        void optionalSchema() {
            var schema = optional(string());
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
        }

        @Test
        void nullableSchema() {
            var schema = nullable(string());
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello").success());
            assertTrue(imported.safeParse(null).success());
        }

        @Test
        void refSchema() {
            var doc = Exporter.exportSchema(ref("#/definitions/Name"));
            @SuppressWarnings("unchecked")
            var root = (Map<String, Object>) doc.get("root");
            assertEquals("ref", root.get("kind"));
            assertEquals("#/definitions/Name", root.get("ref"));
        }
    }

    // ---- JSON String Round-Trip ----
    @Nested
    class JsonRoundTripTests {
        @Test
        void exportAndImportJson() {
            var schema = string().minLength(1);
            String json = Exporter.exportSchemaJson(schema);
            Schema imported = Importer.importSchema(json);
            assertTrue(imported.safeParse("hello").success());
            assertFalse(imported.safeParse("").success());
        }

        @Test
        void objectSchemaJsonRoundTrip() {
            var schema = object_(Map.of(
                    "name", string().minLength(1),
                    "age", int_().min(0)
            ));
            String json = Exporter.exportSchemaJson(schema);
            Schema imported = Importer.importSchema(json);
            assertTrue(imported.safeParse(Map.of("name", "Alice", "age", 30)).success());
        }
    }

    // ---- Definitions Round-Trip ----
    @Nested
    class DefinitionsTests {
        @Test
        void exportWithDefinitions() {
            var nameSchema = string().minLength(1);
            Map<String, Schema> defs = Map.of("Name", nameSchema);
            var rootSchema = object_(Map.of("name", ref("#/definitions/Name")));

            var doc = Exporter.exportSchema(rootSchema, ExportMode.PORTABLE, defs, null);
            assertNotNull(doc.get("definitions"));

            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(Map.of("name", "Alice")).success());
            assertFalse(imported.safeParse(Map.of("name", "")).success());
        }
    }

    // ---- Extended export ----
    @Nested
    class ExtendedExportTests {
        @Test
        void extendedIncludesExtensions() {
            var schema = string();
            Map<String, Object> extensions = Map.of("java", Map.of("customValidator", "myValidator"));
            var doc = Exporter.exportSchema(schema, ExportMode.EXTENDED, null, extensions);
            @SuppressWarnings("unchecked")
            var ext = (Map<String, Object>) doc.get("extensions");
            assertNotNull(ext);
            assertTrue(ext.containsKey("java"));
        }

        @Test
        void portableOmitsExtensions() {
            var schema = string();
            Map<String, Object> extensions = Map.of("java", Map.of("custom", "data"));
            var doc = Exporter.exportSchema(schema, ExportMode.PORTABLE, null, extensions);
            // In portable mode, extensions should not be included
            assertFalse(doc.containsKey("extensions"));
        }
    }

    // ---- Coercion Round-Trip ----
    @Nested
    class CoercionRoundTripTests {
        @Test
        void coercionExportImport() {
            var schema = string().coerce(CoercionConfig.builder().trim(true).lower(true).build());
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertEquals("hello", imported.parse("  HELLO  "));
        }

        @Test
        void coercionToIntRoundTrip() {
            var schema = int_().coerce(CoercionConfig.builder().toInt(true).build());
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertEquals(42L, imported.parse("42"));
        }
    }

    // ---- Default Round-Trip ----
    @Nested
    class DefaultRoundTripTests {
        @Test
        void defaultExportImport() {
            var schema = string().withDefault("hello");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.hasDefault);
        }
    }

    // ---- Document Structure ----
    @Nested
    class DocumentStructureTests {
        @Test
        void documentHasRequiredFields() {
            var doc = string().export();
            assertEquals("1.0", doc.get("anyvaliVersion"));
            assertEquals("1", doc.get("schemaVersion"));
            assertNotNull(doc.get("root"));
        }

        @Test
        void documentFromMap() {
            var map = new LinkedHashMap<String, Object>();
            map.put("anyvaliVersion", "1.0");
            map.put("schemaVersion", "1");
            map.put("root", Map.of("kind", "string"));
            var doc = AnyValiDocument.fromMap(map);
            assertEquals("1.0", doc.anyvaliVersion());
            assertEquals("string", doc.root().get("kind"));
        }
    }

    // ---- String format constraints round-trip ----
    @Nested
    class StringConstraintRoundTripTests {
        @Test
        void patternRoundTrip() {
            var schema = string().pattern("^[a-z]+$");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("abc").success());
            assertFalse(imported.safeParse("ABC").success());
        }

        @Test
        void startsWithRoundTrip() {
            var schema = string().startsWith("hello");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello world").success());
            assertFalse(imported.safeParse("world").success());
        }

        @Test
        void endsWithRoundTrip() {
            var schema = string().endsWith("world");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("hello world").success());
        }

        @Test
        void includesRoundTrip() {
            var schema = string().includes("mid");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("in middle").success());
        }

        @Test
        void formatRoundTrip() {
            var schema = string().format("email");
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse("test@example.com").success());
        }
    }

    // ---- Number constraints round-trip ----
    @Nested
    class NumberConstraintRoundTripTests {
        @Test
        void exclusiveMinMaxRoundTrip() {
            var schema = number().exclusiveMin(0).exclusiveMax(100);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(50.0).success());
            assertFalse(imported.safeParse(0.0).success());
            assertFalse(imported.safeParse(100.0).success());
        }

        @Test
        void multipleOfRoundTrip() {
            var schema = number().multipleOf(5);
            var doc = Exporter.exportSchema(schema);
            Schema imported = Importer.importSchema(doc);
            assertTrue(imported.safeParse(15.0).success());
            assertFalse(imported.safeParse(7.0).success());
        }
    }

    // ---- Unsupported schema kind ----
    @Nested
    class UnsupportedKindTests {
        @Test
        void unsupportedKindThrows() {
            var map = new LinkedHashMap<String, Object>();
            map.put("anyvaliVersion", "1.0");
            map.put("schemaVersion", "1");
            map.put("root", Map.of("kind", "custom_thing"));
            assertThrows(IllegalArgumentException.class,
                    () -> Importer.importSchema(map));
        }

        @Test
        void missingKindThrows() {
            var map = new LinkedHashMap<String, Object>();
            map.put("anyvaliVersion", "1.0");
            map.put("schemaVersion", "1");
            map.put("root", Map.of());
            assertThrows(Exception.class,
                    () -> Importer.importSchema(map));
        }

        @Test
        void nullEmptyRootThrows() {
            var emptyDoc = new LinkedHashMap<String, Object>();
            emptyDoc.put("anyvaliVersion", "1.0");
            emptyDoc.put("schemaVersion", "1");
            assertThrows(Exception.class,
                    () -> Importer.importSchema(emptyDoc));

            var nullRoot = new LinkedHashMap<String, Object>();
            nullRoot.put("anyvaliVersion", "1.0");
            nullRoot.put("schemaVersion", "1");
            nullRoot.put("root", null);
            assertThrows(Exception.class,
                    () -> Importer.importSchema(nullRoot));
        }
    }
}

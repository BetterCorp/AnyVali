package com.anyvali.interchange;

import com.anyvali.AnyValiDocument;
import com.anyvali.Schema;
import com.anyvali.UnknownKeyMode;
import com.anyvali.parse.CoercionConfig;
import com.anyvali.schemas.*;

import java.util.*;

/**
 * Import schemas from AnyVali interchange format.
 */
public final class Importer {
    private Importer() {}

    /**
     * Import a schema from an AnyVali document map or JSON string.
     * Returns the root schema.
     */
    public static Schema importSchema(Object source) {
        Map<String, Object> map;
        if (source instanceof String json) {
            map = JsonHelper.parseJsonObject(json);
        } else if (source instanceof Map) {
            @SuppressWarnings("unchecked")
            Map<String, Object> m = (Map<String, Object>) source;
            map = m;
        } else {
            throw new IllegalArgumentException("Expected Map or JSON String");
        }

        AnyValiDocument doc = AnyValiDocument.fromMap(map);

        // Build definitions
        Map<String, Schema> definitions = new LinkedHashMap<>();
        @SuppressWarnings("unchecked")
        Map<String, Object> defNodes = doc.definitions();
        for (var entry : defNodes.entrySet()) {
            @SuppressWarnings("unchecked")
            Map<String, Object> node = (Map<String, Object>) entry.getValue();
            definitions.put(entry.getKey(), importNode(node, definitions));
        }

        Schema root = importNode(doc.root(), definitions);

        // Resolve all refs
        for (Schema defSchema : definitions.values()) {
            resolveRefs(defSchema, definitions);
        }
        resolveRefs(root, definitions);

        return root;
    }

    @SuppressWarnings("unchecked")
    private static Schema importNode(Map<String, Object> node, Map<String, Schema> definitions) {
        String kind = (String) node.getOrDefault("kind", "");
        Schema schema = buildSchema(kind, node, definitions);

        // Apply common fields
        if (node.containsKey("default")) {
            schema.applyImportedDefault(node.get("default"));
        }

        if (node.containsKey("coerce")) {
            schema.applyImportedCoercion(parseCoercion(node.get("coerce")));
        }

        return schema;
    }

    @SuppressWarnings("unchecked")
    private static CoercionConfig parseCoercion(Object raw) {
        if (raw == null) return null;

        if (raw instanceof Map<?, ?> map) {
            return new CoercionConfig(
                    Boolean.TRUE.equals(map.get("toInt")),
                    Boolean.TRUE.equals(map.get("toNumber")),
                    Boolean.TRUE.equals(map.get("toBool")),
                    Boolean.TRUE.equals(map.get("trim")),
                    Boolean.TRUE.equals(map.get("lower")),
                    Boolean.TRUE.equals(map.get("upper"))
            );
        }

        List<String> tokens;
        if (raw instanceof String s) {
            tokens = List.of(s);
        } else if (raw instanceof List<?> list) {
            tokens = (List<String>) (List<?>) list;
        } else {
            return null;
        }

        boolean toInt = false, toNumber = false, toBool = false;
        boolean trim = false, lower = false, upper = false;
        for (String token : tokens) {
            switch (token) {
                case "string->int" -> toInt = true;
                case "string->number" -> toNumber = true;
                case "string->bool" -> toBool = true;
                case "trim" -> trim = true;
                case "lower" -> lower = true;
                case "upper" -> upper = true;
            }
        }

        if (!toInt && !toNumber && !toBool && !trim && !lower && !upper) {
            return null;
        }
        return new CoercionConfig(toInt, toNumber, toBool, trim, lower, upper);
    }

    @SuppressWarnings("unchecked")
    private static Schema buildSchema(String kind, Map<String, Object> node,
                                       Map<String, Schema> definitions) {
        return switch (kind) {
            case "string" -> {
                var s = new StringSchema();
                if (node.containsKey("minLength"))
                    s = s.minLength(toInt(node.get("minLength")));
                if (node.containsKey("maxLength"))
                    s = s.maxLength(toInt(node.get("maxLength")));
                if (node.containsKey("pattern"))
                    s = s.pattern((String) node.get("pattern"));
                if (node.containsKey("startsWith"))
                    s = s.startsWith((String) node.get("startsWith"));
                if (node.containsKey("endsWith"))
                    s = s.endsWith((String) node.get("endsWith"));
                if (node.containsKey("includes"))
                    s = s.includes((String) node.get("includes"));
                if (node.containsKey("format"))
                    s = s.format((String) node.get("format"));
                yield s;
            }
            case "number" -> buildNumberSchema(new NumberSchema(), node);
            case "float64" -> buildNumberSchema(new Float64Schema(), node);
            case "float32" -> buildNumberSchema(new Float32Schema(), node);
            case "int" -> buildIntSchema(new IntSchema(), node);
            case "int8" -> buildIntSchema(new IntSchema.Int8Schema(), node);
            case "int16" -> buildIntSchema(new IntSchema.Int16Schema(), node);
            case "int32" -> buildIntSchema(new IntSchema.Int32Schema(), node);
            case "int64" -> buildIntSchema(new IntSchema.Int64Schema(), node);
            case "uint8" -> buildIntSchema(new IntSchema.Uint8Schema(), node);
            case "uint16" -> buildIntSchema(new IntSchema.Uint16Schema(), node);
            case "uint32" -> buildIntSchema(new IntSchema.Uint32Schema(), node);
            case "uint64" -> buildIntSchema(new IntSchema.Uint64Schema(), node);
            case "bool" -> new BoolSchema();
            case "null" -> new NullSchema();
            case "any" -> new AnySchema();
            case "unknown" -> new UnknownSchema();
            case "never" -> new NeverSchema();
            case "literal" -> new LiteralSchema(node.get("value"));
            case "enum" -> new EnumSchema((List<Object>) node.get("values"));
            case "array" -> {
                Schema items = importNode((Map<String, Object>) node.get("items"), definitions);
                Integer minItems = node.containsKey("minItems") ? toInt(node.get("minItems")) : null;
                Integer maxItems = node.containsKey("maxItems") ? toInt(node.get("maxItems")) : null;
                yield new ArraySchema(items, minItems, maxItems);
            }
            case "tuple" -> {
                List<Map<String, Object>> elements = (List<Map<String, Object>>) node.getOrDefault("elements",
                        node.getOrDefault("items", List.of()));
                List<Schema> items = new ArrayList<>();
                for (Map<String, Object> elem : elements) {
                    items.add(importNode(elem, definitions));
                }
                yield new TupleSchema(items);
            }
            case "object" -> {
                Map<String, Object> propsRaw = (Map<String, Object>) node.getOrDefault("properties", Map.of());
                Map<String, Schema> props = new LinkedHashMap<>();
                for (var entry : propsRaw.entrySet()) {
                    props.put(entry.getKey(), importNode((Map<String, Object>) entry.getValue(), definitions));
                }
                Set<String> required = null;
                if (node.containsKey("required")) {
                    required = new LinkedHashSet<>((List<String>) node.get("required"));
                }
                UnknownKeyMode ukm = UnknownKeyMode.fromString(
                        (String) node.getOrDefault("unknownKeys", "reject"));
                yield new ObjectSchema(props, required, ukm);
            }
            case "record" -> {
                Schema valueSchema = importNode((Map<String, Object>) node.get("values"), definitions);
                yield new RecordSchema(valueSchema);
            }
            case "union" -> {
                List<Map<String, Object>> variants = (List<Map<String, Object>>) node.getOrDefault("variants",
                        node.getOrDefault("schemas", List.of()));
                List<Schema> schemas = new ArrayList<>();
                for (Map<String, Object> v : variants) {
                    schemas.add(importNode(v, definitions));
                }
                yield new UnionSchema(schemas);
            }
            case "intersection" -> {
                List<Map<String, Object>> allOf = (List<Map<String, Object>>) node.getOrDefault("allOf",
                        node.getOrDefault("schemas", List.of()));
                List<Schema> schemas = new ArrayList<>();
                for (Map<String, Object> s : allOf) {
                    schemas.add(importNode(s, definitions));
                }
                yield new IntersectionSchema(schemas);
            }
            case "optional" -> {
                Schema inner = importNode((Map<String, Object>) node.get("schema"), definitions);
                yield new OptionalSchema(inner);
            }
            case "nullable" -> {
                Schema inner = importNode((Map<String, Object>) node.get("schema"), definitions);
                yield new NullableSchema(inner);
            }
            case "ref" -> new RefSchema((String) node.get("ref"));
            default -> throw new IllegalArgumentException("Unsupported schema kind: " + kind);
        };
    }

    private static NumberSchema buildNumberSchema(NumberSchema base, Map<String, Object> node) {
        NumberSchema s = base;
        if (node.containsKey("min")) s = s.min(toDouble(node.get("min")));
        if (node.containsKey("max")) s = s.max(toDouble(node.get("max")));
        if (node.containsKey("exclusiveMin")) s = s.exclusiveMin(toDouble(node.get("exclusiveMin")));
        if (node.containsKey("exclusiveMax")) s = s.exclusiveMax(toDouble(node.get("exclusiveMax")));
        if (node.containsKey("multipleOf")) s = s.multipleOf(toDouble(node.get("multipleOf")));
        return s;
    }

    private static IntSchema buildIntSchema(IntSchema base, Map<String, Object> node) {
        IntSchema s = base;
        if (node.containsKey("min")) s = s.min(toLong(node.get("min")));
        if (node.containsKey("max")) s = s.max(toLong(node.get("max")));
        if (node.containsKey("exclusiveMin")) s = s.exclusiveMin(toLong(node.get("exclusiveMin")));
        if (node.containsKey("exclusiveMax")) s = s.exclusiveMax(toLong(node.get("exclusiveMax")));
        if (node.containsKey("multipleOf")) s = s.multipleOf(toLong(node.get("multipleOf")));
        return s;
    }

    private static void resolveRefs(Schema schema, Map<String, Schema> definitions) {
        if (schema instanceof RefSchema ref) {
            ref.setDefinitions(definitions);
            String refName = ref.getRef();
            if (refName.startsWith("#/definitions/")) {
                refName = refName.substring("#/definitions/".length());
            }
            if (definitions.containsKey(refName)) {
                ref.resolve(definitions.get(refName));
            }
        } else if (schema instanceof OptionalSchema opt) {
            resolveRefs(opt.getInner(), definitions);
        } else if (schema instanceof NullableSchema nul) {
            resolveRefs(nul.getInner(), definitions);
        } else if (schema instanceof ArraySchema arr) {
            resolveRefs(arr.getItems(), definitions);
        } else if (schema instanceof TupleSchema tup) {
            for (Schema item : tup.getItems()) {
                resolveRefs(item, definitions);
            }
        } else if (schema instanceof ObjectSchema obj) {
            for (Schema propSchema : obj.getProperties().values()) {
                resolveRefs(propSchema, definitions);
            }
        } else if (schema instanceof RecordSchema rec) {
            resolveRefs(rec.getValueSchema(), definitions);
        } else if (schema instanceof UnionSchema u) {
            for (Schema s : u.getSchemas()) {
                resolveRefs(s, definitions);
            }
        } else if (schema instanceof IntersectionSchema i) {
            for (Schema s : i.getSchemas()) {
                resolveRefs(s, definitions);
            }
        }
    }

    private static int toInt(Object v) {
        if (v instanceof Number n) return n.intValue();
        return Integer.parseInt(v.toString());
    }

    private static long toLong(Object v) {
        if (v instanceof Number n) return n.longValue();
        return Long.parseLong(v.toString());
    }

    private static double toDouble(Object v) {
        if (v instanceof Number n) return n.doubleValue();
        return Double.parseDouble(v.toString());
    }
}

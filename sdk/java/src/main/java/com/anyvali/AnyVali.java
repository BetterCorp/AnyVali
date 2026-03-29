package com.anyvali;

import com.anyvali.interchange.Exporter;
import com.anyvali.interchange.Importer;
import com.anyvali.schemas.*;

import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * AnyVali Java SDK - portable schema validation.
 * <p>
 * Usage:
 * <pre>
 *     import static com.anyvali.AnyVali.*;
 *
 *     Schema schema = object_(Map.of(
 *         "name", string().minLength(1),
 *         "age", int_().min(0)
 *     ));
 *
 *     ParseResult result = schema.safeParse(Map.of("name", "Alice", "age", 30));
 *     assert result.success();
 * </pre>
 */
public final class AnyVali {
    private AnyVali() {}

    public static final String VERSION = "0.0.1";

    // ---- Builder functions ----

    public static StringSchema string() {
        return new StringSchema();
    }

    public static NumberSchema number() {
        return new NumberSchema();
    }

    public static Float32Schema float32() {
        return new Float32Schema();
    }

    public static Float64Schema float64() {
        return new Float64Schema();
    }

    public static IntSchema int_() {
        return new IntSchema();
    }

    public static IntSchema.Int8Schema int8() {
        return new IntSchema.Int8Schema();
    }

    public static IntSchema.Int16Schema int16() {
        return new IntSchema.Int16Schema();
    }

    public static IntSchema.Int32Schema int32() {
        return new IntSchema.Int32Schema();
    }

    public static IntSchema.Int64Schema int64() {
        return new IntSchema.Int64Schema();
    }

    public static IntSchema.Uint8Schema uint8() {
        return new IntSchema.Uint8Schema();
    }

    public static IntSchema.Uint16Schema uint16() {
        return new IntSchema.Uint16Schema();
    }

    public static IntSchema.Uint32Schema uint32() {
        return new IntSchema.Uint32Schema();
    }

    public static IntSchema.Uint64Schema uint64() {
        return new IntSchema.Uint64Schema();
    }

    public static BoolSchema bool_() {
        return new BoolSchema();
    }

    public static NullSchema null_() {
        return new NullSchema();
    }

    public static AnySchema any_() {
        return new AnySchema();
    }

    public static UnknownSchema unknown() {
        return new UnknownSchema();
    }

    public static NeverSchema never() {
        return new NeverSchema();
    }

    public static LiteralSchema literal(Object value) {
        return new LiteralSchema(value);
    }

    public static EnumSchema enum_(List<Object> values) {
        return new EnumSchema(values);
    }

    public static ArraySchema array(Schema items) {
        return new ArraySchema(items);
    }

    public static TupleSchema tuple_(List<Schema> items) {
        return new TupleSchema(items);
    }

    public static ObjectSchema object_(Map<String, Schema> properties) {
        return new ObjectSchema(properties);
    }

    public static ObjectSchema object_(Map<String, Schema> properties, Set<String> required) {
        return new ObjectSchema(properties, required);
    }

    public static ObjectSchema object_(Map<String, Schema> properties, Set<String> required,
                                       UnknownKeyMode unknownKeys) {
        return new ObjectSchema(properties, required, unknownKeys);
    }

    public static RecordSchema record(Schema valueSchema) {
        return new RecordSchema(valueSchema);
    }

    public static UnionSchema union(List<Schema> schemas) {
        return new UnionSchema(schemas);
    }

    public static IntersectionSchema intersection(List<Schema> schemas) {
        return new IntersectionSchema(schemas);
    }

    public static OptionalSchema optional(Schema schema) {
        return new OptionalSchema(schema);
    }

    public static NullableSchema nullable(Schema schema) {
        return new NullableSchema(schema);
    }

    public static RefSchema ref(String reference) {
        return new RefSchema(reference);
    }

    // ---- Interchange ----

    public static Map<String, Object> exportSchema(Schema schema) {
        return Exporter.exportSchema(schema);
    }

    public static Map<String, Object> exportSchema(Schema schema, ExportMode mode,
                                                    Map<String, Schema> definitions,
                                                    Map<String, Object> extensions) {
        return Exporter.exportSchema(schema, mode, definitions, extensions);
    }

    public static String exportSchemaJson(Schema schema) {
        return Exporter.exportSchemaJson(schema);
    }

    public static Schema importSchema(Object source) {
        return Importer.importSchema(source);
    }
}

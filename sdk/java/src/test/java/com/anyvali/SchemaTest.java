package com.anyvali;

import com.anyvali.parse.CoercionConfig;
import com.anyvali.schemas.*;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.*;

import static com.anyvali.AnyVali.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Comprehensive unit tests for ALL schema types.
 */
class SchemaTest {

    // ---- String Schema ----
    @Nested
    class StringSchemaTests {
        @Test
        void parsesValidString() {
            assertEquals("hello", string().parse("hello"));
        }

        @Test
        void rejectsNonString() {
            var r = string().safeParse(123);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_TYPE, r.issues().get(0).code());
        }

        @Test
        void rejectsNull() {
            var r = string().safeParse(null);
            assertFalse(r.success());
        }

        @Test
        void rejectsBoolean() {
            var r = string().safeParse(true);
            assertFalse(r.success());
            assertEquals("boolean", r.issues().get(0).received());
        }

        @Test
        void minLength() {
            var s = string().minLength(3);
            assertEquals("abc", s.parse("abc"));
            assertEquals("abcd", s.parse("abcd"));
            var r = s.safeParse("ab");
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_SMALL, r.issues().get(0).code());
        }

        @Test
        void maxLength() {
            var s = string().maxLength(3);
            assertEquals("abc", s.parse("abc"));
            var r = s.safeParse("abcd");
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_LARGE, r.issues().get(0).code());
        }

        @Test
        void pattern() {
            var s = string().pattern("^[a-z]+$");
            assertEquals("abc", s.parse("abc"));
            var r = s.safeParse("ABC");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_STRING, r.issues().get(0).code());
        }

        @Test
        void startsWith() {
            var s = string().startsWith("hello");
            assertEquals("hello world", s.parse("hello world"));
            var r = s.safeParse("world");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_STRING, r.issues().get(0).code());
        }

        @Test
        void endsWith() {
            var s = string().endsWith("world");
            assertEquals("hello world", s.parse("hello world"));
            var r = s.safeParse("hello");
            assertFalse(r.success());
        }

        @Test
        void includes() {
            var s = string().includes("mid");
            assertEquals("in middle", s.parse("in middle"));
            var r = s.safeParse("nothing");
            assertFalse(r.success());
        }

        @Test
        void format() {
            var s = string().format("email");
            assertEquals("test@example.com", s.parse("test@example.com"));
            var r = s.safeParse("notanemail");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_STRING, r.issues().get(0).code());
        }

        @Test
        void chainingIsImmutable() {
            var base = string();
            var withMin = base.minLength(5);
            // base should not be affected
            assertTrue(base.safeParse("hi").success());
            assertFalse(withMin.safeParse("hi").success());
        }

        @Test
        void multipleConstraints() {
            var s = string().minLength(2).maxLength(5).pattern("^[a-z]+$");
            assertEquals("abc", s.parse("abc"));
            var r1 = s.safeParse("a");
            assertFalse(r1.success());
            var r2 = s.safeParse("ABCDEF");
            assertFalse(r2.success());
        }

        @Test
        void parseThrowsValidationError() {
            assertThrows(ValidationError.class, () -> string().parse(42));
        }
    }

    // ---- Number Schema ----
    @Nested
    class NumberSchemaTests {
        @Test
        void parsesValidNumber() {
            assertEquals(42.0, number().parse(42.0));
            assertEquals(42, number().parse(42));
        }

        @Test
        void rejectsString() {
            var r = number().safeParse("42");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_TYPE, r.issues().get(0).code());
        }

        @Test
        void rejectsBoolean() {
            var r = number().safeParse(true);
            assertFalse(r.success());
            assertEquals("boolean", r.issues().get(0).received());
        }

        @Test
        void rejectsNaN() {
            var r = number().safeParse(Double.NaN);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_NUMBER, r.issues().get(0).code());
        }

        @Test
        void rejectsInfinity() {
            var r = number().safeParse(Double.POSITIVE_INFINITY);
            assertFalse(r.success());
        }

        @Test
        void min() {
            var s = number().min(10);
            assertEquals(10.0, s.parse(10.0));
            var r = s.safeParse(9.0);
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_SMALL, r.issues().get(0).code());
        }

        @Test
        void max() {
            var s = number().max(10);
            assertEquals(10.0, s.parse(10.0));
            var r = s.safeParse(11.0);
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_LARGE, r.issues().get(0).code());
        }

        @Test
        void exclusiveMin() {
            var s = number().exclusiveMin(10);
            assertEquals(11.0, s.parse(11.0));
            var r = s.safeParse(10.0);
            assertFalse(r.success());
        }

        @Test
        void exclusiveMax() {
            var s = number().exclusiveMax(10);
            assertEquals(9.0, s.parse(9.0));
            var r = s.safeParse(10.0);
            assertFalse(r.success());
        }

        @Test
        void multipleOf() {
            var s = number().multipleOf(3);
            assertEquals(9.0, s.parse(9.0));
            var r = s.safeParse(10.0);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_NUMBER, r.issues().get(0).code());
        }
    }

    // ---- Float64 Schema ----
    @Nested
    class Float64SchemaTests {
        @Test
        void parsesValidFloat64() {
            assertEquals(3.14, float64().parse(3.14));
        }

        @Test
        void rejectsNonNumber() {
            assertFalse(float64().safeParse("hello").success());
        }

        @Test
        void constraintsWork() {
            var s = float64().min(0).max(100);
            assertEquals(50.0, s.parse(50.0));
            assertFalse(s.safeParse(-1.0).success());
        }
    }

    // ---- Float32 Schema ----
    @Nested
    class Float32SchemaTests {
        @Test
        void parsesValidFloat32() {
            assertEquals(3.14, float32().parse(3.14));
        }

        @Test
        void rejectsOutOfRange() {
            var r = float32().safeParse(1e39);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_NUMBER, r.issues().get(0).code());
        }

        @Test
        void acceptsZero() {
            assertEquals(0.0, float32().parse(0.0));
        }
    }

    // ---- Int Schema ----
    @Nested
    class IntSchemaTests {
        @Test
        void parsesValidInt() {
            assertEquals(42L, int_().parse(42));
        }

        @Test
        void rejectsBoolean() {
            var r = int_().safeParse(true);
            assertFalse(r.success());
        }

        @Test
        void rejectsFloat() {
            var r = int_().safeParse(3.14);
            assertFalse(r.success());
        }

        @Test
        void acceptsWholeFloat() {
            assertEquals(42L, int_().parse(42.0));
        }

        @Test
        void rejectsString() {
            var r = int_().safeParse("42");
            assertFalse(r.success());
        }

        @Test
        void minMax() {
            var s = int_().min(0).max(100);
            assertEquals(50L, s.parse(50));
            assertFalse(s.safeParse(-1).success());
            assertFalse(s.safeParse(101).success());
        }

        @Test
        void exclusiveMinMax() {
            var s = int_().exclusiveMin(0).exclusiveMax(10);
            assertEquals(5L, s.parse(5));
            assertFalse(s.safeParse(0).success());
            assertFalse(s.safeParse(10).success());
        }

        @Test
        void multipleOf() {
            var s = int_().multipleOf(5);
            assertEquals(10L, s.parse(10));
            assertFalse(s.safeParse(7).success());
        }
    }

    // ---- Int8 ----
    @Nested
    class Int8Tests {
        @Test
        void validRange() {
            assertEquals(-128L, int8().parse(-128));
            assertEquals(127L, int8().parse(127));
        }

        @Test
        void outOfRange() {
            assertFalse(int8().safeParse(128).success());
            assertFalse(int8().safeParse(-129).success());
        }
    }

    // ---- Int16 ----
    @Nested
    class Int16Tests {
        @Test
        void validRange() {
            assertEquals(-32768L, int16().parse(-32768));
            assertEquals(32767L, int16().parse(32767));
        }

        @Test
        void outOfRange() {
            assertFalse(int16().safeParse(32768).success());
            assertFalse(int16().safeParse(-32769).success());
        }
    }

    // ---- Int32 ----
    @Nested
    class Int32Tests {
        @Test
        void validRange() {
            assertEquals((long) Integer.MIN_VALUE, int32().parse(Integer.MIN_VALUE));
            assertEquals((long) Integer.MAX_VALUE, int32().parse(Integer.MAX_VALUE));
        }

        @Test
        void outOfRange() {
            assertFalse(int32().safeParse((long) Integer.MAX_VALUE + 1).success());
            assertFalse(int32().safeParse((long) Integer.MIN_VALUE - 1).success());
        }
    }

    // ---- Int64 ----
    @Nested
    class Int64Tests {
        @Test
        void validRange() {
            assertEquals(Long.MIN_VALUE, int64().parse(Long.MIN_VALUE));
            assertEquals(Long.MAX_VALUE, int64().parse(Long.MAX_VALUE));
        }
    }

    // ---- Uint8 ----
    @Nested
    class Uint8Tests {
        @Test
        void validRange() {
            assertEquals(0L, uint8().parse(0));
            assertEquals(255L, uint8().parse(255));
        }

        @Test
        void outOfRange() {
            assertFalse(uint8().safeParse(-1).success());
            assertFalse(uint8().safeParse(256).success());
        }
    }

    // ---- Uint16 ----
    @Nested
    class Uint16Tests {
        @Test
        void validRange() {
            assertEquals(0L, uint16().parse(0));
            assertEquals(65535L, uint16().parse(65535));
        }

        @Test
        void outOfRange() {
            assertFalse(uint16().safeParse(-1).success());
            assertFalse(uint16().safeParse(65536).success());
        }
    }

    // ---- Uint32 ----
    @Nested
    class Uint32Tests {
        @Test
        void validRange() {
            assertEquals(0L, uint32().parse(0));
            assertEquals(4294967295L, uint32().parse(4294967295L));
        }

        @Test
        void outOfRange() {
            assertFalse(uint32().safeParse(-1).success());
            assertFalse(uint32().safeParse(4294967296L).success());
        }
    }

    // ---- Uint64 ----
    @Nested
    class Uint64Tests {
        @Test
        void validRange() {
            assertEquals(0L, uint64().parse(0));
            assertEquals(Long.MAX_VALUE, uint64().parse(Long.MAX_VALUE));
        }

        @Test
        void outOfRange() {
            assertFalse(uint64().safeParse(-1).success());
        }
    }

    // ---- Bool Schema ----
    @Nested
    class BoolSchemaTests {
        @Test
        void parsesTrue() {
            assertEquals(true, bool_().parse(true));
        }

        @Test
        void parsesFalse() {
            assertEquals(false, bool_().parse(false));
        }

        @Test
        void rejectsString() {
            assertFalse(bool_().safeParse("true").success());
        }

        @Test
        void rejectsInt() {
            assertFalse(bool_().safeParse(1).success());
        }

        @Test
        void rejectsNull() {
            assertFalse(bool_().safeParse(null).success());
        }
    }

    // ---- Null Schema ----
    @Nested
    class NullSchemaTests {
        @Test
        void parsesNull() {
            assertNull(null_().parse(null));
        }

        @Test
        void rejectsString() {
            assertFalse(null_().safeParse("null").success());
        }

        @Test
        void rejectsInt() {
            assertFalse(null_().safeParse(0).success());
        }

        @Test
        void rejectsBool() {
            assertFalse(null_().safeParse(false).success());
        }
    }

    // ---- Any Schema ----
    @Nested
    class AnySchemaTests {
        @Test
        void acceptsEverything() {
            assertEquals("hello", any_().parse("hello"));
            assertEquals(42, any_().parse(42));
            assertEquals(true, any_().parse(true));
            assertNull(any_().parse(null));
        }

        @Test
        void acceptsList() {
            var list = List.of(1, 2, 3);
            assertEquals(list, any_().parse(list));
        }

        @Test
        void acceptsMap() {
            var map = Map.of("a", 1);
            assertEquals(map, any_().parse(map));
        }
    }

    // ---- Unknown Schema ----
    @Nested
    class UnknownSchemaTests {
        @Test
        void acceptsEverything() {
            assertEquals("hello", unknown().parse("hello"));
            assertNull(unknown().parse(null));
            assertEquals(42, unknown().parse(42));
        }
    }

    // ---- Never Schema ----
    @Nested
    class NeverSchemaTests {
        @Test
        void rejectsEverything() {
            assertFalse(never().safeParse("hello").success());
            assertFalse(never().safeParse(null).success());
            assertFalse(never().safeParse(42).success());
            assertFalse(never().safeParse(true).success());
        }

        @Test
        void issueCode() {
            var r = never().safeParse("x");
            assertEquals(IssueCodes.INVALID_TYPE, r.issues().get(0).code());
            assertEquals("never", r.issues().get(0).expected());
        }
    }

    // ---- Literal Schema ----
    @Nested
    class LiteralSchemaTests {
        @Test
        void matchesStringLiteral() {
            assertEquals("hello", literal("hello").parse("hello"));
        }

        @Test
        void matchesIntLiteral() {
            assertEquals(42, literal(42).parse(42));
        }

        @Test
        void matchesBoolLiteral() {
            assertEquals(true, literal(true).parse(true));
        }

        @Test
        void matchesNullLiteral() {
            assertNull(literal(null).parse(null));
        }

        @Test
        void rejectsDifferentValue() {
            var r = literal("hello").safeParse("world");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_LITERAL, r.issues().get(0).code());
        }

        @Test
        void rejectsDifferentType() {
            // String "42" should not match int 42
            assertFalse(literal(42).safeParse("42").success());
        }

        @Test
        void nullLiteralAcceptsNull() {
            assertTrue(literal(null).acceptsNullValue());
        }

        @Test
        void nonNullLiteralRejectsNull() {
            assertFalse(literal("hello").safeParse(null).success());
        }
    }

    // ---- Enum Schema ----
    @Nested
    class EnumSchemaTests {
        @Test
        void matchesValidValue() {
            var s = enum_(List.of("a", "b", "c"));
            assertEquals("a", s.parse("a"));
            assertEquals("c", s.parse("c"));
        }

        @Test
        void rejectsInvalidValue() {
            var s = enum_(List.of("a", "b", "c"));
            var r = s.safeParse("d");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_TYPE, r.issues().get(0).code());
        }

        @Test
        void typeStrictness() {
            // String "1" should not match integer 1
            var s = enum_(List.of(1, 2, 3));
            assertFalse(s.safeParse("1").success());
        }

        @Test
        void mixedTypes() {
            var s = enum_(List.of("a", 1, true));
            assertEquals("a", s.parse("a"));
            assertEquals(1, s.parse(1));
            assertEquals(true, s.parse(true));
        }
    }

    // ---- Array Schema ----
    @Nested
    class ArraySchemaTests {
        @Test
        void parsesValidArray() {
            var s = array(string());
            assertEquals(List.of("a", "b"), s.parse(List.of("a", "b")));
        }

        @Test
        void rejectsNonArray() {
            var r = array(string()).safeParse("not array");
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_TYPE, r.issues().get(0).code());
        }

        @Test
        void validatesElements() {
            var s = array(string());
            var r = s.safeParse(List.of("a", 42, "c"));
            assertFalse(r.success());
            assertEquals(1, r.issues().get(0).path().get(0));
        }

        @Test
        void minItems() {
            var s = array(string()).minItems(2);
            var r = s.safeParse(List.of("a"));
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_SMALL, r.issues().get(0).code());
        }

        @Test
        void maxItems() {
            var s = array(string()).maxItems(2);
            var r = s.safeParse(List.of("a", "b", "c"));
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_LARGE, r.issues().get(0).code());
        }

        @Test
        void emptyArrayValid() {
            assertEquals(List.of(), array(string()).parse(List.of()));
        }

        @Test
        void nestedArrays() {
            var s = array(array(int_()));
            var input = List.of(List.of(1, 2), List.of(3, 4));
            var result = s.parse(input);
            assertNotNull(result);
        }
    }

    // ---- Tuple Schema ----
    @Nested
    class TupleSchemaTests {
        @Test
        void parsesValidTuple() {
            var s = tuple_(List.of(string(), int_()));
            var result = (List<?>) s.parse(List.of("hello", 42));
            assertEquals("hello", result.get(0));
            assertEquals(42L, result.get(1));
        }

        @Test
        void rejectsTooFew() {
            var s = tuple_(List.of(string(), int_()));
            var r = s.safeParse(List.of("hello"));
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_SMALL, r.issues().get(0).code());
        }

        @Test
        void rejectsTooMany() {
            var s = tuple_(List.of(string(), int_()));
            var r = s.safeParse(List.of("hello", 42, "extra"));
            assertFalse(r.success());
            assertEquals(IssueCodes.TOO_LARGE, r.issues().get(0).code());
        }

        @Test
        void rejectsNonArray() {
            var r = tuple_(List.of(string())).safeParse("not tuple");
            assertFalse(r.success());
        }

        @Test
        void validatesElementTypes() {
            var s = tuple_(List.of(string(), int_()));
            var r = s.safeParse(List.of(42, "hello"));
            assertFalse(r.success());
        }
    }

    // ---- Object Schema ----
    @Nested
    class ObjectSchemaTests {
        @Test
        void parsesValidObject() {
            var s = object_(new LinkedHashMap<>(Map.of("name", string(), "age", int_())));
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of("name", "Alice", "age", 30));
            assertEquals("Alice", result.get("name"));
            assertEquals(30L, result.get("age"));
        }

        @Test
        void rejectsNonObject() {
            var r = object_(Map.of("x", string())).safeParse("string");
            assertFalse(r.success());
        }

        @Test
        void rejectsUnknownKeysByDefault() {
            var s = object_(Map.of("name", string()));
            var r = s.safeParse(Map.of("name", "Alice", "extra", "value"));
            assertFalse(r.success());
            assertEquals(IssueCodes.UNKNOWN_KEY, r.issues().get(0).code());
        }

        @Test
        void stripsUnknownKeys() {
            var s = object_(Map.of("name", string()), null, UnknownKeyMode.STRIP);
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of("name", "Alice", "extra", "value"));
            assertEquals(1, result.size());
            assertEquals("Alice", result.get("name"));
        }

        @Test
        void allowsUnknownKeys() {
            var s = object_(Map.of("name", string()), null, UnknownKeyMode.ALLOW);
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of("name", "Alice", "extra", "value"));
            assertEquals(2, result.size());
            assertEquals("value", result.get("extra"));
        }

        @Test
        void requiresMissingField() {
            var s = object_(Map.of("name", string()));
            var r = s.safeParse(Map.of());
            assertFalse(r.success());
            assertEquals(IssueCodes.REQUIRED, r.issues().get(0).code());
            assertEquals("name", r.issues().get(0).path().get(0));
        }

        @Test
        void optionalField() {
            var s = object_(Map.of("name", string()), Set.of());
            var result = s.parse(Map.of());
            assertNotNull(result);
        }

        @Test
        void validatesPropertyTypes() {
            var s = object_(Map.of("age", int_()));
            var r = s.safeParse(Map.of("age", "not a number"));
            assertFalse(r.success());
            assertEquals("age", r.issues().get(0).path().get(0));
        }
    }

    // ---- Record Schema ----
    @Nested
    class RecordSchemaTests {
        @Test
        void parsesValidRecord() {
            var s = record(int_());
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of("a", 1, "b", 2));
            assertEquals(1L, result.get("a"));
            assertEquals(2L, result.get("b"));
        }

        @Test
        void rejectsNonObject() {
            assertFalse(record(string()).safeParse("not a map").success());
        }

        @Test
        void validatesValues() {
            var s = record(int_());
            var r = s.safeParse(Map.of("a", "not int"));
            assertFalse(r.success());
        }

        @Test
        void emptyRecord() {
            assertEquals(Map.of(), record(string()).parse(Map.of()));
        }
    }

    // ---- Union Schema ----
    @Nested
    class UnionSchemaTests {
        @Test
        void matchesFirstVariant() {
            var s = union(List.of(string(), int_()));
            assertEquals("hello", s.parse("hello"));
        }

        @Test
        void matchesSecondVariant() {
            var s = union(List.of(string(), int_()));
            assertEquals(42L, s.parse(42));
        }

        @Test
        void rejectsNoMatch() {
            var s = union(List.of(string(), int_()));
            var r = s.safeParse(true);
            assertFalse(r.success());
            assertEquals(IssueCodes.INVALID_UNION, r.issues().get(0).code());
        }

        @Test
        void acceptsNullIfAnyVariantDoes() {
            var s = union(List.of(string(), null_()));
            assertNull(s.parse(null));
        }
    }

    // ---- Intersection Schema ----
    @Nested
    class IntersectionSchemaTests {
        @Test
        void mergesObjectResults() {
            var s1 = object_(Map.of("a", string()), null, UnknownKeyMode.STRIP);
            var s2 = object_(Map.of("b", int_()), null, UnknownKeyMode.STRIP);
            var s = intersection(List.of(s1, s2));
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of("a", "hello", "b", 42));
            assertEquals("hello", result.get("a"));
            assertEquals(42L, result.get("b"));
        }

        @Test
        void failsIfAnyFails() {
            var s1 = object_(Map.of("a", string()), null, UnknownKeyMode.STRIP);
            var s2 = object_(Map.of("b", int_()), null, UnknownKeyMode.STRIP);
            var s = intersection(List.of(s1, s2));
            var r = s.safeParse(Map.of("a", "hello", "b", "not int"));
            assertFalse(r.success());
        }

        @Test
        void nonObjectReturnsLast() {
            var s = intersection(List.of(string().minLength(1), string().maxLength(10)));
            assertEquals("hello", s.parse("hello"));
        }
    }

    // ---- Optional Schema ----
    @Nested
    class OptionalSchemaTests {
        @Test
        void acceptsAbsent() {
            var s = optional(string());
            assertNull(s.runPipeline(Schema.ABSENT, new ValidationContext()));
        }

        @Test
        void validatesPresent() {
            var s = optional(string());
            assertEquals("hello", s.parse("hello"));
        }

        @Test
        void rejectsInvalidPresent() {
            var s = optional(string());
            var r = s.safeParse(42);
            assertFalse(r.success());
        }

        @Test
        void withDefault() {
            var s = optional(string().withDefault("fallback"));
            var ctx = new ValidationContext();
            assertEquals("fallback", s.runPipeline(Schema.ABSENT, ctx));
        }

        @Test
        void modifierForm() {
            var s = string().optional();
            assertNull(s.runPipeline(Schema.ABSENT, new ValidationContext()));
        }
    }

    // ---- Nullable Schema ----
    @Nested
    class NullableSchemaTests {
        @Test
        void acceptsNull() {
            assertNull(nullable(string()).parse(null));
        }

        @Test
        void validatesNonNull() {
            assertEquals("hello", nullable(string()).parse("hello"));
        }

        @Test
        void rejectsInvalidNonNull() {
            var r = nullable(string()).safeParse(42);
            assertFalse(r.success());
        }

        @Test
        void modifierForm() {
            assertNull(string().nullable().parse(null));
        }
    }

    // ---- Ref Schema ----
    @Nested
    class RefSchemaTests {
        @Test
        void resolvesReference() {
            var ref = ref("#/definitions/Name");
            ref.resolve(string().minLength(1));
            assertEquals("hello", ref.parse("hello"));
        }

        @Test
        void unresolvedRefFails() {
            var ref = ref("#/definitions/Missing");
            var r = ref.safeParse("hello");
            assertFalse(r.success());
        }

        @Test
        void lazyResolution() {
            var ref = ref("#/definitions/Name");
            Map<String, Schema<?>> defs = new HashMap<>();
            defs.put("Name", string());
            ref.setDefinitions(defs);
            assertEquals("hello", ref.parse("hello"));
        }
    }

    // ---- Defaults ----
    @Nested
    class DefaultTests {
        @Test
        void appliesDefaultWhenAbsent() {
            var s = string().withDefault("fallback");
            var ctx = new ValidationContext();
            assertEquals("fallback", s.runPipeline(Schema.ABSENT, ctx));
        }

        @Test
        void doesNotOverridePresent() {
            var s = string().withDefault("fallback");
            assertEquals("hello", s.parse("hello"));
        }

        @Test
        void invalidDefaultProducesDefaultInvalid() {
            // Default of 42 on a string schema is invalid
            var s = string().withDefault(42);
            var ctx = new ValidationContext();
            s.runPipeline(Schema.ABSENT, ctx);
            assertTrue(ctx.hasIssues());
            assertEquals(IssueCodes.DEFAULT_INVALID, ctx.getIssues().get(0).code());
        }

        @Test
        void defaultInObjectSchema() {
            var s = object_(Map.of("name", string().withDefault("unknown")));
            @SuppressWarnings("unchecked")
            var result = (Map<String, Object>) s.parse(Map.of());
            assertEquals("unknown", result.get("name"));
        }
    }

    // ---- Export ----
    @Nested
    class ExportTests {
        @Test
        void exportStringSchema() {
            var node = string().minLength(1).maxLength(100).toPortableNode();
            assertEquals("string", node.get("kind"));
            assertEquals(1, node.get("minLength"));
            assertEquals(100, node.get("maxLength"));
        }

        @Test
        void exportObjectSchema() {
            var s = object_(Map.of("name", string()));
            var doc = s.export();
            assertNotNull(doc.get("root"));
            assertEquals("1.0", doc.get("anyvaliVersion"));
        }

        @Test
        void exportIntSchema() {
            var node = int_().min(0).max(100).toPortableNode();
            assertEquals("int", node.get("kind"));
            assertEquals(0L, node.get("min"));
            assertEquals(100L, node.get("max"));
        }

        @Test
        void exportWithDefault() {
            var node = string().withDefault("hello").toPortableNode();
            assertEquals("hello", node.get("default"));
        }

        @Test
        void exportWithCoercion() {
            var s = string().coerce(CoercionConfig.builder().trim(true).lower(true).build());
            var node = s.toPortableNode();
            @SuppressWarnings("unchecked")
            var coerce = (Map<String, Object>) node.get("coerce");
            assertEquals(true, coerce.get("trim"));
            assertEquals(true, coerce.get("lower"));
        }
    }

    // ---- Utility ----
    @Nested
    class UtilityTests {
        @Test
        void anyvaliTypeName() {
            assertEquals("null", Schema.anyvaliTypeName(null));
            assertEquals("boolean", Schema.anyvaliTypeName(true));
            assertEquals("integer", Schema.anyvaliTypeName(42));
            assertEquals("integer", Schema.anyvaliTypeName(42L));
            assertEquals("number", Schema.anyvaliTypeName(3.14));
            assertEquals("string", Schema.anyvaliTypeName("hello"));
            assertEquals("array", Schema.anyvaliTypeName(List.of()));
            assertEquals("object", Schema.anyvaliTypeName(Map.of()));
        }

        @Test
        void deepCopyValue() {
            var list = new ArrayList<>(List.of(1, 2, 3));
            var copy = (List<?>) Schema.deepCopyValue(list);
            assertEquals(list, copy);
            list.add(4);
            assertNotEquals(list, copy);
        }

        @Test
        void deepCopyMap() {
            var map = new LinkedHashMap<String, Object>();
            map.put("a", new ArrayList<>(List.of(1, 2)));
            @SuppressWarnings("unchecked")
            var copy = (Map<String, Object>) Schema.deepCopyValue(map);
            assertEquals(map, copy);
            ((List<Object>) map.get("a")).add(3);
            assertNotEquals(map, copy);
        }
    }
}

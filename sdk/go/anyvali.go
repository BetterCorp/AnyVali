// Package anyvali provides a native Go validation library with portable schema interchange.
package anyvali

import "math"

// String creates a new StringSchema.
func String() *StringSchema { return newStringSchema() }

// Number creates a new Float64Schema (number = float64).
func Number() *Float64Schema { return newFloat64Schema("number") }

// Float64 creates a new Float64Schema.
func Float64() *Float64Schema { return newFloat64Schema("float64") }

// Float32 creates a new Float32Schema with float32 range constraints.
func Float32() *Float32Schema { return newFloat32Schema() }

// Int creates a new Int64Schema (int = int64).
func Int() *IntSchema { return newIntSchema("int", math.MinInt64, math.MaxInt64) }

// Int8 creates a new IntSchema for int8 range.
func Int8() *IntSchema { return newIntSchema("int8", math.MinInt8, math.MaxInt8) }

// Int16 creates a new IntSchema for int16 range.
func Int16() *IntSchema { return newIntSchema("int16", math.MinInt16, math.MaxInt16) }

// Int32 creates a new IntSchema for int32 range.
func Int32() *IntSchema { return newIntSchema("int32", math.MinInt32, math.MaxInt32) }

// Int64 creates a new IntSchema for int64 range.
func Int64() *IntSchema { return newIntSchema("int64", math.MinInt64, math.MaxInt64) }

// Uint8 creates a new IntSchema for uint8 range.
func Uint8() *IntSchema { return newUintSchema("uint8", 0, math.MaxUint8) }

// Uint16 creates a new IntSchema for uint16 range.
func Uint16() *IntSchema { return newUintSchema("uint16", 0, math.MaxUint16) }

// Uint32 creates a new IntSchema for uint32 range.
func Uint32() *IntSchema { return newUintSchema("uint32", 0, math.MaxUint32) }

// Uint64 creates a new IntSchema for uint64 range.
func Uint64() *IntSchema { return newUintSchema("uint64", 0, math.MaxUint64) }

// Bool creates a new BoolSchema.
func Bool() *BoolSchema { return newBoolSchema() }

// Null creates a new NullSchema.
func Null() *NullSchema { return newNullSchema() }

// Any creates a new AnySchema.
func Any() *AnySchema { return newAnySchema() }

// Unknown creates a new UnknownSchema.
func Unknown() *UnknownSchema { return newUnknownSchema() }

// Never creates a new NeverSchema.
func Never() *NeverSchema { return newNeverSchema() }

// Literal creates a new LiteralSchema for the given value.
func Literal(v any) *LiteralSchema { return newLiteralSchema(v) }

// Enum creates a new EnumSchema for the given values.
func Enum(values ...any) *EnumSchema { return newEnumSchema(values) }

// Array creates a new ArraySchema with the given item schema.
func Array(item Schema) *ArraySchema { return newArraySchema(item) }

// Tuple creates a new TupleSchema with the given item schemas.
func Tuple(items ...Schema) *TupleSchema { return newTupleSchema(items) }

// Object creates a new ObjectSchema with the given properties.
func Object(props map[string]Schema) *ObjectSchema { return newObjectSchema(props) }

// Record creates a new RecordSchema with the given value schema.
func Record(value Schema) *RecordSchema { return newRecordSchema(value) }

// Union creates a new UnionSchema.
func Union(schemas ...Schema) *UnionSchema { return newUnionSchema(schemas) }

// Intersection creates a new IntersectionSchema.
func Intersection(schemas ...Schema) *IntersectionSchema { return newIntersectionSchema(schemas) }

// Optional wraps a schema to make it optional.
func Optional(s Schema) *OptionalSchema { return newOptionalSchema(s) }

// Nullable wraps a schema to allow null values.
func Nullable(s Schema) *NullableSchema { return newNullableSchema(s) }

@file:JvmName("AnyVali")
package com.anyvali

import com.anyvali.schemas.*
import com.anyvali.interchange.Importer
import kotlinx.serialization.json.*

// Top-level builder functions

fun string() = StringSchema()
fun number() = NumberSchema(schemaKind = "number")
fun float32() = NumberSchema(schemaKind = "float32")
fun float64() = NumberSchema(schemaKind = "float64")
fun int_() = IntSchema(schemaKind = "int")
fun int8() = IntSchema(schemaKind = "int8")
fun int16() = IntSchema(schemaKind = "int16")
fun int32() = IntSchema(schemaKind = "int32")
fun int64() = IntSchema(schemaKind = "int64")
fun uint8() = IntSchema(schemaKind = "uint8")
fun uint16() = IntSchema(schemaKind = "uint16")
fun uint32() = IntSchema(schemaKind = "uint32")
fun uint64() = IntSchema(schemaKind = "uint64")
fun bool() = BoolSchema()
fun null_() = NullSchema()
fun any_() = AnySchema()
fun unknown() = UnknownSchema()
fun never() = NeverSchema()
fun literal(value: Any?) = LiteralSchema(value)
fun enum_(vararg values: Any?) = EnumSchema(values.toList())
fun array(items: Schema) = ArraySchema(items)
fun tuple(vararg elements: Schema) = TupleSchema(elements.toList())
fun record(values: Schema) = RecordSchema(values)
fun union(vararg variants: Schema) = UnionSchema(variants.toList())
fun intersection(vararg schemas: Schema) = IntersectionSchema(schemas.toList())
fun optional(schema: Schema) = OptionalSchema(schema)
fun nullable(schema: Schema) = NullableSchema(schema)
fun ref(ref: String) = RefSchema(ref)

fun obj(
    properties: Map<String, Schema>,
    required: Set<String> = emptySet(),
    unknownKeys: UnknownKeyMode = UnknownKeyMode.REJECT
) = ObjectSchema(properties, required, unknownKeys)

fun import_(jsonStr: String): Pair<Schema, Map<String, Schema>> = Importer.importFromJson(jsonStr)

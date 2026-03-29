// ---- Re-export types ----
export type {
  SchemaNode,
  SchemaKind,
  ParseResult,
  ValidationIssue,
  AnyValiDocument,
  ExportMode,
  UnknownKeyMode,
  CoercionConfig,
  StringFormat,
  ParseContext,
} from "./types.js";

export { ISSUE_CODES } from "./issue-codes.js";
export type { IssueCode } from "./issue-codes.js";

export { ValidationError } from "./errors.js";

// ---- Re-export schema classes ----
export {
  BaseSchema,
  ABSENT,
  StringSchema,
  NumberSchema,
  Float32Schema,
  Float64Schema,
  IntSchema,
  Int8Schema,
  Int16Schema,
  Int32Schema,
  Int64Schema,
  Uint8Schema,
  Uint16Schema,
  Uint32Schema,
  Uint64Schema,
  BoolSchema,
  NullSchema,
  AnySchema,
  UnknownSchema,
  NeverSchema,
  LiteralSchema,
  EnumSchema,
  ArraySchema,
  TupleSchema,
  ObjectSchema,
  RecordSchema,
  UnionSchema,
  IntersectionSchema,
  OptionalSchema,
  NullableSchema,
  RefSchema,
} from "./schemas/index.js";

// ---- Builder functions ----

import { StringSchema } from "./schemas/string.js";
import { NumberSchema, Float32Schema, Float64Schema } from "./schemas/number.js";
import {
  IntSchema,
  Int8Schema,
  Int16Schema,
  Int32Schema,
  Int64Schema,
  Uint8Schema,
  Uint16Schema,
  Uint32Schema,
  Uint64Schema,
} from "./schemas/int.js";
import { BoolSchema } from "./schemas/bool.js";
import { NullSchema } from "./schemas/null.js";
import { AnySchema } from "./schemas/any.js";
import { UnknownSchema } from "./schemas/unknown.js";
import { NeverSchema } from "./schemas/never.js";
import { LiteralSchema } from "./schemas/literal.js";
import { EnumSchema } from "./schemas/enum.js";
import { ArraySchema } from "./schemas/array.js";
import { TupleSchema } from "./schemas/tuple.js";
import { ObjectSchema } from "./schemas/object.js";
import { RecordSchema } from "./schemas/record.js";
import { UnionSchema } from "./schemas/union.js";
import { IntersectionSchema } from "./schemas/intersection.js";
import { OptionalSchema } from "./schemas/optional.js";
import { NullableSchema } from "./schemas/nullable.js";
import { BaseSchema } from "./schemas/base.js";

import type { UnknownKeyMode, ParseResult, AnyValiDocument, ExportMode } from "./types.js";

/** Create a string schema */
export function string(): StringSchema {
  return new StringSchema();
}

/** Create a number (float64) schema */
export function number(): NumberSchema {
  return new NumberSchema();
}

/** Create a float32 schema */
export function float32(): Float32Schema {
  return new Float32Schema();
}

/** Create a float64 schema */
export function float64(): Float64Schema {
  return new Float64Schema();
}

/** Create an int (int64) schema */
export function int(): IntSchema {
  return new IntSchema();
}

/** Create an int8 schema */
export function int8(): Int8Schema {
  return new Int8Schema();
}

/** Create an int16 schema */
export function int16(): Int16Schema {
  return new Int16Schema();
}

/** Create an int32 schema */
export function int32(): Int32Schema {
  return new Int32Schema();
}

/** Create an int64 schema */
export function int64(): Int64Schema {
  return new Int64Schema();
}

/** Create a uint8 schema */
export function uint8(): Uint8Schema {
  return new Uint8Schema();
}

/** Create a uint16 schema */
export function uint16(): Uint16Schema {
  return new Uint16Schema();
}

/** Create a uint32 schema */
export function uint32(): Uint32Schema {
  return new Uint32Schema();
}

/** Create a uint64 schema */
export function uint64(): Uint64Schema {
  return new Uint64Schema();
}

/** Create a boolean schema */
export function bool(): BoolSchema {
  return new BoolSchema();
}

/** Create a null schema. Named null_ to avoid conflict with the null keyword. */
export function null_(): NullSchema {
  return new NullSchema();
}

/** Create an any schema */
export function any(): AnySchema {
  return new AnySchema();
}

/** Create an unknown schema */
export function unknown(): UnknownSchema {
  return new UnknownSchema();
}

/** Create a never schema */
export function never(): NeverSchema {
  return new NeverSchema();
}

/** Create a literal schema */
export function literal<T extends string | number | boolean | null>(
  value: T
): LiteralSchema<T> {
  return new LiteralSchema(value);
}

/** Create an enum schema. Named enum_ to avoid conflict with the enum keyword. */
export function enum_(values: (string | number)[]): EnumSchema {
  return new EnumSchema(values);
}

/** Create an array schema */
export function array<T extends BaseSchema>(items: T): ArraySchema<T> {
  return new ArraySchema(items);
}

/** Create a tuple schema */
export function tuple(items: BaseSchema[]): TupleSchema {
  return new TupleSchema(items);
}

/** Create an object schema */
export function object(
  shape: Record<string, BaseSchema>,
  options?: { unknownKeys?: UnknownKeyMode }
): ObjectSchema {
  return new ObjectSchema(shape, options);
}

/** Create a record schema */
export function record(valueSchema: BaseSchema): RecordSchema {
  return new RecordSchema(valueSchema);
}

/** Create a union schema */
export function union(variants: BaseSchema[]): UnionSchema {
  return new UnionSchema(variants);
}

/** Create an intersection schema */
export function intersection(schemas: BaseSchema[]): IntersectionSchema {
  return new IntersectionSchema(schemas);
}

/** Wrap a schema as optional */
export function optional<T extends BaseSchema>(schema: T): OptionalSchema<T> {
  return new OptionalSchema(schema);
}

/** Wrap a schema as nullable */
export function nullable<T extends BaseSchema>(schema: T): NullableSchema<T> {
  return new NullableSchema(schema);
}

// ---- Top-level parse functions ----

/** Parse input using the given schema. Throws ValidationError on failure. */
export function parse<T>(schema: BaseSchema<unknown, T>, input: unknown): T {
  return schema.parse(input);
}

/** Parse input using the given schema. Returns a result object. */
export function safeParse<T>(
  schema: BaseSchema<unknown, T>,
  input: unknown
): ParseResult<T> {
  return schema.safeParse(input);
}

// ---- Interchange functions ----

import { exportSchema as _exportSchema } from "./interchange/exporter.js";
import { importSchema as _importSchema } from "./interchange/importer.js";

/** Export a schema to an AnyValiDocument */
export function exportSchema(
  schema: BaseSchema,
  mode: ExportMode = "portable"
): AnyValiDocument {
  return _exportSchema(schema, mode);
}

/** Import an AnyValiDocument to a live schema */
export function importSchema(doc: AnyValiDocument): BaseSchema {
  return _importSchema(doc);
}

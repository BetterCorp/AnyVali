import type {
  AnyValiDocument,
  SchemaNode,
} from "../types.js";
import { BaseSchema } from "../schemas/base.js";
import { StringSchema } from "../schemas/string.js";
import {
  NumberSchema,
  Float32Schema,
  Float64Schema,
} from "../schemas/number.js";
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
} from "../schemas/int.js";
import { BoolSchema } from "../schemas/bool.js";
import { NullSchema } from "../schemas/null.js";
import { AnySchema } from "../schemas/any.js";
import { UnknownSchema } from "../schemas/unknown.js";
import { NeverSchema } from "../schemas/never.js";
import { LiteralSchema } from "../schemas/literal.js";
import { EnumSchema } from "../schemas/enum.js";
import { ArraySchema } from "../schemas/array.js";
import { TupleSchema } from "../schemas/tuple.js";
import { ObjectSchema } from "../schemas/object.js";
import { RecordSchema } from "../schemas/record.js";
import { UnionSchema } from "../schemas/union.js";
import { IntersectionSchema } from "../schemas/intersection.js";
import { OptionalSchema } from "../schemas/optional.js";
import { NullableSchema } from "../schemas/nullable.js";
import { RefSchema } from "../schemas/ref.js";
import { normalizeCoercionConfig } from "../parse/coerce.js";
import type { StringFormat, UnknownKeyMode } from "../types.js";

/**
 * Import an AnyValiDocument back into a live Schema.
 */
export function importSchema(doc: AnyValiDocument): BaseSchema {
  const definitions = doc.definitions ?? {};
  const resolvedDefs = new Map<string, BaseSchema>();

  function importNode(node: any): BaseSchema {
    let schema: BaseSchema;

    switch (node.kind) {
      case "string": {
        let s = new StringSchema();
        if (node.minLength !== undefined) s = s.minLength(node.minLength);
        if (node.maxLength !== undefined) s = s.maxLength(node.maxLength);
        if (node.pattern !== undefined) s = s.pattern(node.pattern);
        if (node.startsWith !== undefined) s = s.startsWith(node.startsWith);
        if (node.endsWith !== undefined) s = s.endsWith(node.endsWith);
        if (node.includes !== undefined) s = s.includes(node.includes);
        if (node.format !== undefined)
          s = s.format(node.format as StringFormat);
        schema = s;
        break;
      }

      case "number":
      case "float64": {
        let s =
          node.kind === "float64" ? new Float64Schema() : new NumberSchema();
        schema = applyNumericConstraints(s, node);
        break;
      }

      case "float32": {
        schema = applyNumericConstraints(new Float32Schema(), node);
        break;
      }

      case "int":
      case "int64": {
        schema = applyNumericConstraints(
          node.kind === "int64" ? new Int64Schema() : new IntSchema(),
          node
        );
        break;
      }

      case "int8":
        schema = applyNumericConstraints(new Int8Schema(), node);
        break;
      case "int16":
        schema = applyNumericConstraints(new Int16Schema(), node);
        break;
      case "int32":
        schema = applyNumericConstraints(new Int32Schema(), node);
        break;
      case "uint8":
        schema = applyNumericConstraints(new Uint8Schema(), node);
        break;
      case "uint16":
        schema = applyNumericConstraints(new Uint16Schema(), node);
        break;
      case "uint32":
        schema = applyNumericConstraints(new Uint32Schema(), node);
        break;
      case "uint64":
        schema = applyNumericConstraints(new Uint64Schema(), node);
        break;

      case "bool":
        schema = new BoolSchema();
        break;

      case "null":
        schema = new NullSchema();
        break;

      case "any":
        schema = new AnySchema();
        break;

      case "unknown":
        schema = new UnknownSchema();
        break;

      case "never":
        schema = new NeverSchema();
        break;

      case "literal": {
        schema = new LiteralSchema(node.value);
        break;
      }

      case "enum": {
        schema = new EnumSchema(node.values);
        break;
      }

      case "array": {
        let s = new ArraySchema(importNode(node.items));
        if (node.minItems !== undefined) s = s.minItems(node.minItems);
        if (node.maxItems !== undefined) s = s.maxItems(node.maxItems);
        schema = s;
        break;
      }

      case "tuple": {
        // Corpus uses "elements", our export uses "items"
        const elements = node.elements ?? node.items;
        schema = new TupleSchema(
          elements.map((i: any) => importNode(i))
        );
        break;
      }

      case "object": {
        const shape: Record<string, BaseSchema> = {};
        const requiredSet = new Set<string>(node.required ?? []);
        for (const [key, propNode] of Object.entries(
          node.properties ?? {}
        )) {
          let propSchema = importNode(propNode);
          if (!requiredSet.has(key)) {
            propSchema = new OptionalSchema(propSchema);
          }
          shape[key] = propSchema;
        }
        schema = new ObjectSchema(shape, {
          unknownKeys:
            (node.unknownKeys as UnknownKeyMode) ?? "reject",
        });
        break;
      }

      case "record": {
        // Corpus uses "values", our export uses "valueSchema"
        const valueNode = node.values ?? node.valueSchema;
        schema = new RecordSchema(importNode(valueNode));
        break;
      }

      case "union": {
        schema = new UnionSchema(
          node.variants.map((v: any) => importNode(v))
        );
        break;
      }

      case "intersection": {
        schema = new IntersectionSchema(
          node.allOf.map((s: any) => importNode(s))
        );
        break;
      }

      case "optional": {
        // Corpus uses "schema", our export uses "inner"
        const innerNode = node.schema ?? node.inner;
        schema = new OptionalSchema(importNode(innerNode));
        break;
      }

      case "nullable": {
        // Corpus uses "schema", our export uses "inner"
        const innerNode = node.schema ?? node.inner;
        schema = new NullableSchema(importNode(innerNode));
        break;
      }

      case "ref": {
        const refPath = node.ref as string;
        const defName = refPath.replace("#/definitions/", "");
        schema = new RefSchema(refPath, () => {
          if (resolvedDefs.has(defName)) {
            return resolvedDefs.get(defName)!;
          }
          const defNode = definitions[defName];
          if (!defNode) {
            throw new Error(`Unresolved definition: ${defName}`);
          }
          const resolved = importNode(defNode);
          resolvedDefs.set(defName, resolved);
          return resolved;
        });
        break;
      }

      default:
        throw new Error(`Unsupported schema kind: ${node.kind}`);
    }

    // Apply default
    if (node.default !== undefined) {
      schema = schema.default(node.default as any);
    }

    // Apply coercion config - handle both string and object formats
    if (node.coerce !== undefined) {
      const config = normalizeCoercionConfig(node.coerce);
      schema = schema.coerce(config);
    }

    return schema;
  }

  return importNode(doc.root);
}

function applyNumericConstraints<T extends NumberSchema>(
  schema: T,
  node: any
): T {
  let s = schema;
  if (node.min !== undefined) s = s.min(node.min) as T;
  if (node.max !== undefined) s = s.max(node.max) as T;
  if (node.exclusiveMin !== undefined)
    s = s.exclusiveMin(node.exclusiveMin) as T;
  if (node.exclusiveMax !== undefined)
    s = s.exclusiveMax(node.exclusiveMax) as T;
  if (node.multipleOf !== undefined) s = s.multipleOf(node.multipleOf) as T;
  return s;
}

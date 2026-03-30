import type { ParseContext, SchemaNode, UnknownKeyMode } from "../types.js";
import { BaseSchema, ABSENT } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";
import type { OptionalSchema } from "./optional.js";

/** Flatten an intersection into a clean single-level type. */
type Prettify<T> = { [K in keyof T]: T[K] } & {};

/** Map a shape record to its inferred output type, separating required from optional fields. */
export type InferShape<
  T extends Record<string, BaseSchema<any, any>>,
> = Prettify<
  {
    [K in keyof T as T[K] extends OptionalSchema<any> ? never : K]: T[K]["_output"];
  } & {
    [K in keyof T as T[K] extends OptionalSchema<any>
      ? K
      : never]?: T[K]["_output"];
  }
>;

interface PropertyDef {
  schema: BaseSchema;
  required: boolean;
}

export class ObjectSchema<
  TShape extends Record<string, BaseSchema<any, any>> = Record<
    string,
    BaseSchema
  >,
> extends BaseSchema<Record<string, unknown>, InferShape<TShape>> {
  private _properties: Map<string, PropertyDef>;
  private _unknownKeys: UnknownKeyMode;

  constructor(
    shape: TShape,
    options?: { unknownKeys?: UnknownKeyMode },
  ) {
    super();
    this._properties = new Map();
    this._unknownKeys = options?.unknownKeys ?? "reject";

    for (const [key, schema] of Object.entries(shape)) {
      // Check if the schema is an OptionalSchema wrapper
      const isOptional = (schema as any)._isOptionalWrapper === true;
      this._properties.set(key, {
        schema,
        required: !isOptional,
      });
    }
  }

  unknownKeys(mode: UnknownKeyMode): ObjectSchema<TShape> {
    const clone = this._clone();
    clone._unknownKeys = mode;
    return clone;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (typeof input !== "object" || input === null || Array.isArray(input)) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected object, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: "object",
        received: describeType(input),
      });
      return undefined;
    }

    const obj = input as Record<string, unknown>;
    const result: Record<string, unknown> = {};
    const inputKeys = new Set(Object.keys(obj));

    // Validate declared properties
    for (const [key, prop] of this._properties) {
      ctx.path.push(key);
      const hasKey = Object.prototype.hasOwnProperty.call(obj, key);
      inputKeys.delete(key);

      if (!hasKey) {
        // Check if required
        if (prop.required && prop.schema._defaultValue === ABSENT) {
          const expectedKind = prop.schema._toNode().kind;
          ctx.issues.push({
            code: ISSUE_CODES.REQUIRED,
            message: `Required property "${key}" is missing`,
            path: [...ctx.path],
            expected: expectedKind,
            received: "undefined",
          });
          ctx.path.pop();
          continue;
        }
      }

      const rawValue = hasKey ? obj[key] : undefined;
      const val = prop.schema._runPipeline(rawValue, ctx);

      // Only include in result if value is not undefined or it was explicitly present
      if (val !== undefined || hasKey || prop.schema._defaultValue !== ABSENT) {
        result[key] = val;
      }

      ctx.path.pop();
    }

    // Handle unknown keys
    for (const key of inputKeys) {
      switch (this._unknownKeys) {
        case "reject":
          ctx.issues.push({
            code: ISSUE_CODES.UNKNOWN_KEY,
            message: `Unknown key "${key}"`,
            path: [...ctx.path, key],
            expected: "undefined",
            received: key,
          });
          break;
        case "allow":
          result[key] = obj[key];
          break;
        case "strip":
          // Just ignore it
          break;
      }
    }

    return result;
  }

  _toNode(): SchemaNode {
    const properties: Record<string, SchemaNode> = {};
    const required: string[] = [];

    for (const [key, prop] of this._properties) {
      properties[key] = prop.schema._toNode();
      if (prop.required) {
        required.push(key);
      }
    }

    const node = {
      kind: "object" as const,
      properties,
      required,
      unknownKeys: this._unknownKeys,
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

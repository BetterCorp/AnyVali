import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";

export class IntersectionSchema extends BaseSchema<unknown, unknown> {
  private _schemas: BaseSchema[];

  constructor(schemas: BaseSchema[]) {
    super();
    this._schemas = schemas;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    let result: unknown = input;
    let anyFailed = false;

    for (const schema of this._schemas) {
      const innerCtx: ParseContext = {
        path: [...ctx.path],
        issues: [],
      };
      const validated = schema._runPipeline(input, innerCtx);

      if (innerCtx.issues.length > 0) {
        ctx.issues.push(...innerCtx.issues);
        anyFailed = true;
      } else {
        // Merge object results
        if (
          typeof result === "object" &&
          result !== null &&
          typeof validated === "object" &&
          validated !== null &&
          !Array.isArray(result) &&
          !Array.isArray(validated)
        ) {
          result = {
            ...(result as Record<string, unknown>),
            ...(validated as Record<string, unknown>),
          };
        } else {
          result = validated;
        }
      }
    }

    if (anyFailed) {
      return undefined;
    }

    return result;
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "intersection" as const,
      allOf: this._schemas.map((s) => s._toNode()),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

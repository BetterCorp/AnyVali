import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class RecordSchema extends BaseSchema<
  Record<string, unknown>,
  Record<string, unknown>
> {
  private _valueSchema: BaseSchema;

  constructor(valueSchema: BaseSchema) {
    super();
    this._valueSchema = valueSchema;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (typeof input !== "object" || input === null || Array.isArray(input)) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected record, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: "record",
        received: describeType(input),
      });
      return undefined;
    }

    const obj = input as Record<string, unknown>;
    const result: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(obj)) {
      ctx.path.push(key);
      result[key] = this._valueSchema._runPipeline(value, ctx);
      ctx.path.pop();
    }

    return result;
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "record" as const,
      valueSchema: this._valueSchema._toNode(),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

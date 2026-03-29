import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";

export class EnumSchema extends BaseSchema<string | number, string | number> {
  private _values: (string | number)[];

  constructor(values: (string | number)[]) {
    super();
    this._values = values;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (!this._values.includes(input as string | number)) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected one of enum(${this._values.join(",")}), received ${String(input)}`,
        path: [...ctx.path],
        expected: `enum(${this._values.join(",")})`,
        received: String(input),
      });
      return undefined;
    }
    return input;
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "enum" as const,
      values: [...this._values],
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

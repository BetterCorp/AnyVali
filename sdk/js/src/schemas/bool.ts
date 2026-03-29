import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class BoolSchema extends BaseSchema<boolean, boolean> {
  _getCoercionTarget(): string {
    return "bool";
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (typeof input !== "boolean") {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected boolean, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: "bool",
        received: describeType(input),
      });
      return undefined;
    }
    return input;
  }

  _toNode(): SchemaNode {
    const node: SchemaNode = { kind: "bool" } as SchemaNode;
    this._addDefault(node);
    return node;
  }
}

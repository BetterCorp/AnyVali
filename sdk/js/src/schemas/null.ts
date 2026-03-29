import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class NullSchema extends BaseSchema<null, null> {
  _validate(input: unknown, ctx: ParseContext): unknown {
    if (input !== null) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected null, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: "null",
        received: describeType(input),
      });
      return undefined;
    }
    return null;
  }

  _toNode(): SchemaNode {
    const node: SchemaNode = { kind: "null" } as SchemaNode;
    this._addDefault(node);
    return node;
  }
}

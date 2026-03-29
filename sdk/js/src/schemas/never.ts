import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class NeverSchema extends BaseSchema<never, never> {
  _validate(input: unknown, ctx: ParseContext): unknown {
    ctx.issues.push({
      code: ISSUE_CODES.INVALID_TYPE,
      message: `Expected never (no value is valid)`,
      path: [...ctx.path],
      expected: "never",
      received: describeType(input),
    });
    return undefined;
  }

  _toNode(): SchemaNode {
    return { kind: "never" } as SchemaNode;
  }
}

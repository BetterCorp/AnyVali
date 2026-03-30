import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class UnionSchema<
  T extends BaseSchema<any, any>[] = BaseSchema[],
> extends BaseSchema<unknown, T[number]["_output"]> {
  private _variants: T;

  constructor(variants: [...T]) {
    super();
    this._variants = variants as T;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    for (const variant of this._variants) {
      const innerCtx: ParseContext = {
        path: [...ctx.path],
        issues: [],
      };
      const result = variant._runPipeline(input, innerCtx);
      if (innerCtx.issues.length === 0) {
        return result;
      }
    }

    const variantKinds = this._variants.map((v) => v._toNode().kind);
    ctx.issues.push({
      code: ISSUE_CODES.INVALID_UNION,
      message: `Input did not match any variant of the union`,
      path: [...ctx.path],
      expected: variantKinds.join(" | "),
      received: describeType(input),
    });

    return undefined;
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "union" as const,
      variants: this._variants.map((v) => v._toNode()),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

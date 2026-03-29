import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";

export class RefSchema extends BaseSchema<unknown, unknown> {
  private _ref: string;
  private _resolver?: () => BaseSchema;

  constructor(ref: string, resolver?: () => BaseSchema) {
    super();
    this._ref = ref;
    this._resolver = resolver;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (this._resolver) {
      const resolved = this._resolver();
      return resolved._validate(input, ctx);
    }

    ctx.issues.push({
      code: ISSUE_CODES.UNSUPPORTED_SCHEMA_KIND,
      message: `Unresolved ref: ${this._ref}`,
      path: [...ctx.path],
    });
    return undefined;
  }

  _toNode(): SchemaNode {
    return {
      kind: "ref" as const,
      ref: this._ref,
    } as unknown as SchemaNode;
  }
}

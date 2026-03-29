import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";

type LiteralValue = string | number | boolean | null;

export class LiteralSchema<T extends LiteralValue> extends BaseSchema<T, T> {
  private _value: T;

  constructor(value: T) {
    super();
    this._value = value;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (input !== this._value) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_LITERAL,
        message: `Expected literal ${String(this._value)}, received ${String(input)}`,
        path: [...ctx.path],
        expected: String(this._value),
        received: String(input),
      });
      return undefined;
    }
    return input;
  }

  _toNode(): SchemaNode {
    const node = { kind: "literal" as const, value: this._value };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
